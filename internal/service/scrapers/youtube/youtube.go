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

	state       entity.ScraperState
	stateMx     *sync.Mutex
	watchCancel context.CancelFunc

	messages  []entity.Message
	messageMx *sync.Mutex
}

func New(
	userName string,
	youtubeClient YoutubeClient,
) *Service {
	scraper := &Service{
		userName:      userName,
		youtubeClient: youtubeClient,
		messageMx:     &sync.Mutex{},
		stateMx:       &sync.Mutex{},
		state:         entity.ScraperStateStopped,
	}

	return scraper
}

func (s *Service) Run(ctx context.Context) {
	logger.Info("start youtube scraper")
	s.stateMx.Lock()
	newCtx, cancel := context.WithCancel(ctx)
	s.watchCancel = cancel
	go s.watchChat(newCtx)
}

func (s *Service) Stop() {
	logger.Info("stop youtube scraper")
	s.stateMx.Lock()
	defer s.stateMx.Unlock()

	s.watchCancel()
	s.state = entity.ScraperStateStopped
}

func (s *Service) Status() entity.ScraperState {
	return s.state
}

func (s *Service) GetMessages() []entity.Message {
	s.messageMx.Lock()
	defer s.messageMx.Unlock()
	result := slices.Clone(s.messages)
	s.messages = nil

	return result
}

func (s *Service) watchChat(ctx context.Context) {
	firstTry := true
	const secondsToWait = 5
	for {
		select {
		case <-ctx.Done():
			return
		default:
			if !firstTry {
				// Чтобы не забанили
				time.Sleep(secondsToWait * time.Second)
				s.stateMx.Lock()
			}
			firstTry = false

			err := s.initChat(ctx)
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
	}
}

func (s *Service) initChat(ctx context.Context) error {
	defer s.stateMx.Unlock()

	streamKey, err := s.youtubeClient.GetLastTranslationID(ctx, s.userName)
	if err != nil {
		return err
	}

	err = s.youtubeClient.Init(streamKey)
	if err != nil {
		return err
	}

	s.state = entity.ScraperStateActive

	return nil
}

func (s *Service) scrap(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
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
