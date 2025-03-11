package twitter

import (
	"crypto-analytics/models/entities"
	"crypto-analytics/utils/databases"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

func New(db databases.SqlConnection) *Impl {
	return &Impl{db: db}
}

func (repo *Impl) SaveOrUpdate(tweet entities.Tweet) error {
	var existingTweet entities.Tweet

	result := repo.db.GetDB().Where("id = ?", tweet.ID).First(&existingTweet)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			if err := repo.db.GetDB().Create(&tweet).Error; err != nil {
				return fmt.Errorf("failed to create tweet: %w", err)
			}
		} else {
			return fmt.Errorf("failed to check tweet existence: %w", result.Error)
		}
	} else {
		tweet.Timestamp = existingTweet.Timestamp // To check if necessary
		if err := repo.db.GetDB().Model(&existingTweet).Updates(tweet).Error; err != nil {
			return fmt.Errorf("failed to update tweet: %w", err)
		}
	}

	return nil
}

func (repo *Impl) GetTweetBetweenTimestamps(startTimestamp int64, endTimestamp int64) ([]entities.Tweet, error) {
	var tweets []entities.Tweet

	res := repo.db.GetDB().Find(&tweets).
		Where("timestamp > ?", startTimestamp).
		Where("timestamp < ?", endTimestamp)

	return tweets, res.Error
}

func (repo *Impl) Count() int64 {
	count := new(int64)
	repo.db.GetDB().Model(&entities.Tweet{}).Count(count)

	return *count
}
