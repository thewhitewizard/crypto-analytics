package telegram

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

func (repo *Impl) FetchAll() ([]entities.TelegramUser, error) {
	var users []entities.TelegramUser
	result := repo.db.GetDB().Find(&users)

	return users, result.Error
}

func (repo *Impl) SaveOrUpdate(user entities.TelegramUser) error {
	var existingUser entities.TelegramUser

	result := repo.db.GetDB().Where("chat_id = ?", user.ChatID).First(&existingUser)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			if err := repo.db.GetDB().Create(&user).Error; err != nil {
				return fmt.Errorf("failed to create user: %w", err)
			}
		} else {
			return fmt.Errorf("failed to check tweet existence: %w", result.Error)
		}
	}

	return nil
}

func (repo *Impl) Delete(user entities.TelegramUser) error {
	result := repo.db.GetDB().Delete(&entities.TelegramUser{}, user.ChatID)
	return result.Error
}
