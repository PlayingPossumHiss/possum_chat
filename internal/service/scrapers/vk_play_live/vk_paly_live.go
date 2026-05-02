package vk_play_live

import (
	"context"
	"errors"
	"slices"
	"strconv"
	"sync"
	"time"

	"github.com/PlayingPossumHiss/possum_chat/internal/entity"
	app_errors "github.com/PlayingPossumHiss/possum_chat/internal/errors"
	"github.com/PlayingPossumHiss/possum_chat/internal/service/logger"
)

type Service struct {
	streamKey     string
	vkPlayLiveApi VkPlayLiveApi
	vkPlayLiveWs  VkPlayLiveWs

	state       entity.ScraperState
	stateMx     *sync.Mutex
	watchCancel context.CancelFunc

	messages  []entity.Message
	messageMx *sync.Mutex
}

func New(
	ctx context.Context,
	streamKey string,
	vkPlayLiveApi VkPlayLiveApi,
	vkPlayLiveWs VkPlayLiveWs,
) (*Service, error) {
	userID, err := vkPlayLiveApi.GetUserID(ctx, streamKey)
	if err != nil {
		return nil, err
	}
	scraper := &Service{
		streamKey:     strconv.Itoa(userID),
		vkPlayLiveApi: vkPlayLiveApi,
		vkPlayLiveWs:  vkPlayLiveWs,
		messageMx:     &sync.Mutex{},
		stateMx:       &sync.Mutex{},
		state:         entity.ScraperStateStopped,
	}

	return scraper, nil
}

func (s *Service) Run(ctx context.Context) {
	logger.Info("start vk play live scraper")
	s.stateMx.Lock()
	newCtx, cancel := context.WithCancel(ctx)
	s.watchCancel = cancel
	go s.watchChat(newCtx)
}

func (s *Service) Stop() {
	logger.Info("stop vk play live scraper")
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
	firstRun := true
	for {
		select {
		case <-ctx.Done():
			logger.Warn("vk play live watcher is stopped by contex cancel")

			return
		default:
			err := s.scrap(ctx, firstRun)
			if err != nil {
				logger.Error(err)
			}
			firstRun = false

			// чтобы не словить бан
			time.Sleep(time.Second)
		}
	}
}

func (s *Service) scrap(
	ctx context.Context,
	firstRun bool,
) error {
	if !firstRun {
		s.stateMx.Lock()
	}

	token, err := s.vkPlayLiveApi.GetWsToken(ctx)
	if err != nil {
		return err
	}

	err = s.vkPlayLiveWs.Init(ctx, token, s.streamKey)
	if err != nil {
		return err
	}
	s.state = entity.ScraperStateActive
	s.stateMx.Unlock()

	defer func() {
		s.vkPlayLiveWs.Close()
	}()

	return s.doScrapCycle(ctx)
}

func (s *Service) doScrapCycle(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			chatMessage, err := s.vkPlayLiveWs.ReadMessage()
			if errors.Is(err, app_errors.ErrIsPing) {
				err = s.vkPlayLiveWs.WritePong()
				if err != nil {
					return err
				}

				continue
			} else if err != nil {
				return err
			}
			if chatMessage == nil {
				continue
			}
			s.messageMx.Lock()
			s.messages = append(s.messages, *chatMessage)
			s.messageMx.Unlock()
		}
	}
}
