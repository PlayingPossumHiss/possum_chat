package ui

import (
	"context"
	"fmt"
	"log"

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
) {
	ui := &UI{
		app:      app.New(),
		scrapers: scrapers,
	}

	ui.newMainWindow()

	ui.mainWindow.ShowAndRun()
}

func (ui *UI) newMainWindow() {
	mainWindow := ui.app.NewWindow("Possum Chat")

	youtubeLabel := widget.NewLabel("Youtube")
	youtubeButton := widget.NewButton("Run", func() {
		log.Println("tapped")
	})

	vkLabelContent := binding.NewString()
	vkLabelContent.Set(getLabelText(entity.SourceVkPlayLive, false))
	vkLabel := widget.NewLabelWithData(vkLabelContent)
	vkButton := widget.NewButton(
		"Turn",
		ui.turnButtonHandler(
			entity.SourceVkPlayLive,
			vkLabelContent,
		),
	)

	twitchContent := binding.NewString()
	twitchContent.Set(getLabelText(entity.SourceTwitch, false))
	twitchLabel := widget.NewLabelWithData(twitchContent)
	twitchButton := widget.NewButton(
		"Turn",
		ui.turnButtonHandler(
			entity.SourceTwitch,
			twitchContent,
		),
	)

	donationAlertsLabel := widget.NewLabel("Donation Alerts")
	donationAlertsButton := widget.NewButton("Run", func() {
		log.Println("tapped")
	})

	grid := container.New(
		layout.NewGridLayout(2),
		youtubeLabel, youtubeButton,
		vkLabel, vkButton,
		twitchLabel, twitchButton,
		donationAlertsLabel, donationAlertsButton,
	)
	mainWindow.SetContent(grid)

	ui.mainWindow = mainWindow
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
			label.Set(getLabelText(source, true))
		} else {
			scraper.Stop()
			label.Set(getLabelText(source, false))
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
