package middleware

import (
    "net/http"
    "strings"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/golang-jwt/jwt/v4"
    "github.com/inquisitivefrog/ecommerce-app/config"
    "github.com/inquisitivefrog/ecommerce-app/models"
    "github.com/inquisitivefrog/ecommerce-app/utils"
)

func GenerateJWT(userID uint, cfg *config.Config) (string, error) {
    if len(cfg.JWTSecret) < 32 {
        return "", utils.NewAPIError(http.StatusInternalServerError, "JWT secret too short", "INVALID_JWT_SECRET")
    }
    claims := jwt.MapClaims{
        "user_id": userID,
        "exp":     time.Now().Add(time.Hour * 24).Unix(),
        "iat":     time.Now().Unix(),
    }
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString([]byte(cfg.JWTSecret))
}

func AuthMiddleware(cfg *config.Config) gin.HandlerFunc {
    return func(c *gin.Context) {
        authHeader := c.GetHeader("Authorization")
        if authHeader == "" {
            err := utils.NewAPIError(http.StatusUnauthorized, "Authorization header required", "NO_AUTH_HEADER")
            c.Error(err) // Set error for middleware
            utils.RespondWithError(c, http.StatusUnauthorized, "Authorization header required")
            c.Abort()
            return
        }

        parts := strings.Split(authHeader, " ")
        if len(parts) != 2 || parts[0] != "Bearer" {
            err := utils.NewAPIError(http.StatusUnauthorized, "Invalid Authorization header", "INVALID_AUTH_HEADER")
            c.Error(err)
            utils.RespondWithError(c, http.StatusUnauthorized, "Invalid Authorization header")
            c.Abort()
            return
        }

        token, err := jwt.Parse(parts[1], func(token *jwt.Token) (interface{}, error) {
            if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
                return nil, jwt.ErrSignatureInvalid
            }
            return []byte(cfg.JWTSecret), nil
        })
        if err != nil {
            err := utils.NewAPIError(http.StatusUnauthorized, "Invalid token", "INVALID_TOKEN")
            c.Error(err)
            utils.RespondWithError(c, http.StatusUnauthorized, "Invalid token")
            c.Abort()
            return
        }

        if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
            userID, ok := claims["user_id"].(float64)
            if !ok {
                err := utils.NewAPIError(http.StatusUnauthorized, "Invalid token claims", "INVALID_CLAIMS")
                c.Error(err)
                utils.RespondWithError(c, http.StatusUnauthorized, "Invalid token claims")
                c.Abort()
                return
            }

            var user models.User
            if err := cfg.DB.Where("id = ?", uint(userID)).First(&user).Error; err != nil {
                err := utils.NewAPIError(http.StatusUnauthorized, "User not found", "USER_NOT_FOUND")
                c.Error(err)
                utils.RespondWithError(c, http.StatusUnauthorized, "User not found")
                c.Abort()
                return
            }
            c.Set("user", user)
            c.Next()
        } else {
            err := utils.NewAPIError(http.StatusUnauthorized, "Invalid token", "INVALID_TOKEN")
            c.Error(err)
            utils.RespondWithError(c, http.StatusUnauthorized, "Invalid token")
            c.Abort()
        }
    }
}

func AdminMiddleware(cfg *config.Config) gin.HandlerFunc {
    return func(c *gin.Context) {
        user, exists := c.Get("user")
        if !exists {
            err := utils.NewAPIError(http.StatusUnauthorized, "User not authenticated", "NO_USER")
            c.Error(err)
            utils.RespondWithError(c, http.StatusUnauthorized, "User not authenticated")
            c.Abort()
            return
        }

        u, ok := user.(models.User)
        if !ok || u.Role != "admin" {
            err := utils.NewAPIError(http.StatusForbidden, "Admin access required", "FORBIDDEN")
            c.Error(err)
            utils.RespondWithError(c, http.StatusForbidden, "Admin access required")
            c.Abort()
            return
        }
        c.Next()
    }
}
