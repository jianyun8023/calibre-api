package errors

import (
	"errors"
	"net/http"
	"testing"
)

func TestNew(t *testing.T) {
	err := New(CodeInvalidRequest, "invalid parameter", http.StatusBadRequest)

	if err.Code != CodeInvalidRequest {
		t.Errorf("expected code %s, got %s", CodeInvalidRequest, err.Code)
	}
	if err.Message != "invalid parameter" {
		t.Errorf("expected message 'invalid parameter', got '%s'", err.Message)
	}
	if err.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, err.StatusCode)
	}
}

func TestWrap(t *testing.T) {
	original := errors.New("original error")
	wrapped := Wrap(original, CodeInternalError, "wrapped error", http.StatusInternalServerError)

	if wrapped.Err != original {
		t.Error("wrapped error should contain original error")
	}
	if !errors.Is(wrapped, original) {
		t.Error("errors.Is should work with wrapped errors")
	}
}

func TestWithContext(t *testing.T) {
	err := NewBookNotFoundError("123")
	err = err.WithContext("user_id", "user-456")

	if err.Context["user_id"] != "user-456" {
		t.Error("context should be set correctly")
	}
	if err.Context["book_id"] != "123" {
		t.Error("initial context should be preserved")
	}
}

func TestAsAppError(t *testing.T) {
	appErr := NewInvalidRequestError("test")

	converted, ok := AsAppError(appErr)
	if !ok {
		t.Error("should convert AppError successfully")
	}
	if converted != appErr {
		t.Error("converted error should be the same as original")
	}

	stdErr := errors.New("standard error")
	_, ok = AsAppError(stdErr)
	if ok {
		t.Error("should not convert standard error")
	}
}

func TestGetHTTPStatus(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected int
	}{
		{
			name:     "app error",
			err:      NewBookNotFoundError("123"),
			expected: http.StatusNotFound,
		},
		{
			name:     "standard error",
			err:      errors.New("test"),
			expected: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status := GetHTTPStatus(tt.err)
			if status != tt.expected {
				t.Errorf("expected status %d, got %d", tt.expected, status)
			}
		})
	}
}
