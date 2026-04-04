package vk_play_live_ws

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"slices"
	"time"

	"github.com/PlayingPossumHiss/possum_chat/internal/entity"
	app_errors "github.com/PlayingPossumHiss/possum_chat/internal/errors"
	"github.com/PlayingPossumHiss/possum_chat/internal/service/logger"
	"github.com/gorilla/websocket"
)

type Client struct {
	client *websocket.Conn
	token  string
	userID string
}

func New() *Client {
	return &Client{}
}

func (c *Client) Init(
	ctx context.Context,
	token string,
	userID string,
) error {
	c.token = token
	c.userID = userID
	dialer := websocket.DefaultDialer
	dialer.EnableCompression = true
	client, _, err := dialer.DialContext( //nolint // закрывается в клозере
		ctx,
		"wss://pubsub.live.vkvideo.ru/connection/websocket?cf_protocol_version=v2",
		http.Header{
			"Origin": []string{"https://live.vkvideo.ru"},
		},
	)
	if err != nil {
		err = fmt.Errorf("failed to create ws request for vk play live: %w", err)
		return err
	}

	c.client = client
	err = c.connectToChat()
	if err != nil {
		c.Close()

		err = fmt.Errorf("failed to connect to chat for vk play live: %w", err)
		return err
	}

	return nil
}

func (c *Client) Close() {
	err := c.client.Close()
	if err != nil {
		err = fmt.Errorf("failed close ws connect for vk play live: %w", err)
		logger.Error(err.Error())
	}
	c.client = nil
}

func (c *Client) connectToChat() error {
	err := c.client.WriteMessage(
		websocket.TextMessage,
		[]byte(fmt.Sprintf(`{"connect":{"token":"%s","name":"js"},"id":1}`, c.token)),
	)
	if err != nil {
		err = fmt.Errorf("failed to write connect to chat request for vk play live: %w", err)
		return err
	}
	_, _, err = c.client.ReadMessage()
	if err != nil {
		err = fmt.Errorf("failed to read after write connect to chat request for vk play live: %w", err)
		return err
	}
	err = c.client.WriteMessage(
		websocket.TextMessage,
		[]byte(fmt.Sprintf(`{"subscribe":{"channel":"channel-chat:%s"},"id":2}`, c.userID)),
	)
	if err != nil {
		err = fmt.Errorf("failed to write subscribe to chat request for vk play live: %w", err)
		return err
	}

	return nil
}

func (c *Client) ReadMessage() (entity.Message, error) {
	_, rawMsg, err := c.client.ReadMessage()
	if err != nil {
		err = fmt.Errorf("failed to read from ws chat for vk play live: %w", err)
		return entity.Message{}, err
	}

	if slices.Equal(rawMsg, []byte("{}")) {
		return entity.Message{}, app_errors.ErrIsPing
	}

	return getMessageFromBytes(rawMsg)
}

func (c *Client) WritePong() error {
	err := c.client.WriteMessage(websocket.TextMessage, []byte("{}"))
	if err != nil {
		err = fmt.Errorf("failed to send ws pong for vk play live: %w", err)
		return err
	}

	return nil
}

func getMessageFromBytes(rawMsg []byte) (entity.Message, error) {
	msg := message{}
	err := json.Unmarshal(rawMsg, &msg)
	if err != nil {
		err = fmt.Errorf("failed parse message for vk play live: %w", err)
		return entity.Message{}, err
	}
	if msg.Push.Pub.Data.Type != "message" {
		return entity.Message{}, nil
	}
	chatMessage := entity.Message{
		ID:        fmt.Sprintf("vk_play_live_%d", msg.Push.Pub.Data.Data.ID),
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
