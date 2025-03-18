package feedsources

import (
	"crypto-analytics/models/entities"
	"crypto-analytics/utils/databases"
)

type Repository interface {
	GetFeedSources() ([]entities.FeedSource, error)
	Create(feedSource entities.FeedSource) error
	Save(feedSource entities.FeedSource) error
	Count() int64
}

type Impl struct {
	db databases.SqlConnection
}
