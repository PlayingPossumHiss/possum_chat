package get_online

import "github.com/PlayingPossumHiss/possum_chat/internal/entity"

type Scraper interface {
	GetOnline() int64
	Status() entity.ScraperState
}

type OnlineGetter struct {
	scrapers map[entity.Source]Scraper
}

func New(
	scrapers map[entity.Source]Scraper,
) *OnlineGetter {
	return &OnlineGetter{
		scrapers: scrapers,
	}
}

func (og *OnlineGetter) GetOnline() map[entity.Source]int64 {
	result := make(map[entity.Source]int64)
	for source, scraper := range og.scrapers {
		if scraper.Status() != entity.ScraperStateActive {
			continue
		}

		result[source] = scraper.GetOnline()
	}

	return result
}
