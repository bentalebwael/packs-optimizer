package errors

import (
	"errors"
	"testing"
)

func TestError_Error(t *testing.T) {
	t.Run("error without wrapped error", func(t *testing.T) {
		err := &Error{
			Type:    ErrorTypeInvalidRequest,
			Message: "invalid input",
		}
		expected := "INVALID_REQUEST: invalid input"
		if err.Error() != expected {
			t.Errorf("Error() = %v, want %v", err.Error(), expected)
		}
	})

	t.Run("error with wrapped error", func(t *testing.T) {
		innerErr := errors.New("inner error")
		err := &Error{
			Type:    ErrorTypeInvalidRequest,
			Message: "invalid input",
			Err:     innerErr,
		}
		expected := "INVALID_REQUEST: invalid input: inner error"
		if err.Error() != expected {
			t.Errorf("Error() = %v, want %v", err.Error(), expected)
		}
	})
}

func TestNewError(t *testing.T) {
	t.Run("creates error with all fields", func(t *testing.T) {
		innerErr := errors.New("inner error")
		err := NewError(ErrorTypeInternal, "test message", innerErr)

		if err.Type != ErrorTypeInternal {
			t.Errorf("Type = %v, want %v", err.Type, ErrorTypeInternal)
		}
		if err.Message != "test message" {
			t.Errorf("Message = %v, want %v", err.Message, "test message")
		}
		if err.Err != innerErr {
			t.Errorf("Err = %v, want %v", err.Err, innerErr)
		}
	})
}

func TestNewInternalErrorWrap(t *testing.T) {
	innerErr := errors.New("database error")
	err := NewInternalErrorWrap("failed to save", innerErr)

	if err.Type != ErrorTypeInternal {
		t.Errorf("Type = %v, want %v", err.Type, ErrorTypeInternal)
	}
	if err.Message != "failed to save" {
		t.Errorf("Message = %v, want %v", err.Message, "failed to save")
	}
	if err.Err != innerErr {
		t.Errorf("Err = %v, want %v", err.Err, innerErr)
	}
}

func TestNewValidationErrorWrap(t *testing.T) {
	innerErr := errors.New("validation failed")
	err := NewValidationErrorWrap("invalid input", innerErr)

	if err.Type != ErrorTypeInvalidRequest {
		t.Errorf("Type = %v, want %v", err.Type, ErrorTypeInvalidRequest)
	}
	if err.Message != "invalid input" {
		t.Errorf("Message = %v, want %v", err.Message, "invalid input")
	}
	if !errors.Is(innerErr, err.Err) {
		t.Errorf("Err = %v, want %v", err.Err, innerErr)
	}
}

func TestNewValidationError(t *testing.T) {
	err := NewValidationError("invalid input")

	if err.Type != ErrorTypeInvalidRequest {
		t.Errorf("Type = %v, want %v", err.Type, ErrorTypeInvalidRequest)
	}
	if err.Message != "invalid input" {
		t.Errorf("Message = %v, want %v", err.Message, "invalid input")
	}
	if err.Err == nil {
		t.Error("Err should not be nil")
	}
	if err.Err.Error() != "invalid input" {
		t.Errorf("Err.Error() = %v, want %v", err.Err.Error(), "invalid input")
	}
}

func TestNewInternalError(t *testing.T) {
	err := NewInternalError("server error")

	if err.Type != ErrorTypeInternal {
		t.Errorf("Type = %v, want %v", err.Type, ErrorTypeInternal)
	}
	if err.Message != "server error" {
		t.Errorf("Message = %v, want %v", err.Message, "server error")
	}
	if err.Err == nil {
		t.Error("Err should not be nil")
	}
	if err.Err.Error() != "server error" {
		t.Errorf("Err.Error() = %v, want %v", err.Err.Error(), "server error")
	}
}
