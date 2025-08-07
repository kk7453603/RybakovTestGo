package grpc

import (
	"fmt"
	"net/http"
	"strings"
	"time"
)

func TimeValidationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		timeParams := []string{"startTime", "endTime", "timestamp"}

		query := r.URL.Query()
		for _, param := range timeParams {
			if timeStr := query.Get(param); timeStr != "" {
				if err := validateTimeFormat(timeStr); err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
			}
		}

		next.ServeHTTP(w, r)
	})
}

func validateTimeFormat(timeStr string) error {
	// Поддерживаемые форматы времени
	supportedFormats := []string{
		time.RFC3339,           // "2006-01-02T15:04:05Z07:00"
		time.RFC3339Nano,       // "2006-01-02T15:04:05.999999999Z07:00"
		"2006-01-02T15:04:05Z", // UTC формат
		"2006-01-02",           // Только дата
		"15:04:05",             // Только время
	}

	var parseErrors []string

	for _, format := range supportedFormats {
		if _, err := time.Parse(format, timeStr); err == nil {
			return nil
		} else {
			parseErrors = append(parseErrors, format+": "+err.Error())
		}
	}

	return &InvalidTimeFormatError{
		Value:            timeStr,
		SupportedFormats: supportedFormats,
		ParseErrors:      parseErrors,
	}
}

type InvalidTimeFormatError struct {
	Value            string
	SupportedFormats []string
	ParseErrors      []string
}

func (e *InvalidTimeFormatError) Error() string {
	return fmt.Sprintf(
		"invalid time format: '%s'. Supported formats: %s",
		e.Value,
		strings.Join(e.SupportedFormats, ", "),
	)
}
