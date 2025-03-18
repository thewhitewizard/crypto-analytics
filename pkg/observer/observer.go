package observer

import "github.com/mmcdole/gofeed"

type EventType int

const (
	TrendingEvent        EventType = 1
	RankingEvent         EventType = 2
	PriceEvent           EventType = 3
	MarketIndicatorEvent EventType = 4
	RSSEvent             EventType = 5
)

type Event struct {
	E    EventType
	Feed *gofeed.Item
}

func NewRSSEvent(item *gofeed.Item) Event {
	return Event{Feed: item, E: RSSEvent}
}

type Observer interface {
	OnNotify(Event)
}

type Notifier interface {
	Register(Observer)
	Notify(Event)
}
