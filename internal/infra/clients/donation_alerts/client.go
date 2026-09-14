package donation_alerts

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/PlayingPossumHiss/possum_chat/internal/entity"
	app_errors "github.com/PlayingPossumHiss/possum_chat/internal/errors"
	"github.com/PlayingPossumHiss/possum_chat/internal/service/logger"
	utils_time "github.com/PlayingPossumHiss/possum_chat/internal/utils/time"
	"github.com/gorilla/websocket"
)

const (
	donationAlertsBaseURL = "https://www.donationalerts.com"

	centrifugoConnectName = "js"

	centrifugoMethodSubscribe = 1
	centrifugoMethodPing      = 7

	centrifugoPushPublication = 0

	donationAlertsChannelPrefix = "$alerts:donation_"

	httpRequestTimeoutSeconds = 5
	replyWaitTimeoutSeconds   = 5
	pingIntervalSeconds       = 25
	messagesBufferSize        = 16
)

type Client struct {
	clock utils_time.Clock

	wsMx    *sync.Mutex
	ws      *websocket.Conn
	writeMx *sync.Mutex

	messages chan entity.Message
	done     chan error

	stopCh   chan struct{}
	stopOnce *sync.Once

	idMx   *sync.Mutex
	nextID uint64

	pendingMx *sync.Mutex
	pending   map[uint64]chan centrifugoReply
}

func New(
	clock utils_time.Clock,
) *Client {
	return &Client{
		clock:     clock,
		wsMx:      &sync.Mutex{},
		writeMx:   &sync.Mutex{},
		stopCh:    make(chan struct{}),
		stopOnce:  &sync.Once{},
		idMx:      &sync.Mutex{},
		pendingMx: &sync.Mutex{},
		pending:   make(map[uint64]chan centrifugoReply),
	}
}

type bootstrapData struct {
	endpoint          string
	subscribeEndpoint string
	apiToken          string
	userID            int64
	socketToken       string
}

func (c *Client) Close() {
	c.stopOnce.Do(func() {
		close(c.stopCh)
	})
	c.wsMx.Lock()
	defer c.wsMx.Unlock()
	if c.ws != nil {
		_ = c.ws.Close()
	}
}

func (c *Client) Done() error {
	select {
	case err := <-c.done:
		return err
	case <-c.stopCh:
		return fmt.Errorf("%w: donation alerts client is closed", app_errors.ErrScraperStoped)
	}
}

func (c *Client) Init(
	ctx context.Context,
	token string,
) (chan entity.Message, error) {
	bootstrap, err := c.bootstrap(ctx, token)
	if err != nil {
		return nil, err
	}

	conn, _, err := websocket.DefaultDialer.DialContext(
		ctx,
		bootstrap.endpoint,
		http.Header{"Origin": {donationAlertsBaseURL}},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to dial donation alerts websocket: %w", err)
	}

	messages := make(chan entity.Message, messagesBufferSize)
	done := make(chan error, 1)
	connStop := make(chan struct{})

	c.messages = messages
	c.done = done

	c.wsMx.Lock()
	c.ws = conn
	c.wsMx.Unlock()

	go c.readLoop(conn, connStop, messages, done)

	connectResult, err := c.connect(conn, bootstrap.socketToken)
	if err != nil {
		_ = conn.Close()

		return nil, err
	}

	channel := donationAlertsChannelPrefix + strconv.FormatInt(bootstrap.userID, 10)
	subscriptionToken, err := c.fetchSubscriptionToken(ctx, bootstrap, connectResult.Client, channel)
	if err != nil {
		_ = conn.Close()

		return nil, err
	}

	if err := c.subscribe(conn, channel, subscriptionToken); err != nil {
		_ = conn.Close()

		return nil, err
	}

	go c.pingLoop(conn, connStop)

	return messages, nil
}

func (c *Client) bootstrap(
	ctx context.Context,
	widgetToken string,
) (*bootstrapData, error) {
	front, err := c.fetchEnvFront(ctx)
	if err != nil {
		return nil, err
	}

	apiToken, err := c.fetchWidgetToken(ctx, widgetToken)
	if err != nil {
		return nil, err
	}

	user, err := c.fetchUserWidget(ctx, apiToken)
	if err != nil {
		return nil, err
	}

	return &bootstrapData{
		endpoint:          front.Data.Centrifugo.Endpoint,
		subscribeEndpoint: front.Data.Centrifugo.SubscribeEndpoint,
		apiToken:          apiToken,
		userID:            user.Data.ID,
		socketToken:       user.Data.SocketConnectionToken,
	}, nil
}

func (c *Client) fetchEnvFront(ctx context.Context) (*envFrontResponse, error) {
	out := &envFrontResponse{}
	if err := c.getJSON(ctx, donationAlertsBaseURL+"/api/v1/env/front", "", out); err != nil {
		return nil, fmt.Errorf("failed to get donation alerts env: %w", err)
	}

	return out, nil
}

func (c *Client) fetchWidgetToken(
	ctx context.Context,
	widgetToken string,
) (string, error) {
	out := &widgetTokenResponse{}
	url := donationAlertsBaseURL + "/api/v1/token/widget?token=" + widgetToken
	if err := c.getJSON(ctx, url, "", out); err != nil {
		return "", fmt.Errorf("failed to get donation alerts api token: %w", err)
	}

	return out.Data.Token, nil
}

func (c *Client) fetchUserWidget(
	ctx context.Context,
	apiToken string,
) (*userWidgetResponse, error) {
	out := &userWidgetResponse{}
	if err := c.getJSON(ctx, donationAlertsBaseURL+"/api/v1/user/widget", apiToken, out); err != nil {
		return nil, fmt.Errorf("failed to get donation alerts user: %w", err)
	}

	return out, nil
}

func (c *Client) getJSON(
	ctx context.Context,
	url string,
	bearer string,
	out any,
) error {
	requestCtx, cancel := context.WithTimeout(ctx, time.Duration(httpRequestTimeoutSeconds)*time.Second)
	defer cancel()

	request, err := http.NewRequestWithContext(requestCtx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	request.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:151.0) Gecko/20100101 Firefox/151.0")
	if bearer != "" {
		request.Header.Set("Authorization", "Bearer "+bearer)
	}

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return err
	}
	defer func() {
		_ = response.Body.Close()
	}()

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("%w: code %d", app_errors.ErrRequestFail, response.StatusCode)
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return err
	}

	return json.Unmarshal(body, out)
}

func (c *Client) fetchSubscriptionToken(
	ctx context.Context,
	bootstrap *bootstrapData,
	clientID string,
	channel string,
) (string, error) {
	body, err := json.Marshal(subscribeRequest{
		Client:   clientID,
		Channels: []string{channel},
	})
	if err != nil {
		return "", err
	}

	requestCtx, cancel := context.WithTimeout(ctx, time.Duration(httpRequestTimeoutSeconds)*time.Second)
	defer cancel()

	request, err := http.NewRequestWithContext(
		requestCtx,
		http.MethodPost,
		bootstrap.subscribeEndpoint,
		strings.NewReader(string(body)),
	)
	if err != nil {
		return "", err
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", "Bearer "+bootstrap.apiToken)

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return "", fmt.Errorf("failed to fetch donation alerts subscription token: %w", err)
	}
	defer func() {
		_ = response.Body.Close()
	}()

	if response.StatusCode != http.StatusOK {
		return "", fmt.Errorf(
			"%w: subscribe endpoint code %d",
			app_errors.ErrRequestFail,
			response.StatusCode,
		)
	}

	respBody, err := io.ReadAll(response.Body)
	if err != nil {
		return "", err
	}

	var subResp subscribeResponse
	if err := json.Unmarshal(respBody, &subResp); err != nil {
		return "", err
	}

	if len(subResp.Channels) == 0 {
		return "", fmt.Errorf("donation alerts subscription token is empty: %w", app_errors.ErrNoData)
	}

	return subResp.Channels[0].Token, nil
}

func (c *Client) connect(
	conn *websocket.Conn,
	socketToken string,
) (centrifugoConnectResult, error) {
	reply, err := c.request(conn, nil, centrifugoConnectParams{
		Token: socketToken,
		Name:  centrifugoConnectName,
	})
	if err != nil {
		return centrifugoConnectResult{}, err
	}

	if reply.Error != nil {
		return centrifugoConnectResult{}, fmt.Errorf(
			"donation alerts connect error %d: %s",
			reply.Error.Code,
			reply.Error.Message,
		)
	}

	var result centrifugoConnectResult
	if err := json.Unmarshal(reply.Result, &result); err != nil {
		return centrifugoConnectResult{}, fmt.Errorf("failed to parse donation alerts connect result: %w", err)
	}

	return result, nil
}

func (c *Client) subscribe(
	conn *websocket.Conn,
	channel string,
	token string,
) error {
	method := centrifugoMethodSubscribe
	reply, err := c.request(conn, &method, centrifugoSubscribeParams{
		Channel: channel,
		Token:   token,
	})
	if err != nil {
		return err
	}

	if reply.Error != nil {
		return fmt.Errorf(
			"donation alerts subscribe error %d: %s",
			reply.Error.Code,
			reply.Error.Message,
		)
	}

	return nil
}

func (c *Client) request(
	conn *websocket.Conn,
	method *int,
	params any,
) (centrifugoReply, error) {
	c.idMx.Lock()
	c.nextID++
	commandID := c.nextID
	c.idMx.Unlock()

	replyCh := make(chan centrifugoReply, 1)
	c.pendingMx.Lock()
	c.pending[commandID] = replyCh
	c.pendingMx.Unlock()

	defer c.removePending(commandID)

	cmd := centrifugoCommand{ID: commandID, Method: method, Params: params}
	data, err := json.Marshal(cmd)
	if err != nil {
		return centrifugoReply{}, fmt.Errorf("failed to marshal donation alerts command: %w", err)
	}

	if err := c.write(conn, data); err != nil {
		return centrifugoReply{}, fmt.Errorf("failed to write donation alerts command: %w", err)
	}

	select {
	case reply := <-replyCh:
		return reply, nil
	case <-c.stopCh:
		return centrifugoReply{}, fmt.Errorf("%w: donation alerts client is closed", app_errors.ErrScraperStoped)
	case <-time.After(time.Duration(replyWaitTimeoutSeconds) * time.Second):
		return centrifugoReply{}, fmt.Errorf("donation alerts request timeout: %w", app_errors.ErrRequestFail)
	}
}

func (c *Client) write(
	conn *websocket.Conn,
	data []byte,
) error {
	c.writeMx.Lock()
	defer c.writeMx.Unlock()

	return conn.WriteMessage(websocket.TextMessage, data)
}

func (c *Client) removePending(commandID uint64) {
	c.pendingMx.Lock()
	delete(c.pending, commandID)
	c.pendingMx.Unlock()
}

func (c *Client) failAllPending(err error) {
	c.pendingMx.Lock()
	defer c.pendingMx.Unlock()

	for commandID, replyChan := range c.pending {
		replyChan <- centrifugoReply{Error: &centrifugoError{Message: err.Error()}}
		delete(c.pending, commandID)
	}
}

func (c *Client) readLoop(
	conn *websocket.Conn,
	connStop chan struct{},
	messages chan entity.Message,
	done chan error,
) {
	var readErr error
	defer func() {
		c.failAllPending(readErr)
		close(messages)
		done <- readErr
		close(connStop)
	}()

	for {
		_, data, err := conn.ReadMessage()
		if err != nil {
			readErr = fmt.Errorf("donation alerts connection is closed: %w", err)

			return
		}

		for _, line := range strings.Split(string(data), "\n") {
			if len(line) == 0 {
				continue
			}
			c.handleLine(line, messages)
		}
	}
}

func (c *Client) handleLine(
	line string,
	messages chan entity.Message,
) {
	var reply centrifugoReply
	if err := json.Unmarshal([]byte(line), &reply); err != nil {
		logger.Error(fmt.Sprintf("failed to parse donation alerts reply: %s", err.Error()))

		return
	}

	if reply.ID > 0 {
		c.deliverReply(reply)

		return
	}

	c.handlePush(reply.Result, messages)
}

func (c *Client) deliverReply(reply centrifugoReply) {
	c.pendingMx.Lock()
	replyChan, found := c.pending[reply.ID]
	if found {
		delete(c.pending, reply.ID)
	}
	c.pendingMx.Unlock()

	if found {
		replyChan <- reply
	}
}

func (c *Client) handlePush(
	result json.RawMessage,
	messages chan entity.Message,
) {
	if len(result) == 0 {
		return
	}

	var push centrifugoPush
	if err := json.Unmarshal(result, &push); err != nil {
		logger.Error(fmt.Sprintf("failed to parse donation alerts push: %s", err.Error()))

		return
	}

	if push.Type != centrifugoPushPublication {
		return
	}

	var publication centrifugoPublication
	if err := json.Unmarshal(push.Data, &publication); err != nil {
		logger.Error(fmt.Sprintf("failed to parse donation alerts publication: %s", err.Error()))

		return
	}

	var donation alert
	if err := json.Unmarshal(publication.Data, &donation); err != nil {
		logger.Error(fmt.Sprintf("failed to parse donation alerts alert: %s", err.Error()))

		return
	}

	c.onAlert(donation, messages)
}

func (c *Client) onAlert(
	donation alert,
	messages chan entity.Message,
) {
	logger.Debug(fmt.Sprintf("message from donation alerts: %s", donation.Message))

	if donation.Message == "" {
		return
	}

	message := entity.Message{
		ID:        "donation_alerts_" + strconv.FormatInt(donation.ID, 10),
		Source:    entity.SourceDonationAlerts,
		User:      donation.Username,
		CreatedAt: c.clock.Now(),
		Content: []entity.MessageContentItem{
			{
				Type:  entity.MessageContentItemTypeText,
				Value: fmt.Sprintf("(%s %s) %s", donation.Amount.String(), donation.Currency, donation.Message),
			},
		},
	}

	select {
	case messages <- message:
	case <-c.stopCh:
	}
}

func (c *Client) pingLoop(
	conn *websocket.Conn,
	connStop chan struct{},
) {
	ticker := time.NewTicker(time.Duration(pingIntervalSeconds) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-connStop:
			return
		case <-c.stopCh:
			return
		case <-ticker.C:
			method := centrifugoMethodPing
			if _, err := c.request(conn, &method, nil); err != nil {
				logger.Error(fmt.Errorf("donation alerts ping failed: %w", err))
				_ = conn.Close()

				return
			}
		}
	}
}
