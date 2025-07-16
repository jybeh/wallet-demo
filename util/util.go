package util

import (
	"encoding/base64"
	"encoding/json"
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

type DataCursor struct {
	LastTimestamp time.Time `json:"lastTimestamp"`
	LastID        string    `json:"lastID"`
}

// EncodeNextToken creates a Base64 cursor
func EncodeNextToken(cursor DataCursor) string {
	b, _ := json.Marshal(cursor)
	return base64.StdEncoding.EncodeToString(b)
}

// DecodeNextToken parses Base64 back into a cursor
func DecodeNextToken(token string) (*DataCursor, error) {
	if token == "" {
		return nil, nil
	}

	decoded, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		return nil, err
	}

	var cursor DataCursor
	if unmarshalErr := json.Unmarshal(decoded, &cursor); unmarshalErr != nil {
		return nil, err
	}

	return &cursor, nil
}
