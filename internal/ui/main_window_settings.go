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
	app_errors "github.com/PlayingPossumHiss/possum_chat/internal/errors"
	"github.com/PlayingPossumHiss/possum_chat/internal/service/logger"
)

func (ui *UI) getSettingsTabContent() *fyne.Container {
	const itemsInLine = 2

	settingsContent := make([]fyne.CanvasObject, 0, itemsInLine*len(ui.scrapers))

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

	settingsContent = append(
		settingsContent,
		ui.getLangSettingsView()...,
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
		widget.NewLabel(ui.languageProvider.Local(entity.LanguageTextConstantSettingsTimeToHideMessage)),
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
		widget.NewLabel(ui.languageProvider.Local(entity.LanguageTextConstantSettingsTimeToDeleteMessage)),
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
		widget.NewLabel(ui.languageProvider.Local(entity.LanguageTextConstantSettingsPort)),
		portField,
	}
}

func (ui *UI) getLangSettingsView() []fyne.CanvasObject {
	langField := widget.NewSelectEntry([]string{"en", "ru"})
	langField.SetText(langFromEntity(ui.configStorage.Config().UI.Lang))
	langField.OnChanged = func(s string) {
		newValue, err := langToEntity(s)
		if err != nil {
			logger.Warn(fmt.Sprintf("try to change lang to invalid value %s", langField.Text))
			langField.SetText(langFromEntity(ui.configStorage.Config().UI.Lang))

			return
		}
		err = ui.configStorage.UpdateConfig([]entity.ConfigUpdateOption{
			func(target *entity.Config) {
				target.UI.Lang = newValue
			},
		})
		if err != nil {
			logger.Error(fmt.Errorf("failed to update config: %w", err))
		}
	}

	return []fyne.CanvasObject{
		widget.NewLabel(ui.languageProvider.Local(entity.LanguageTextConstantSettingsLang)),
		langField,
	}
}

func langToEntity(src string) (entity.ConfigLang, error) {
	switch src {
	case "en":
		return entity.ConfigLangEn, nil
	case "ru":
		return entity.ConfigLangRu, nil
	}

	return 0, app_errors.ErrInvalidConfig
}

func langFromEntity(src entity.ConfigLang) string {
	if src == entity.ConfigLangRu {
		return "ru"
	}

	return "en"
}

func (ui *UI) getCssTabContent() *widget.Entry {
	cssField := widget.NewMultiLineEntry()
	cssField.SetText(ui.configStorage.Config().View.CssStyle)
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
