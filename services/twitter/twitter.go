package twitter

import (
	repo "crypto-analytics/repositories/twitter"
	"sort"
	"sync"

	"crypto-analytics/models/constants"
	"crypto-analytics/models/entities"

	"github.com/go-co-op/gocron/v2"
	twitterscraper "github.com/n0madic/twitter-scraper"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

func New(scheduler gocron.Scheduler,
	repository repo.Repository,
	accounts []constants.TwitterAccount) (*Impl, error) {
	service := &Impl{
		accounts:   accounts,
		repository: repository,
		authToken:  viper.GetString(constants.TwitterAuthToken),
		csrfToken:  viper.GetString(constants.TwitterCSRFToken),
		tweetCount: viper.GetInt(constants.TwitterTweetCount),
		scraper:    twitterscraper.New(),
	}

	service.fetchAndSaveTweets()

	_, errJob := scheduler.NewJob(
		gocron.CronJob("*/15 * * * *", true),
		gocron.NewTask(func() { service.fetchAndSaveTweets() }),
		gocron.WithName("Fetch twitter accounts"),
	)
	if errJob != nil {
		return nil, errJob
	}

	return service, nil
}

func (service *Impl) fetchAndSaveTweets() {
	log.Info().Msg("Start fetching twitter accounts")
	var wg sync.WaitGroup
	for _, account := range service.accounts {
		wg.Add(1)
		go func(twitterAccount constants.TwitterAccount) {
			defer wg.Done()
			service.checkTwitterAccount(twitterAccount)
		}(account)
	}

	wg.Wait()
	log.Info().Msg("End fetching twitter accounts")
}

func (service *Impl) checkTwitterAccount(account constants.TwitterAccount) {
	log.Info().
		Str(constants.LogTwitterName, account.Name).
		Str(constants.LogTwitterID, account.ID).
		Msgf("Reading tweets...")

	tweets, _, err := service.scraper.FetchTweets(account.Name, service.tweetCount, "")

	if err != nil {
		log.Error().Err(err).
			Str(constants.LogTwitterID, account.ID).
			Msgf("Cannot retrieve tweets from account, ignored")
		return
	}

	tweets = service.keepInterestingTweets(tweets)

	for _, tweet := range tweets {
		tweetToSave := entities.Tweet{ID: tweet.ID}
		tweetToSave.ConversationID = tweet.ConversationID
		tweetToSave.IsPin = tweet.IsPin
		tweetToSave.IsQuoted = tweet.IsQuoted
		tweetToSave.Mentions = len(tweet.Mentions)
		tweetToSave.IsReply = tweet.IsQuoted
		tweetToSave.IsRetweet = tweet.IsQuoted
		tweetToSave.IsSelfThread = tweet.IsQuoted
		tweetToSave.Likes = tweet.Likes
		tweetToSave.Name = tweet.Name
		tweetToSave.PermanentURL = tweet.PermanentURL
		tweetToSave.Replies = tweet.Replies
		tweetToSave.Retweets = tweet.Retweets
		//tweetToSave.Text = tweet.Text
		tweetToSave.Timestamp = tweet.Timestamp
		tweetToSave.UserID = tweet.UserID
		tweetToSave.Views = tweet.Views
		service.repository.SaveOrUpdate(tweetToSave)

	}

}

func (service *Impl) keepInterestingTweets(tweets []*twitterscraper.Tweet) []*twitterscraper.Tweet {
	result := make([]*twitterscraper.Tweet, 0)

	for _, tweet := range tweets {
		// Exclude RTs
		if tweet.RetweetedStatus != nil {
			continue
		}

		result = append(result, tweet)
	}

	sort.SliceStable(result, func(i, j int) bool {
		return result[i].Timestamp < result[j].Timestamp
	})

	return result
}
