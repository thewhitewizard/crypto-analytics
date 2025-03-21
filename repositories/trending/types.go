package trending

import (
	"crypto-analytics/models/entities"
	"crypto-analytics/utils/databases"
)

type Repository interface {
	Save(crypto entities.TrendingCrypto) error
	Count() int64
	IsCryptoTrendyAtDay(symbol string, day string) (entities.TrendingCrypto, error)
}

type Impl struct {
	db databases.SqlConnection
}
