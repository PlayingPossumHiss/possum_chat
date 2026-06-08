package ui

import (
	"context"
	"fmt"
	"net/url"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/PlayingPossumHiss/possum_chat/internal/entity"
	"github.com/PlayingPossumHiss/possum_chat/internal/service/logger"
)

func (ui *UI) getConnectionTabContent() (*fyne.Container, error) {
	switchesContent := make([]fyne.CanvasObject, 0, connectionTabRows*(1+len(ui.scrapers)))
	switchesContent = append(
		switchesContent,
		widget.NewLabel(ui.languageProvider.Local(entity.LanguageTextConstantConnectionSwitchesHead)),
		widget.NewLabel(ui.languageProvider.Local(entity.LanguageTextConstantConnectionSourcesHead)),
		widget.NewLabel(ui.languageProvider.Local(entity.LanguageTextConstantConnectionKeysHead)),
	)

	connectionsOrder := []entity.Source{
		entity.SourceYoutube,
		entity.SourceTwitch,
		entity.SourceKick,
		entity.SourceVkPlayLive,
		entity.SourceDonationAlerts,
	}
	for _, source := range connectionsOrder {
		rowItems, err := ui.getConnectionRow(source)
		if err != nil {
			return nil, err
		}
		switchesContent = append(switchesContent, rowItems...)
	}

	switchesContent = append(switchesContent, ui.getLinkButtons()...)

	grid := container.NewVBox(
		container.New(
			layout.NewGridLayout(connectionTabRows),
			switchesContent...,
		),
	)

	return grid, nil
}

func (ui *UI) getLinkButtons() []fyne.CanvasObject {
	linksViews := make([]fyne.CanvasObject, 0, connectionTabRows)
	linksViews = append(
		linksViews,
		widget.NewHyperlink(
			ui.languageProvider.Local(entity.LanguageTextConstantWidgetOBS),
			mustParseUrl(fmt.Sprintf(
				"http://127.0.0.1:%d/messages.html",
				ui.configStorage.Config().Port),
			),
		),
		widget.NewHyperlink(
			ui.languageProvider.Local(entity.LanguageTextConstantMessagePanel),
			mustParseUrl(fmt.Sprintf(
				"http://127.0.0.1:%d/messages.html?for_last=1h&use_scroll=true",
				ui.configStorage.Config().Port),
			),
		),
		widget.NewHyperlink(
			ui.languageProvider.Local(entity.LanguageTextConstantMyGithub),
			mustParseUrl("https://github.com/PlayingPossumHiss/possum_chat"),
		),
	)

	return linksViews
}

func mustParseUrl(src string) *url.URL {
	link, err := url.Parse(src)
	if err != nil {
		logger.Error(fmt.Errorf("error on mustParseUrl: %w", err))
	}

	return link
}

func (ui *UI) getConnectionRow(source entity.Source) ([]fyne.CanvasObject, error) {
	// Заголовок
	scraperContent := binding.NewString()
	err := scraperContent.Set(ui.getLabelText(source, false))
	if err != nil {
		return nil, err
	}
	scraperLabel := widget.NewLabelWithData(scraperContent)

	// Кнопка переключения
	scraperButton := widget.NewButton(
		ui.languageProvider.Local(entity.LanguageTextConstantConnectionSwitchButton),
		ui.turnButtonHandler(
			source,
			scraperContent,
		),
	)

	// Строка с конфигом
	scraperConfig := widget.NewEntry()
	scraperConfig.SetText(ui.getConnectionConfig(source))
	scraperConfig.Password = source.KeyIsSecret()
	scraperConfig.SetPlaceHolder(ui.connectionPlaceholder(source))
	scraperConfig.OnChanged = func(s string) {
		err = ui.configStorage.UpdateConfig([]entity.ConfigUpdateOption{
			connectionConfigUpdateOption(source, scraperConfig.Text),
		})
		if err != nil {
			logger.Error(fmt.Errorf("failed to update config: %w", err))
		}
	}

	return []fyne.CanvasObject{
		scraperButton,
		scraperLabel,
		scraperConfig,
	}, nil
}

func (ui *UI) getConnectionConfig(src entity.Source) string {
	switch src {
	case entity.SourceYoutube:
		return ui.configStorage.Config().Connections.Youtube.ChannelName
	case entity.SourceTwitch:
		return ui.configStorage.Config().Connections.Twitch.ChannelName
	case entity.SourceKick:
		return ui.configStorage.Config().Connections.Kick.ChannelName
	case entity.SourceVkPlayLive:
		return ui.configStorage.Config().Connections.VkPlayLive.ChannelName
	case entity.SourceDonationAlerts:
		return ui.configStorage.Config().Connections.DonationAlerts.Token
	}

	return ""
}

func (ui *UI) connectionPlaceholder(src entity.Source) string {
	switch src {
	case entity.SourceDonationAlerts:
		return ui.languageProvider.Local(entity.LanguageTextConstantDaConnPlaceholder)
	case entity.SourceTwitch, entity.SourceVkPlayLive, entity.SourceKick:
		return ui.languageProvider.Local(entity.LanguageTextConstantTwitchConnPlaceholder)
	case entity.SourceYoutube:
		return ui.languageProvider.Local(entity.LanguageTextConstantYoutubeConnPlaceholder)
	}

	return ""
}

func (ui *UI) turnButtonHandler(
	source entity.Source,
	label binding.String,
) func() {
	return func() {
		scraper, ok := ui.scrapers[source]
		if !ok {
			logger.Error("can't scraper")

			return
		}
		if scraper.Status() == entity.ScraperStateStopped {
			scraper.Run(context.Background())
			err := label.Set(ui.getLabelText(source, true))
			if err != nil {
				logger.Error(fmt.Errorf("failed to change label: %w", err))

				return
			}
		} else {
			scraper.Stop()
			err := label.Set(ui.getLabelText(source, false))
			if err != nil {
				logger.Error(fmt.Errorf("failed to change label: %w", err))

				return
			}
		}
	}
}

func (ui *UI) getLabelText(
	source entity.Source,
	isActive bool,
) string {
	var (
		serviceName string
		statusName  string
	)

	switch source {
	case entity.SourceDonationAlerts:
		serviceName = "Donation Alerts"
	case entity.SourceTwitch:
		serviceName = "Twitch"
	case entity.SourceKick:
		serviceName = "Kick"
	case entity.SourceVkPlayLive:
		serviceName = "VK Play Live"
	case entity.SourceYoutube:
		serviceName = "Youtube"
	default:
		serviceName = "Unknown"
	}

	if isActive {
		statusName = ui.languageProvider.Local(entity.LanguageTextConstantUnknownScraperIsOn)
	} else {
		statusName = ui.languageProvider.Local(entity.LanguageTextConstantUnknownScraperIsOff)
	}

	return fmt.Sprintf("%s (%s)", serviceName, statusName)
}

func connectionConfigUpdateOption(source entity.Source, newValue string) entity.ConfigUpdateOption {
	switch source {
	case entity.SourceYoutube:
		return func(c *entity.Config) {
			c.Connections.Youtube.ChannelName = newValue
		}
	case entity.SourceDonationAlerts:
		return func(c *entity.Config) {
			c.Connections.DonationAlerts.Token = newValue
		}
	case entity.SourceTwitch:
		return func(c *entity.Config) {
			c.Connections.Twitch.ChannelName = newValue
		}
	case entity.SourceVkPlayLive:
		return func(c *entity.Config) {
			c.Connections.VkPlayLive.ChannelName = newValue
		}
	case entity.SourceKick:
		return func(c *entity.Config) {
			c.Connections.Kick.ChannelName = newValue
		}
	}

	return func(c *entity.Config) {}
}
