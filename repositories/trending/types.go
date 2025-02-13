package trending

import (
	"crypto-analytics/models/entities"
	"crypto-analytics/utils/databases"
)

type Repository interface {
	Save(crypto entities.TrendingCrypto) error
	Count() int64
}

type Impl struct {
	db databases.SqlConnection
}
