package ui

import (
	"context"

	"github.com/PlayingPossumHiss/possum_chat/internal/entity"
)

type Scraper interface {
	Run(ctx context.Context)
	Stop()
	Status() entity.ScraperState
	GetConnectionConfig() string
	ConnectionConfigUpdateOption(string) entity.ConfigUpdateOption
}

type ConfigStorage interface {
	UpdateConfig(opts []entity.ConfigUpdateOption) error
	Config() entity.Config
}
