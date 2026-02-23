package models

import "errors"

// ErrNotFound is returned by repository methods when the requested record does
// not exist.  HTTP handlers map this to 404 Not Found.
var ErrNotFound = errors.New("not found")

// ErrConflict is returned when a unique constraint would be violated (e.g. a
// duplicate username).  HTTP handlers map this to 409 Conflict.
var ErrConflict = errors.New("conflict")
