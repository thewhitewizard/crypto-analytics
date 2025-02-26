package observer

type EventType int

const (
	TrendingEvent EventType = 1
	RankingEvent  EventType = 1
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
