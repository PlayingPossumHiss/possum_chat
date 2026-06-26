package youtube_scraper

import (
	"context"

	"github.com/PlayingPossumHiss/possum_chat/internal/entity"
)

type YoutubeClient interface {
	Init(streamKey string) error
	GetMessages() ([]entity.Message, error)
	GetLastTranslationID(ctx context.Context, userName string) (string, error)
	GetOnline(ctx context.Context, liveID string) (int64, error)
}

type ConfigStorage interface {
	Config() entity.Config
}
