package application

import (
	"crypto-analytics/models/constants"
	"crypto-analytics/models/entities"
	historicalRepo "crypto-analytics/repositories/historical"
	trendingRepo "crypto-analytics/repositories/trending"
	twitterRepo "crypto-analytics/repositories/twitter"
	"crypto-analytics/services/coinmarketcap"
	"crypto-analytics/services/health"
	"crypto-analytics/services/twitter"
	databases "crypto-analytics/utils/databases"
	"crypto-analytics/utils/insights"
	"time"

	"github.com/go-co-op/gocron/v2"
	"github.com/rs/zerolog/log"
)

func New() (*Impl, error) {
	db := databases.New()
	if errDB := db.Run(); errDB != nil {
		return nil, errDB
	}

	errMigration := db.GetDB().AutoMigrate(&entities.Historical{}, &entities.TrendingCrypto{}, &entities.Tweet{})
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

	twitterService, errTwitter := twitter.New(scheduler, twitterRepo, constants.GetTwitterAccounts())
	if errTwitter != nil {
		return nil, errTwitter
	}
	coinmarketcapService, errCMC := coinmarketcap.New(scheduler, trendRepo, histoRepo)
	if errCMC != nil {
		return nil, errCMC
	}
	healthService, errHealthService := health.New(scheduler)
	if errHealthService != nil {
		return nil, errHealthService
	}

	return &Impl{
		scheduler:            scheduler,
		healthService:        healthService,
		probes:               probes,
		coinmarketcapServoce: coinmarketcapService,
		twitterService:       twitterService,
	}, nil
}

func (app *Impl) Run() {
	app.scheduler.Start()
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
