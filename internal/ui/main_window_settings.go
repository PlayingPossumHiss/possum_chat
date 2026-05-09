package ui

import (
	"fmt"
	"strconv"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/PlayingPossumHiss/possum_chat/internal/entity"
	"github.com/PlayingPossumHiss/possum_chat/internal/service/logger"
)

func (ui *UI) getSettingsTabContent() *fyne.Container {
	const itemsInLine = 2

	settingsContent := make([]fyne.CanvasObject, 0, itemsInLine*(1+len(ui.scrapers)))
	settingsContent = append(
		settingsContent,
		widget.NewLabel("Key"),
		widget.NewLabel("Value"),
	)

	settingsContent = append(
		settingsContent,
		ui.getTimeToHideSettingsView()...,
	)

	settingsContent = append(
		settingsContent,
		ui.getTimeToDeleteSettingsView()...,
	)

	settingsContent = append(
		settingsContent,
		ui.getPortSettingsView()...,
	)

	grid := container.New(
		layout.NewGridLayout(itemsInLine),
		settingsContent...,
	)

	return grid
}

func (ui *UI) getTimeToHideSettingsView() []fyne.CanvasObject {
	timeToHideField := widget.NewEntry()
	timeToHideField.SetText(strconv.Itoa(int(ui.configStorage.Config().View.TimeToHideMessage.Seconds())))
	ui.bindTimeToHideSettingsViewHandler(timeToHideField)

	return []fyne.CanvasObject{
		widget.NewLabel("Time to hide message (sec)"),
		timeToHideField,
	}
}

func (ui *UI) bindTimeToHideSettingsViewHandler(timeToHideField *widget.Entry) {
	timeToHideField.OnChanged = func(s string) {
		newValue, err := strconv.Atoi(timeToHideField.Text)
		if err != nil {
			logger.Warn(fmt.Sprintf("try to change time to hide to invalid value %s", timeToHideField.Text))
			timeToHideField.SetText(strconv.Itoa(int(ui.configStorage.Config().View.TimeToHideMessage.Seconds())))

			return
		}
		err = ui.configStorage.UpdateConfig([]entity.ConfigUpdateOption{
			func(target *entity.Config) {
				target.View.TimeToHideMessage = time.Second * time.Duration(newValue)
			},
		})
		if err != nil {
			logger.Error(fmt.Errorf("failed to update config: %w", err))
		}
	}
}

func (ui *UI) getTimeToDeleteSettingsView() []fyne.CanvasObject {
	timeToDeleteField := widget.NewEntry()
	timeToDeleteField.SetText(strconv.Itoa(int(ui.configStorage.Config().View.TimeToDeleteMessage.Minutes())))
	ui.bindTimeToDeleteSettingsViewHandler(timeToDeleteField)

	return []fyne.CanvasObject{
		widget.NewLabel("Time to delete message (min.)"),
		timeToDeleteField,
	}
}

func (ui *UI) bindTimeToDeleteSettingsViewHandler(timeToDeleteField *widget.Entry) {
	timeToDeleteField.OnChanged = func(s string) {
		newValue, err := strconv.Atoi(timeToDeleteField.Text)
		if err != nil {
			logger.Warn(fmt.Sprintf("try to change time to delete to invalid value %s", timeToDeleteField.Text))
			timeToDeleteField.SetText(strconv.Itoa(int(ui.configStorage.Config().View.TimeToDeleteMessage.Minutes())))

			return
		}
		err = ui.configStorage.UpdateConfig([]entity.ConfigUpdateOption{
			func(target *entity.Config) {
				target.View.TimeToDeleteMessage = time.Minute * time.Duration(newValue)
			},
		})
		if err != nil {
			logger.Error(fmt.Errorf("failed to update config: %w", err))
		}
	}
}

func (ui *UI) getPortSettingsView() []fyne.CanvasObject {
	portField := widget.NewEntry()
	portField.SetText(strconv.Itoa(ui.configStorage.Config().Port))
	portField.OnChanged = func(s string) {
		newValue, err := strconv.Atoi(portField.Text)
		if err != nil {
			logger.Warn(fmt.Sprintf("try to change port to invalid value %s", portField.Text))
			portField.SetText(strconv.Itoa(ui.configStorage.Config().Port))

			return
		}
		err = ui.configStorage.UpdateConfig([]entity.ConfigUpdateOption{
			func(target *entity.Config) {
				target.Port = newValue
			},
		})
		if err != nil {
			logger.Error(fmt.Errorf("failed to update config: %w", err))
		}
	}

	return []fyne.CanvasObject{
		widget.NewLabel("Port (need restart)"),
		portField,
	}
}

func (ui *UI) getCssTabContent() *widget.Entry {
	cssField := widget.NewEntry()
	cssField.SetText(ui.configStorage.Config().View.CssStyle)
	cssField.MultiLine = true
	cssField.OnChanged = func(s string) {
		err := ui.configStorage.UpdateConfig([]entity.ConfigUpdateOption{
			func(target *entity.Config) {
				target.View.CssStyle = cssField.Text
			},
		})
		if err != nil {
			logger.Error(fmt.Errorf("failed to update config: %w", err))
		}
	}

	return cssField
}
