package historical

import (
	"crypto-analytics/models/entities"
	"crypto-analytics/utils/databases"
)

type Repository interface {
	Save(crypto entities.Historical) error
	Count() int64
	FetchForSymbolYesterday(symbol string, day string) (entities.Historical, error)
}

type Impl struct {
	db databases.SqlConnection
}
