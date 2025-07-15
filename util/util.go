package util

import (
	"errors"
	"time"
)

// Retry retries fn up to maxRetries if error is not in knownErrors
func Retry(fn func() error, maxRetries int, delay time.Duration, knownErrors ...error) error {
	for i := 0; i <= maxRetries; i++ {
		err := fn()
		if err == nil {
			return nil
		}

		// Skip retry if error is known
		for _, known := range knownErrors {
			if errors.Is(err, known) {
				return err
			}
		}

		// Only retry if not the last attempt
		if i < maxRetries {
			time.Sleep(delay)
		} else {
			return err // exhausted all retries
		}
	}
	return nil
}
