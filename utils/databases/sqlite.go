package databases

import (
	"crypto-analytics/models/constants"

	"github.com/glebarez/sqlite"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type sqliteConnection struct {
	dsn string
	db  *gorm.DB
}

func New() SqlConnection {
	return &sqliteConnection{
		dsn: viper.GetString(constants.SqliteURL),
	}
}

func (c *sqliteConnection) GetDB() *gorm.DB {
	return c.db
}

func (c *sqliteConnection) IsConnected() bool {
	if c.db == nil {
		return false
	}

	dbSQL, errSQL := c.db.DB()
	if errSQL != nil {
		return false
	}

	if errPing := dbSQL.Ping(); errPing != nil {
		return false
	}

	return true
}

func (c *sqliteConnection) Run() error {
	db, err := gorm.Open(sqlite.Open(c.dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		return err
	}

	c.db = db
	log.Info().Msg("Connected to Sqlite")
	return nil
}

func (c *sqliteConnection) Shutdown() {
	log.Info().Msg("Shutdown the connection to Sqlite")
	dbSQL, err := c.db.DB()
	if err != nil {
		log.Error().Err(err).Msgf("Failed to shutdown database connection")
		return
	}

	if errClose := dbSQL.Close(); errClose != nil {
		log.Error().Err(errClose).Msgf("Failed to shutdown database connection")
	}
}
