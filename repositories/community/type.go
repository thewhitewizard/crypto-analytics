package community

import (
	"crypto-analytics/models/entities"
	"crypto-analytics/utils/databases"
)

type Repository interface {
	Save(crypto entities.CommunityData) error
	FetchForSymbolYesterday(id int, day string) (entities.CommunityData, error)
	Count() int64
}

type Impl struct {
	db databases.SqlConnection
}
