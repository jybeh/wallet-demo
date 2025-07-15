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
