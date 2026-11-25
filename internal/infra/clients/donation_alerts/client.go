package donation_alerts

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/PlayingPossumHiss/possum_chat/internal/entity"
	app_errors "github.com/PlayingPossumHiss/possum_chat/internal/errors"
	"github.com/PlayingPossumHiss/possum_chat/internal/service/logger"
	utils_time "github.com/PlayingPossumHiss/possum_chat/internal/utils/time"
	"github.com/gorilla/websocket"
)

type Client struct {
	userID  int64
	token   string
	conn    *websocket.Conn
	clock   utils_time.Clock
	errChan chan error
}

func New(
	clock utils_time.Clock,
) *Client {
	return &Client{clock: clock}
}

const (
	daHost                = "socket.donationalerts.ru"
	daPort                = 443
	defaultDedlineSeconds = 5
)

func (c *Client) Close() {
	if c.conn == nil {
		return
	}
	c.conn.Close()
}

func (c *Client) Done() error {
	return <-c.errChan
}

func (c *Client) Init(
	ctx context.Context,
	initToken string,
) (chan entity.Message, error) {
	token, err := c.getWsToken(ctx, initToken)
	if err != nil {
		return nil, err
	}
	c.token = token
	userID, err := c.getUserID(ctx, token)
	if err != nil {
		return nil, err
	}
	c.userID = userID

	// создаем коненкт
	conn, resp, err := websocket.DefaultDialer.Dial("wss://centrifugo.donationalerts.com/connection/websocket", nil)
	if err != nil {
		return nil, err
	}
	resp.Body.Close()

	c.conn = conn
	c.errChan = make(chan error)
	result := make(chan entity.Message)

	err = c.joinChannelByID()
	if err != nil {
		return nil, fmt.Errorf("error on join DA channel: %w", err)
	}

	go func() {
		var id int64 = 3
		for {
			ping := wsMessagePing{
				ID:     id,
				Method: 7,
			}

			msg, err := json.Marshal(ping)
			if err != nil {
				logger.Error(fmt.Errorf("marshal ping for DA error: %w", err))
				return
			}

			err = c.conn.WriteMessage(websocket.TextMessage, msg)
			if err != nil {
				logger.Error(fmt.Errorf("send ping for DA error: %w", err))
				return
			}
			id++

			time.Sleep(time.Second * 15)
		}
	}()

	go func() {
		for {
			_, msg, err := c.conn.ReadMessage()
			if err != nil {
				logger.Error(fmt.Errorf("read msg from DA error: %w", err))
				return
			}
			logger.Warn(string(msg))
		}
	}()

	return result, nil
}

func (c *Client) joinChannelByID() error {
	greet := wsMessageGreet{
		ID: 1,
		Params: wsMessageGreetParams{
			Token: c.token,
			Name:  "js",
		},
	}

	msg, err := json.Marshal(greet)
	if err != nil {
		return fmt.Errorf("marshal error: %w", err)
	}

	err = c.conn.WriteMessage(websocket.TextMessage, msg)
	if err != nil {
		return fmt.Errorf("error greet: %w", err)
	}

	_, _, err = c.conn.ReadMessage()
	if err != nil {
		return fmt.Errorf("error read greet: %w", err)
	}

	subscribeMessage := wsMessageSubscribe{
		ID: 2,
		Params: wsMessageSubscribeParams{
			Token:   c.token,
			Channel: fmt.Sprintf("$alerts:donation_%d", c.userID),
		},
	}

	msg, err = json.Marshal(subscribeMessage)
	if err != nil {
		return fmt.Errorf("marshal error: %w", err)
	}

	err = c.conn.WriteMessage(websocket.TextMessage, msg)
	if err != nil {
		return fmt.Errorf("error joining channel: %w", err)
	}

	_, msg, err = c.conn.ReadMessage()
	if err != nil {
		return fmt.Errorf("error read joining channel: %w", err)
	}
	logger.Debug(string(msg))

	_, msg, err = c.conn.ReadMessage()
	if err != nil {
		return err
	}
	logger.Debug(string(msg))

	return nil
}

func (c *Client) getWsToken(ctx context.Context, initToken string) (string, error) {
	bodyBytes, err := c.do(
		ctx,
		http.MethodGet,
		fmt.Sprintf("https://www.donationalerts.com/api/v1/token/widget?token=%s", initToken),
		nil,
		nil,
	)
	if err != nil {
		return "", fmt.Errorf("error on get ws token: %w", err)
	}

	parsedMessage := &tokenResponse{}
	json.Unmarshal(bodyBytes, parsedMessage)

	return parsedMessage.Data.Token, nil
}

func (c *Client) getUserID(ctx context.Context, token string) (int64, error) {
	bodyBytes, err := c.do(
		ctx,
		http.MethodGet,
		"https://www.donationalerts.com/api/v1/user/widget",
		nil,
		headerWithAuth(token),
	)
	if err != nil {
		return 0, fmt.Errorf("error on get user ID: %w", err)
	}

	parsedMessage := &userInfoResponse{}
	json.Unmarshal(bodyBytes, parsedMessage)

	return parsedMessage.Data.ID, nil
}

func headerWithAuth(token string) map[string][]string {
	return map[string][]string{
		"Authorization": {
			fmt.Sprintf("Bearer %s", token),
		},
	}
}

func (c *Client) do(
	ctx context.Context,
	method string,
	url string,
	body io.Reader,
	headers map[string][]string,
) ([]byte, error) {
	defaultDedlineCtx, cancel := context.WithTimeout(ctx, defaultDedlineSeconds*time.Second)
	defer cancel()

	request, err := http.NewRequestWithContext(
		defaultDedlineCtx,
		method,
		url,
		body,
	)
	if err != nil {
		err = fmt.Errorf("failed to create api request for alerts: %w", err)

		return nil, err
	}
	request.Header.Add("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:151.0) Gecko/20100101 Firefox/151.0")
	for key, values := range headers {
		for _, velue := range values {
			request.Header.Add(key, velue)
		}
	}

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		err = fmt.Errorf("failed to do api request for donation alerts: %w", err)

		return nil, err
	}
	defer func() {
		dErr := response.Body.Close()
		if dErr != nil {
			logger.Error(dErr)
		}
	}()
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(
			"%w: error on alerts request: code %d",
			app_errors.ErrRequestFail,
			response.StatusCode,
		)
	}

	bodyBytes, err := io.ReadAll(response.Body)
	if err != nil {
		err = fmt.Errorf("failed to read api request from alerts: %w", err)

		return nil, err
	}

	return bodyBytes, nil
}
