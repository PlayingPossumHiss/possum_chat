package ui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/PlayingPossumHiss/possum_chat/internal/entity"
	app_errors "github.com/PlayingPossumHiss/possum_chat/internal/errors"
	"github.com/PlayingPossumHiss/possum_chat/internal/service/logger"
)

func (ui *UI) getCssTabContent() *fyne.Container {
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

	mainStyleRow := container.New(
		layout.NewGridLayout(connectionTabRows),
		ui.getMainStyleSettingsView()...,
	)

	content := container.New(
		layout.NewVBoxLayout(),
		mainStyleRow,
		cssField,
	)

	return content
}

func (ui *UI) getMainStyleSettingsView() []fyne.CanvasObject {
	mainStyleField := widget.NewSelectEntry([]string{
		configMainStyleSimpleBlock,
		configMainStyleSimpleNoBg,
	})
	mainStyleField.SetText(mainStyleFromConsig(ui.configStorage.Config().View.CssMainStyle))
	mainStyleField.OnChanged = func(s string) {
		newValue, err := mainStyleToConfig(s)
		if err != nil {
			logger.Warn(fmt.Sprintf("try to change lang to invalid value %s", mainStyleField.Text))
			mainStyleField.SetText(mainStyleFromConsig(ui.configStorage.Config().View.CssMainStyle))

			return
		}
		err = ui.configStorage.UpdateConfig([]entity.ConfigUpdateOption{
			func(target *entity.Config) {
				target.View.CssMainStyle = newValue
			},
		})
		if err != nil {
			logger.Error(fmt.Errorf("failed to update config: %w", err))
		}
	}

	testMessageButton := widget.NewButton(
		ui.languageProvider.Local(entity.LanguageTextConstantTestMessageButton),
		ui.sendTestMessage,
	)

	return []fyne.CanvasObject{
		widget.NewLabel(ui.languageProvider.Local(entity.LanguageTextConstantMainStyleField)),
		mainStyleField,
		testMessageButton,
	}
}

const (
	configMainStyleSimpleBlock = "simple_block"
	configMainStyleSimpleNoBg  = "simple_no_bg"
)

func mainStyleToConfig(src string) (entity.ConfigMainStyle, error) {
	switch src {
	case configMainStyleSimpleBlock:
		return entity.ConfigMainStyleSimpleBlock, nil
	case configMainStyleSimpleNoBg:
		return entity.ConfigMainStyleSimpleNoBg, nil
	}

	return 0, app_errors.ErrInvalidConfig
}

func mainStyleFromConsig(src entity.ConfigMainStyle) string {
	if src == entity.ConfigMainStyleSimpleNoBg {
		return configMainStyleSimpleNoBg
	}

	return configMainStyleSimpleBlock
}

func (ui *UI) sendTestMessage() {
	content := ui.languageProvider.Local(entity.LanguageTextConstantTestMessageContent)
	ui.sendTestMessagesUseCase.SendTestMessages(content)
}
