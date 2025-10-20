package api

import (
	"errors"
	"fmt"
	"strings"
	"sync"

	app_errors "flow-ai/backend/internal/errors"

	"github.com/go-playground/validator/v10"
)

// This file provides a centralized, singleton-based validation helper for API request bodies.
// Using a singleton for the validator is a performance best practice, as it avoids
// the costly process of recreating the validator instance on every request.

var (
	// validate holds the single instance of the validator.
	validate *validator.Validate
	// once ensures that the validator is initialized only one time.
	once sync.Once
)

// getInstance uses sync.Once to safely initialize and return the validator singleton.
func getInstance() *validator.Validate {
	once.Do(func() {
		validate = validator.New()
	})
	return validate
}

// validateRequest checks a given payload struct against the validation rules
// defined in its field tags (e.g., `validate:"required,min=1"`).
// If validation fails, it returns a wrapped `app_errors.ErrValidation` with a
// user-friendly, detailed message.
func validateRequest(payload interface{}) error {
	v := getInstance()
	err := v.Struct(payload)
	if err == nil {
		return nil
	}

	// Try to cast the error to `validator.ValidationErrors`. If this fails, it's not a
	// validation error from the library but some other unexpected issue.
	var validationErrors validator.ValidationErrors
	if !errors.As(err, &validationErrors) {
		return fmt.Errorf("%w: an unexpected error occurred during validation: %s", app_errors.ErrValidation, err.Error())
	}

	// Format the validation errors into a clean, readable string.
	var errorMessages []string
	for _, fieldErr := range validationErrors {
		// Example output: "Field 'Content' failed on the 'required' tag."
		errMsg := fmt.Sprintf("Field '%s' failed on the '%s' tag", fieldErr.Field(), fieldErr.Tag())
		errorMessages = append(errorMessages, errMsg)
	}

	// Return a single, well-structured validation error that can be displayed to the user.
	return fmt.Errorf("%w: %s", app_errors.ErrValidation, strings.Join(errorMessages, "; "))
}
