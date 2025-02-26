package telegram

import (
	"crypto-analytics/models/entities"
	"crypto-analytics/utils/databases"
)

type Repository interface {
	SaveOrUpdate(user entities.TelegramUser) error
	Delete(user entities.TelegramUser) error
	FetchAll() ([]entities.TelegramUser, error)
}

type Impl struct {
	db databases.SqlConnection
}
