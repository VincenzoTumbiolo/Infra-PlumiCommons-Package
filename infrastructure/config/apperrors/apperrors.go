// See https://app.clickup.com/20558920/v/dc/kkd28-45712

package apperrors

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gofiber/fiber/v2"
)

const UNKNOWN_ERROR_MESSAGE = "Unknown error"

// AppError defines the methods for every type of error
type AppError interface {
	error
	Code() uint
	Subject() string
	StatusCode() int
}

// FullCode gives the full error identifier. The schema is ERR-123
func FullCode(err AppError) string {
	return fmt.Sprintf("%s-%03d", err.Subject(), err.Code())
}

// ExtractStatus gives the corresponding HTTP error code
func ExtractStatus(err error) int {
	var status int
	var appErr AppError
	var fibErr *fiber.Error
	if errors.As(err, &appErr) {
		status = appErr.StatusCode()
	} else if errors.As(err, &fibErr) {
		status = fibErr.Code
	} else {
		status = http.StatusInternalServerError
	}

	return status
}
