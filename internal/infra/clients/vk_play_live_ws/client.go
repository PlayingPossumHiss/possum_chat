package vk_play_live_ws

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"slices"
	"time"

	"github.com/PlayingPossumHiss/possum_chat/internal/entity"
	app_errors "github.com/PlayingPossumHiss/possum_chat/internal/errors"
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
		return err
	}

	c.client = client
	err = c.connectToChat()
	if err != nil {
		c.Close()

		return err
	}

	return nil
}

func (c *Client) Close() {
	err := c.client.Close()
	if err != nil {
		log.Println(err)
	}
	c.client = nil
}

func (c *Client) connectToChat() error {
	err := c.client.WriteMessage(
		websocket.TextMessage,
		[]byte(fmt.Sprintf(`{"connect":{"token":"%s","name":"js"},"id":1}`, c.token)),
	)
	if err != nil {
		return err
	}
	_, _, err = c.client.ReadMessage()
	if err != nil {
		return err
	}
	err = c.client.WriteMessage(
		websocket.TextMessage,
		[]byte(fmt.Sprintf(`{"subscribe":{"channel":"channel-chat:%s"},"id":2}`, c.userID)),
	)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) ReadMessage() (entity.Message, error) {
	_, rawMsg, err := c.client.ReadMessage()
	if err != nil {
		return entity.Message{}, err
	}

	if slices.Equal(rawMsg, []byte("{}")) {
		return entity.Message{}, app_errors.ErrIsPing
	}

	return getMessageFromBytes(rawMsg)
}

func (c *Client) WritePong() error {
	return c.client.WriteMessage(websocket.TextMessage, []byte("{}"))
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
