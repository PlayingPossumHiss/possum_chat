package api

import (
	"time"

	"github.com/PlayingPossumHiss/possum_chat/internal/entity"
)

type GetStyleUC interface {
	GetStyle() string
}

type ListMessagesUC interface {
	ListMessages(forLast *time.Duration) []entity.Message
}
