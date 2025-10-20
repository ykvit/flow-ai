package errors

import "errors"

// This package defines a centralized set of sentinel errors for the application.
// Using sentinel errors allows services to return specific, recognizable error types
// without coupling them to implementation details like HTTP status codes. The API
// layer can then use `errors.Is()` to check for these specific errors and map
// them to the correct HTTP responses. This is a core principle of Clean Architecture.

var (
	// ErrNotFound signifies that a requested resource could not be located.
	// This is typically mapped to a 404 Not Found HTTP status.
	ErrNotFound = errors.New("resource not found")

	// ErrValidation signifies that input data provided by a client failed
	// business rule validation.
	// This is typically mapped to a 400 Bad Request HTTP status.
	ErrValidation = errors.New("validation failed")

	// ErrConflict signifies that an operation could not be completed because
	// it conflicts with the current state of a resource (e.g., creating a
	// resource that already exists).
	// This is typically mapped to a 409 Conflict HTTP status.
	ErrConflict = errors.New("resource conflict")

	// ErrPermission signifies that the authenticated user is not authorized
	// to perform the requested action.
	// This is typically mapped to a 403 Forbidden HTTP status.
	ErrPermission = errors.New("permission denied")

	// ErrInternal signifies an unexpected error on the server. This is a generic
	// error used to prevent leaking sensitive implementation details to the client.
	// This is typically mapped to a 500 Internal Server Error HTTP status.
	ErrInternal = errors.New("internal server error")
)
