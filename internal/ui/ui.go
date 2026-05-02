package ui

import (
	"context"
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
	vkLabelContent.Set("VK Play Live (stopped)")
	vkLabel := widget.NewLabelWithData(vkLabelContent)
	vkButton := widget.NewButton("Turn", func() {
		scraper, ok := ui.scrapers[entity.SourceVkPlayLive]
		if !ok {
			logger.Error("can't find VK Play Live scraper")
			return
		}
		if scraper.Status() == entity.ScraperStateStopped {
			scraper.Run(context.Background())
			vkLabelContent.Set("VK Play Live (active)")
		} else {
			scraper.Stop()
			vkLabelContent.Set("VK Play Live (stopped)")
		}
	})

	twitchLabel := widget.NewLabel("Twitch")
	twitchButton := widget.NewButton("Run", func() {
		log.Println("tapped")
	})

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
