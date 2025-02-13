package databases

import "gorm.io/gorm"

type SqlConnection interface {
	GetDB() *gorm.DB
	IsConnected() bool
	Run() error
	Shutdown()
}
