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
	app           fyne.App
	mainWindow    fyne.Window
	configStorage ConfigStorage
	scrapers      map[entity.Source]Scraper
}

type onSaveCallback func() entity.ConfigUpdateOption

func New(
	scrapers map[entity.Source]Scraper,
	configStorage ConfigStorage,
) error {
	newUI := &UI{
		app:           app.New(),
		scrapers:      scrapers,
		configStorage: configStorage,
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
	tabs.Append(container.NewTabItem("Connections", connectionTabContent))

	tabs.Append(container.NewTabItem("Styles", ui.getCssTabContent()))

	mainWindow.SetContent(tabs)

	ui.mainWindow = mainWindow

	return nil
}

func (ui *UI) getCssTabContent() fyne.CanvasObject {
	cssField := widget.NewEntry()
	cssField.SetText(ui.configStorage.Config().View.CssStyle)
	cssField.MultiLine = true
	cssField.OnChanged = func(s string) {
		ui.configStorage.UpdateConfig([]entity.ConfigUpdateOption{
			func(target *entity.Config) {
				target.View.CssStyle = cssField.Text
			},
		})
	}

	return cssField
}

func (ui *UI) getConnectionTabContent() (fyne.CanvasObject, error) {
	const itemsInLine = 3

	switchesContent := make([]fyne.CanvasObject, 0, itemsInLine*(1+len(ui.scrapers)))
	switchesContent = append(
		switchesContent,
		widget.NewLabel("Switch"),
		widget.NewLabel("Sources"),
		widget.NewLabel("Key"),
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
	err := scraperContent.Set(getLabelText(source, false))
	if err != nil {
		return nil, err
	}
	scraperLabel := widget.NewLabelWithData(scraperContent)

	// Кнопка переключения
	scraperButton := widget.NewButton(
		"Turn",
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
		ui.configStorage.UpdateConfig([]entity.ConfigUpdateOption{
			scraper.ConnectionConfigUpdateOption(scraperConfig.Text),
		})
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
			err := label.Set(getLabelText(source, true))
			if err != nil {
				logger.Error(fmt.Errorf("failed to change label: %w", err))

				return
			}
		} else {
			scraper.Stop()
			err := label.Set(getLabelText(source, false))
			if err != nil {
				logger.Error(fmt.Errorf("failed to change label: %w", err))

				return
			}
		}
	}
}

func getLabelText(
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
		statusName = "active"
	} else {
		statusName = "stopped"
	}

	return fmt.Sprintf("%s (%s)", serviceName, statusName)
}
