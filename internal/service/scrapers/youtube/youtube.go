package youtube_scraper

import (
	"context"
	"slices"
	"sync"
	"time"

	"github.com/PlayingPossumHiss/possum_chat/internal/entity"
	"github.com/PlayingPossumHiss/possum_chat/internal/service/logger"
)

type Service struct {
	userName      string
	youtubeClient YoutubeClient

	messages  []entity.Message
	messageMx *sync.Mutex
}

func New(
	ctx context.Context,
	userName string,
	youtubeClient YoutubeClient,
) *Service {
	scraper := &Service{
		userName:      userName,
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
			streamKey, err := s.youtubeClient.GetLastTranslationID(ctx, s.userName)
			if err != nil {
				logger.Error(err)

				continue
			}

			err = s.youtubeClient.Init(streamKey)
			if err != nil {
				logger.Error(err)

				continue
			}

			// TODO: надо отсюда выходить, если трансляция закончилась
			// https://github.com/PlayingPossumHiss/possum_chat/issues/25
			err = s.scrap(ctx)
			if err != nil {
				logger.Error(err)
			}
		}

		// Чтобы не забанили
		time.Sleep(time.Second)
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
			}
		}
	}
}
