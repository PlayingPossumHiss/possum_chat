package run_watch_scrapers

import (
	"context"
)

// UseCase обходит все скрейперы комментов и собрав с каждого его отдает это в хранилище
type UseCase struct {
	scrapers            []Scraper
	messageQueueService MessageQueueService
}

func New(
	scrapers []Scraper,
	messageQueueService MessageQueueService,
) *UseCase {
	return &UseCase{
		scrapers:            scrapers,
		messageQueueService: messageQueueService,
	}
}

func (uc *UseCase) Run(ctx context.Context) error {
	for _, scraper := range uc.scrapers {
		messages := scraper.GetMessages()
		uc.messageQueueService.PushMessages(messages)
	}

	return nil
}
