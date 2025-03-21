package historical

import (
	"crypto-analytics/models/entities"
	"crypto-analytics/utils/databases"
)

func New(db databases.SqlConnection) *Impl {
	return &Impl{db: db}
}

func (repo *Impl) Save(crypto entities.Historical) error {
	return repo.db.GetDB().Save(&crypto).Error
}

func (repo *Impl) Count() int64 {
	count := new(int64)
	repo.db.GetDB().Model(&entities.Historical{}).Count(count)

	return *count
}

func (repo *Impl) FetchForSymbolForDay(symbol string, day string) (entities.Historical, error) {
	var existingHistorical entities.Historical
	result := repo.db.GetDB().Where("symbol = ?", symbol).Where("day = ?", day).First(&existingHistorical)

	return existingHistorical, result.Error
}

func (repo *Impl) FetchForDay(day string) ([]entities.Historical, error) {
	var existingHistorical []entities.Historical
	result := repo.db.GetDB().Where("day = ?", day).First(&existingHistorical)

	return existingHistorical, result.Error
}
