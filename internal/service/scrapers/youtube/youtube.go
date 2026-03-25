package youtube_scraper

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"slices"
	"sync"
	"time"

	"github.com/PlayingPossumHiss/possum_chat/internal/entity"
	yt_chat "github.com/epjane/youtube-live-chat-downloader/v2"
	"github.com/google/uuid"
)

type Service struct {
	streamKey string
	ctx       context.Context
	cooldown  time.Duration

	messages  []entity.Message
	messageMx *sync.Mutex
}

func New(
	ctx context.Context,
	streamKey string,
	cooldown time.Duration,
) *Service {
	scraper := &Service{
		streamKey: streamKey,
		cooldown:  cooldown,
		ctx:       ctx,
		messageMx: &sync.Mutex{},
	}

	go scraper.watchChat()

	return scraper
}

func (s *Service) GetMessages() []entity.Message {
	s.messageMx.Lock()
	defer s.messageMx.Unlock()
	result := slices.Clone(s.messages)
	s.messages = nil
	return result
}

func (s *Service) watchChat() {
	for {
		customCookies := []*http.Cookie{
			{Name: "PREF",
				Value:  "tz=Europe.Rome",
				MaxAge: 300},
			{Name: "CONSENT",
				Value:  fmt.Sprintf("YES+yt.432048971.it+FX+%d", 100+rand.Intn(999-100+1)),
				MaxAge: 300},
		}
		yt_chat.AddCookies(customCookies)

		continuation, cfg, err := yt_chat.ParseInitialData(fmt.Sprintf("https://www.youtube.com/watch?v=%s", s.streamKey))
		if err != nil {
			log.Panicln(err)
			continue
		}

		err = s.scrap(continuation, cfg)
		if err != nil {
			log.Panicln(err)
		}
	}

}

func (s *Service) scrap(
	continuation string,
	cfg yt_chat.YtCfg,
) error {
	for {
		select {
		case <-s.ctx.Done():
			return s.ctx.Err()
		default:
			for {
				chat, newContinuation, error := yt_chat.FetchContinuationChat(continuation, cfg)
				if error == yt_chat.ErrLiveStreamOver {
					log.Fatal("Live stream over")
				}
				if error != nil {
					log.Print(error)
					continue
				}
				// set the newly received continuation
				continuation = newContinuation

				comments := make([]entity.Message, 0, len(chat))
				for _, msg := range chat {
					id, err := uuid.NewV7()
					if err != nil {
						return err
					}
					comments = append(comments, entity.Message{
						Text:      msg.Message,
						Source:    entity.SourceYoutube,
						User:      msg.AuthorName,
						CreatedAt: msg.Timestamp.UTC(),
						ID:        id.String(),
					})
				}

				s.messageMx.Lock()
				s.messages = append(s.messages, comments...)
				s.messageMx.Unlock()

				time.Sleep(s.cooldown)
			}
		}
	}
}
