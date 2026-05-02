package twitch_irc_client

import (
	"fmt"
	"time"

	"github.com/PlayingPossumHiss/possum_chat/internal/entity"
	"github.com/gempir/go-twitch-irc/v4"
)

type Client struct {
	wsConnect *twitch.Client
}

func New() *Client {
	wsConnect := twitch.NewAnonymousClient()

	return &Client{
		wsConnect: wsConnect,
	}
}

func (c *Client) Close() error {
	return c.wsConnect.Disconnect()
}

func (c *Client) Listen(
	callback func(entity.Message),
	channelName string,
) error {
	c.wsConnect.OnPrivateMessage(func(message twitch.PrivateMessage) {
		callback(entity.Message{
			ID:     fmt.Sprintf("twitch_%s", message.ID),
			Source: entity.SourceTwitch,
			User:   message.User.DisplayName,
			Content: []entity.MessageContentItem{
				{
					Type:  entity.MessageContentItemTypeText,
					Value: message.Message,
				},
			},
			CreatedAt: time.Now(),
		})
	})

	c.wsConnect.Join(channelName)

	// TODO: Сюда надо добавить реконект
	// https://github.com/PlayingPossumHiss/possum_chat/issues/26
	err := c.wsConnect.Connect()
	if err != nil {
		err = fmt.Errorf("failed to connect to twitch ws chat: %w", err)

		return err
	}

	return nil
}
