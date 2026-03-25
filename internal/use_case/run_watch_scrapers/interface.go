package run_watch_scrapers

import "github.com/PlayingPossumHiss/possum_chat/internal/entity"

type MessageQueueService interface {
	PushMessages(messages []entity.Message)
}

type Scraper interface {
	GetMessages() []entity.Message
}

type ConfigStorage interface {
	Config() entity.Config
}
