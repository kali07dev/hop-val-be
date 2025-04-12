package utils

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/hopekali04/valuations/schema"
	"gorm.io/gorm"
)

// NewAPIError creates a structured API error response.
func NewAPIError(statusCode int, message string, details ...string) *schema.CustomError {
	err := &schema.CustomError{
		StatusCode: statusCode,
		Message:    message,
	}
	if len(details) > 0 {
		err.Details = details[0]
	}
	return err
}

// HandleError centralizes error handling for handlers.
func HandleError(c *fiber.Ctx, err error) error {
	// Default error
	apiError := NewAPIError(http.StatusInternalServerError, "An unexpected error occurred", err.Error())

	// Check for specific error types
	var ve validator.ValidationErrors
	if errors.As(err, &ve) {
		// Provide more specific validation error messages
		// For simplicity, just taking the first error. Consider iterating for more detail.
		firstErr := ve[0]
		errMsg := fmt.Sprintf("Validation failed on field '%s', condition '%s'", firstErr.Namespace(), firstErr.Tag())
		apiError = NewAPIError(http.StatusBadRequest, "Invalid request data", errMsg)
	} else if errors.Is(err, gorm.ErrRecordNotFound) {
		apiError = NewAPIError(http.StatusNotFound, "Resource not found")
	} else {
		// Check if it's already a CustomError we created
		var customErr *schema.CustomError
		if errors.As(err, &customErr) {
			apiError = customErr // Use the existing custom error
		}
		// Could add more specific checks here (e.g., duplicate key errors from DB)
	}

	return c.Status(apiError.StatusCode).JSON(apiError)
}


var ErrPropertyExists = NewAPIError(http.StatusConflict, "Property already exists", "A property with the provided ID already exists in the database.")

func NewBadRequestError(details string) *schema.CustomError {
	return NewAPIError(http.StatusBadRequest, "Bad Request", details)
}

func NewNotFoundError(resource string) *schema.CustomError {
	return NewAPIError(http.StatusNotFound, fmt.Sprintf("%s not found", resource))
}
