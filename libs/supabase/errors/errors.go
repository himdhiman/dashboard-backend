package errors

import (
	"errors"
	"fmt"
)

// Common Supabase errors
var (
	ErrMissingURL         = errors.New("supabase URL is required")
	ErrMissingKey         = errors.New("supabase key is required")
	ErrInvalidConfig      = errors.New("invalid configuration")
	ErrConnectionFailed   = errors.New("failed to connect to Supabase")
	ErrRequestFailed      = errors.New("request failed")
	ErrUnauthorized       = errors.New("unauthorized")
	ErrForbidden          = errors.New("forbidden")
	ErrNotFound           = errors.New("not found")
	ErrConflict           = errors.New("conflict")
	ErrValidationFailed   = errors.New("validation failed")
	ErrInvalidResponse    = errors.New("invalid response")
	ErrTimeout            = errors.New("request timeout")
	ErrRateLimited        = errors.New("rate limited")
	ErrInternalError      = errors.New("internal server error")
	ErrBadRequest         = errors.New("bad request")
	ErrMethodNotAllowed   = errors.New("method not allowed")
	ErrServiceUnavailable = errors.New("service unavailable")
)

// Auth specific errors
var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserNotFound       = errors.New("user not found")
	ErrEmailNotConfirmed  = errors.New("email not confirmed")
	ErrInvalidToken       = errors.New("invalid token")
	ErrTokenExpired       = errors.New("token expired")
	ErrWeakPassword       = errors.New("weak password")
	ErrEmailAlreadyExists = errors.New("email already exists")
	ErrPhoneAlreadyExists = errors.New("phone already exists")
	ErrInvalidEmail       = errors.New("invalid email")
	ErrInvalidPhone       = errors.New("invalid phone number")
)

// Database specific errors
var (
	ErrDatabaseError       = errors.New("database error")
	ErrTableNotFound       = errors.New("table not found")
	ErrColumnNotFound      = errors.New("column not found")
	ErrConstraintViolation = errors.New("constraint violation")
	ErrDuplicateKey        = errors.New("duplicate key")
	ErrForeignKeyViolation = errors.New("foreign key violation")
	ErrTransactionFailed   = errors.New("transaction failed")
	ErrQueryFailed         = errors.New("query failed")
	ErrInsertFailed        = errors.New("insert failed")
	ErrUpdateFailed        = errors.New("update failed")
	ErrDeleteFailed        = errors.New("delete failed")
	ErrSelectFailed        = errors.New("select failed")
)

// Storage specific errors
var (
	ErrStorageError     = errors.New("storage error")
	ErrBucketNotFound   = errors.New("bucket not found")
	ErrFileNotFound     = errors.New("file not found")
	ErrFileTooLarge     = errors.New("file too large")
	ErrInvalidFileType  = errors.New("invalid file type")
	ErrUploadFailed     = errors.New("upload failed")
	ErrDownloadFailed   = errors.New("download failed")
	ErrDeleteFileFailed = errors.New("delete file failed")
	ErrMoveFileFailed   = errors.New("move file failed")
	ErrCopyFileFailed   = errors.New("copy file failed")
)

// Realtime specific errors
var (
	ErrRealtimeError      = errors.New("realtime error")
	ErrChannelNotFound    = errors.New("channel not found")
	ErrSubscriptionFailed = errors.New("subscription failed")
	ErrBroadcastFailed    = errors.New("broadcast failed")
	ErrConnectionClosed   = errors.New("connection closed")
)

// Function specific errors
var (
	ErrFunctionError    = errors.New("function error")
	ErrFunctionNotFound = errors.New("function not found")
	ErrFunctionTimeout  = errors.New("function timeout")
	ErrFunctionFailed   = errors.New("function execution failed")
)

// Edge Functions specific errors
var (
	ErrEdgeFunctionError    = errors.New("edge function error")
	ErrEdgeFunctionNotFound = errors.New("edge function not found")
	ErrEdgeFunctionTimeout  = errors.New("edge function timeout")
	ErrEdgeFunctionFailed   = errors.New("edge function execution failed")
)

// NewError creates a new error with context
func NewError(err error, context string) error {
	return fmt.Errorf("%s: %w", context, err)
}

// NewErrorf creates a new formatted error
func NewErrorf(format string, args ...interface{}) error {
	return fmt.Errorf(format, args...)
}

// IsError checks if an error is of a specific type
func IsError(err, target error) bool {
	return errors.Is(err, target)
}

// AsError extracts an error of a specific type
func AsError(err error, target interface{}) bool {
	return errors.As(err, target)
}
