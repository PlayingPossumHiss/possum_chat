package get_style

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
