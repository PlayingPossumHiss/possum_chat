package vk_play_live

import (
	"context"
	"encoding/json"
	"log"
	"slices"
	"strconv"
	"sync"
	"time"

	"github.com/PlayingPossumHiss/possum_chat/internal/entity"
)

type Service struct {
	streamKey     string
	ctx           context.Context
	cooldown      time.Duration
	vkPlayLiveApi VkPlayLiveApi
	vkPlayLiveWs  VkPlayLiveWs

	messages  []entity.Message
	messageMx *sync.Mutex
}

func New(
	ctx context.Context,
	streamKey string,
	cooldown time.Duration,
	vkPlayLiveApi VkPlayLiveApi,
	vkPlayLiveWs VkPlayLiveWs,
) (*Service, error) {
	userID, err := vkPlayLiveApi.GetUserID(ctx, streamKey)
	if err != nil {
		return nil, err
	}
	scraper := &Service{
		streamKey:     strconv.Itoa(userID),
		cooldown:      cooldown,
		ctx:           ctx,
		vkPlayLiveApi: vkPlayLiveApi,
		vkPlayLiveWs:  vkPlayLiveWs,
		messageMx:     &sync.Mutex{},
	}

	go scraper.watchChat()

	return scraper, nil
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
		err := s.scrap()
		if err != nil {
			log.Println(err)
		}
	}
}

func (s *Service) scrap() error {
	select {
	case <-s.ctx.Done():
		return s.ctx.Err()
	default:
		token, err := s.vkPlayLiveApi.GetWsToken(s.ctx)
		if err != nil {
			return err
		}

		err = s.vkPlayLiveWs.Init(s.ctx, token, s.streamKey)
		if err != nil {
			return err
		}

		defer func() {
			s.vkPlayLiveWs.Close()
		}()

		for {
			select {
			case <-s.ctx.Done():
				return s.ctx.Err()
			default:
				rawMsg, err := s.vkPlayLiveWs.ReadMessage()
				if err != nil {
					return err
				}
				if slices.Equal(rawMsg, []byte("{}")) {
					err = s.vkPlayLiveWs.WriteMessage(rawMsg)
					if err != nil {
						return err
					}
					continue
				}
				chatMessage, err := getMessageFromBytes(rawMsg)
				if err != nil {
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
}

func getMessageFromBytes(rawMsg []byte) (entity.Message, error) {
	msg := message{}
	err := json.Unmarshal(rawMsg, &msg)
	if err != nil {
		return entity.Message{}, err
	}
	if msg.Push.Pub.Data.Type != "message" {
		return entity.Message{}, nil
	}
	chatMessage := entity.Message{
		ID:        strconv.Itoa(msg.Push.Pub.Data.Data.ID),
		Source:    entity.SourceVkPlayLive,
		User:      msg.Push.Pub.Data.Data.Author.Name,
		CreatedAt: time.Now(),
	}
	for _, textPart := range msg.Push.Pub.Data.Data.Data {
		if textPart.Type != "text" {
			continue
		}
		testPartContent := []any{}
		err = json.Unmarshal([]byte(textPart.Content), &testPartContent)
		if err != nil {
			continue
		}
		if len(testPartContent) > 0 {
			subText, ok := testPartContent[0].(string)
			if !ok {
				continue
			}
			chatMessage.Text += subText
		}
	}
	if chatMessage.Text == "" {
		return entity.Message{}, nil
	}

	return chatMessage, nil
}

type message struct {
	Push struct {
		Pub struct {
			Data struct {
				Type string `json:"type"` // message
				Data struct {
					ID        int   `json:"id"`
					CreatedAt int64 `json:"createdAt"`
					Author    struct {
						Name string `json:"displayName"`
					} `json:"author"`
					Data []struct {
						Content string `json:"content"`
						Type    string `json:"type"` // text
					} `json:"data"`
				} `json:"data"`
			} `json:"data"`
		} `json:"pub"`
	} `json:"push"`
}
