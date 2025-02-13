package main

import (
	"crypto-analytics/application"
	"crypto-analytics/models/constants"
	"os"
	"os/signal"
	"syscall"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

func init() {
	initConfig()
	initLog()
}

func initLog() {
	zerolog.SetGlobalLevel(constants.LogLevelFallback)

	logLevel, err := zerolog.ParseLevel(viper.GetString(constants.LogLevel))
	if err != nil {
		log.Warn().Err(err).Msgf("Log level not set, continue with %s...", constants.LogLevelFallback)
	} else {
		zerolog.SetGlobalLevel(logLevel)
		log.Debug().Msgf("Logger level set to '%s'", logLevel)
	}
}

func initConfig() {
	viper.SetConfigFile(constants.ConfigFileName)

	for configName, defaultValue := range constants.GetDefaultConfigValues() {
		viper.SetDefault(configName, defaultValue)
	}

	err := viper.ReadInConfig()
	if err != nil {
		log.Debug().Str(constants.LogFileName, constants.ConfigFileName).Msgf("Failed to read config file, continue...")
	}

	viper.AutomaticEnv()
}

func main() {
	app, err := application.New()
	if err != nil {
		log.Fatal().Err(err).Msgf("Shutting down after failing to instantiate application")
	}

	app.Run()

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	log.Info().Msgf("%s v%s is now running. Press CTRL-C to exit.", constants.ExternalName, constants.Version)
	<-sc

	log.Info().Msgf("Gracefully shutting down %s...", constants.ExternalName)
	app.Shutdown()
}
