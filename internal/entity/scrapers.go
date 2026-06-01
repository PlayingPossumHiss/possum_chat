package entity

type ScraperState byte

const (
	ScraperStateActive ScraperState = iota + 1
	ScraperStateStarting
	ScraperStateStopped
)
