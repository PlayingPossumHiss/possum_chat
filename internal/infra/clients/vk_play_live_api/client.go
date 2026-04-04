package vk_play_live_api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/PlayingPossumHiss/possum_chat/internal/service/logger"
)

const host = "https://api.live.vkvideo.ru"

type Client struct{}

func New() *Client {
	return &Client{}
}

func (c *Client) GetWsToken(ctx context.Context) (string, error) {
	bodyBytes, err := c.do(
		ctx,
		http.MethodGet,
		fmt.Sprintf("%s/v1/ws/connect", host),
		nil,
	)
	if err != nil {
		err = fmt.Errorf("failed to get ws token for vk play live: %w", err)

		return "", err
	}

	type respContract struct {
		Token string `json:"token"`
	}
	tokenResponse := &respContract{}
	err = json.Unmarshal(bodyBytes, tokenResponse)
	if err != nil {
		err = fmt.Errorf("failed to parse ws token for vk play live: %w", err)

		return "", err
	}

	return tokenResponse.Token, nil
}

func (c *Client) GetUserID(ctx context.Context, userName string) (int, error) {
	bodyBytes, err := c.do(
		ctx,
		http.MethodGet,
		fmt.Sprintf("%s/v1/channel/%s", host, userName),
		nil,
	)
	if err != nil {
		err = fmt.Errorf("failed to get user id for vk play live: %w", err)

		return 0, err
	}

	type respContract struct {
		Data struct {
			Channel struct {
				Owner struct {
					ID int `json:"id"`
				} `json:"owner"`
			} `json:"channel"`
		} `json:"data"`
	}

	userResponse := &respContract{}
	err = json.Unmarshal(bodyBytes, userResponse)
	if err != nil {
		err = fmt.Errorf("failed to parse user id for vk play live: %w", err)

		return 0, err
	}

	return userResponse.Data.Channel.Owner.ID, nil
}

func (c *Client) do(ctx context.Context, method string, url string, body io.Reader) ([]byte, error) {
	request, err := http.NewRequestWithContext(
		ctx,
		method,
		url,
		body,
	)
	if err != nil {
		err = fmt.Errorf("failed to create api request for vk play live: %w", err)

		return nil, err
	}

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		err = fmt.Errorf("failed to do api request for vk play live: %w", err)

		return nil, err
	}
	defer func() {
		dErr := response.Body.Close()
		if dErr != nil {
			logger.Error(dErr)
		}
	}()
	bodyBytes, err := io.ReadAll(response.Body)
	if err != nil {
		err = fmt.Errorf("failed to read api request from vk play live: %w", err)

		return nil, err
	}

	return bodyBytes, nil
}
