package entity

type ScraperState byte

const (
	ScraperStateRunning = iota + 1
	ScraperStateActive
	ScraperStateStopped
)
