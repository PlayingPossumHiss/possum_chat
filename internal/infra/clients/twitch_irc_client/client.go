package twitch_irc_client

import (
	"fmt"
	"sort"
	"time"

	"github.com/PlayingPossumHiss/possum_chat/internal/entity"
	"github.com/PlayingPossumHiss/possum_chat/internal/service/logger"
	"github.com/gempir/go-twitch-irc/v4"
)

type Client struct {
	wsConnect *twitch.Client
}

type emotePosition struct {
	start   int
	end     int
	emoteID string
}

const twitchEmoteURLTemplate = "https://static-cdn.jtvnw.net/emoticons/v2/%s/default/light/1.0#e=0"

func New() *Client {
	wsConnect := twitch.NewAnonymousClient()

	return &Client{
		wsConnect: wsConnect,
	}
}

func (c *Client) Close() error {
	return c.wsConnect.Disconnect()
}

func (c *Client) Listen(channelName string) chan entity.Message {
	result := make(chan entity.Message)
	c.wsConnect.OnPrivateMessage(func(message twitch.PrivateMessage) {
		logger.Debug(fmt.Sprintf("message from youtube: %s", message.Raw))

		result <- entity.Message{
			ID:        fmt.Sprintf("twitch_%s", message.ID),
			Source:    entity.SourceTwitch,
			User:      message.User.DisplayName,
			Content:   buildMessageContent(message.Message, message.Emotes),
			CreatedAt: time.Now(),
		}
	})

	c.wsConnect.Join(channelName)

	go func() {
		defer close(result)
		err := c.wsConnect.Connect()
		if err != nil {
			err = fmt.Errorf("failed to connect to twitch ws chat: %w", err)

			logger.Error(err)
		}
	}()

	return result
}

func collectEmotePositions(emotes []*twitch.Emote) []emotePosition {
	positions := make([]emotePosition, 0, len(emotes))

	for _, emote := range emotes {
		for _, emotePos := range emote.Positions {
			positions = append(positions, emotePosition{
				start:   emotePos.Start,
				end:     emotePos.End,
				emoteID: emote.ID,
			})
		}
	}

	sort.Slice(positions, func(i, j int) bool {
		return positions[i].start < positions[j].start
	})

	return positions
}

func buildMessageContent(message string, emotes []*twitch.Emote) []entity.MessageContentItem {
	positions := collectEmotePositions(emotes)

	if len(positions) == 0 {
		return []entity.MessageContentItem{
			{
				Type:  entity.MessageContentItemTypeText,
				Value: message,
			},
		}
	}

	runes := []rune(message)
	content := make([]entity.MessageContentItem, 0, len(positions)*2+1)
	cursor := 0

	for _, position := range positions {
		if position.start < cursor || position.start < 0 || position.end < position.start ||
			position.start >= len(runes) || position.end >= len(runes) {
			continue
		}

		if position.start > cursor {
			content = append(content, entity.MessageContentItem{
				Type:  entity.MessageContentItemTypeText,
				Value: string(runes[cursor:position.start]),
			})
		}

		content = append(content, entity.MessageContentItem{
			Type:  entity.MessageContentItemTypeImage,
			Value: fmt.Sprintf(twitchEmoteURLTemplate, position.emoteID),
		})

		cursor = position.end + 1
	}

	if cursor < len(runes) {
		content = append(content, entity.MessageContentItem{
			Type:  entity.MessageContentItemTypeText,
			Value: string(runes[cursor:]),
		})
	}

	return content
}
