package youtube_client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"regexp"
	"strings"
	"time"

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

const (
	initialDataRegex = `(?:window\s*\[\s*["\']ytInitialData["\']\s*\]|ytInitialData)\s*=\s*({.+?})\s*;\s*(?:var\s+meta|</script|\n)` //nolint
)

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

func (c *Client) GetMessages() ([]entity.Message, error) {
	chat, newContinuation, err := yt_chat.FetchContinuationChat(c.continuation, c.cfg)
	if err != nil {
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

func (c *Client) GetLastTranslationID(ctx context.Context, userName string) (string, error) {
	const defaultDedlineSeconds = 5
	defaultDedlineCtx, cancel := context.WithTimeout(ctx, defaultDedlineSeconds*time.Second)
	defer cancel()

	request, err := http.NewRequestWithContext(
		defaultDedlineCtx,
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
	initialDataArr := regexSearch(initialDataRegex, bodyBytes)
	initialDataRaw := bytes.Trim(initialDataArr[0], "ytInitialData = ")
	initialDataRaw = bytes.Trim(initialDataRaw, ";</script")

	initialData := &liveListInitialData{}
	err = json.Unmarshal(initialDataRaw, initialData)
	if err != nil {
		err = fmt.Errorf("failed to parse response of get last live id for youtube: %w", err)

		return "", err
	}

	// Получим первое же отрисовываемое видео и попробуем получить из него айдишник
	// так же проверим не завершенна ли она
	for _, tab := range initialData.Contents.TwoColumnBrowseResultsRenderer.Tabs {
		for _, liveData := range tab.TabRenderer.Content.RichGridRenderer.Contents {
			if strings.HasPrefix(
				liveData.RichItemRenderer.Content.VideoRenderer.PublishedTimeText.SimpleText,
				"Трансляция закончилась",
			) {
				return "", fmt.Errorf("last live id for youtube is finished: %w", app_errors.ErrNoData)
			}

			return liveData.RichItemRenderer.Content.VideoRenderer.VideoId, nil //nolint
		}
	}

	return "", fmt.Errorf("failed to get live id from json for youtube: %w", app_errors.ErrNoData)
}

func regexSearch(regex string, str []byte) [][]byte {
	r, _ := regexp.Compile(regex)
	matches := r.FindAll(str, -1)

	return matches
}
