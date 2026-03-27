package vk_play_live_api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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
		return "", err
	}

	type respContract struct {
		Token string `json:"token"`
	}
	tokenResponse := &respContract{}
	err = json.Unmarshal(bodyBytes, tokenResponse)
	if err != nil {
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
		return nil, err
	}

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	bodyBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	return bodyBytes, nil
}
