package twitter

import (
	repo "crypto-analytics/repositories/twitter"
	"crypto-analytics/utils/dates"
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

	if viper.GetBool(constants.Production) {
		service.fetchAndSaveTweets()
	}

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

func (service *Impl) GetYesterdayTweets() ([]entities.Tweet, error) {
	start, end := dates.GetYesterdayTimestamps()
	log.Info().Int64("start", start).Int64("end", end).Msg("Start fetching tweets")
	tweets, err := service.repository.GetTweetBetweenTimestamps(start, end)
	if err != nil {
		return tweets, err
	}

	return filterRootTweets(tweets), nil
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
		tweetToSave := MapTweetToEntity(tweet)
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

func MapTweetToEntity(tweet *twitterscraper.Tweet) entities.Tweet {
	return entities.Tweet{
		ID:             tweet.ID,
		ConversationID: tweet.ConversationID,
		IsPin:          tweet.IsPin,
		IsQuoted:       tweet.IsQuoted,
		Mentions:       len(tweet.Mentions),
		IsReply:        tweet.IsQuoted,
		IsRetweet:      tweet.IsQuoted,
		IsSelfThread:   tweet.IsQuoted,
		Likes:          tweet.Likes,
		Name:           tweet.Name,
		PermanentURL:   tweet.PermanentURL,
		Replies:        tweet.Replies,
		Retweets:       tweet.Retweets,
		Timestamp:      tweet.Timestamp,
		UserID:         tweet.UserID,
		Views:          tweet.Views,
	}
}

func filterRootTweets(tweets []entities.Tweet) []entities.Tweet {
	if len(tweets) == 0 {
		return nil
	}

	// Étape 1 : On filtre pour ne garder que le plus ancien tweet (root) par conversation_id
	oldestPerConversation := make(map[string]entities.Tweet)

	for _, tweet := range tweets {
		if existing, exists := oldestPerConversation[tweet.ConversationID]; !exists || tweet.Timestamp < existing.Timestamp {
			oldestPerConversation[tweet.ConversationID] = tweet
		}
	}

	// Étape 2 : On cherche le plus récent des roots
	var mostRecentRoot *entities.Tweet
	for _, tweet := range oldestPerConversation {
		if mostRecentRoot == nil || tweet.Timestamp > mostRecentRoot.Timestamp {
			mostRecentRoot = &tweet
		}
	}

	return []entities.Tweet{*mostRecentRoot}
}
