package util

import (
	"errors"
	"testing"
	"time"
)

func TestRetry(t *testing.T) {
	errKnown := errors.New("known error")
	errUnknown := errors.New("unknown error")

	tests := []struct {
		name          string
		fn            func() error
		maxRetries    int
		delay         time.Duration
		knownErrors   []error
		expectedError error
	}{
		{
			name: "success on first try",
			fn: func() error {
				return nil
			},
			maxRetries:    3,
			delay:         10 * time.Millisecond,
			knownErrors:   nil,
			expectedError: nil,
		},
		{
			name: "success after retries",
			fn: func() func() error {
				retries := 0
				return func() error {
					if retries < 2 {
						retries++
						return errUnknown
					}
					return nil
				}
			}(),
			maxRetries:    3,
			delay:         10 * time.Millisecond,
			knownErrors:   nil,
			expectedError: nil,
		},
		{
			name: "exceed retries with non-known error",
			fn: func() error {
				return errUnknown
			},
			maxRetries:    2,
			delay:         10 * time.Millisecond,
			knownErrors:   nil,
			expectedError: errUnknown,
		},
		{
			name: "stop retrying on known error",
			fn: func() error {
				return errKnown
			},
			maxRetries:    3,
			delay:         10 * time.Millisecond,
			knownErrors:   []error{errKnown},
			expectedError: errKnown,
		},
		{
			name: "different error not in knownErrors",
			fn: func() func() error {
				retries := 0
				return func() error {
					if retries < 3 {
						retries++
						return errUnknown
					}
					return nil
				}
			}(),
			maxRetries:    5,
			delay:         10 * time.Millisecond,
			knownErrors:   []error{errKnown},
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Retry(tt.fn, tt.maxRetries, tt.delay, tt.knownErrors...)
			if !errors.Is(err, tt.expectedError) {
				t.Errorf("expected error: %v, got: %v", tt.expectedError, err)
			}
		})
	}
}

func TestEncodeNextToken(t *testing.T) {
	tests := []struct {
		name     string
		cursor   DataCursor
		expected string
	}{
		{
			name: "normal case",
			cursor: DataCursor{
				LastTimestamp: time.Date(2025, 7, 15, 10, 0, 0, 0, time.UTC),
				LastID:        "12345",
			},
			expected: "eyJsYXN0VGltZXN0YW1wIjoiMjAyNS0wNy0xNVQxMDowMDowMFoiLCJsYXN0SUQiOiIxMjM0NSJ9",
		},
		{
			name: "empty fields",
			cursor: DataCursor{
				LastTimestamp: time.Time{},
				LastID:        "",
			},
			expected: "eyJsYXN0VGltZXN0YW1wIjoiMDAwMS0wMS0wMVQwMDowMDowMFoiLCJsYXN0SUQiOiIifQ==",
		},
		{
			name: "special characters in ID",
			cursor: DataCursor{
				LastTimestamp: time.Date(2025, 7, 15, 10, 0, 0, 0, time.UTC),
				LastID:        "ID@#&!%",
			},
			expected: "eyJsYXN0VGltZXN0YW1wIjoiMjAyNS0wNy0xNVQxMDowMDowMFoiLCJsYXN0SUQiOiJJREAjXHUwMDI2ISUifQ==",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := EncodeNextToken(tt.cursor)
			if result != tt.expected {
				t.Errorf("expected: %v, got: %v", tt.expected, result)
			}
		})
	}
}

func TestDecodeNextToken(t *testing.T) {
	tests := []struct {
		name     string
		token    string
		expected *DataCursor
		err      error
	}{
		{
			name:  "valid token",
			token: "eyJsYXN0VGltZXN0YW1wIjoiMjAyNS0wNy0xNVQxMDowMDowMFoiLCJsYXN0SUQiOiIxMjM0NSJ9",
			expected: &DataCursor{
				LastTimestamp: time.Date(2025, 7, 15, 10, 0, 0, 0, time.UTC),
				LastID:        "12345",
			},
			err: nil,
		},
		{
			name:     "empty token",
			token:    "",
			expected: nil,
			err:      nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := DecodeNextToken(tt.token)
			if (tt.err == nil && err != nil) || (tt.err != nil && err == nil) || (tt.err != nil && !errors.Is(err, tt.err)) {
				t.Errorf("expected error: %v, got: %v", tt.err, err)
			}
			if (result == nil) != (tt.expected == nil) || (result != nil && *result != *tt.expected) {
				t.Errorf("expected: %v, got: %v", tt.expected, result)
			}
		})
	}
}
