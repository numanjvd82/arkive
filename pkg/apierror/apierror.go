package apierror

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"arkive/pkg/validation"
)

type ErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
}

type Response struct {
	Error ErrorBody `json:"error"`
}

func Write(c *gin.Context, status int, code, message string, details any) {
	c.JSON(status, Response{
		Error: ErrorBody{
			Code:    code,
			Message: message,
			Details: details,
		},
	})
}

func InvalidPayload(c *gin.Context) {
	Write(c, http.StatusBadRequest, "invalid_payload", "Invalid payload", nil)
}

func Unauthorized(c *gin.Context) {
	Write(c, http.StatusUnauthorized, "unauthorized", "Unauthorized", nil)
}

func Forbidden(c *gin.Context, message string) {
	if message == "" {
		message = "Forbidden"
	}
	Write(c, http.StatusForbidden, "forbidden", message, nil)
}

func Internal(c *gin.Context, message string) {
	if message == "" {
		message = "Internal error"
	}
	Write(c, http.StatusInternalServerError, "internal_error", message, nil)
}

func Validation(c *gin.Context, errors validation.Errors) {
	Write(c, http.StatusBadRequest, "validation_failed", "Validation failed", gin.H{
		"fields": errors,
	})
}

func RequestTooLarge(c *gin.Context) {
	Write(c, http.StatusRequestEntityTooLarge, "request_too_large", "Request too large", nil)
}
