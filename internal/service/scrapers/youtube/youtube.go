package youtube_scraper

import (
	"context"
	"log"
	"slices"
	"sync"
	"time"

	"github.com/PlayingPossumHiss/possum_chat/internal/entity"
)

type Service struct {
	streamKey     string
	cooldown      time.Duration
	youtubeClient YoutubeClient

	messages  []entity.Message
	messageMx *sync.Mutex
}

func New(
	ctx context.Context,
	streamKey string,
	cooldown time.Duration,
	youtubeClient YoutubeClient,
) *Service {
	scraper := &Service{
		streamKey:     streamKey,
		cooldown:      cooldown,
		youtubeClient: youtubeClient,
		messageMx:     &sync.Mutex{},
	}

	go scraper.watchChat(ctx)

	return scraper
}

func (s *Service) GetMessages() []entity.Message {
	s.messageMx.Lock()
	defer s.messageMx.Unlock()
	result := slices.Clone(s.messages)
	s.messages = nil

	return result
}

func (s *Service) watchChat(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			err := s.youtubeClient.Init(s.streamKey)
			if err != nil {
				log.Println(err)

				continue
			}

			err = s.scrap(ctx)
			if err != nil {
				log.Println(err)
			}
		}
	}
}

func (s *Service) scrap(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			for {
				comments, err := s.youtubeClient.GetMessages()
				if err != nil {
					return err
				}

				s.messageMx.Lock()
				s.messages = append(s.messages, comments...)
				s.messageMx.Unlock()

				time.Sleep(s.cooldown)
			}
		}
	}
}
