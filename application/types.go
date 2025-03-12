package application

import (
	cmcService "crypto-analytics/services/coinmarketcap"
	"crypto-analytics/services/cryptorank"
	telegramService "crypto-analytics/services/telegram"
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
	coinmarketcapService cmcService.Service
	telegramService      telegramService.Service
	twitterService       twitter.Service
	cryptorankService    cryptorank.Service
	db                   databases.SqlConnection
	probes               insights.Probes
}
