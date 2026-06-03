package get_style

import (
	"io"
	"os"

	"github.com/PlayingPossumHiss/possum_chat/internal/entity"
	app_errors "github.com/PlayingPossumHiss/possum_chat/internal/errors"
)

// UseCase получение кастомного стиля, тут все в целом максимально примитивно: берем его из конфига и отдаем
type UseCase struct {
	configStorage ConfigStorage
}

func New(
	configStorage ConfigStorage,
) *UseCase {
	return &UseCase{
		configStorage: configStorage,
	}
}

func (uc *UseCase) GetCustomStyle() string {
	return uc.configStorage.Config().View.CssStyle
}

func (uc *UseCase) GetMainStyle() (string, error) {
	styleKey := uc.configStorage.Config().View.CssMainStyle

	path, err := getStylePath(styleKey)
	if err != nil {
		return "", err
	}

	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		return "", err
	}

	return string(content), nil
}

func getStylePath(style entity.ConfigMainStyle) (string, error) {
	switch style {
	case entity.ConfigMainStyleSimpleBlock:
		return "./static/css/simple_block.css", nil
	case entity.ConfigMainStyleSimpleNoBg:
		return "./static/css/simple_no_bg.css", nil
	}

	return "", app_errors.ErrInvalidConfig
}
