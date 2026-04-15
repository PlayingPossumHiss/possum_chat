package container

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/PlayingPossumHiss/possum_chat/internal/api"
	"github.com/PlayingPossumHiss/possum_chat/internal/entity"
	donation_alerts_client "github.com/PlayingPossumHiss/possum_chat/internal/infra/clients/donation_alerts"
	"github.com/PlayingPossumHiss/possum_chat/internal/infra/clients/twitch_irc_client"
	"github.com/PlayingPossumHiss/possum_chat/internal/infra/clients/vk_play_live_api"
	"github.com/PlayingPossumHiss/possum_chat/internal/infra/clients/vk_play_live_ws"
	youtube_client "github.com/PlayingPossumHiss/possum_chat/internal/infra/clients/youtube"
	"github.com/PlayingPossumHiss/possum_chat/internal/service/logger"
	"github.com/PlayingPossumHiss/possum_chat/internal/service/message_queue"
	"github.com/PlayingPossumHiss/possum_chat/internal/service/scrapers/donation_alerts"
	"github.com/PlayingPossumHiss/possum_chat/internal/service/scrapers/twitch"
	"github.com/PlayingPossumHiss/possum_chat/internal/service/scrapers/vk_play_live"
	youtube_scraper "github.com/PlayingPossumHiss/possum_chat/internal/service/scrapers/youtube"
	"github.com/PlayingPossumHiss/possum_chat/internal/service/settings"
	"github.com/PlayingPossumHiss/possum_chat/internal/use_case/get_style"
	"github.com/PlayingPossumHiss/possum_chat/internal/use_case/list_messages"
	"github.com/PlayingPossumHiss/possum_chat/internal/use_case/run_watch_scrapers"
	utils_time "github.com/PlayingPossumHiss/possum_chat/internal/utils/time"
	"github.com/go-co-op/gocron/v2"
	"github.com/google/uuid"
)

type Container struct {
	ctx context.Context //nolint

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

	// инфра
	vkPlayLiveApi *vk_play_live_api.Client

	// шедулер
	scheduler gocron.Scheduler
}

const (
	messageWatcherRunMs                   = 30
	messageQueueServiceCleanOldMessagesMs = 30
)

func New(ctx context.Context) (*Container, error) {
	container := &Container{
		ctx: ctx,
	}

	config, err := container.getConfig()
	if err != nil {
		return nil, err
	}

	err = logger.Init(config)
	if err != nil {
		return nil, err
	}

	return container, nil
}

func (c *Container) Run() error {
	scheduler, err := gocron.NewScheduler()
	if err != nil {
		return err
	}
	c.scheduler = scheduler

	messageWatcher, err := c.getWatchSubscribersRunner()
	if err != nil {
		return err
	}

	err = c.addJobToScheduler(
		messageWatcher.Run,
		"ask_watchers_for_messages",
		time.Millisecond*messageWatcherRunMs,
	)
	if err != nil {
		return err
	}

	api, err := c.getSelfApi()
	if err != nil {
		return err
	}

	c.scheduler.Start()

	return api.Run()
}

func (c *Container) addJobToScheduler(
	job func(ctx context.Context) error,
	jobName string,
	interval time.Duration,
) error {
	logger.Info(fmt.Sprintf("starting bg task %s", jobName))

	_, err := c.scheduler.NewJob(
		gocron.DurationJob(interval),
		gocron.NewTask(job),
		gocron.WithContext(c.ctx),
		gocron.WithName(jobName),
		gocron.WithEventListeners(
			gocron.AfterJobRunsWithError(func(jobID uuid.UUID, jobName string, jobErr error) {
				log.Printf("job %s fails %s\n", jobName, jobErr.Error())
			}),
		),
	)

	return err
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
		configService,
		&utils_time.DefaultClock{},
	)

	err = c.addJobToScheduler(
		c.messageQueueService.CleanOldMessages,
		"clean_old_messages",
		time.Millisecond*messageQueueServiceCleanOldMessagesMs,
	)
	if err != nil {
		return nil, err
	}

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
			logger.Info("init youtube scraper")
			result = append(result, c.getYoutubeScraper(connection))
		case entity.SourceVkPlayLive:
			logger.Info("init vk play live scraper")
			scraper, err := c.getVkPlayLiveScraper(connection)
			if err != nil {
				return nil, err
			}
			result = append(result, scraper)
		case entity.SourceTwitch:
			logger.Info("init twitch scraper")
			result = append(result, c.getTwitchScraper(connection))
		case entity.SourceDonationAlerts:
			logger.Info("init donation alerts scraper")
			scraper, err := c.getDonationAlertsSubscraper(connection)
			if err != nil {
				return nil, err
			}
			result = append(result, scraper)
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
		c.getYoutubeClient(),
	)
}

func (c *Container) getVkPlayLiveScraper(
	configConnection entity.ConfigConnection,
) (*vk_play_live.Service, error) {
	return vk_play_live.New(
		c.ctx,
		configConnection.Key,
		c.getVkPalyLiveApi(),
		c.getVkPalyLiveWs(),
	)
}

func (c *Container) getDonationAlertsSubscraper(
	configConnection entity.ConfigConnection,
) (*donation_alerts.Service, error) {
	return donation_alerts.New(
		c.ctx,
		c.getDonationAlertsClient(),
		configConnection.Key,
	)
}

func (c *Container) getTwitchScraper(
	configConnection entity.ConfigConnection,
) *twitch.Service {
	return twitch.New(
		c.ctx,
		c.getTwitchClient(),
		configConnection.Key,
	)
}

func (c *Container) getVkPalyLiveApi() *vk_play_live_api.Client {
	if c.vkPlayLiveApi != nil {
		return c.vkPlayLiveApi
	}

	c.vkPlayLiveApi = vk_play_live_api.New()

	return c.vkPlayLiveApi
}

func (c *Container) getVkPalyLiveWs() *vk_play_live_ws.Client {
	// тут отдельный коннект на каждое соединение
	return vk_play_live_ws.New()
}

func (c *Container) getYoutubeClient() *youtube_client.Client {
	// тут отдельный коннект на каждое соединение
	return youtube_client.New()
}

func (c *Container) getTwitchClient() *twitch_irc_client.Client {
	// тут отдельный коннект на каждое соединение
	return twitch_irc_client.New()
}

func (c *Container) getDonationAlertsClient() *donation_alerts_client.Client {
	// тут отдельный коннект на каждое соединение
	return donation_alerts_client.New(&utils_time.DefaultClock{})
}
