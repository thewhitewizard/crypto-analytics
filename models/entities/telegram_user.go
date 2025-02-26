package entities

type TelegramUser struct {
	ChatID int64  `json:"id" gorm:"primaryKey"`
	Name   string `json:"name,omitempty"`
}
