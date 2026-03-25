package api

import "github.com/PlayingPossumHiss/possum_chat/internal/entity"

type GetStyleUC interface {
	GetStyle() string
}

type ListMessagesUC interface {
	ListMessages() []entity.Message
}
