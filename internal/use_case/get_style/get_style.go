package get_style

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

func (uc *UseCase) GetStyle() string {
	return uc.configStorage.Config().View.CssStyle
}
