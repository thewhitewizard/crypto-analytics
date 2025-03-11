package twitter

import (
	"crypto-analytics/models/constants"
	"crypto-analytics/models/entities"
	repo "crypto-analytics/repositories/twitter"

	twitterscraper "github.com/n0madic/twitter-scraper"
)

type Service interface {
	GetYesterdayTweets() ([]entities.Tweet, error)
}

type Impl struct {
	authToken  string
	csrfToken  string
	tweetCount int
	scraper    *twitterscraper.Scraper
	repository repo.Repository
	accounts   []constants.TwitterAccount
}
