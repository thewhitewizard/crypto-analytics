package constants

import "github.com/rs/zerolog"

const (
	ConfigFileName = ".env"

	//nolint:gosec // False positive.
	// Auth token used when logged in to Twitter.
	TwitterAuthToken = "TWITTER_AUTH_TOKEN"

	//nolint:gosec // False positive.
	// CSRF token used when logged in to Twitter.
	TwitterCSRFToken = "TWITTER_CSRF_TOKEN"

	// Number of tweets retrieved per call.
	TwitterTweetCount = "TWEET_COUNT"

	// SQLITE_URL URL.
	SqliteURL = "SQLITE_URL"

	// Zerolog values from [trace, debug, info, warn, error, fatal, panic].
	LogLevel = "LOG_LEVEL"

	// Probe port.
	ProbePort = "PROBE_PORT"

	// Boolean; used to register commands at development guild level or globally.
	Production = "PRODUCTION"

	// Cron tab to health.
	HealthCronTab = "HEALTH_CRON_TAB"

	// Cron tab to trending.
	TrendingCryptoCronTab = "TRENDING_CRYPTO_CRON_TAB"

	// Cron tab to historical.
	HistoricalCryptoCronTab = "HISTORICAL_CRYPTO_CRON_TAB"

	defaultTwitterAuthToken         = ""
	defaultTwitterCSRFToken         = ""
	defaultTwitterTweetCount        = 20
	defaultProbePort                = 9090
	defaultSqliteURL                = "crypto-analytics.db"
	defaultHealthCrontab            = "* * * * *"
	defaultTrendingCryptoCrontTab   = "0 */6 * * *"
	defaultHistoricalCryptoCrontTab = "0 3 * * *"
	defaultLogLevel                 = zerolog.InfoLevel
	defaultProduction               = false
)

func GetDefaultConfigValues() map[string]any {
	return map[string]any{
		TwitterAuthToken:        defaultTwitterAuthToken,
		TwitterCSRFToken:        defaultTwitterCSRFToken,
		TwitterTweetCount:       defaultTwitterTweetCount,
		ProbePort:               defaultProbePort,
		SqliteURL:               defaultSqliteURL,
		LogLevel:                defaultLogLevel.String(),
		Production:              defaultProduction,
		HealthCronTab:           defaultHealthCrontab,
		TrendingCryptoCronTab:   defaultTrendingCryptoCrontTab,
		HistoricalCryptoCronTab: defaultHistoricalCryptoCrontTab,
	}
}
