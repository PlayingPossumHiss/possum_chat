package youtube_scraper

import (
	"context"
	"fmt"
	"log"
	"slices"
	"sync"
	"time"

	"github.com/PlayingPossumHiss/possum_chat/internal/entity"
)

type Service struct {
	streamKey     string
	ctx           context.Context
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
		ctx:           ctx,
		youtubeClient: youtubeClient,
		messageMx:     &sync.Mutex{},
	}

	go scraper.watchChat()

	return scraper
}

func (s *Service) GetMessages() []entity.Message {
	s.messageMx.Lock()
	defer s.messageMx.Unlock()
	result := slices.Clone(s.messages)
	s.messages = nil
	return result
}

func (s *Service) watchChat() {
	for {
		err := s.youtubeClient.Init(s.streamKey)
		if err != nil {
			fmt.Println(err)
			continue
		}

		err = s.scrap()
		if err != nil {
			log.Println(err)
		}
	}
}

func (s *Service) scrap() error {
	for {
		select {
		case <-s.ctx.Done():
			return s.ctx.Err()
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
