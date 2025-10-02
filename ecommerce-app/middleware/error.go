package middleware

import (
    "github.com/gin-gonic/gin"
    "github.com/inquisitivefrog/ecommerce-app/utils"
    "net/http"
)

func ErrorHandler() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Next()

        if len(c.Errors) > 0 {
            err := c.Errors.Last()
            apiErr, ok := err.Err.(*utils.APIError)
            if !ok {
                apiErr = utils.NewAPIError(http.StatusInternalServerError, "Internal server error", "INTERNAL_ERROR")
            }
            c.JSON(apiErr.Status, apiErr)
        }
    }
}
