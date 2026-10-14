package youtube_client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"regexp"
	"strconv"
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
			Name:     "PREF",
			Secure:   true,
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
			Value:    "tz=Europe.Rome",
			MaxAge:   maxAge,
		},
		{
			Name:     "CONSENT",
			Secure:   true,
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
			Value:    fmt.Sprintf("YES+yt.432048971.it+FX+%d", 100+rand.Intn(999-100+1)), //nolint
			MaxAge:   maxAge,
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
			Content:   parseMessage(msg.Message),
			Source:    entity.SourceYoutube,
			User:      msg.AuthorName,
			CreatedAt: msg.Timestamp.UTC(),
			ID:        fmt.Sprintf("youtube_%s", newMsgId.String()),
		})
	}

	return comments, nil
}

func parseMessage(src string) []entity.MessageContentItem {
	logger.Debug(fmt.Sprintf("message from youtube: %s", src))

	emojiPrefix := "https://yt3.ggpht.com"
	srcElements := strings.Split(src, " ")
	result := []entity.MessageContentItem{}

	var buffer []string
	for _, srcElement := range srcElements {
		if strings.HasPrefix(srcElement, emojiPrefix) {
			if len(buffer) != 0 {
				result = append(result, entity.MessageContentItem{
					Type:  entity.MessageContentItemTypeText,
					Value: strings.Join(buffer, " "),
				})
				buffer = nil
			}
			result = append(result, entity.MessageContentItem{
				Type:  entity.MessageContentItemTypeImage,
				Value: srcElement,
			})
		} else {
			buffer = append(buffer, srcElement)
		}
	}
	if len(buffer) != 0 {
		result = append(result, entity.MessageContentItem{
			Type:  entity.MessageContentItemTypeText,
			Value: strings.Join(buffer, " "),
		})
	}

	return result
}

func (c *Client) GetOnline(ctx context.Context, liveID string) (int64, error) {
	const (
		retryLimit      = 3
		secondsForRetry = 15
	)

	var lastErr error

	for i := 0; i < retryLimit; i++ {
		online, err := c.getOnline(ctx, liveID)
		if err == nil {
			return online, nil
		}
		lastErr = err

		if i == retryLimit-1 {
			break
		}

		timer := time.NewTimer(time.Second * secondsForRetry)
		defer timer.Stop()

		select {
		case <-ctx.Done():
			return 0, ctx.Err()
		case <-timer.C:
		}
	}

	return 0, fmt.Errorf("error on get youtube online event with retry: %w", lastErr)
}

func (c *Client) getOnline(ctx context.Context, liveID string) (int64, error) {
	initialDataRaw, err := c.getInitDataFrom(
		ctx,
		fmt.Sprintf("https://www.youtube.com/watch?v=%s", liveID),
	)
	if err != nil {
		return 0, fmt.Errorf("get online translation error %w", err)
	}

	initialData := &liveListInitialData{}
	err = json.Unmarshal(initialDataRaw, initialData)
	if err != nil {
		err = fmt.Errorf("failed to parse response of get online for youtube: %w", err)

		return 0, err
	}

	twoColumnWatchNextResultsResults := initialData.Contents.TwoColumnWatchNextResults.Results.Results.Contents
	if len(twoColumnWatchNextResultsResults) == 0 {
		return 0, fmt.Errorf("no view count renderer on youtube watch page: %w", app_errors.ErrNoData)
	}

	online, err := strconv.ParseInt(
		twoColumnWatchNextResultsResults[0].VideoPrimaryInfoRenderer.ViewCount.VideoViewCountRenderer.OriginalViewCount,
		10, 64,
	)
	if err != nil {
		return 0, fmt.Errorf("error on parse youtube online %w", err)
	}

	return online, nil
}

func (c *Client) GetLastTranslationID(ctx context.Context, userName string) (string, error) {
	initialDataRaw, err := c.getInitDataFrom(
		ctx,
		fmt.Sprintf("https://www.youtube.com/@%s/streams", userName),
	)
	if err != nil {
		if errors.Is(err, app_errors.ErrNoData) {
			return userName, nil
		}

		return "", fmt.Errorf("get last translation error %w", err)
	}

	initialData := &liveListInitialData{}
	err = json.Unmarshal(initialDataRaw, initialData)
	if err != nil {
		err = fmt.Errorf("failed to parse response of get last live id for youtube: %w", err)

		return "", err
	}

	streamID := getStreamIDFormParsedData(initialData)
	if streamID != "" {
		return streamID, nil
	}

	return "", fmt.Errorf("failed to get live id from json for youtube: %w", app_errors.ErrNoData)
}

const liveBadgeStyle = "THUMBNAIL_OVERLAY_BADGE_STYLE_LIVE"

func getStreamIDFormParsedData(initialData *liveListInitialData) string {
	for _, tab := range initialData.Contents.TwoColumnBrowseResultsRenderer.Tabs {
		for _, liveData := range tab.TabRenderer.Content.RichGridRenderer.Contents {
			if liveData.isLiveStream() {
				return liveData.RichItemRenderer.Content.LockupViewModel.ContentId
			}
		}
	}

	return ""
}

func (c liveListInitialDataTabContent) isLiveStream() bool {
	overlays := c.RichItemRenderer.Content.LockupViewModel.ContentImage.ThumbnailViewModel.Overlays

	for _, overlay := range overlays {
		for _, badge := range overlay.ThumbnailBottomOverlayViewModel.Badges {
			if badge.ThumbnailBadgeViewModel.BadgeStyle == liveBadgeStyle {
				return true
			}
		}
	}

	return false
}

func regexSearch(regex string, str []byte) [][]byte {
	r, _ := regexp.Compile(regex)
	matches := r.FindAll(str, -1)

	return matches
}

func (c *Client) getInitDataFrom(ctx context.Context, url string) ([]byte, error) {
	const defaultDedlineSeconds = 5
	defaultDedlineCtx, cancel := context.WithTimeout(ctx, defaultDedlineSeconds*time.Second)
	defer cancel()

	request, err := http.NewRequestWithContext(
		defaultDedlineCtx,
		http.MethodGet,
		url,
		nil,
	)
	if err != nil {
		err = fmt.Errorf("failed to create request for youtube: %w", err)

		return nil, err
	}

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		err = fmt.Errorf("failed to do request for youtube: %w", err)

		return nil, err
	}
	defer func() {
		dErr := response.Body.Close()
		if dErr != nil {
			logger.Error(dErr)
		}
	}()

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to do request for youtube: %w", app_errors.ErrNoData)
	}

	bodyBytes, err := io.ReadAll(response.Body)
	if err != nil {
		err = fmt.Errorf("failed to read response for youtube: %w", err)

		return nil, err
	}

	// Вырезать нужный кусок резуляркой - костыль
	// но я хочу поскорее это докатить и вообще
	// работает - не трож
	initialDataArr := regexSearch(initialDataRegex, bodyBytes)
	if len(initialDataArr) == 0 {
		return nil, fmt.Errorf("failed to find ytInitialData on youtube page: %w", app_errors.ErrNoData)
	}
	initialDataRaw := bytes.Trim(initialDataArr[0], "ytInitialData = ")

	return bytes.Trim(initialDataRaw, ";</script"), nil
}
