package container

import (
	"context"
	"time"

	"github.com/PlayingPossumHiss/possum_chat/internal/api"
	"github.com/PlayingPossumHiss/possum_chat/internal/entity"
	"github.com/PlayingPossumHiss/possum_chat/internal/service/message_queue"
	youtube_scraper "github.com/PlayingPossumHiss/possum_chat/internal/service/scrapers/youtube"
	"github.com/PlayingPossumHiss/possum_chat/internal/service/settings"
	"github.com/PlayingPossumHiss/possum_chat/internal/use_case/get_style"
	"github.com/PlayingPossumHiss/possum_chat/internal/use_case/list_messages"
	"github.com/PlayingPossumHiss/possum_chat/internal/use_case/run_watch_scrapers"
	utils_time "github.com/PlayingPossumHiss/possum_chat/internal/utils/time"
)

type Container struct {
	ctx context.Context

	// сервисы
	configService       *settings.Service
	messageQueueService *message_queue.Service
	scrapers            []run_watch_scrapers.Scraper

	// юзкейсы
	watchSubscribersRunner *run_watch_scrapers.UseCase
	messagesLister         *list_messages.UseCase
	styleGetter            *get_style.UseCase

	// апишка (своя)
	selfApi *api.Api
}

func New(ctx context.Context) *Container {
	container := &Container{
		ctx: ctx,
	}

	return container
}

func (c *Container) Run() error {
	messageWatcher, err := c.getWatchSubscribersRunner()
	if err != nil {
		return err
	}

	// TODO: норм воркеры сделать
	go func() {
		for {
			messageWatcher.Run(c.ctx)
			time.Sleep(time.Millisecond * 150)
		}
	}()

	api, err := c.getSelfApi()
	if err != nil {
		return err
	}

	return api.Run()
}

func (c *Container) getSelfApi() (*api.Api, error) {
	if c.selfApi != nil {
		return c.selfApi, nil
	}

	configService, err := c.getConfig()
	if err != nil {
		return nil, err
	}
	config := configService.Config()

	styleGetter, err := c.getStyleGetter()
	if err != nil {
		return nil, err
	}

	messageLister, err := c.getMessagesLister()
	if err != nil {
		return nil, err
	}

	c.selfApi = api.New(
		config.Port,
		styleGetter,
		messageLister,
	)

	return c.selfApi, nil
}

func (c *Container) getStyleGetter() (*get_style.UseCase, error) {
	if c.styleGetter != nil {
		return c.styleGetter, nil
	}

	configService, err := c.getConfig()
	if err != nil {
		return nil, err
	}

	c.styleGetter = get_style.New(configService)
	return c.styleGetter, nil
}

func (c *Container) getMessagesLister() (*list_messages.UseCase, error) {
	if c.messagesLister != nil {
		return c.messagesLister, nil
	}

	messageQueueService, err := c.getMessageQueueService()
	if err != nil {
		return nil, err
	}
	c.messagesLister = list_messages.New(messageQueueService)

	return c.messagesLister, nil
}

func (c *Container) getWatchSubscribersRunner() (*run_watch_scrapers.UseCase, error) {
	if c.watchSubscribersRunner != nil {
		return c.watchSubscribersRunner, nil
	}

	scrapers, err := c.getScrapers()
	if err != nil {
		return nil, err
	}
	messageQueueService, err := c.getMessageQueueService()
	if err != nil {
		return nil, err
	}

	c.watchSubscribersRunner = run_watch_scrapers.New(
		scrapers,
		messageQueueService,
	)

	return c.watchSubscribersRunner, nil
}

func (c *Container) getMessageQueueService() (*message_queue.Service, error) {
	if c.messageQueueService != nil {
		return c.messageQueueService, nil
	}

	configService, err := c.getConfig()
	if err != nil {
		return nil, err
	}

	c.messageQueueService = message_queue.New(
		c.ctx,
		configService,
		&utils_time.DefaultClock{},
	)

	return c.messageQueueService, nil
}

func (c *Container) getConfig() (*settings.Service, error) {
	if c.configService != nil {
		return c.configService, nil
	}

	configService, err := settings.New()
	if err != nil {
		return nil, err
	}

	c.configService = configService
	return c.configService, nil
}

func (c *Container) getScrapers() ([]run_watch_scrapers.Scraper, error) {
	if c.scrapers != nil {
		return c.scrapers, nil
	}

	configService, err := c.getConfig()
	if err != nil {
		return nil, err
	}

	result := []run_watch_scrapers.Scraper{}
	for _, connection := range configService.Config().Connections {
		switch connection.Source {
		case entity.SourceYoutube:
			result = append(result, c.getYoutubeScraper(connection))
		}
	}

	c.scrapers = result
	return c.scrapers, nil
}

func (c *Container) getYoutubeScraper(
	configConnection entity.ConfigConnection,
) *youtube_scraper.Service {
	return youtube_scraper.New(
		c.ctx,
		configConnection.Key,
		configConnection.RefreshTime,
	)
}
