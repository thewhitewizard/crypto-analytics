package coinmarketcap

import (
	"bytes"
	"crypto-analytics/models/constants"
	"crypto-analytics/models/entities"
	"crypto-analytics/pkg/observer"
	communityRepo "crypto-analytics/repositories/community"
	historicalRepo "crypto-analytics/repositories/historical"
	trendingRepo "crypto-analytics/repositories/trending"
	"crypto-analytics/utils/dates"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"time"

	"github.com/go-co-op/gocron/v2"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

func New(scheduler gocron.Scheduler,
	trending trendingRepo.Repository,
	historical historicalRepo.Repository,
	community communityRepo.Repository) (*Impl, error) {
	service := &Impl{
		baseURL: cmcBaseAPI,
		client: &http.Client{
			Timeout: clientHTTPTimeout,
		},
		trendRepo:     trending,
		histoRepo:     historical,
		communityRepo: community,
	}

	if viper.GetBool(constants.Production) {
		service.FetchAndSaveTrendingCrypto()
		if service.communityRepo.Count() == 0 {
			service.fetchAndSaveCommunityData(true)
		} else {
			service.fetchAndSaveCommunityData(false)
		}

	}
	if service.histoRepo.Count() == 0 {
		service.fetchAndSaveHistoricalSinceHalving()
	}

	_, errTrendingJob := scheduler.NewJob(
		gocron.CronJob(viper.GetString(constants.TrendingCryptoCronTab), true),
		gocron.NewTask(func() { service.FetchAndSaveTrendingCrypto() }),
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

	_, errCommunityData := scheduler.NewJob(
		gocron.CronJob("0 * * * *", true),
		gocron.NewTask(func() { service.fetchAndSaveCommunityData(false) }),
		gocron.WithName("Fetch community data"),
	)
	if errCommunityData != nil {
		return nil, errCommunityData
	}

	service.observers = map[observer.Observer]struct{}{}

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

func (service *Impl) fetchAndSaveHistoricalSinceHalving() {
	log.Info().Msg("Start fetching historical crypto since halving")
	from, _ := dates.StringToDate(halvingDate, dates.DateFormat)
	rangeDates := dates.GenerateDatesBetweenTwoDates(from, time.Now())
	var startSteps = []int{1, 201, 401, 601, 801}

	for _, date := range rangeDates {
		log.Info().Time("date", date).Msg("fetching for date")
		for _, start := range startSteps {
			data, err := service.fetchHistoricalPaginate(date.Format(dates.DateFormat), start, limitDataPerCall)
			if err != nil {
				continue
			}
			for _, d := range data.CryptoCurrencies {
				crypto := entities.Historical{ID: d.ID, Rank: d.CmcRank, Slug: d.Slug, Tags: d.KeepOnlyRelevantsTags(), Name: d.Name, Symbol: d.Symbol, Day: date.Format(dates.DateFormat), Price: d.Quotes[0].Price, Marketcap: d.Quotes[0].MaketCap}
				service.histoRepo.Save(crypto)
			}
			time.Sleep(delayBetweenCall)
		}
	}
	log.Info().Msg("End fetching historical crypto since halving")
}

func (service *Impl) fetchWatcherData(cryptoID int) (*LiteResponse, error) {
	endpoint := fmt.Sprintf("%s/data-api/v3/cryptocurrency/detail/lite?id=%v", service.baseURL, cryptoID)
	resp, err := http.Get(endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch data: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status: %d", resp.StatusCode)
	}

	var result LiteResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return &result, nil
}

func (service *Impl) fetchAndSaveCommunityData(first bool) {
	log.Info().Msg("Start fetching community data")
	cryptocurrencies := constants.GetCrytoWatch()
	day := time.Now().Format(dates.DateFormat)
	if first {
		day = time.Now().AddDate(0, 0, -1).Format(dates.DateFormat)
	}
	for _, cryptoccryptocurrency := range cryptocurrencies {
		log.Info().Str("symbol", cryptoccryptocurrency.Symbol).Msg("Fetching community data")
		entity := entities.CommunityData{Cid: cryptoccryptocurrency.CryptoId, Symbol: cryptoccryptocurrency.Symbol, Day: day, Followers: "0", WatchCount: "0"}
		profileData, errProfile := service.fetchProfileData(cryptoccryptocurrency.Handle)
		watchData, errWatch := service.fetchWatcherData(cryptoccryptocurrency.CryptoId)
		if errProfile == nil {
			entity.Followers = profileData.Data.Account.Followers
		}
		if errWatch == nil {
			entity.WatchCount = watchData.Data.WatchCount
		}
		err := service.communityRepo.Save(entity)
		if err != nil {
			log.Error().Err(err).Str("symbol", cryptoccryptocurrency.Symbol).Msg("Fetching community data")
		}
	}
	log.Info().Msg("End fetching community data")
	service.notify(observer.Event{E: observer.RankingEvent})
}

func (service *Impl) fetchProfileData(handle string) (*ProfileResponse, error) {
	endpoint := fmt.Sprintf("%s/gravity/v3/gravity/profile/query", service.baseURL)

	values := map[string]string{"handle": handle}
	jsonValue, err := json.Marshal(values)

	if err != nil {
		return nil, fmt.Errorf("failed to prepared data: %w", err)
	}

	resp, err := http.Post(endpoint, "application/json", bytes.NewBuffer(jsonValue))
	if err != nil {
		return nil, fmt.Errorf("failed to fetch data: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status: %d", resp.StatusCode)
	}

	var result ProfileResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return &result, nil
}

/*
*

	func (service *Impl) fetchVoteData() (*VoteResponse, error) {
		endpoint := fmt.Sprintf("%s/gravity/v3/gravity/crypto/queryVoteResult", service.baseURL)

		values := map[string]int{"cryptoId": 1637}
		jsonValue, err := json.Marshal(values)

		if err != nil {
			return nil, fmt.Errorf("failed to prepared data: %w", err)
		}

		resp, err := http.Post(endpoint, "application/json", bytes.NewBuffer(jsonValue))
		if err != nil {
			return nil, fmt.Errorf("failed to fetch data: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("API request failed with status: %d", resp.StatusCode)
		}

		var result VoteResponse
		err = json.NewDecoder(resp.Body).Decode(&result)
		if err != nil {
			return nil, fmt.Errorf("failed to parse response: %w", err)
		}
		return &result, nil
	}

*
*/
func (service *Impl) fetchAndSaveHistorical() {
	log.Info().Msg("Start fetching historical crypto")
	var startSteps = []int{1, 201, 401, 601, 801}
	yesterday := time.Now().AddDate(0, 0, -1).Format(dates.DateFormat)
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
	service.notify(observer.Event{E: observer.RankingEvent})
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

func (service *Impl) FetchAndSaveTrendingCrypto() {
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

	today := time.Now().Format(dates.DateFormat)
	for _, d := range trendingResponse.Data.CryptoCurrencies {
		crypto := entities.TrendingCrypto{
			ID: d.ID, Slug: d.Slug, Name: d.Name, Symbol: d.Symbol, Day: today,
		}
		service.trendRepo.Save(crypto)
	}
	service.notify(observer.Event{E: observer.TrendingEvent})
	log.Info().Msg("End fetching trending crypto")
}

func (service *Impl) IsCryptoTrendyYersterday(symbol string) bool {
	yesterday := time.Now().AddDate(0, 0, -1).Format(dates.DateFormat)
	v, err := service.trendRepo.IsCryptoTrendyAtDay(symbol, yesterday)
	if err != nil || v.Name == "" {
		return false
	}
	return true
}

func (service *Impl) FetchForSymbolYesterday(symbol string) (entities.Historical, error) {
	yesterday := time.Now().AddDate(0, 0, -1).Format(dates.DateFormat)

	return service.histoRepo.FetchForSymbolForDay(symbol, yesterday)
}

func (service *Impl) FetchForSymbolForTwoDaysAgo(symbol string) (entities.Historical, error) {
	twoDays := time.Now().AddDate(0, 0, -2).Format(dates.DateFormat)

	return service.histoRepo.FetchForSymbolForDay(symbol, twoDays)
}

func (service *Impl) FetchForSymbol7DaysAgo(symbol string) (entities.Historical, error) {
	sevenDaysAgo := time.Now().AddDate(0, 0, -8).Format(dates.DateFormat)

	return service.histoRepo.FetchForSymbolForDay(symbol, sevenDaysAgo)
}

func (service *Impl) IsCryptoTrendyToday(symbol string) bool {
	today := time.Now().Format(dates.DateFormat)
	v, err := service.trendRepo.IsCryptoTrendyAtDay(symbol, today)
	if err != nil || v.Name == "" {
		return false
	}
	return true
}

func (service *Impl) FetchCommunityDataForSymbolYesterday(id int) (entities.CommunityData, error) {
	yesterday := time.Now().AddDate(0, 0, -1).Format(dates.DateFormat)

	return service.communityRepo.FetchForSymbolYesterday(id, yesterday)
}

func (service *Impl) GetTopGainers() ([]Gainer, error) {

	twoDays := time.Now().AddDate(0, 0, -2).Format(dates.DateFormat)
	yesterday := time.Now().AddDate(0, 0, -1).Format(dates.DateFormat)
	yesterdayData, err := service.histoRepo.FetchForDay(yesterday)
	twoDaysData, err2 := service.histoRepo.FetchForDay(twoDays)
	if err != nil {
		log.Error().Err(err).Msg("failed to fetch yesterdayData")
		return nil, err
	}
	if err2 != nil {
		log.Error().Err(err2).Msg("failed to fetch twoDaysData")
		return nil, err2
	}
	twoDaysMap := make(map[string]float64)
	for _, data := range twoDaysData {
		twoDaysMap[data.Symbol] = data.Price
	}

	// Compute percentage change
	var gainers []Gainer
	for _, yData := range yesterdayData {
		oldPrice, exists := twoDaysMap[yData.Symbol]
		if !exists || oldPrice == 0 {
			continue // Skip if no data or invalid price
		}

		priceChange := yData.Price - oldPrice
		percentChange := (priceChange / oldPrice) * 100

		// Consider only positive changes (gainers)
		if percentChange > 0 {
			gainers = append(gainers, Gainer{
				Symbol:        yData.Symbol,
				PriceChange:   priceChange,
				PercentChange: percentChange,
			})
		}
	}

	// Sort gainers by highest percentage increase
	sort.Slice(gainers, func(i, j int) bool {
		return gainers[i].PercentChange > gainers[j].PercentChange
	})

	// Return top 3 gainers
	if len(gainers) > 3 {
		return gainers[:3], nil
	}
	return gainers, nil
}
