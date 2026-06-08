package kick

import (
	"context"

	"github.com/PlayingPossumHiss/possum_chat/internal/entity"
)

type Client interface {
	GetRoomIDByUserName(ctx context.Context, userName string) (int64, error)
	Listen(
		callback func(entity.Message),
		roomID int64,
	) error
	Close() error
}

type ConfigStorage interface {
	Config() entity.Config
}
