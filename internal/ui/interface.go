package ui

import (
	"context"

	"github.com/PlayingPossumHiss/possum_chat/internal/entity"
)

type Scraper interface {
	Run(ctx context.Context)
	Stop()
	Status() entity.ScraperState
}

type ConfigStorage interface {
	UpdateConfig(opts []entity.ConfigUpdateOption) error
	Config() entity.Config
}

type LanguageProvider interface {
	Local(name entity.LanguageTextConstant) string
}

type SendTestMessagesUseCase interface {
	SendTestMessages(messageText string)
}
