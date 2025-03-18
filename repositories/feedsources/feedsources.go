package feedsources

import (
	"crypto-analytics/models/entities"
	"crypto-analytics/utils/databases"
)

func New(db databases.SqlConnection) *Impl {
	return &Impl{db: db}
}

func (repo *Impl) GetFeedSources() ([]entities.FeedSource, error) {
	var feedSources []entities.FeedSource
	response := repo.db.GetDB().Model(&entities.FeedSource{}).Find(&feedSources)
	return feedSources, response.Error
}

func (repo *Impl) Create(feedSource entities.FeedSource) error {
	return repo.db.GetDB().Create(&feedSource).Error
}

func (repo *Impl) Save(feedSource entities.FeedSource) error {
	return repo.db.GetDB().
		Model(&feedSource).
		Where("feed_type_id = ? ",
			feedSource.FeedTypeID).
		Update("last_update", feedSource.LastUpdate).
		Error
}

func (repo *Impl) Count() int64 {
	count := new(int64)
	repo.db.GetDB().Model(&entities.FeedSource{}).Count(count)

	return *count
}
