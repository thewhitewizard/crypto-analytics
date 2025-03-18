package feeds

import (
	"context"
	"crypto-analytics/models/constants"
	"crypto-analytics/models/entities"
	"crypto-analytics/pkg/observer"
	"crypto-analytics/repositories/feedsources"
	"sort"
	"time"

	"github.com/go-co-op/gocron/v2"
	"github.com/mmcdole/gofeed"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

func New(feedSourceRepo feedsources.Repository, scheduler gocron.Scheduler) (*Impl, error) {
	fp := gofeed.NewParser()
	fp.UserAgent = viper.GetString(constants.UserAgent)
	service := &Impl{
		feedParser:     fp,
		timeout:        time.Duration(viper.GetInt(constants.RSSTimeout)) * time.Second,
		feedSourceRepo: feedSourceRepo,
	}
	service.observers = map[observer.Observer]struct{}{}

	_, errJob := scheduler.NewJob(
		gocron.CronJob("0 8 * * *", true),
		gocron.NewTask(func() { service.FetchFeeds() }),
		gocron.WithName("Fetch feeds"),
	)
	if errJob != nil {
		return nil, errJob
	}

	if service.feedSourceRepo.Count() == 0 {
		err := service.feedSourceRepo.Create(entities.FeedSource{FeedTypeID: "cointelegraph", URL: "https://cointelegraph.com/rss", LastUpdate: time.Now().AddDate(0, 0, -5)})
		if err != nil {
			log.Error().Err(err).Msg("Error on save feed")
		}
	}

	return service, nil

}

func (service *Impl) RegisterObserver(o observer.Observer) {
	service.observers[o] = struct{}{}
}

func (service *Impl) FetchFeeds() error {
	log.Info().Msgf("Checking feeds...")

	feedSources, err := service.feedSourceRepo.GetFeedSources()
	if err != nil {
		return err
	}

	for _, feedSource := range feedSources {

		service.checkFeed(feedSource)
	}
	return nil
}

func (service *Impl) checkFeed(source entities.FeedSource) {
	log.Info().
		Str(constants.LogFeedURL, source.URL).
		Str(constants.LogFeedType, source.FeedTypeID).
		Msgf("Reading feed source...")

	feed, err := service.readFeed(source.URL)
	if err != nil {
		log.Error().
			Err(err).
			Str(constants.LogFeedType, source.FeedTypeID).
			Str(constants.LogFeedURL, source.URL).
			Msgf("Cannot parse URL, source ignored")
		return
	}

	publishedFeeds := 0
	lastUpdate := source.LastUpdate
	for _, feedItem := range feed.Items {
		if feedItem.PublishedParsed.UTC().After(lastUpdate.UTC()) {
			errPublish := service.publishFeedItem(feedItem, feed.Copyright, source)
			if errPublish != nil {
				log.Error().Err(err).
					Str(constants.LogFeedType, source.FeedTypeID).
					Str(constants.LogFeedURL, source.URL).
					Str(constants.LogFeedItemID, feedItem.GUID).
					Msgf("Impossible to publish RSS feed, breaking loop")
				break
			}

			if feedItem.PublishedParsed.UTC().After(source.LastUpdate.UTC()) {
				source.LastUpdate = feedItem.PublishedParsed.UTC()
			}
			err = service.feedSourceRepo.Save(source)
			if err != nil {
				log.Error().Err(err).
					Str(constants.LogFeedType, source.FeedTypeID).
					Str(constants.LogFeedURL, source.URL).
					Str(constants.LogFeedItemID, feedItem.GUID).
					Msgf("Impossible to update feed source, breaking loop; this feed might be published again next time")
				break
			}

			publishedFeeds++
		}
	}

	log.Info().
		Str(constants.LogFeedType, source.FeedTypeID).
		Str(constants.LogFeedURL, source.URL).
		Int(constants.LogFeedNumber, publishedFeeds).
		Msgf("Feed(s) read and published")
}

func (service *Impl) readFeed(url string) (*gofeed.Feed, error) {
	ctx, cancel := context.WithTimeout(context.Background(), service.timeout)
	defer cancel()
	feed, err := service.feedParser.ParseURLWithContext(url, ctx)
	if err != nil {
		return nil, err
	}

	sort.SliceStable(feed.Items, func(i, j int) bool {
		return feed.Items[i].PublishedParsed.Before(*feed.Items[j].PublishedParsed)
	})

	return feed, nil
}

func (service *Impl) publishFeedItem(item *gofeed.Item, source string,
	feedSource entities.FeedSource) error {
	for o := range service.observers {
		o.OnNotify(observer.NewRSSEvent(item))
	}

	return nil
}
