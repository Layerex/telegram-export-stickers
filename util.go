package main

import (
	"time"
)

func FormatDate(date int32) string {
	return time.Unix(int64(date), 0).UTC().Format(time.RFC3339)
}

func Now() string {
	return time.Now().UTC().Format(time.RFC3339)
}
