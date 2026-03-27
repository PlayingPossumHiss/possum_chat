package youtube_client

import (
	"fmt"
	"math/rand"
	"net/http"

	"github.com/PlayingPossumHiss/possum_chat/internal/entity"
	yt_chat "github.com/epjane/youtube-live-chat-downloader/v2"
	"github.com/google/uuid"
)

type Client struct {
	continuation string
	cfg          yt_chat.YtCfg
}

func New() *Client {
	return &Client{}
}

func (c *Client) Init(streamKey string) error {
	customCookies := []*http.Cookie{
		{Name: "PREF",
			Value:  "tz=Europe.Rome",
			MaxAge: 300},
		{Name: "CONSENT",
			Value:  fmt.Sprintf("YES+yt.432048971.it+FX+%d", 100+rand.Intn(999-100+1)),
			MaxAge: 300},
	}
	yt_chat.AddCookies(customCookies)

	continuation, cfg, err := yt_chat.ParseInitialData(fmt.Sprintf("https://www.youtube.com/watch?v=%s", streamKey))
	if err != nil {
		return err
	}

	c.continuation = continuation
	c.cfg = cfg

	return nil
}

func (c *Client) GetMessages() ([]entity.Message, error) {
	chat, newContinuation, err := yt_chat.FetchContinuationChat(c.continuation, c.cfg)
	if err == yt_chat.ErrLiveStreamOver {
		return nil, err
	}

	// set the newly received continuation
	c.continuation = newContinuation

	comments := make([]entity.Message, 0, len(chat))
	for _, msg := range chat {
		id, err := uuid.NewV7()
		if err != nil {
			return nil, err
		}
		comments = append(comments, entity.Message{
			Text:      msg.Message,
			Source:    entity.SourceYoutube,
			User:      msg.AuthorName,
			CreatedAt: msg.Timestamp.UTC(),
			ID:        id.String(),
		})
	}

	return comments, nil
}
