package vk_play_live_ws

import (
	"context"
	"fmt"
	"log"
	"net/http"

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
	client, _, err := dialer.DialContext(
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
	err := c.client.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf(`{"connect":{"token":"%s","name":"js"},"id":1}`, c.token)))
	if err != nil {
		return err
	}
	_, _, err = c.client.ReadMessage()
	if err != nil {
		return err
	}
	err = c.client.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf(`{"subscribe":{"channel":"channel-chat:%s"},"id":2}`, c.userID)))
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) ReadMessage() ([]byte, error) {
	_, msg, err := c.client.ReadMessage()
	return msg, err
}

func (c *Client) WriteMessage(rawMsg []byte) error {
	return c.client.WriteMessage(websocket.TextMessage, rawMsg)
}
