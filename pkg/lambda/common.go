package lambda

import (
	"os"
	"strings"
)

func tryString(s1, s2 string) string {
	if s1 != "" {
		return s1
	}
	return s2
}

func stringToSlice(s1 string) []string {
	if s1 != "" {
		return strings.Split(s1, ",")
	}
	return nil
}

func maxInt(x, y int) int {
	if x > y {
		return x
	}
	return y
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
