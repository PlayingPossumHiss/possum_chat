package utils_time

import "time"

type Clock interface {
	Now() time.Time
}

type DefaultClock struct{}

func (_ *DefaultClock) Now() time.Time {
	return time.Now()
}
