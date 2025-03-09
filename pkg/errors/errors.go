package errors

import (
	"errors"
	"fmt"
)

type ErrorType string

const (
	ErrorTypeInvalidRequest ErrorType = "INVALID_REQUEST"
	ErrorTypeInternal       ErrorType = "INTERNAL"
)

type Error struct {
	Type    ErrorType
	Message string
	Err     error
}

func (e *Error) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s: %v", e.Type, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

func NewError(errType ErrorType, message string, err error) *Error {
	return &Error{
		Type:    errType,
		Message: message,
		Err:     err,
	}
}

func NewInternalErrorWrap(message string, err error) *Error {
	return NewError(ErrorTypeInternal, message, err)
}

func NewValidationErrorWrap(message string, err error) *Error {
	return NewError(ErrorTypeInvalidRequest, message, err)
}

func NewValidationError(message string) *Error {
	return NewError(ErrorTypeInvalidRequest, message, errors.New(message))
}

func NewInternalError(message string) *Error {
	return NewError(ErrorTypeInternal, message, errors.New(message))
}
