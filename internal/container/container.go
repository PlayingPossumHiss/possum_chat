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
	"github.com/PlayingPossumHiss/possum_chat/internal/service/language_provider"
	"github.com/PlayingPossumHiss/possum_chat/internal/service/logger"
	"github.com/PlayingPossumHiss/possum_chat/internal/service/message_queue"
	"github.com/PlayingPossumHiss/possum_chat/internal/service/scrapers/donation_alerts"
	"github.com/PlayingPossumHiss/possum_chat/internal/service/scrapers/twitch"
	"github.com/PlayingPossumHiss/possum_chat/internal/service/scrapers/vk_play_live"
	youtube_scraper "github.com/PlayingPossumHiss/possum_chat/internal/service/scrapers/youtube"
	"github.com/PlayingPossumHiss/possum_chat/internal/service/settings"
	"github.com/PlayingPossumHiss/possum_chat/internal/ui"
	"github.com/PlayingPossumHiss/possum_chat/internal/use_case/get_style"
	"github.com/PlayingPossumHiss/possum_chat/internal/use_case/list_messages"
	"github.com/PlayingPossumHiss/possum_chat/internal/use_case/run_watch_scrapers"
	utils_time "github.com/PlayingPossumHiss/possum_chat/internal/utils/time"
	"github.com/go-co-op/gocron/v2"
	"github.com/google/uuid"
)

type Container struct {
	// сервисы
	configService       *settings.Service
	messageQueueService *message_queue.Service
	scrapers            map[entity.Source]Scraper
	languageProvider    *language_provider.LanguageProvider

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
	container := &Container{}

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

	scrapers, err := c.getScrapers()
	if err != nil {
		return err
	}

	configService, err := c.getConfig()
	if err != nil {
		return err
	}

	languageProvider, err := c.getLanguageProvider()
	if err != nil {
		return err
	}

	c.scheduler.Start()

	api.Run()

	uiScrapers := make(map[entity.Source]ui.Scraper, len(scrapers))
	for source, scraper := range scrapers {
		uiScrapers[source] = scraper
	}
	err = ui.New(
		languageProvider,
		uiScrapers,
		configService,
	)
	if err != nil {
		return err
	}

	return nil
}

func (c *Container) getLanguageProvider() (*language_provider.LanguageProvider, error) {
	if c.languageProvider != nil {
		return c.languageProvider, nil
	}

	configService, err := c.getConfig()
	if err != nil {
		return nil, err
	}

	c.languageProvider = language_provider.New(configService.Config().UI.Lang)

	return c.languageProvider, nil
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

	scrapersForRun := make([]run_watch_scrapers.Scraper, 0, len(scrapers))
	for _, scraper := range scrapers {
		scrapersForRun = append(scrapersForRun, scraper)
	}
	c.watchSubscribersRunner = run_watch_scrapers.New(
		scrapersForRun,
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

func (c *Container) getScrapers() (map[entity.Source]Scraper, error) {
	if c.scrapers != nil {
		return c.scrapers, nil
	}

	result := map[entity.Source]Scraper{}
	ytScraper, err := c.getYoutubeScraper()
	if err != nil {
		return nil, err
	}
	result[entity.SourceYoutube] = ytScraper
	twitchScraper, err := c.getTwitchScraper()
	if err != nil {
		return nil, err
	}
	result[entity.SourceTwitch] = twitchScraper
	vkScraper, err := c.getVkPlayLiveScraper()
	if err != nil {
		return nil, err
	}
	result[entity.SourceVkPlayLive] = vkScraper
	daScraper, err := c.getDonationAlertsScraper()
	if err != nil {
		return nil, err
	}
	result[entity.SourceDonationAlerts] = daScraper

	c.scrapers = result

	return c.scrapers, nil
}

func (c *Container) getYoutubeScraper() (*youtube_scraper.Service, error) {
	configService, err := c.getConfig()
	if err != nil {
		return nil, err
	}

	return youtube_scraper.New(
		configService,
		c.getYoutubeClient(),
	), nil
}

func (c *Container) getVkPlayLiveScraper() (*vk_play_live.Service, error) {
	configService, err := c.getConfig()
	if err != nil {
		return nil, err
	}

	return vk_play_live.New(
		configService,
		c.getVkPalyLiveApi(),
		c.getVkPalyLiveWs(),
	)
}

func (c *Container) getDonationAlertsScraper() (*donation_alerts.Service, error) {
	configService, err := c.getConfig()
	if err != nil {
		return nil, err
	}

	return donation_alerts.New(
		c.getDonationAlertsClient(),
		configService,
	)
}

func (c *Container) getTwitchScraper() (*twitch.Service, error) {
	configService, err := c.getConfig()
	if err != nil {
		return nil, err
	}

	return twitch.New(
		c.getTwitchClient(),
		configService,
	), nil
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
