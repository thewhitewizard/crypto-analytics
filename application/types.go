package application

import (
	cmcService "crypto-analytics/services/coinmarketcap"
	"crypto-analytics/services/health"
	"crypto-analytics/services/twitter"
	databases "crypto-analytics/utils/databases"
	"crypto-analytics/utils/insights"

	"github.com/go-co-op/gocron/v2"
)

type Application interface {
	Run()
	Shutdown()
}

type Impl struct {
	scheduler            gocron.Scheduler
	healthService        health.Service
	coinmarketcapServoce cmcService.Service
	twitterService       twitter.Service
	db                   databases.SqlConnection
	probes               insights.Probes
}
