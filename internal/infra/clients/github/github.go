package github

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/PlayingPossumHiss/possum_chat/internal/entity"
)

type Client struct{}

func New() *Client {
	return &Client{}
}

func (c *Client) LatestVersion(ctx context.Context) (*entity.SourceVersion, error) {
	client := http.Client{
		Timeout: time.Second,
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		"https://api.github.com/repos/PlayingPossumHiss/possum_chat/releases",
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("error on get request for app version in github: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error on do request for app version in github: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error on get app version in github: unexpected status %d", resp.StatusCode) //nolint:err113
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error on read request body for app version in github: %w", err)
	}

	var releases []release
	if err := json.Unmarshal(body, &releases); err != nil {
		return nil, fmt.Errorf("error on unmarshal releases for app version in github: %w", err)
	}
	if len(releases) == 0 {
		return nil, nil
	}

	var versionURL string
	for _, attachment := range releases[0].Assets {
		if attachment.Name == "possum_chat.tar.gz" {
			versionURL = attachment.DownloadURL

			break
		}
	}
	if versionURL == "" {
		return nil, nil
	}

	return &entity.SourceVersion{
		Version:     releases[0].Version,
		DownloadURL: versionURL,
	}, nil
}
