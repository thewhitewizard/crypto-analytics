package observer

type EventType int

const (
	TrendingEvent EventType = 1
	RankingEvent  EventType = 2
	PriceEvent    EventType = 3
)

type Event struct {
	E EventType
}

type Observer interface {
	OnNotify(Event)
}

type Notifier interface {
	Register(Observer)
	Notify(Event)
}
