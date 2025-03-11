package community

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

func (repo *Impl) Save(crypto entities.CommunityData) error {

	var existing entities.CommunityData

	result := repo.db.GetDB().Where("cid = ?", crypto.Cid).Where("day = ?", crypto.Day).First(&existing)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			if err := repo.db.GetDB().Create(&crypto).Error; err != nil {
				return fmt.Errorf("failed to create CommunityData: %w", err)
			}
		} else {
			return fmt.Errorf("failed to check CommunityData existence: %w", result.Error)
		}
	} else {
		if err := repo.db.GetDB().Model(&existing).Updates(crypto).Error; err != nil {
			return fmt.Errorf("failed to update CommunityData: %w", err)
		}
	}

	return repo.db.GetDB().Save(&crypto).Error
}

func (repo *Impl) Count() int64 {
	count := new(int64)
	repo.db.GetDB().Model(&entities.CommunityData{}).Count(count)

	return *count
}

func (repo *Impl) FetchForSymbolYesterday(id int, day string) (entities.CommunityData, error) {
	var existing entities.CommunityData
	result := repo.db.GetDB().Where("cid = ?", id).Where("day = ?", day).First(&existing)

	return existing, result.Error
}
