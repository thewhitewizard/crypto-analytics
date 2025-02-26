package trending

import (
	"crypto-analytics/models/entities"
	"crypto-analytics/utils/databases"
)

func New(db databases.SqlConnection) *Impl {
	return &Impl{db: db}
}

func (repo *Impl) Save(crypto entities.TrendingCrypto) error {
	return repo.db.GetDB().Save(&crypto).Error
}

func (repo *Impl) Count() int64 {
	count := new(int64)
	repo.db.GetDB().Model(&entities.TrendingCrypto{}).Count(count)

	return *count
}

func (repo *Impl) IsCryptoTrendyAtDay(symbol string, day string) (entities.TrendingCrypto, error) {
	var existingTrendingCrypto entities.TrendingCrypto
	result := repo.db.GetDB().Where("symbol = ?", symbol).Where("day = ?", day).First(&existingTrendingCrypto)

	return existingTrendingCrypto, result.Error
}
