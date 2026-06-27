package api

import (
	"time"

	"github.com/PlayingPossumHiss/possum_chat/internal/entity"
)

type GetStyleUC interface {
	GetCustomStyle() string
	GetMainStyle() (string, error)
}

type ListMessagesUC interface {
	ListMessages(forLast *time.Duration) []entity.Message
}

type OnlineGetter interface {
	GetOnline() map[entity.Source]int64
}
