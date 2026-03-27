package youtube_scraper

import "github.com/PlayingPossumHiss/possum_chat/internal/entity"

type YoutubeClient interface {
	Init(streamKey string) error
	GetMessages() ([]entity.Message, error)
}
