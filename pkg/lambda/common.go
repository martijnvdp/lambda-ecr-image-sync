package lambda

import (
	"strings"
	"time"
)

// tryString returns the first non empty string
func tryString(s1, s2 string) string {
	if s1 != "" {
		return s1
	}
	return s2
}

// maxInt returns the max int
func maxInt(x, y int) int {
	if x > y {
		return x
	}
	return y
}

// stringToSlice returns a slice of strings from a comma separated string
func stringToSlice(s1 string) []string {
	if s1 != "" {
		return strings.Split(s1, ",")
	}
	return nil
}

// parseTime parses a string to time.Time
func parseTime(s string) time.Time {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		panic(err)
	}
	return t
}
