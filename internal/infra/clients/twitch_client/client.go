package twitch_client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	app_errors "github.com/PlayingPossumHiss/possum_chat/internal/errors"
	"github.com/PlayingPossumHiss/possum_chat/internal/service/logger"
)

const defaultDeadlineSeconds = 5

type Client struct {
}

func New() *Client {
	return &Client{}
}

func (c *Client) GetClientID(ctx context.Context, channelName string) (string, error) {
	return "", nil
}

// GetOnline возвращает текущее число зрителей трансляции канала.
// Значение берётся из того же GraphQL-запроса, которым сайт twitch.tv
// заполняет блок со зрителями на странице канала.
func (c *Client) GetOnline(ctx context.Context, clientID string, channelName string) (int64, error) {
	requestBody, err := json.Marshal(gqlRequest{
		Query:     getOnlineQuery,
		Variables: map[string]any{"login": channelName},
	})
	if err != nil {
		return 0, fmt.Errorf("failed to marshal twitch online request: %w", err)
	}

	bodyBytes, err := c.do(ctx, clientID, requestBody)
	if err != nil {
		return 0, fmt.Errorf("failed to get twitch online: %w", err)
	}

	response := &streamResponse{}
	err = json.Unmarshal(bodyBytes, response)
	if err != nil {
		return 0, fmt.Errorf("failed to parse twitch online response: %w", err)
	}

	if response.Data.User.Stream == nil {
		return 0, nil
	}

	return response.Data.User.Stream.ViewersCount, nil
}

func (c *Client) do(ctx context.Context, clientID string, body []byte) ([]byte, error) {
	deadlineCtx, cancel := context.WithTimeout(ctx, defaultDeadlineSeconds*time.Second)
	defer cancel()

	request, err := http.NewRequestWithContext(
		deadlineCtx,
		http.MethodPost,
		gqlURL,
		bytes.NewReader(body),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create twitch api request: %w", err)
	}

	request.Header.Add("Client-Id", clientID)
	request.Header.Add("Content-Type", "application/json")
	request.Header.Add("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:151.0) Gecko/20100101 Firefox/151.0")

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("failed to do twitch api request: %w", err)
	}
	defer func() {
		dErr := response.Body.Close()
		if dErr != nil {
			logger.Error(dErr)
		}
	}()

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(
			"%w: error on twitch request: code %d",
			app_errors.ErrRequestFail,
			response.StatusCode,
		)
	}

	bodyBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read twitch api response: %w", err)
	}

	return bodyBytes, nil
}
