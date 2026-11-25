package ws

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
	"uuid"

	"github.com/PlayingPossumHiss/possum_chat/internal/entity"
	"github.com/PlayingPossumHiss/possum_chat/internal/service/logger"
	"github.com/coder/websocket"
)

type Service struct {
	port         int
	waiters      map[string]chan entity.Message
	onlineGetter OnlineGetter
	mx           *sync.Mutex
}

type OnlineGetter interface {
	GetOnline() map[entity.Source]int64
}

func New(
	onlineGetter OnlineGetter,
	port int,
) *Service {
	return &Service{
		mx:           &sync.Mutex{},
		port:         port,
		onlineGetter: onlineGetter,
		waiters:      map[string]chan entity.Message{},
	}
}

func (s *Service) Handle(ctx context.Context, message entity.Message) error {
	s.mx.Lock()
	defer s.mx.Unlock()

	for key := range s.waiters {
		select {
		case s.waiters[key] <- message:
		default:
			// Буфер клиента переполнен — пропускаем сообщение для него,
			// чтобы не блокировать раздачу остальным и не ловить дедлок.
		}
	}

	return nil
}

func (s *Service) Listen() {
	http.HandleFunc("/chat_events", s.handleWS)
	logger.Info("ws chat_events started")
	go func() {
		err := http.ListenAndServe(fmt.Sprintf(":%d", s.port), nil)
		if err != nil {
			logger.Error(fmt.Errorf("error on listed ws: %w", err))
		}
	}()
}

func (s *Service) handleWS(response http.ResponseWriter, request *http.Request) {
	conn, err := websocket.Accept(response, request, &websocket.AcceptOptions{
		InsecureSkipVerify: true, // страха нет
	})
	if err != nil {
		log.Printf("Accept failed: %v", err)
		return
	}
	waiterKey := uuid.NewV7().String()
	s.mx.Lock()
	s.waiters[waiterKey] = make(chan entity.Message, 1000)
	s.mx.Unlock()
	ctx, cancel := context.WithCancel(request.Context())
	defer cancel()
	defer conn.CloseNow()
	defer func() {
		s.mx.Lock()
		defer s.mx.Unlock()

		close(s.waiters[waiterKey])
		delete(s.waiters, waiterKey)
	}()

	// Передаем сообщения
	go func() {
		for msg := range s.waiters[waiterKey] {
			messageData, err := json.Marshal(msgToApiEvent(msg))
			if err != nil {
				logger.Error(fmt.Errorf("error on marshal ws message: %w", err))
				continue
			}
			err = conn.Write(ctx, websocket.MessageText, messageData)
			if err != nil {
				logger.Error(fmt.Errorf("error on send ws message: %w", err))
			}
		}
	}()
	// передаем онлайн и статусы
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				time.Sleep(time.Second)
				online := s.onlineGetter.GetOnline()

				onlineData, err := json.Marshal(onlineToApiEvent(online))
				if err != nil {
					logger.Error(fmt.Errorf("error on marshal ws online: %w", err))
					continue
				}
				err = conn.Write(ctx, websocket.MessageText, onlineData)
				if err != nil {
					logger.Error(fmt.Errorf("error on send ws online: %w", err))
				}

				status := logger.GetStatus()
				loggingStatusData, err := json.Marshal(loggingStatusToApiEvent(status))
				if err != nil {
					logger.Error(fmt.Errorf("error on marshal ws logging status: %w", err))
					continue
				}
				err = conn.Write(ctx, websocket.MessageText, loggingStatusData)
				if err != nil {
					logger.Error(fmt.Errorf("error on send ws logging status: %w", err))
				}
			}
		}
	}()
	// играем в пинг-понг
	for {
		if ctx.Err() != nil {
			return
		}

		readCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
		_, messageData, err := conn.Read(readCtx)
		cancel()

		if err != nil {
			logger.Error(fmt.Errorf("error on read ws message: %w", err))
			return
		}

		if bytes.Equal(messageData, []byte("ping")) {
			err = conn.Write(ctx, websocket.MessageText, []byte("pong"))
			if err != nil {
				logger.Error(fmt.Errorf("error on send ws pong: %w", err))
				return
			}
		}
	}
}
