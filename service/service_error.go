package service

import (
	"fmt"

	"github.com/pkg/errors"
)

// ErrorCode is code of service error
type ErrorCode string

// Definition of ErrorCode
const (
	ErrorCodeUnexpected            ErrorCode = "UnexpectedError"
	ErrorCodeBadRequest            ErrorCode = "BadRequest"
	ErrorCodeInvalidArguments      ErrorCode = "InvalidArguments"
	ErrorCodeInvalidlStatus        ErrorCode = "InvalidStatusError"
	ErrorCodeDB                    ErrorCode = "DBError"
	ErrorCodeNotFound              ErrorCode = "NotFound"
	ErrorCodeAlreadyExist          ErrorCode = "AlreadyExist"
	ErrorCodeOptimisticLockFailure ErrorCode = "OptimisticLockFailure"
	ErrorCodePreconditionInvalid   ErrorCode = "PreconditionInvalid"
	ErrorCodeUnauthenticated       ErrorCode = "Unauthenticated"
)

// SvcError presents error of logic service, This has error code, message and cause error.
type SvcError struct {
	Code    ErrorCode // Error code
	Message string    // message
	Details []string  // details
	Cause   error     // Cause error, when not exists, set nil
}

// SvcError returns the message of error
func (e *SvcError) Error() string {
	return e.Message
}

// NewSvcErrorWithDetails creates new server error with details
func NewSvcErrorWithDetails(code ErrorCode, err error, message string, details []string) error {
	if err != nil {
		err = errors.WithStack(err)
	}
	return &SvcError{
		Code:    code,
		Message: message,
		Details: details,
		Cause:   err,
	}
}

// NewSvcErrorWithDetailsf creates new server error with details which has formatted message
func NewSvcErrorWithDetailsf(code ErrorCode, err error, format string, details []string, values ...interface{}) error {
	return NewSvcErrorWithDetails(code, err, fmt.Sprintf(format, values...), details)
}

// NewSvcError creates new server error
func NewSvcError(code ErrorCode, err error, message string) error {
	return NewSvcErrorWithDetails(code, err, message, []string{})
}

// NewSvcErrorf creates new server error which has formatted message
func NewSvcErrorf(code ErrorCode, err error, format string, values ...interface{}) error {
	return NewSvcError(code, err, fmt.Sprintf(format, values...))
}

// NewBadRequestError creates new server error for bad request
func NewBadRequestError(err error) error {
	return NewSvcError(ErrorCodeBadRequest, err, "Failed to parse request")
}

// NewPathParameterError creates new server error of url path parameter
func NewPathParameterError(key string) error {
	return NewSvcErrorf(ErrorCodeBadRequest, nil, "Path parameter [%s] is not specifed in api path", key)
}

// NewDBCommitError creates new server error of commit transaction
func NewDBCommitError(err error) error {
	return NewSvcError(ErrorCodeDB, err, "Failed to commit transaction")
}
