package health

import (
	"crypto-analytics/models/constants"

	"github.com/go-co-op/gocron/v2"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

func New(scheduler gocron.Scheduler) (*Impl, error) {
	service := Impl{}

	_, errJob := scheduler.NewJob(
		gocron.CronJob(viper.GetString(constants.HealthCronTab), true),
		gocron.NewTask(func() { service.echo() }),
		gocron.WithName("Check app running"),
	)
	if errJob != nil {
		return nil, errJob
	}

	return &service, nil
}

func (service *Impl) echo() {
	log.Info().Msgf("Application is running")
}
