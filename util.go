package main

import (
	"time"
)

func IsHex(s string) bool {
	for _, ch := range s {
		if '0' <= ch && ch <= '9' || 'a' <= ch && ch <= 'f' || 'A' <= ch && ch <= 'F' {
			continue
		}
		return false
	}
	return true
}

func FormatDate(date int32) string {
	return time.Unix(int64(date), 0).UTC().Format(time.RFC3339)
}

func Now() string {
	return time.Now().UTC().Format(time.RFC3339)
}
