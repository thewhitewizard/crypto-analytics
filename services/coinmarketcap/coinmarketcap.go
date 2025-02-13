package coinmarketcap

import (
	"crypto-analytics/models/constants"
	"crypto-analytics/models/entities"
	historicalRepo "crypto-analytics/repositories/historical"
	trendingRepo "crypto-analytics/repositories/trending"
	"crypto-analytics/utils/dates"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-co-op/gocron/v2"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

func New(scheduler gocron.Scheduler,
	trending trendingRepo.Repository,
	historical historicalRepo.Repository) (*Impl, error) {
	service := &Impl{
		baseURL: cmcBaseAPI,
		client: &http.Client{
			Timeout: clientHTTPTimeout,
		},
		trendRepo: trending,
		histoRepo: historical,
	}

	service.fetchAndSaveTrendingCrypto()

	if service.histoRepo.Count() == 0 {
		service.fetchAndSaveHistoricalSinceHalving()
	}

	_, errTrendingJob := scheduler.NewJob(
		gocron.CronJob(viper.GetString(constants.TrendingCryptoCronTab), true),
		gocron.NewTask(func() { service.fetchAndSaveTrendingCrypto() }),
		gocron.WithName("Fetch trending crypto"),
	)
	if errTrendingJob != nil {
		return nil, errTrendingJob
	}

	_, errHistoricalJob := scheduler.NewJob(
		gocron.CronJob(viper.GetString(constants.HistoricalCryptoCronTab), true),
		gocron.NewTask(func() { service.fetchAndSaveHistorical() }),
		gocron.WithName("Fetch historical crypto"),
	)
	if errHistoricalJob != nil {
		return nil, errHistoricalJob
	}

	return service, nil
}

func (service *Impl) fetchAndSaveHistoricalSinceHalving() {
	log.Info().Msg("Start fetching historical crypto since halving")
	from, _ := dates.StringToDate(halvingDate, dateFormat)
	rangeDates := dates.GenerateDatesBetweenTwoDates(from, time.Now())
	var startSteps = []int{1, 201, 401, 601, 801}

	for _, date := range rangeDates {
		log.Info().Time("date", date).Msg("fetching for date")
		for _, start := range startSteps {
			data, err := service.fetchHistoricalPaginate(date.Format(dateFormat), start, limitDataPerCall)
			if err != nil {
				continue
			}
			for _, d := range data.CryptoCurrencies {
				crypto := entities.Historical{ID: d.ID, Rank: d.CmcRank, Slug: d.Slug, Tags: d.KeepOnlyRelevantsTags(), Name: d.Name, Symbol: d.Symbol, Day: date.Format(dateFormat), Price: d.Quotes[0].Price, Marketcap: d.Quotes[0].MaketCap}
				service.histoRepo.Save(crypto)
			}
			time.Sleep(delayBetweenCall)
		}
	}
	log.Info().Msg("End fetching historical crypto since halving")
}

func (service *Impl) fetchAndSaveHistorical() {
	log.Info().Msg("Start fetching historical crypto")
	var startSteps = []int{1, 201, 401, 601, 801}
	yesterday := time.Now().AddDate(0, 0, -1).Format(dateFormat)
	for _, start := range startSteps {
		data, err := service.fetchHistoricalPaginate(yesterday, start, limitDataPerCall)
		if err != nil {
			continue
		}
		for _, d := range data.CryptoCurrencies {
			crypto := entities.Historical{ID: d.ID, Rank: d.CmcRank, Tags: d.KeepOnlyRelevantsTags(), Slug: d.Slug, Name: d.Name, Symbol: d.Symbol, Day: yesterday, Price: d.Quotes[0].Price, Marketcap: d.Quotes[0].MaketCap}
			service.histoRepo.Save(crypto)
		}
		time.Sleep(delayBetweenCall)
	}
	log.Info().Msg("End fetching historical crypto")
}

func (service *Impl) fetchHistoricalPaginate(date string, start int, limit int) (*HistoricalResponse, error) {
	endpoint := fmt.Sprintf("%s/data-api/v3/cryptocurrency/listings/historical", service.baseURL)

	url := fmt.Sprintf("%s?convertId=%s&date=%s&limit=%d&start=%d",
		endpoint, convertIDs, date, limit, start)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch data: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status: %d", resp.StatusCode)
	}

	var result HistoricalResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return &result, nil
}

func (service *Impl) fetchAndSaveTrendingCrypto() {
	log.Info().Msg("Start fetching trending crypto")

	url := fmt.Sprintf("%s/data-api/v3/cryptocurrency/listing?start=1&limit=50&sortBy=trending_24h&sortType=desc&cryptoType=all&tagType=all&audited=false", service.baseURL)
	resp, err := http.Get(url)
	if err != nil {
		log.Error().Err(err).Msg("failed to make API request")
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Error().Msgf("API request failed with status: %d", resp.StatusCode)
		return
	}

	var trendingResponse TrendingResponse
	err = json.NewDecoder(resp.Body).Decode(&trendingResponse)
	if err != nil {
		log.Error().Err(err).Msg("failed to decode JSON response")
		return
	}

	today := time.Now().Format(dateFormat)
	for _, d := range trendingResponse.Data.CryptoCurrencies {
		crypto := entities.TrendingCrypto{
			ID: d.ID, Slug: d.Slug, Name: d.Name, Symbol: d.Symbol, Day: today,
		}
		service.trendRepo.Save(crypto)
	}

	log.Info().Msg("End fetching trending crypto")
}
