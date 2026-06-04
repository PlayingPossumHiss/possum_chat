package ui

import (
	"fmt"
	"os"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"github.com/PlayingPossumHiss/possum_chat/internal/entity"
)

var connectionTabRows = 3

type UI struct {
	app              fyne.App
	mainWindow       fyne.Window
	configStorage    ConfigStorage
	languageProvider LanguageProvider
	messageQueue     MessageQueue
	scrapers         map[entity.Source]Scraper
}

func New(
	languageProvider LanguageProvider,
	scrapers map[entity.Source]Scraper,
	configStorage ConfigStorage,
	messageQueue MessageQueue,
) error {
	newUI := &UI{
		app:              app.New(),
		scrapers:         scrapers,
		languageProvider: languageProvider,
		configStorage:    configStorage,
		messageQueue:     messageQueue,
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
