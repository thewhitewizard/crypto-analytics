package coinmarketcap

import (
	"crypto-analytics/models/entities"
	"crypto-analytics/pkg/observer"
	communityRepo "crypto-analytics/repositories/community"
	historicalRepo "crypto-analytics/repositories/historical"
	trendingRepo "crypto-analytics/repositories/trending"
	"net/http"
	"strings"
	"time"
)

const (
	cmcBaseAPI        = "https://api.coinmarketcap.com"
	convertIDs        = "2781,1"
	halvingDate       = "2024-04-19"
	delayBetweenCall  = 2 * time.Second
	clientHTTPTimeout = 15 * time.Second
	limitDataPerCall  = 200
)

type ProfileResponse struct {
	Data ProfileData `json:"data"`
}

type ProfileData struct {
	Account GravityAccount `json:"gravityAccount"`
}

type GravityAccount struct {
	Handle    string `json:"handle"`
	Followers string `json:"followers"`
}

/*
*

	type VoteResponse struct {
		Data VoteData `json:"data"`
	}

	type VoteData struct {
		Bullish int32 `json:"id"`
		Bearish int32 `json:"watchCount"`
	}

*
*/
type LiteResponse struct {
	Data LiteData `json:"data"`
}

type LiteData struct {
	ID         int    `json:"id"`
	WatchCount string `json:"watchCount"`
}

type HistoricalResponse struct {
	CryptoCurrencies []CryptoCurrency `json:"data"`
}
type TrendingResponse struct {
	Data Data `json:"data"`
}

type Data struct {
	CryptoCurrencies []CryptoCurrency `json:"cryptoCurrencyList,omitempty"`
}

type Quotes struct {
	Price    float64 `json:"price"`
	MaketCap float64 `json:"marketCap"`
}

type CryptoCurrency struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Symbol      string    `json:"symbol"`
	Slug        string    `json:"slug"`
	CmcRank     int       `json:"cmcRank"`
	LastUpdated time.Time `json:"lastUpdated"`
	Quotes      []Quotes  `json:"quotes"`
	Tags        []string  `json:"tags"`
}

type Gainer struct {
	Symbol        string
	PriceChange   float64
	PercentChange float64
}

func (c *CryptoCurrency) KeepOnlyRelevantsTags() string {
	tags := ""
	if len(c.Tags) == 0 {
		return tags
	}

	for _, t := range c.Tags {
		search := strings.ToLower(t)
		if strings.Contains(search, "ai-") || strings.Contains(search, "-ai") || strings.Contains(search, "depin") || strings.Contains(search, "distributed-computing") {
			if len(tags) > 0 {
				tags += ";"
			}
			tags += search
		}
	}

	return tags

}

type Service interface {
	IsCryptoTrendyToday(symbol string) bool
	IsCryptoTrendyYersterday(symbol string) bool
	FetchForSymbolYesterday(symbol string) (entities.Historical, error)
	FetchForSymbol7DaysAgo(symbol string) (entities.Historical, error)
	FetchForSymbolForTwoDaysAgo(symbol string) (entities.Historical, error)
	FetchCommunityDataForSymbolYesterday(id int) (entities.CommunityData, error)
	FetchAndSaveTrendingCrypto()
	GetTopGainers() ([]Gainer, error)
	RegisterObserver(o observer.Observer)
}

type Impl struct {
	baseURL       string
	client        *http.Client
	trendRepo     trendingRepo.Repository
	histoRepo     historicalRepo.Repository
	communityRepo communityRepo.Repository
	observers     map[observer.Observer]struct{}
}
