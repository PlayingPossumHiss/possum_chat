package vk_play_live

import (
	"context"
	"encoding/json"
	"log"
	"net/url"
	"slices"
	"strconv"
	"sync"
	"time"

	"github.com/PlayingPossumHiss/possum_chat/internal/entity"
	"github.com/gorilla/websocket"
)

type Service struct {
	streamKey string
	ctx       context.Context
	cooldown  time.Duration

	messages  []entity.Message
	messageMx *sync.Mutex
}

func New(
	ctx context.Context,
	streamKey string,
	cooldown time.Duration,
) *Service {
	scraper := &Service{
		streamKey: streamKey,
		cooldown:  cooldown,
		ctx:       ctx,
		messageMx: &sync.Mutex{},
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
		chatUrl := url.URL{
			Scheme: "ws",
			Host:   "pubsub.live.vkvideo.ru",
			Path:   "/connection/websocket?cf_protocol_version=v2",
		}

		client, _, err := websocket.DefaultDialer.Dial(
			chatUrl.String(),
			nil,
		)

		if err != nil {
			return err
		}
		defer client.Close()

		for {
			select {
			case <-s.ctx.Done():
				return s.ctx.Err()
			default:
				_, rawMsg, err := client.ReadMessage()
				if err != nil {
					return err
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
		ID:     strconv.Itoa(msg.Push.Pub.Data.Data.ID),
		Source: entity.SourceVkPlayLive,
		User:   msg.Push.Pub.Data.Data.Author.Name,
	}
	for _, textPart := range msg.Push.Pub.Data.Data.Data {
		if textPart.Type != "text" {
			continue
		}
		testPartContent := []any{}
		err = json.Unmarshal([]byte(textPart.Content), &testPartContent)
		if err != nil {
			return entity.Message{}, err
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
