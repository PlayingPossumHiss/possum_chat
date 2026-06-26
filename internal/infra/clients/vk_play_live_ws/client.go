package vk_play_live_ws

import (
	"context"
	"encoding/json"
	"errors"
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
) (entity.VkStreamData, error) {
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

		return entity.VkStreamData{}, err
	}

	c.client = client
	err = c.connectToChat()
	if err != nil {
		c.Close()

		err = fmt.Errorf("failed to connect to chat for vk play live: %w", err)

		return entity.VkStreamData{}, err
	}

	result := entity.VkStreamData{
		MessageCh: make(chan entity.Message),
		Online:    make(chan int64),
		Error:     make(chan error),
	}

	go c.doScrapCycle(ctx, result)

	return result, nil
}

func (c *Client) doScrapCycle(
	ctx context.Context,
	result entity.VkStreamData,
) {
	defer func() {
		close(result.Error)
		close(result.MessageCh)
		close(result.Online)
	}()

	for {
		select {
		case <-ctx.Done():
			result.Error <- ctx.Err()

			return
		default:
			err := c.readMessage(result)
			if errors.Is(err, app_errors.ErrIsPing) {
				err = c.writePong()
				if err != nil {
					result.Error <- err

					return
				}

				continue
			} else if err != nil {
				result.Error <- err

				return
			}
		}
	}
}

func (c *Client) Close() {
	if c.client == nil {
		return
	}

	err := c.client.Close()
	if err != nil {
		err = fmt.Errorf("failed close ws connect for vk play live: %w", err)
		logger.Error(err)
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
	err = c.client.WriteMessage(
		websocket.TextMessage,
		[]byte(fmt.Sprintf(`{"subscribe":{"channel":"channel-info:%s"},"id":3}`, c.userID)),
	)
	if err != nil {
		err = fmt.Errorf("failed to write subscribe to channel viewers request for vk play live: %w", err)

		return err
	}

	return nil
}

func (c *Client) readMessage(result entity.VkStreamData) error {
	_, rawMsg, err := c.client.ReadMessage()
	if err != nil {
		err = fmt.Errorf("failed to read from ws chat for vk play live: %w", err)

		return err
	}

	if slices.Equal(rawMsg, []byte("{}")) {
		return app_errors.ErrIsPing
	}

	logger.Debug(fmt.Sprintf("message from vk paly live: %s", string(rawMsg)))

	return getMessageFromBytes(result, rawMsg)
}

func (c *Client) writePong() error {
	err := c.client.WriteMessage(websocket.TextMessage, []byte("{}"))
	if err != nil {
		err = fmt.Errorf("failed to send ws pong for vk play live: %w", err)

		return err
	}

	return nil
}

func getMessageFromBytes(result entity.VkStreamData, rawMsg []byte) error {
	msg := message{}
	err := json.Unmarshal(rawMsg, &msg)
	if err != nil {
		err = fmt.Errorf("failed parse message for vk play live: %w", err)

		return err
	}

	switch msg.Push.Pub.Data.Type {
	case "message":
		chatMessage := entity.Message{
			ID:        fmt.Sprintf("vk_play_live_%d", msg.Push.Pub.Data.Data.ID),
			Source:    entity.SourceVkPlayLive,
			User:      msg.Push.Pub.Data.Data.Author.Name,
			CreatedAt: time.Now(),
			Content:   getMessageContent(msg.Push.Pub.Data.Data.Data),
		}

		if len(chatMessage.Content) == 0 {
			logger.Warn("can't parse message for vk play")

			return nil
		}

		result.MessageCh <- chatMessage
	case "stream_slot_online_status":
		result.Online <- msg.Push.Pub.Data.Data.Stream.Viewers
	default:
		fmt.Println(string(rawMsg))
	}

	return nil
}

func getMessageContent(messageData []messageData) []entity.MessageContentItem {
	result := make([]entity.MessageContentItem, 0, len(messageData))

	for _, messagePart := range messageData {
		switch messagePart.Type {
		case "text", "link":
			testPartContent := []any{}
			err := json.Unmarshal([]byte(messagePart.Content), &testPartContent)
			if err != nil {
				continue
			}
			if len(testPartContent) > 0 {
				subText, ok := testPartContent[0].(string)
				if !ok {
					continue
				}
				result = append(result, entity.MessageContentItem{
					Type:  entity.MessageContentItemTypeText,
					Value: subText,
				})
			}
		case "smile":
			result = append(result, entity.MessageContentItem{
				Type:  entity.MessageContentItemTypeImage,
				Value: messagePart.SmallUrl,
			})
		default:
			continue
		}
	}

	return result
}
