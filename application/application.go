package application

import (
	"crypto-analytics/models/constants"
	"crypto-analytics/models/entities"
	communityRepo "crypto-analytics/repositories/community"
	historicalRepo "crypto-analytics/repositories/historical"
	telegramRepo "crypto-analytics/repositories/telegram"
	trendingRepo "crypto-analytics/repositories/trending"
	twitterRepo "crypto-analytics/repositories/twitter"
	"crypto-analytics/services/coinmarketcap"
	"crypto-analytics/services/telegram"

	"crypto-analytics/services/twitter"
	databases "crypto-analytics/utils/databases"
	"crypto-analytics/utils/insights"
	"time"

	"github.com/go-co-op/gocron/v2"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

func New() (*Impl, error) {
	db := databases.New()
	if errDB := db.Run(); errDB != nil {
		return nil, errDB
	}

	errMigration := db.GetDB().AutoMigrate(&entities.CommunityData{}, &entities.TelegramUser{}, &entities.Historical{}, &entities.TrendingCrypto{}, &entities.Tweet{})
	if errMigration != nil {
		return nil, errMigration
	}

	probes := insights.NewProbes(db.IsConnected)
	frenchLocation, err := time.LoadLocation(constants.FrenchTimezone)
	if err != nil {
		return nil, err
	}

	scheduler, errScheduler := gocron.NewScheduler(gocron.WithLocation(frenchLocation))
	if errScheduler != nil {
		return nil, errScheduler
	}

	// Repositories
	histoRepo := historicalRepo.New(db)
	trendRepo := trendingRepo.New(db)
	twitterRepo := twitterRepo.New(db)
	telegramRepo := telegramRepo.New(db)
	communityRepo := communityRepo.New(db)

	twitterService, errTwitter := twitter.New(scheduler, twitterRepo, constants.GetTwitterAccounts())
	if errTwitter != nil {
		return nil, errTwitter
	}
	coinmarketcapService, errCMC := coinmarketcap.New(scheduler, trendRepo, histoRepo, communityRepo)
	if errCMC != nil {
		return nil, errCMC
	}

	telegramService, errTg := telegram.New(scheduler, viper.GetString(constants.TelegramBotToken), telegramRepo, coinmarketcapService, twitterService)
	if errTg != nil {
		return nil, errTg
	}

	coinmarketcapService.RegisterObserver(telegramService)

	return &Impl{
		scheduler:            scheduler,
		probes:               probes,
		coinmarketcapService: coinmarketcapService,
		telegramService:      telegramService,
		twitterService:       twitterService,
		db:                   db,
	}, nil
}

func (app *Impl) Run() {
	app.scheduler.Start()
	go app.telegramService.ListenAndDispatch()
	for _, job := range app.scheduler.Jobs() {
		scheduledTime, err := job.NextRun()
		if err == nil {
			log.Info().Msgf("%v scheduled at %v", job.Name(), scheduledTime)
		}
	}

	app.probes.ListenAndServe()
}

func (app *Impl) Shutdown() {
	if err := app.scheduler.Shutdown(); err != nil {
		log.Error().Err(err).Msg("Cannot shutdown scheduler, continuing...")
	}
	app.db.Shutdown()
	log.Info().Msgf("Application is no longer running")
}
