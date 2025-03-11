package twitter

import (
	"crypto-analytics/models/entities"
	"crypto-analytics/utils/databases"
)

type Repository interface {
	SaveOrUpdate(tweet entities.Tweet) error
	GetTweetBetweenTimestamps(startTimestamp int64, endTimestamp int64) ([]entities.Tweet, error)
	Count() int64
}

type Impl struct {
	db databases.SqlConnection
}
