package container

import (
	"context"

	"github.com/PlayingPossumHiss/possum_chat/internal/entity"
)

type Scraper interface {
	GetMessages() []entity.Message
	Run(ctx context.Context)
	Stop()
	Status() entity.ScraperState
}
