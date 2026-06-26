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

	online    int64
	streamKey string
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

func (s *Service) Run(ctx context.Context) {
	logger.Info("start youtube scraper")
	newCtx, cancel := context.WithCancel(ctx)
	s.watchCancel = cancel
	s.state = entity.ScraperStateStarting
	go s.watchChat(newCtx)
	go s.watchOnline(newCtx)
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

func (s *Service) GetOnline() int64 {
	return s.online
}

func (s *Service) watchOnline(
	ctx context.Context,
) {
	for {
		select {
		case <-ctx.Done():
			logger.Warn("youtube online watcher is stopped by contex cancel")

			return
		default:
			if s.streamKey == "" {
				time.Sleep(time.Second)

				continue
			}

			online, err := s.youtubeClient.GetOnline(ctx, s.streamKey)
			if err != nil {
				logger.Error(fmt.Errorf("error on get youtube online, %w", err))
			}
			s.online = online
			time.Sleep(time.Minute)
		}
	}
}

func (s *Service) watchChat(ctx context.Context) {
	firstTry := true
	const secondsToWait = 5
	for {
		select {
		case <-ctx.Done():
			return
		default:
			// Чтобы не забанили
			if !firstTry {
				time.Sleep(secondsToWait * time.Second)
			}
			firstTry = false

			// Если остановили, то завершаем попытки
			if s.state == entity.ScraperStateStopped {
				return
			}

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
	s.stateMx.Lock()
	defer s.stateMx.Unlock()

	channelName := s.configStorage.Config().Connections.Youtube.ChannelName
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
	s.streamKey = streamKey

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
