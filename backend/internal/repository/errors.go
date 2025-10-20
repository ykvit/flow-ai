package repository

import "errors"

// This file defines custom errors specific to the repository layer.
// This allows the repository to communicate outcomes in a database-agnostic way.

// ErrNotFound is a repository-specific sentinel error. It is returned when a
// query for a single entity (e.g., GetChatByID) finds no rows.
//
// The service layer should check for this specific error and translate it into
// a domain-level error (like `app_errors.ErrNotFound`), thus decoupling the
// business logic from the data access implementation. This abstracts away the
// underlying database driver's error (e.g., `sql.ErrNoRows`).
var ErrNotFound = errors.New("repository: not found")
