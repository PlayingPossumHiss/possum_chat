package youtube_scraper

import (
	"context"
	"fmt"
	"slices"
	"sync"
	"time"

	"github.com/PlayingPossumHiss/possum_chat/internal/entity"
	app_errors "github.com/PlayingPossumHiss/possum_chat/internal/errors"
	"github.com/PlayingPossumHiss/possum_chat/internal/service/logger"
)

type Service struct {
	configStorage ConfigStorage
	youtubeClient YoutubeClient

	state       entity.ScraperState
	stateMx     *sync.Mutex
	watchCancel context.CancelFunc

	messages  []entity.Message
	messageMx *sync.Mutex
}

func New(
	configStorage ConfigStorage,
	youtubeClient YoutubeClient,
) *Service {
	scraper := &Service{
		configStorage: configStorage,
		youtubeClient: youtubeClient,
		messageMx:     &sync.Mutex{},
		stateMx:       &sync.Mutex{},
		state:         entity.ScraperStateStopped,
	}

	return scraper
}

func (s *Service) GetConnectionConfig() string {
	return s.configStorage.Config().Connections.Youtube.ChannelName
}

func (s *Service) ConnectionConfigUpdateOption(newValue string) entity.ConfigUpdateOption {
	return func(c *entity.Config) {
		c.Connections.Youtube.ChannelName = newValue
	}
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

	channelName := s.GetConnectionConfig()
	if len(channelName) == 0 {
		return fmt.Errorf("%w: can't get channel name for vk play live", app_errors.ErrInvalidConfig)
	}
	streamKey, err := s.youtubeClient.GetLastTranslationID(
		ctx,
		channelName,
	)
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
