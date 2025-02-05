package errors

import "errors"

var (
	ErrInsertFailed     = errors.New("failed to insert document")
	ErrParseInsertedID  = errors.New("failed to parse inserted ID")
	ErrDocumentNotFound = errors.New("document not found")
	ErrUpdateFailed     = errors.New("failed to update document")
	ErrDeleteFailed     = errors.New("failed to delete document")
	ErrCountFailed      = errors.New("failed to count documents")
	ErrInvalidObjectID  = errors.New("invalid object ID")
)
