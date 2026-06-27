package kick

import (
	"context"

	"github.com/PlayingPossumHiss/possum_chat/internal/entity"
)

type Client interface {
	GetRoomIDByUserName(ctx context.Context, userName string) (int64, error)
	GetOnline(ctx context.Context, userName string) (int64, error)
	Listen(
		roomID int64,
	) (chan entity.Message, error)
	Close() error
}

type ConfigStorage interface {
	Config() entity.Config
}
