package main

import (
	carbon "github.com/golang-module/carbon/v2"
	"time"
)

// labels

func main() {
	// Convert between Carbon and Time
	// Convert Time.time into Carbon
	carbon.FromStdTime(time.Now())
	// Convert Carbon into Time.time
	carbon.Now().ToStdTime()

	// Whether is now time
	carbon.Now().IsNow() // true
	// Whether is future time
	carbon.Tomorrow().IsFuture() // true
	// Whether is pass time
	carbon.Yesterday().IsPast() // true
}
