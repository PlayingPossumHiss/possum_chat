package vk_play_live

import (
	"context"

	"github.com/PlayingPossumHiss/possum_chat/internal/entity"
)

type VkPlayLiveApi interface {
	GetWsToken(ctx context.Context) (string, error)
	GetUserID(ctx context.Context, userName string) (int, error)
}

type VkPlayLiveWs interface {
	Init(
		ctx context.Context,
		token string,
		userID string,
	) error
	ReadMessage() (*entity.Message, error)
	WritePong() error
	Close()
}
