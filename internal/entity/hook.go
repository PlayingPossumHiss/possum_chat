package entity

import (
	"context"
)

type Hook interface {
	Handle(ctx context.Context, message Message) error
}
