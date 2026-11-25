package donation_alerts

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/PlayingPossumHiss/possum_chat/internal/entity"
	app_errors "github.com/PlayingPossumHiss/possum_chat/internal/errors"
	"github.com/PlayingPossumHiss/possum_chat/internal/service/logger"
	utils_time "github.com/PlayingPossumHiss/possum_chat/internal/utils/time"
	"github.com/gorilla/websocket"
)

const (
	apiBaseURL             = "https://www.donationalerts.com/api/v1"
	centrifugoEndpoint     = "wss://centrifugo.donationalerts.com/connection/websocket"
	centrifugoSubscribeURL = "https://www.donationalerts.com/api/v1/centrifuge/subscribe"
	defaultDedlineSeconds  = 5
	pingInterval           = 15 * time.Second
	userAgent              = "Mozilla/5.0 (X11; Linux x86_64; rv:151.0) Gecko/20100101 Firefox/151.0"

	connectCommandID   = int64(1)
	subscribeCommandID = int64(2)
	pingCommandIDStart = int64(3)
)

type Client struct {
	conn  *websocket.Conn
	clock utils_time.Clock

	writeMu sync.Mutex
	errChan chan error
}

func New(
	clock utils_time.Clock,
) *Client {
	return &Client{clock: clock}
}

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
	apiToken, err := c.getWsToken(ctx, initToken)
	if err != nil {
		return nil, err
	}

	userData, err := c.getUserData(ctx, apiToken)
	if err != nil {
		return nil, err
	}

	conn, resp, err := websocket.DefaultDialer.Dial(centrifugoEndpoint, nil)
	if err != nil {
		return nil, err
	}
	resp.Body.Close()
	c.conn = conn

	channel := fmt.Sprintf("$alerts:donation_%d", userData.Data.ID)
	err = c.connectAndSubscribe(ctx, apiToken, userData.Data.SocketConnectionToken, channel)
	if err != nil {
		conn.Close()

		return nil, fmt.Errorf("error on join DA channel: %w", err)
	}

	c.errChan = make(chan error, 1)
	result := make(chan entity.Message)

	go c.pingLoop()
	go c.readLoop(result)

	return result, nil
}

// connectAndSubscribe выполняет подключение к Centrifugo и подписку на канал донатов.
func (c *Client) connectAndSubscribe(
	ctx context.Context,
	apiToken string,
	socketToken string,
	channel string,
) error {
	err := c.sendCommand(centrifugoCommand{
		ID:     connectCommandID,
		Method: centrifugoMethodConnect,
		Params: centrifugoConnectParams{
			Token: socketToken,
			Name:  "js",
		},
	})
	if err != nil {
		return fmt.Errorf("error greet: %w", err)
	}

	reply, err := c.readReply()
	if err != nil {
		return fmt.Errorf("error read greet: %w", err)
	}
	if reply.Error != nil {
		return fmt.Errorf("greet error: %s", reply.Error.Message)
	}

	connectResult := &centrifugoConnectResult{}
	err = json.Unmarshal(reply.Result, connectResult)
	if err != nil {
		return fmt.Errorf("parse connect result: %w", err)
	}

	channelToken, err := c.getSubscribeToken(ctx, apiToken, connectResult.Client, channel)
	if err != nil {
		return err
	}

	err = c.sendCommand(centrifugoCommand{
		ID:     subscribeCommandID,
		Method: centrifugoMethodSubscribe,
		Params: centrifugoSubscribeParams{
			Channel: channel,
			Token:   channelToken,
		},
	})
	if err != nil {
		return fmt.Errorf("error joining channel: %w", err)
	}

	reply, err = c.readReply()
	if err != nil {
		return fmt.Errorf("error read joining channel: %w", err)
	}
	if reply.Error != nil {
		return fmt.Errorf("join channel error: %s", reply.Error.Message)
	}

	return nil
}

func (c *Client) sendCommand(cmd centrifugoCommand) error {
	msg, err := json.Marshal(cmd)
	if err != nil {
		return fmt.Errorf("marshal DA command: %w", err)
	}

	c.writeMu.Lock()
	defer c.writeMu.Unlock()

	err = c.conn.WriteMessage(websocket.TextMessage, msg)
	if err != nil {
		return fmt.Errorf("send DA command: %w", err)
	}

	return nil
}

func (c *Client) readReply() (*centrifugoReply, error) {
	_, raw, err := c.conn.ReadMessage()
	if err != nil {
		return nil, err
	}

	reply := &centrifugoReply{}
	err = json.Unmarshal(raw, reply)
	if err != nil {
		return nil, fmt.Errorf("parse DA reply: %w", err)
	}

	return reply, nil
}

func (c *Client) pingLoop() {
	id := pingCommandIDStart
	ticker := time.NewTicker(pingInterval)
	defer ticker.Stop()

	for range ticker.C {
		err := c.sendCommand(centrifugoCommand{
			ID:     id,
			Method: centrifugoMethodPing,
		})
		if err != nil {
			logger.Error(fmt.Errorf("send ping for DA error: %w", err))

			return
		}
		id++
	}
}

func (c *Client) readLoop(result chan entity.Message) {
	err := c.readMessages(result)
	close(result)
	c.errChan <- err
}

func (c *Client) readMessages(result chan entity.Message) error {
	for {
		_, raw, err := c.conn.ReadMessage()
		if err != nil {
			return fmt.Errorf("read msg from DA error: %w", err)
		}

		message, ok := c.parseDonation(raw)
		if ok {
			result <- message
		}
	}
}

func (c *Client) parseDonation(raw []byte) (entity.Message, bool) {
	logger.Error(string(raw))
	reply := &centrifugoReply{}
	err := json.Unmarshal(raw, reply)
	if err != nil {
		logger.Error(fmt.Errorf("parse DA message: %w", err))

		return entity.Message{}, false
	}

	// Push-сообщение приходит без id, ответы на команды — с id.
	if reply.ID != nil {
		return entity.Message{}, false
	}

	push := &centrifugoPush{}
	err = json.Unmarshal(reply.Result, push)
	if err != nil {
		logger.Error(fmt.Errorf("parse DA push: %w", err))

		return entity.Message{}, false
	}

	if push.Type != centrifugoPushPublication {
		return entity.Message{}, false
	}

	publication := &centrifugoPublication{}
	err = json.Unmarshal(push.Data, publication)
	if err != nil {
		logger.Error(fmt.Errorf("parse DA publication: %w", err))

		return entity.Message{}, false
	}

	donationMsg := &donation{}
	err = json.Unmarshal(publication.Data, donationMsg)
	if err != nil {
		logger.Error(fmt.Errorf("parse DA donation: %w", err))

		return entity.Message{}, false
	}

	if donationMsg.Message == "" {
		return entity.Message{}, false
	}

	amount := strconv.FormatFloat(float64(donationMsg.Amount), 'f', -1, 64)

	return entity.Message{
		ID:        "donation_alerts_" + string(donationMsg.ID),
		Source:    entity.SourceDonationAlerts,
		User:      donationMsg.Username,
		CreatedAt: c.clock.Now(),
		Content: []entity.MessageContentItem{
			{
				Type: entity.MessageContentItemTypeText,
				Value: fmt.Sprintf(
					"(%s %s) %s",
					amount,
					donationMsg.Currency,
					donationMsg.Message,
				),
			},
		},
	}, true
}

func (c *Client) getWsToken(ctx context.Context, initToken string) (string, error) {
	bodyBytes, err := c.do(
		ctx,
		http.MethodGet,
		fmt.Sprintf("%s/token/widget?token=%s", apiBaseURL, initToken),
		nil,
		nil,
	)
	if err != nil {
		return "", fmt.Errorf("error on get ws token: %w", err)
	}

	parsedMessage := &tokenResponse{}
	err = json.Unmarshal(bodyBytes, parsedMessage)
	if err != nil {
		return "", fmt.Errorf("parse ws token response: %w", err)
	}

	return parsedMessage.Data.Token, nil
}

func (c *Client) getUserData(ctx context.Context, token string) (*userDataResponse, error) {
	bodyBytes, err := c.do(
		ctx,
		http.MethodGet,
		apiBaseURL+"/user/widget",
		nil,
		headerWithAuth(token),
	)
	if err != nil {
		return nil, fmt.Errorf("error on get user data: %w", err)
	}

	parsedMessage := &userDataResponse{}
	err = json.Unmarshal(bodyBytes, parsedMessage)
	if err != nil {
		return nil, fmt.Errorf("parse user data response: %w", err)
	}

	return parsedMessage, nil
}

func (c *Client) getSubscribeToken(
	ctx context.Context,
	apiToken string,
	clientID string,
	channel string,
) (string, error) {
	body := &subscribeRequest{
		Client:   clientID,
		Channels: []string{channel},
	}
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return "", fmt.Errorf("marshal subscribe request: %w", err)
	}

	respBytes, err := c.do(
		ctx,
		http.MethodPost,
		centrifugoSubscribeURL,
		bytes.NewReader(bodyBytes),
		headerWithJSON(apiToken),
	)
	if err != nil {
		return "", fmt.Errorf("error on get subscribe token: %w", err)
	}

	parsedMessage := &subscribeResponse{}
	err = json.Unmarshal(respBytes, parsedMessage)
	if err != nil {
		return "", fmt.Errorf("parse subscribe response: %w", err)
	}

	for _, subscribedChannel := range parsedMessage.Channels {
		if subscribedChannel.Channel == channel {
			return subscribedChannel.Token, nil
		}
	}

	return "", fmt.Errorf("channel %q not found in subscribe response", channel)
}

func headerWithAuth(token string) map[string][]string {
	return map[string][]string{
		"Authorization": {
			fmt.Sprintf("Bearer %s", token),
		},
	}
}

func headerWithJSON(token string) map[string][]string {
	return map[string][]string{
		"Authorization": {
			fmt.Sprintf("Bearer %s", token),
		},
		"Content-Type": {
			"application/json",
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
	request.Header.Add("User-Agent", userAgent)
	for key, values := range headers {
		for _, value := range values {
			request.Header.Add(key, value)
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
