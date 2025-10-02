package handlers

import (
    "net/http"

    "github.com/gin-gonic/gin"
    "github.com/inquisitivefrog/ecommerce-app/config"
    "github.com/inquisitivefrog/ecommerce-app/middleware"
    "github.com/inquisitivefrog/ecommerce-app/models"
    "github.com/sirupsen/logrus"
    "golang.org/x/crypto/bcrypt"
    "gorm.io/gorm"
)

type UserHandler struct {
    DB     *gorm.DB
    Config *config.Config
    Logger *logrus.Logger
}

func (h *UserHandler) Register(c *gin.Context) {
    var input struct {
        Username string `json:"username"`
        Password string `json:"password"`
        Email    string `json:"email"`
    }
    if err := c.ShouldBindJSON(&input); err != nil {
        h.Logger.WithFields(logrus.Fields{
            "error":      err,
            "error_code": "INVALID_INPUT",
        }).Warn("Invalid input for register")
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
    if err != nil {
        h.Logger.WithFields(logrus.Fields{
            "error":      err,
            "error_code": "HASH_PASSWORD_FAILED",
        }).Error("Failed to hash password")
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to register user"})
        return
    }

    user := models.User{
        Username: input.Username,
        Password: string(hashedPassword),
        Email:    input.Email,
        Role:     "user",
    }
    if err := h.DB.Create(&user).Error; err != nil {
        h.Logger.WithFields(logrus.Fields{
            "error":      err,
            "error_code": "CREATE_USER_FAILED",
        }).Error("Failed to create user")
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to register user"})
        return
    }

    h.Logger.WithFields(logrus.Fields{
        "username": input.Username,
    }).Info("User registered successfully")
    c.JSON(http.StatusCreated, gin.H{"message": "User registered successfully"})
}

func (h *UserHandler) Login(c *gin.Context) {
    var input struct {
        Username string `json:"username"`
        Password string `json:"password"`
    }
    if err := c.ShouldBindJSON(&input); err != nil {
        h.Logger.WithFields(logrus.Fields{
            "error":      err,
            "error_code": "INVALID_INPUT",
        }).Warn("Invalid input for login")
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    h.Logger.WithFields(logrus.Fields{
        "username": input.Username,
    }).Debug("Login input received")

    var user models.User
    if err := h.DB.Where("username = ?", input.Username).First(&user).Error; err != nil {
        h.Logger.WithFields(logrus.Fields{
            "username":   input.Username,
            "error":      err,
            "error_code": "USER_NOT_FOUND",
        }).Warn("User not found")
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
        return
    }

    h.Logger.WithFields(logrus.Fields{
        "username": user.Username,
        "user_id":  user.ID,
    }).Debug("User retrieved from database")

    if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
        h.Logger.WithFields(logrus.Fields{
            "username":   input.Username,
            "error":      err,
            "error_code": "INVALID_PASSWORD",
        }).Warn("Invalid password")
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
        return
    }

    token, err := middleware.GenerateJWT(user.ID, h.Config)
    if err != nil {
        h.Logger.WithFields(logrus.Fields{
            "user_id":    user.ID,
            "error":      err,
            "error_code": "JWT_GENERATION_FAILED",
        }).Error("Failed to generate JWT")
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
        return
    }

    h.Logger.WithFields(logrus.Fields{
        "username": input.Username,
        "user_id":  user.ID,
    }).Info("User logged in successfully")
    c.JSON(http.StatusOK, gin.H{"token": token})
}

func (h *UserHandler) GetProfile(c *gin.Context) {
    user, exists := c.Get("user")
    if !exists {
        h.Logger.WithFields(logrus.Fields{
            "error_code": "USER_NOT_AUTHENTICATED",
        }).Warn("User not authenticated")
        c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
        return
    }

    u, ok := user.(models.User)
    if !ok {
        h.Logger.WithFields(logrus.Fields{
            "error_code": "INVALID_USER_TYPE",
        }).Error("Invalid user type")
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user type"})
        return
    }

    profile := gin.H{
        "id":       u.ID,
        "username": u.Username,
        "email":    u.Email,
        "role":     u.Role,
    }
    h.Logger.WithFields(logrus.Fields{
        "user_id":  u.ID,
        "username": u.Username,
    }).Info("Fetched user profile")
    c.JSON(http.StatusOK, profile)
}
