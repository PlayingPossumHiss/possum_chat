package entity

type ScraperState byte

const (
	ScraperStateActive ScraperState = iota + 1
	ScraperStateStarting
	ScraperStateStopped
)

type VkStreamData struct {
	MessageCh chan Message
	Online    chan int64
	Error     chan error
}
