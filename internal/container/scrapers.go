package container

import (
	donation_alerts_client "github.com/PlayingPossumHiss/possum_chat/internal/infra/clients/donation_alerts"
	"github.com/PlayingPossumHiss/possum_chat/internal/infra/clients/kick_chat_api"
	"github.com/PlayingPossumHiss/possum_chat/internal/infra/clients/twitch_irc_client"
	"github.com/PlayingPossumHiss/possum_chat/internal/infra/clients/vk_play_live_api"
	"github.com/PlayingPossumHiss/possum_chat/internal/infra/clients/vk_play_live_ws"
	youtube_client "github.com/PlayingPossumHiss/possum_chat/internal/infra/clients/youtube"
	"github.com/PlayingPossumHiss/possum_chat/internal/service/scrapers/donation_alerts"
	"github.com/PlayingPossumHiss/possum_chat/internal/service/scrapers/kick"
	"github.com/PlayingPossumHiss/possum_chat/internal/service/scrapers/twitch"
	"github.com/PlayingPossumHiss/possum_chat/internal/service/scrapers/vk_play_live"
	youtube_scraper "github.com/PlayingPossumHiss/possum_chat/internal/service/scrapers/youtube"
	utils_time "github.com/PlayingPossumHiss/possum_chat/internal/utils/time"
)

func (c *Container) getYoutubeScraper() (*youtube_scraper.Service, error) {
	if c.ytScraper != nil {
		return c.ytScraper, nil
	}

	configService, err := c.getConfig()
	if err != nil {
		return nil, err
	}

	c.ytScraper = youtube_scraper.New(
		configService,
		c.getYoutubeClient(),
	)

	return c.ytScraper, nil
}

func (c *Container) getVkPlayLiveScraper() (*vk_play_live.Service, error) {
	if c.vkScraper != nil {
		return c.vkScraper, nil
	}

	configService, err := c.getConfig()
	if err != nil {
		return nil, err
	}

	scraper, err := vk_play_live.New(
		configService,
		c.getVkPalyLiveApi(),
		c.getVkPalyLiveWs(),
	)
	if err != nil {
		return nil, err
	}
	c.vkScraper = scraper

	return c.vkScraper, nil
}

func (c *Container) getDonationAlertsScraper() (*donation_alerts.Service, error) {
	if c.daScraper != nil {
		return c.daScraper, nil
	}

	configService, err := c.getConfig()
	if err != nil {
		return nil, err
	}

	scraper, err := donation_alerts.New(
		c.getDonationAlertsClient(),
		configService,
	)
	if err != nil {
		return nil, err
	}
	c.daScraper = scraper

	return c.daScraper, nil
}

func (c *Container) getTwitchScraper() (*twitch.Service, error) {
	if c.twitchScraper != nil {
		return c.twitchScraper, nil
	}

	configService, err := c.getConfig()
	if err != nil {
		return nil, err
	}

	c.twitchScraper = twitch.New(
		c.getTwitchClient(),
		configService,
	)

	return c.twitchScraper, nil
}

func (c *Container) getKickScraper() (*kick.Service, error) {
	if c.kickScraper != nil {
		return c.kickScraper, nil
	}

	configService, err := c.getConfig()
	if err != nil {
		return nil, err
	}

	c.kickScraper = kick.New(
		c.getKickApi(),
		configService,
	)

	return c.kickScraper, nil
}

func (c *Container) getVkPalyLiveApi() *vk_play_live_api.Client {
	if c.vkPlayLiveApi != nil {
		return c.vkPlayLiveApi
	}

	c.vkPlayLiveApi = vk_play_live_api.New()

	return c.vkPlayLiveApi
}

func (c *Container) getKickApi() *kick_chat_api.Client {
	if c.kickApi != nil {
		return c.kickApi
	}

	c.kickApi = kick_chat_api.New()

	return c.kickApi
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
