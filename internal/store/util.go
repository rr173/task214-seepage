package store

import (
	"time"
)

const timeLayout = time.RFC3339

func fmtTime(t time.Time) string { return t.UTC().Format(timeLayout) }

func parseTime(s string) time.Time {
	if s == "" {
		return time.Time{}
	}
	t, err := time.Parse(timeLayout, s)
	if err != nil {
		return time.Time{}
	}
	return t
}

func parseNullableTime(s interface{}) *time.Time {
	str, ok := s.(string)
	if !ok || str == "" {
		return nil
	}
	t, err := time.Parse(timeLayout, str)
	if err != nil {
		return nil
	}
	return &t
}
