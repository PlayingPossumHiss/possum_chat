package twitch_irc_client

import (
	"fmt"
	"log"
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

func (c *Client) Listen(
	callback func(entity.Message),
	channelName string,
) {
	c.wsConnect.OnPrivateMessage(func(message twitch.PrivateMessage) {
		callback(entity.Message{
			ID:        fmt.Sprintf("twitch_%s", message.ID),
			Source:    entity.SourceTwitch,
			User:      message.User.DisplayName,
			Text:      message.Message,
			CreatedAt: time.Now(),
		})
	})

	c.wsConnect.Join(channelName)

	go func() {
		err := c.wsConnect.Connect()
		if err != nil {
			log.Println(err)

			return
		}
	}()
}
