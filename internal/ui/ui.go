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
	app        fyne.App
	mainWindow fyne.Window
	scrapers   map[entity.Source]Scraper
}

type Scraper interface {
	Run(ctx context.Context)
	Stop()
	Status() entity.ScraperState
}

func New(
	scrapers map[entity.Source]Scraper,
) error {
	newUI := &UI{
		app:      app.New(),
		scrapers: scrapers,
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

	const itemsInLine = 2

	switchesContent := make([]fyne.CanvasObject, 0, itemsInLine*(1+len(ui.scrapers)))
	switchesContent = append(
		switchesContent,
		widget.NewLabel("Sources"),
		widget.NewLabel("Switch"),
	)
	for source := range ui.scrapers {
		scraperContent := binding.NewString()
		err = scraperContent.Set(getLabelText(source, false))
		if err != nil {
			return err
		}
		scraperLabel := widget.NewLabelWithData(scraperContent)
		scraperButton := widget.NewButton(
			"Turn",
			ui.turnButtonHandler(
				source,
				scraperContent,
			),
		)
		switchesContent = append(switchesContent, scraperLabel, scraperButton)
	}

	grid := container.New(
		layout.NewGridLayout(itemsInLine),
		switchesContent...,
	)
	mainWindow.SetContent(grid)

	ui.mainWindow = mainWindow

	return nil
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
