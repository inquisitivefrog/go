package utils

import (
    "github.com/gin-gonic/gin"
    "net/http"
)

// APIError represents a standard API error response
type APIError struct {
    Status    int    `json:"status"`
    Message   string `json:"message"`
    ErrorCode string `json:"error_code"`
}

// Error implements the error interface
func (e *APIError) Error() string {
    return e.Message
}

// NewAPIError creates a new APIError instance
func NewAPIError(status int, message, code string) *APIError {
    return &APIError{Status: status, Message: message, ErrorCode: code}
}

// RespondWithError sends a standardized error response
func RespondWithError(c *gin.Context, status int, message string) {
    err := NewAPIError(status, message, http.StatusText(status))
    c.Error(err) // Set error in context for middleware
    c.JSON(status, err)
}
