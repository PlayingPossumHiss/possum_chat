package youtube_client

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"regexp"

	"github.com/PlayingPossumHiss/possum_chat/internal/entity"
	app_errors "github.com/PlayingPossumHiss/possum_chat/internal/errors"
	"github.com/PlayingPossumHiss/possum_chat/internal/service/logger"
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
	const maxAge = 300
	customCookies := []*http.Cookie{
		{
			Name:   "PREF",
			Value:  "tz=Europe.Rome",
			MaxAge: maxAge,
		},
		{
			Name:   "CONSENT",
			Value:  fmt.Sprintf("YES+yt.432048971.it+FX+%d", 100+rand.Intn(999-100+1)), //nolint
			MaxAge: maxAge,
		},
	}
	yt_chat.AddCookies(customCookies)

	continuation, cfg, err := yt_chat.ParseInitialData(fmt.Sprintf("https://www.youtube.com/watch?v=%s", streamKey))
	if err != nil {
		err = fmt.Errorf("failed init chat for youtube: %w", err)

		return err
	}

	c.continuation = continuation
	c.cfg = cfg

	return nil
}

func (c *Client) GetLastTranslationID(ctx context.Context, userName string) (string, error) {
	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		fmt.Sprintf("https://www.youtube.com/@%s/streams", userName),
		nil,
	)
	if err != nil {
		err = fmt.Errorf("failed to create request for get last live id for youtube: %w", err)

		return "", err
	}

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		err = fmt.Errorf("failed to do request for get last live id for youtube: %w", err)

		return "", err
	}
	defer func() {
		dErr := response.Body.Close()
		if dErr != nil {
			logger.Error(dErr)
		}
	}()
	bodyBytes, err := io.ReadAll(response.Body)
	if err != nil {
		err = fmt.Errorf("failed to read response of get last live id for youtube: %w", err)

		return "", err
	}
	// Вырезать нужный кусок резуляркой - костыль
	// но я хочу поскорее это докатить и вообще
	// работает - не трож
	matcher := regexp.MustCompile(`"videoId":"[^"]+"`)
	possibleKey := matcher.Find(bodyBytes)
	if possibleKey == nil {
		err = fmt.Errorf("can't find last live id in get last live id request for youtube: %w", app_errors.ErrNoData)

		return "", err
	}

	return string(possibleKey[11 : len(possibleKey)-1]), nil
}

func (c *Client) GetMessages() ([]entity.Message, error) {
	chat, newContinuation, err := yt_chat.FetchContinuationChat(c.continuation, c.cfg)
	if errors.Is(err, yt_chat.ErrLiveStreamOver) {
		err = fmt.Errorf("failed to read new messages for youtube: %w", err)

		return nil, err
	}

	// set the newly received continuation
	c.continuation = newContinuation

	comments := make([]entity.Message, 0, len(chat))
	for _, msg := range chat {
		newMsgId, err := uuid.NewV7()
		if err != nil {
			return nil, err
		}
		comments = append(comments, entity.Message{
			Text:      msg.Message,
			Source:    entity.SourceYoutube,
			User:      msg.AuthorName,
			CreatedAt: msg.Timestamp.UTC(),
			ID:        fmt.Sprintf("youtube_%s", newMsgId.String()),
		})
	}

	return comments, nil
}
