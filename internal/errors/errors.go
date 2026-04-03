package errors

import "errors"

var (
	ErrIsPing        = errors.New("is ping")
	ErrNoData        = errors.New("no data")
	ErrInvalidConfig = errors.New("invalid config")
)
