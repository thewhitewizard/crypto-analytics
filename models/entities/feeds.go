package entities

import "time"

type FeedSource struct {
	FeedTypeID string `gorm:"primaryKey"`
	URL        string
	LastUpdate time.Time `gorm:"not null; default:current_timestamp"`
}
