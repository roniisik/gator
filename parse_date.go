package main

import (
	"errors"
	"time"
)

var dateFormats = []string{
	time.RFC1123,       // "Mon, 02 Jan 2006 15:04:05 MST"
	time.RFC1123Z,      // "Mon, 02 Jan 2006 15:04:05 -0700"
	time.RFC3339,       // "2006-01-02T15:04:05Z07:00"
	"2006-01-02 15:04", // Custom format: "YYYY-MM-DD HH:MM"
	"02 Jan 2006",      // Custom format: "DD Mon YYYY"
}

func parseDate(dateStr string) (time.Time, error) {
	for _, format := range dateFormats {
		parsedDate, err := time.Parse(format, dateStr)
		if err == nil {
			return parsedDate, nil // Successfully parsed
		}
	}
	return time.Time{}, errors.New("unsupported date format")
}
