package coinmarketcap

import (
	historicalRepo "crypto-analytics/repositories/historical"
	trendingRepo "crypto-analytics/repositories/trending"
	"net/http"
	"strings"
	"time"
)

const (
	cmcBaseAPI        = "https://api.coinmarketcap.com"
	convertIDs        = "2781,1"
	dateFormat        = "2006-01-02"
	halvingDate       = "2024-04-19"
	delayBetweenCall  = 2 * time.Second
	clientHTTPTimeout = 15 * time.Second
	limitDataPerCall  = 200
)

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
}

type Impl struct {
	baseURL   string
	client    *http.Client
	trendRepo trendingRepo.Repository
	histoRepo historicalRepo.Repository
}
