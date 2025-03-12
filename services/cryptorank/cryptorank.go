package cryptorank

import (
	"crypto-analytics/pkg/observer"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-co-op/gocron/v2"
	"github.com/patrickmn/go-cache"
	"github.com/rs/zerolog/log"
)

func New(scheduler gocron.Scheduler) (*Impl, error) {
	service := &Impl{
		baseURL: cryptorankBaseAPI,
		client: &http.Client{
			Timeout: clientHTTPTimeout,
		},
		cache: cache.New(20*time.Minute, 1*time.Hour),
	}

	_, errJob := scheduler.NewJob(
		gocron.CronJob("*/15 * * * *", true),
		gocron.NewTask(func() { service.fetchAndCache() }),
		gocron.WithName("Fetch market indicator"),
	)
	if errJob != nil {
		return nil, errJob
	}
	service.observers = map[observer.Observer]struct{}{}
	service.fetchAndCache()

	return service, nil
}

func (service *Impl) RegisterObserver(o observer.Observer) {
	service.observers[o] = struct{}{}
}

func (service *Impl) notify(e observer.Event) {
	for o := range service.observers {
		o.OnNotify(e)
	}
}

func (service *Impl) GetMarketIndicator() (MarketIndicator, error) {

	var marketIndicator MarketIndicator
	if x, found := service.cache.Get(marketIndicatorCacheKey); found {
		marketIndicator = x.(MarketIndicator)
	} else {
		return marketIndicator, fmt.Errorf("failed to fetch market indicator from cache")
	}
	return marketIndicator, nil
}

func (service *Impl) fetchAndCache() {

	index, err := service.fetchFearAndGreed()
	dominance, errD := service.fetchDominance()

	if err == nil && index != nil && index.Today > 0 && errD == nil && dominance != nil && len(dominance.Values) > 0 {
		log.Info().Msg("Put market indicator in cache")
		indicator := MarketIndicator{FearGreedIndex: index.Today, FearGreedYesterdayIndex: index.Yesterday, BtcDominance: dominance.Values[len(dominance.Values)-1]}
		service.cache.SetDefault(marketIndicatorCacheKey, indicator)
		service.notify(observer.Event{E: observer.MarketIndicatorEvent})
	} else {
		log.Error().Err(err).Msg("market indicator")
		log.Error().Err(errD).Msg("market indicator")
	}
}

func (service *Impl) fetchFearAndGreed() (*FearGreedIndex, error) {
	log.Info().Msg("Start fetching fear and gred index")

	endpoint := fmt.Sprintf("%s/v0/widgets/fear-and-greed-index", service.baseURL)
	resp, err := http.Get(endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch data: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status: %d", resp.StatusCode)
	}

	var result FearGreedIndex
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return &result, nil
}

func (service *Impl) fetchDominance() (*BtcDominance, error) {
	log.Info().Msg("Start fetching BTC dominance")

	endpoint := fmt.Sprintf("%s/v0/widgets/btc-dominance-chart?period=24H", service.baseURL)
	resp, err := http.Get(endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch data: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status: %d", resp.StatusCode)
	}

	var result BtcDominance

	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return &result, nil
}
