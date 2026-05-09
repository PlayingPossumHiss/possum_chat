package ui

import (
	"context"
	"fmt"
	"os"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/PlayingPossumHiss/possum_chat/internal/entity"
	"github.com/PlayingPossumHiss/possum_chat/internal/service/logger"
)

type UI struct {
	app              fyne.App
	mainWindow       fyne.Window
	configStorage    ConfigStorage
	languageProvider LanguageProvider
	scrapers         map[entity.Source]Scraper
}

func New(
	languageProvider LanguageProvider,
	scrapers map[entity.Source]Scraper,
	configStorage ConfigStorage,
) error {
	newUI := &UI{
		app:              app.New(),
		scrapers:         scrapers,
		languageProvider: languageProvider,
		configStorage:    configStorage,
	}

	err := newUI.newMainWindow()
	if err != nil {
		return err
	}

	newUI.mainWindow.ShowAndRun()

	return nil
}

func (ui *UI) newMainWindow() error {
	mainWindow := ui.app.NewWindow("Possum Chat")

	mainWindowIconData, err := os.ReadFile("./static/img/favicon.ico")
	if err != nil {
		return fmt.Errorf("error on get main window icon: %w", err)
	}
	mainWindowIcon := fyne.NewStaticResource("main_window_icon", mainWindowIconData)
	mainWindow.SetIcon(mainWindowIcon)

	tabs := container.NewAppTabs()

	connectionTabContent, err := ui.getConnectionTabContent()
	if err != nil {
		return fmt.Errorf("error on get main window connection tab content: %w", err)
	}
	tabs.Append(container.NewTabItem(
		ui.languageProvider.Local(entity.LanguageTextConstantConnectionsTab),
		connectionTabContent,
	))

	tabs.Append(container.NewTabItem(
		ui.languageProvider.Local(entity.LanguageTextConstantCSSTab),
		ui.getCssTabContent(),
	))

	tabs.Append(container.NewTabItem(
		ui.languageProvider.Local(entity.LanguageTextConstantSettingsTab),
		ui.getSettingsTabContent(),
	))

	mainWindow.SetContent(tabs)

	ui.mainWindow = mainWindow

	return nil
}

func (ui *UI) getConnectionTabContent() (*fyne.Container, error) {
	const itemsInLine = 3

	switchesContent := make([]fyne.CanvasObject, 0, itemsInLine*(1+len(ui.scrapers)))
	switchesContent = append(
		switchesContent,
		widget.NewLabel(ui.languageProvider.Local(entity.LanguageTextConstantConnectionSwitchesHead)),
		widget.NewLabel(ui.languageProvider.Local(entity.LanguageTextConstantConnectionSourcesHead)),
		widget.NewLabel(ui.languageProvider.Local(entity.LanguageTextConstantConnectionKeysHead)),
	)

	connectionsOrder := []entity.Source{
		entity.SourceYoutube,
		entity.SourceTwitch,
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

	grid := container.New(
		layout.NewGridLayout(itemsInLine),
		switchesContent...,
	)

	return grid, nil
}

func (ui *UI) getConnectionRow(source entity.Source) ([]fyne.CanvasObject, error) {
	scraper := ui.scrapers[source]
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
	scraperConfig.SetText(scraper.GetConnectionConfig())
	scraperConfig.Password = source.KeyIsSecret()
	scraperConfig.OnChanged = func(s string) {
		err = ui.configStorage.UpdateConfig([]entity.ConfigUpdateOption{
			scraper.ConnectionConfigUpdateOption(scraperConfig.Text),
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
