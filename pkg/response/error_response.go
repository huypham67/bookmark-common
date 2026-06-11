package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ErrorResponse represents the structure of an error response returned by the API.
type ErrorResponse struct {
	Error string `json:"error"`
}

// Error sends a JSON response with the specified HTTP status code and error message.
func Error(c *gin.Context, status int, message string) {
	c.JSON(status, ErrorResponse{
		Error: message,
	})
}

// BadRequest sends a 400 Bad Request response with the provided error message.
func BadRequest(c *gin.Context, message string) {
	Error(c, http.StatusBadRequest, message)
}

// Unauthorized sends a 401 Unauthorized response with the provided error message.
func Unauthorized(c *gin.Context, message string) {
	Error(c, http.StatusUnauthorized, message)
}

// Forbidden sends a 403 Forbidden response with the provided error message.
func Forbidden(c *gin.Context, message string) {
	Error(c, http.StatusForbidden, message)
}

// NotFound sends a 404 Not Found response with the provided error message.
func NotFound(c *gin.Context, message string) {
	Error(c, http.StatusNotFound, message)
}

// Conflict sends a 409 Conflict response with the provided error message.
func Conflict(c *gin.Context, message string) {
	Error(c, http.StatusConflict, message)
}

// InternalServerError sends a 500 Internal Server Error response with a generic error message.
func InternalServerError(c *gin.Context) {
	Error(c, http.StatusInternalServerError, "Internal Server Error")
}
