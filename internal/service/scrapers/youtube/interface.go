package youtube_scraper

import (
	"context"

	"github.com/PlayingPossumHiss/possum_chat/internal/entity"
)

type YoutubeClient interface {
	Init(streamKey string) error
	GetMessages() ([]entity.Message, error)
	GetLastTranslationID(ctx context.Context, userName string) (string, error)
}
