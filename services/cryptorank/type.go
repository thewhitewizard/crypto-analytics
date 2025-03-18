package cryptorank

import (
	"crypto-analytics/pkg/observer"
	"net/http"
	"time"

	"github.com/patrickmn/go-cache"
)

const (
	cryptorankBaseAPI       = "https://api.cryptorank.io"
	clientHTTPTimeout       = 15 * time.Second
	marketIndicatorCacheKey = "marketIndicatorCacheKey"
)

type MarketIndicator struct {
	FearGreedIndex          int     `json:"fearGreedIndex,omitempty"`
	FearGreedYesterdayIndex int     `json:"fearGreedYesterdayIndex,omitempty"`
	BtcDominance            float64 `json:"btcDominance,omitempty"`
	TotalMarketCap          int64   `json:"totalMarketCap,omitempty"`
}

type FearGreedIndex struct {
	Today     int `json:"today,omitempty"`
	Yesterday int `json:"yesterday,omitempty"`
	LastWeek  int `json:"lastWeek,omitempty"`
	LastMonth int `json:"lastMonth,omitempty"`
}

type GlobalIndicator struct {
	BtcDominanceChangePercent   float64 `json:"btcDominanceChangePercent,omitempty"`
	TotalVolume24H              int64   `json:"totalVolume24h,omitempty"`
	TotalMarketCapChangePercent float64 `json:"totalMarketCapChangePercent,omitempty"`
	TotalVolume24HChangePercent float64 `json:"totalVolume24hChangePercent,omitempty"`
	BtcDominance                float64 `json:"btcDominance,omitempty"`
	TotalMarketCap              int64   `json:"totalMarketCap,omitempty"`
	EthDominance                float64 `json:"ethDominance,omitempty"`
	EthDominanceChangePercent   float64 `json:"ethDominanceChangePercent,omitempty"`
}

type BtcDominance struct {
	Timestamps                []int64   `json:"timestamps,omitempty"`
	Values                    []float64 `json:"values,omitempty"`
	BtcDominanceChangePercent float64   `json:"btcDominanceChangePercent,omitempty"`
}

type Service interface {
	GetMarketIndicator() (MarketIndicator, error)
	RegisterObserver(o observer.Observer)
}

type Impl struct {
	baseURL   string
	client    *http.Client
	cache     *cache.Cache
	observers map[observer.Observer]struct{}
}
