package feeds

import (
	"crypto-analytics/pkg/observer"
	"crypto-analytics/repositories/feedsources"
	"time"

	"github.com/mmcdole/gofeed"
)

type Service interface {
	RegisterObserver(o observer.Observer)
	FetchFeeds() error
}

type Impl struct {
	feedParser     *gofeed.Parser
	timeout        time.Duration
	feedSourceRepo feedsources.Repository
	observers      map[observer.Observer]struct{}
}
