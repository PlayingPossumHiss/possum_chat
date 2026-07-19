package container

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/PlayingPossumHiss/possum_chat/internal/api"
	"github.com/PlayingPossumHiss/possum_chat/internal/entity"
	"github.com/PlayingPossumHiss/possum_chat/internal/infra/clients/kick_chat_api"
	"github.com/PlayingPossumHiss/possum_chat/internal/infra/clients/vk_play_live_api"
	"github.com/PlayingPossumHiss/possum_chat/internal/service/language_provider"
	"github.com/PlayingPossumHiss/possum_chat/internal/service/logger"
	"github.com/PlayingPossumHiss/possum_chat/internal/service/message_queue"
	"github.com/PlayingPossumHiss/possum_chat/internal/service/scrapers/donation_alerts"
	"github.com/PlayingPossumHiss/possum_chat/internal/service/scrapers/kick"
	"github.com/PlayingPossumHiss/possum_chat/internal/service/scrapers/twitch"
	"github.com/PlayingPossumHiss/possum_chat/internal/service/scrapers/vk_play_live"
	youtube_scraper "github.com/PlayingPossumHiss/possum_chat/internal/service/scrapers/youtube"
	"github.com/PlayingPossumHiss/possum_chat/internal/service/settings"
	"github.com/PlayingPossumHiss/possum_chat/internal/ui"
	"github.com/PlayingPossumHiss/possum_chat/internal/use_case/get_online"
	"github.com/PlayingPossumHiss/possum_chat/internal/use_case/get_style"
	"github.com/PlayingPossumHiss/possum_chat/internal/use_case/list_messages"
	"github.com/PlayingPossumHiss/possum_chat/internal/use_case/run_watch_scrapers"
	"github.com/PlayingPossumHiss/possum_chat/internal/use_case/send_test_messages"
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

	// скрейперы
	ytScraper     *youtube_scraper.Service
	twitchScraper *twitch.Service
	vkScraper     *vk_play_live.Service
	kickScraper   *kick.Service
	daScraper     *donation_alerts.Service

	// апишка (своя)
	selfApi *api.Api

	// инфра
	vkPlayLiveApi *vk_play_live_api.Client
	kickApi       *kick_chat_api.Client

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

	c.scheduler.Start()

	api.Run()

	err = c.startUI()
	if err != nil {
		return err
	}

	return nil
}

func (c *Container) startUI() error {
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

	sendTestMessageUC, err := c.getSendTestMessageUseCase()
	if err != nil {
		return err
	}

	uiScrapers := make(map[entity.Source]ui.Scraper, len(scrapers))
	for source, scraper := range scrapers {
		uiScrapers[source] = scraper
	}
	err = ui.New(
		languageProvider,
		uiScrapers,
		configService,
		sendTestMessageUC,
	)
	if err != nil {
		return err
	}

	return nil
}

func (c *Container) getSendTestMessageUseCase() (*send_test_messages.UseCase, error) {
	messageQueue, err := c.getMessageQueueService()
	if err != nil {
		return nil, err
	}

	return send_test_messages.New(messageQueue), nil
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

	onlineScrapers, err := c.getOnlineScraper()
	if err != nil {
		return nil, err
	}

	c.selfApi = api.New(
		config.Port,
		styleGetter,
		messageLister,
		onlineScrapers,
	)

	return c.selfApi, nil
}

func (c *Container) getOnlineScraper() (*get_online.OnlineGetter, error) {
	onlineScrapers := map[entity.Source]get_online.Scraper{}

	ytScraper, err := c.getYoutubeScraper()
	if err != nil {
		return nil, err
	}
	onlineScrapers[entity.SourceYoutube] = ytScraper
	twitchScraper, err := c.getTwitchScraper()
	if err != nil {
		return nil, err
	}
	onlineScrapers[entity.SourceTwitch] = twitchScraper
	kickScraper, err := c.getKickScraper()
	if err != nil {
		return nil, err
	}
	onlineScrapers[entity.SourceKick] = kickScraper
	vkScraper, err := c.getVkPlayLiveScraper()
	if err != nil {
		return nil, err
	}
	onlineScrapers[entity.SourceVkPlayLive] = vkScraper

	configService, err := c.getConfig()
	if err != nil {
		return nil, err
	}

	return get_online.New(
		onlineScrapers,
		configService,
	), nil
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
	kickScraper, err := c.getKickScraper()
	if err != nil {
		return nil, err
	}
	result[entity.SourceKick] = kickScraper
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
