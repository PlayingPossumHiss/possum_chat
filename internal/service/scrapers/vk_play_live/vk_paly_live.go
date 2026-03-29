package vk_play_live

import (
	"context"
	"errors"
	"log"
	"slices"
	"strconv"
	"sync"

	"github.com/PlayingPossumHiss/possum_chat/internal/entity"
	app_errors "github.com/PlayingPossumHiss/possum_chat/internal/errors"
)

type Service struct {
	streamKey     string
	vkPlayLiveApi VkPlayLiveApi
	vkPlayLiveWs  VkPlayLiveWs

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
	}

	go scraper.watchChat(ctx)

	return scraper, nil
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
		err := s.scrap(ctx)
		if err != nil {
			log.Println(err)
		}
	}
}

func (s *Service) scrap(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		token, err := s.vkPlayLiveApi.GetWsToken(ctx)
		if err != nil {
			return err
		}

		err = s.vkPlayLiveWs.Init(ctx, token, s.streamKey)
		if err != nil {
			return err
		}

		defer func() {
			s.vkPlayLiveWs.Close()
		}()

		return s.doScrapCycle(ctx)
	}
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
			if chatMessage == (entity.Message{}) {
				continue
			}
			s.messageMx.Lock()
			s.messages = append(s.messages, chatMessage)
			s.messageMx.Unlock()
		}
	}
}
