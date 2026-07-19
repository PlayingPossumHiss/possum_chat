package get_online

import "github.com/PlayingPossumHiss/possum_chat/internal/entity"

type Scraper interface {
	GetOnline() int64
	Status() entity.ScraperState
}

type ConfigStorage interface {
	Config() entity.Config
}

type OnlineGetter struct {
	scrapers      map[entity.Source]Scraper
	configStorage ConfigStorage
}

func New(
	scrapers map[entity.Source]Scraper,
	configStorage ConfigStorage,
) *OnlineGetter {
	return &OnlineGetter{
		scrapers:      scrapers,
		configStorage: configStorage,
	}
}

func (og *OnlineGetter) GetOnline() map[entity.Source]int64 {
	result := make(map[entity.Source]int64)
	if !og.configStorage.Config().View.ShowUserCount {
		return result
	}

	for source, scraper := range og.scrapers {
		if scraper.Status() != entity.ScraperStateActive {
			continue
		}

		result[source] = scraper.GetOnline()
	}

	return result
}
