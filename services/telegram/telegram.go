package telegram

import (
	"crypto-analytics/models/constants"
	"crypto-analytics/models/entities"
	"crypto-analytics/pkg/observer"
	telegramRepo "crypto-analytics/repositories/telegram"
	cmcService "crypto-analytics/services/coinmarketcap"
	"crypto-analytics/utils/dates"
	"fmt"
	"time"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"
	"github.com/go-co-op/gocron/v2"
	"github.com/patrickmn/go-cache"
	"github.com/rs/zerolog/log"
)

func New(scheduler gocron.Scheduler, token string, telegramRepo telegramRepo.Repository, cmcService cmcService.Service) (*Impl, error) {

	if token == "" {
		return &Impl{}, ErrTokenIsMissing
	}

	b, err := gotgbot.NewBot(token, nil)
	if err != nil {
		return &Impl{}, ErrBotNotInitialized
	}

	dispatcher := ext.NewDispatcher(&ext.DispatcherOpts{
		Error: func(b *gotgbot.Bot, ctx *ext.Context, err error) ext.DispatcherAction {
			log.Warn().Msg("an error occurred while handling update")
			return ext.DispatcherActionNoop
		},
		MaxRoutines: ext.DefaultMaxRoutines,
	})

	service := Impl{bot: b, telegramRepo: telegramRepo, cmcService: cmcService, cache: cache.New(1*time.Hour, 2*time.Hour)}
	dispatcher.AddHandler(handlers.NewCommand("start", service.startCmd))
	dispatcher.AddHandler(handlers.NewCommand("help", service.helpCmd))
	dispatcher.AddHandler(handlers.NewCommand("report", service.reportCmd))
	dispatcher.AddHandler(handlers.NewCommand("subscribe", service.subscribeCmd))
	dispatcher.AddHandler(handlers.NewCommand("unsubscribe", service.unsubscribeCmd))
	dispatcher.AddHandler(handlers.NewCommand("maintenance", service.maintenanceCmd))
	dispatcher.AddHandler(handlers.NewCommand("", service.unknownCmd))

	service.updater = ext.NewUpdater(dispatcher, nil)

	_, errJob := scheduler.NewJob(
		gocron.CronJob("0 7 * * *", true),
		gocron.NewTask(func() { service.sendDailyReport(-1) }),
		gocron.WithName("Send daily report"),
	)
	if errJob != nil {
		return nil, errJob
	}

	_, errAdminJob := scheduler.NewJob(
		gocron.CronJob("0 14 * * *", true),
		gocron.NewTask(func() { service.dailyAdminReport() }),
		gocron.WithName("Send daily report to admin"),
	)
	if errAdminJob != nil {
		return nil, errAdminJob
	}

	/**
		_, errJobGenerateReport := scheduler.NewJob(
			gocron.CronJob("/2 * * * *", true),
			gocron.NewTask(func() { service.generateReport() }),
			gocron.WithName("Generate daily report"),
		)
		if errJobGenerateReport != nil {
			return nil, errJobGenerateReport
		}

		_, errJobNotify := scheduler.NewJob(
			gocron.CronJob("* * * * *", true),
			gocron.NewTask(func() { service.tendringNotify() }),
			gocron.WithName("Check alert"),
		)
		if errJobNotify != nil {
			return nil, errJobNotify
		}
	**/
	service.generateReport()
	return &service, nil
}

func (service *Impl) ListenAndDispatch() error {

	err := service.updater.StartPolling(service.bot, &ext.PollingOpts{
		DropPendingUpdates: true,
		GetUpdatesOpts: &gotgbot.GetUpdatesOpts{
			Timeout: 9,
			RequestOpts: &gotgbot.RequestOpts{
				Timeout: time.Second * 10,
			},
		},
	})
	if err != nil {
		return ErrFailedToStartListening
	}

	service.updater.Idle()

	time.Sleep(1 * time.Hour)
	return nil
}

func (service *Impl) maintenanceCmd(b *gotgbot.Bot, ctx *ext.Context) error {

	if ctx.EffectiveChat.Id != constants.TelegramAdmin {
		log.Warn().Str("cmd", "maintenance").Int64("chatID", ctx.EffectiveChat.Id).Msg("forbidden usage")
		return nil
	}

	msg := "🚧 *Scheduled Maintenance Alert* ⚙️\n\n"
	msg += "Hey there! Just a heads-up that I'll be undergoing maintenance soon to keep things running smoothly. 🛠️\n\n"
	msg += "🔹 During this time, some features may be temporarily unavailable.\n"
	msg += "🔹 Don't worry—I'll be back online as soon as possible!\n\n"
	msg += "Thanks for your patience and support! 🚀🤖\n"

	users, err := service.telegramRepo.FetchAll()

	if err == nil {
		for _, user := range users {
			log.Info().Str("cmd", "maintenance").Int64("chatID", user.ChatID).Msg("send maintenance")
			service.bot.SendMessage(user.ChatID, msg, &gotgbot.SendMessageOpts{ParseMode: "Markdown"})
		}
	}

	return nil

}

func (service *Impl) startCmd(b *gotgbot.Bot, ctx *ext.Context) error {
	log.Info().Str("cmd", "start").Str("username", ctx.EffectiveChat.Username).Int64("chatID", ctx.EffectiveChat.Id).Msg("command received")
	service.bot.SendMessage(ctx.EffectiveChat.Id, getMessageFromMessageType(MessageTypeWelcome), &gotgbot.SendMessageOpts{ParseMode: "Markdown"})
	return nil
}

func (service *Impl) helpCmd(b *gotgbot.Bot, ctx *ext.Context) error {
	log.Info().Str("cmd", "help").Str("username", ctx.EffectiveChat.Username).Int64("chatID", ctx.EffectiveChat.Id).Msg("command received")
	service.bot.SendMessage(ctx.EffectiveChat.Id, getMessageFromMessageType(MessageTypeHelp), &gotgbot.SendMessageOpts{ParseMode: "Markdown"})
	return nil
}

func (service *Impl) unknownCmd(b *gotgbot.Bot, ctx *ext.Context) error {
	log.Info().Str("cmd", "unknown").Str("username", ctx.EffectiveChat.Username).Int64("chatID", ctx.EffectiveChat.Id).Msg("command received")
	service.bot.SendMessage(ctx.EffectiveChat.Id, getGenericErrorMEssage(), &gotgbot.SendMessageOpts{ParseMode: "Markdown"})
	return nil
}

func (service *Impl) subscribeCmd(b *gotgbot.Bot, ctx *ext.Context) error {
	log.Info().Str("cmd", "subscribe").Str("username", ctx.EffectiveChat.Username).Int64("chatID", ctx.EffectiveChat.Id).Msg("command received")
	err := service.telegramRepo.SaveOrUpdate(entities.TelegramUser{ChatID: ctx.EffectiveChat.Id, Name: ctx.EffectiveChat.Username})
	if err != nil {
		log.Error().Err(err).Str("username", ctx.EffectiveChat.Username).Int64("chatID", ctx.EffectiveChat.Id).Msg("error on saved")
	} else {
		service.notifyAdminOnNewUser(ctx.EffectiveChat.Id)
	}
	service.bot.SendMessage(ctx.EffectiveChat.Id, getMessageFromMessageType(MessageTypeSubscribe), &gotgbot.SendMessageOpts{ParseMode: "Markdown"})

	return nil
}

func (service *Impl) unsubscribeCmd(b *gotgbot.Bot, ctx *ext.Context) error {
	log.Info().Str("cmd", "unsubscribe").Str("username", ctx.EffectiveChat.Username).Int64("chatID", ctx.EffectiveChat.Id).Msg("command received")
	err := service.telegramRepo.Delete(entities.TelegramUser{ChatID: ctx.EffectiveChat.Id})
	if err != nil {
		log.Error().Err(err).Int64("chatID", ctx.EffectiveChat.Id).Msg("error on deleted")
	}
	service.bot.SendMessage(ctx.EffectiveChat.Id, getMessageFromMessageType(MessageTypeUnsubscribe), &gotgbot.SendMessageOpts{ParseMode: "Markdown"})
	return nil
}

func (service *Impl) notifyAdminOnNewUser(chatID int64) {
	if chatID != constants.TelegramAdmin {
		msg := "🆕 *Nouvel abonnement!* 🎉\n\n"
		msg += "Un nouvel utilisateur s'est abonné aux notifications RLC Watchdog. 🚀\n"
		msg += fmt.Sprintf("👤 *User ID:* `%d`\n", chatID)
		msg += fmt.Sprintf("📅 *Date:* `%s`\n", time.Now().Format("2006-01-02 15:04:05"))
		msg += "\nLe bot gagne en popularité ! 📈🔥"

		service.bot.SendMessage(constants.TelegramAdmin, msg, &gotgbot.SendMessageOpts{ParseMode: "Markdown"})
	}
}

func (service *Impl) dailyAdminReport() {
	users, err := service.telegramRepo.FetchAll()
	if err == nil && len(users) > 0 {
		msg := "📢 *Rapport quotidien des abonnés* 📊\n\n"
		msg += fmt.Sprintf("👥 *Nombre total d'abonnés:* `%d`\n", len(users))

		service.bot.SendMessage(constants.TelegramAdmin, msg, &gotgbot.SendMessageOpts{ParseMode: "Markdown"})
	}
}

func (service *Impl) reportCmd(b *gotgbot.Bot, ctx *ext.Context) error {
	log.Info().Str("cmd", "report").Str("username", ctx.EffectiveChat.Username).Int64("chatID", ctx.EffectiveChat.Id).Msg("command received")
	service.sendDailyReport(ctx.EffectiveChat.Id)
	return nil
}

func (service *Impl) tendringNotify() {
	log.Info().Msg("Check trending notification")
	cryptocurrencies := constants.GetCrytoWatch()
	users, _ := service.telegramRepo.FetchAll()
	if len(users) > 0 {
		today := time.Now().Format(dates.DateFormat)
		if len(cryptocurrencies) > 0 {
			for _, crycryptocurrency := range cryptocurrencies {
				trendy := service.cmcService.IsCryptoTrendyToday(crycryptocurrency.Symbol)
				if trendy {
					key := today + crycryptocurrency.Symbol
					_, found := service.cache.Get(key)
					if !found {
						service.cache.Set(key, true, time.Hour*25)

						msg := "🚨 *Trending Alert!* 🚀🔥\n\n"
						msg += "🔍 A cryptocurrency is gaining traction! Check it out:\n\n"
						msg += fmt.Sprintf("🔹 *%s* is now *TRENDING!* 🚀\n", crycryptocurrency.Symbol)
						msg += "\n⚡ Stay ahead of the market!\n"
						for _, user := range users {
							log.Info().Int64("chatID", user.ChatID).Msg("send trending notification")
							service.bot.SendMessage(user.ChatID, msg, &gotgbot.SendMessageOpts{ParseMode: "Markdown"})
						}
					}
				}
			}
		}
	}
}

func (service *Impl) generateReport() {
	log.Info().Msg("Generate daily report")
	cryptocurrencies := constants.GetCrytoWatch()
	ok := false
	if len(cryptocurrencies) > 0 {
		msg := "📢 *Daily Crypto Report* 🚀\n\n"
		for _, crycryptocurrency := range cryptocurrencies {
			msg += "🔹 *" + crycryptocurrency.Desc + "*\n"
			histo, errPrice := service.cmcService.FetchForSymbolYesterday(crycryptocurrency.Symbol)
			trendy := service.cmcService.IsCryptoTrendyYersterday(crycryptocurrency.Symbol)
			community, errCommunity := service.cmcService.FetchCommunityDataForSymbolYesterday(crycryptocurrency.CryptoId)

			if errPrice == nil {
				msg += fmt.Sprintf("💰 Price: `$%.2f`\n", histo.Price)
				msg += fmt.Sprintf("📊 Rank: `#%d`\n", histo.Rank)
				msg += fmt.Sprintf("🏛 Market Cap: `$%.2f`\n", histo.Marketcap)
				ok = true
			}
			if trendy {
				msg += fmt.Sprintf("🔥 Trending: *%s*\n\n", "Yes! 🚀")
			} else {
				msg += fmt.Sprintf("🔥 Trending: *%s*\n\n", "No ❄️")
			}
			if errCommunity == nil {
				msg += fmt.Sprintf("👥 *Followers on CMC:* `%s`\n", community.Followers)
				msg += fmt.Sprintf("⭐ *Watchlist Count:* `%s`\n", community.WatchCount)
			}
			msg += "\n"
		}

		histo, err := service.cmcService.FetchForSymbolYesterday("BTC")
		if err == nil {
			msg += fmt.Sprintf("💰 BTC Price: `$%.2f`\n", histo.Price)
		}

		histo, err = service.cmcService.FetchForSymbolYesterday("ETH")
		if err == nil {
			msg += fmt.Sprintf("💰 ETH Price: `$%.2f`\n", histo.Price)
		}
		msg += "\n"
		msg += "📆 Data from *yesterday*. Stay tuned for more updates! 📈\n"

		if ok {
			service.cache.Set("daily_report", msg, cache.NoExpiration)
		}
	}

}

func (service *Impl) OnNotify(e observer.Event) {
	log.Info().Msg("Received internal notification")
	if e.E == observer.TrendingEvent {
		service.tendringNotify()
	} else if e.E == observer.RankingEvent {
		service.generateReport()
	}

}
func (service *Impl) sendDailyReport(chatID int64) {
	log.Info().Msg("Send daily report")
	var users []entities.TelegramUser
	var err error
	if chatID != -1 {
		users = append(users, entities.TelegramUser{ChatID: chatID})
	} else {
		users, err = service.telegramRepo.FetchAll()
	}

	var message string
	if x, found := service.cache.Get("daily_report"); found {
		message = x.(string)
		if len(message) > 0 && err == nil {
			for _, user := range users {
				log.Info().Str("cmd", "report").Int64("chatID", user.ChatID).Msg("send report")
				service.bot.SendMessage(user.ChatID, message, &gotgbot.SendMessageOpts{ParseMode: "Markdown"})
			}
		}
	} else {
		log.Warn().Str("cmd", "report").Msg("No report")
	}

	/**
	cryptocurrencies := constants.GetCrytoWatch()
	if err == nil && len(users) > 0 && len(cryptocurrencies) > 0 {
		msg := "📢 *Daily Crypto Report* 🚀\n\n"
		for _, crycryptocurrency := range cryptocurrencies {
			msg += "🔹 *" + crycryptocurrency.Desc + "*\n"
			histo, errPrice := service.cmcService.FetchForSymbolYesterday(crycryptocurrency.Symbol)
			trendy := service.cmcService.IsCryptoTrendyYersterday(crycryptocurrency.Symbol)

			if errPrice == nil {
				msg += fmt.Sprintf("💰 Price: `$%.2f`\n", histo.Price)
				msg += fmt.Sprintf("📊 Rank: `#%d`\n", histo.Rank)
				msg += fmt.Sprintf("🏛 Market Cap: `$%.2f`\n", histo.Marketcap)
			}
			if trendy {
				msg += fmt.Sprintf("🔥 Trending: *%s*\n\n", "Yes! 🚀")
			} else {
				msg += fmt.Sprintf("🔥 Trending: *%s*\n\n", "No ❄️")
			}
			msg += "\n"
		}

		msg += "📆 Data from *yesterday*. Stay tuned for more updates! 📈\n"

		for _, user := range users {
			log.Info().Str("cmd", "report").Int64("chatID", user.ChatID).Msg("send report")
			service.bot.SendMessage(user.ChatID, msg, &gotgbot.SendMessageOpts{ParseMode: "Markdown"})
		}

	} else {
		log.Warn().Str("cmd", "report").Msg("No User")
	}
		**/

}

func getGenericErrorMEssage() string {

	msg := "😔 *Oops! Something Went Wrong*\n\n"

	msg += "It looks like I couldn’t complete your request. Don’t worry, it’s not you—it’s me. Here’s what you can try:\n"

	msg += "1️⃣ Double-check the information you provided.\n"
	msg += "2️⃣ Wait a moment and try again.\n\n"

	msg += "Thanks for your patience—I’ll do my best to sort this out! 🤖✨"

	return msg
}

func getMessageFromMessageType(messageType MessageType) string {
	switch messageType {
	case MessageTypeWelcome:
		msg := "👋 Hi! I'm *RLC Watchdog* 🤖\n\n"
		msg += "This bot keeps you updated on RLC's key metrics 📊—trending status, rank, and how it compares to competitors.\n\n"
		msg += "No need to do anything! I'll send you a daily report automatically 📨. Just sit back and stay informed.\n\n"
		msg += "✅ *Want to receive updates?* Type `/subscribe` to start receiving daily reports.\n"
		msg += "❌ *Want to stop updates?* Type `/unsubscribe` at any time.\n\n"
		msg += "💬 *Need help?* Type `/help` for a list of commands."

		return msg

	case MessageTypeHelp:
		msg := "🤖 *RLC Watchdog* – Help Guide 📢\n\n"
		msg += "This bot provides daily updates on RLC’s ranking and trends 📈.\n\n"
		msg += "📝 *Commands available:*\n"
		msg += "✅ `/subscribe` – Start receiving daily reports.\n"
		msg += "❌ `/unsubscribe` – Stop receiving daily reports.\n"
		msg += "📊 `/report` – Get the latest RLC report instantly.\n"
		msg += "💡 `/help` – Show this help message.\n\n"
		msg += "🚀 Stay ahead with the latest RLC data!\n"

		return msg

	case MessageTypeSubscribe:
		msg := "🎉 *Subscription Confirmed!* ✅\n\n"
		msg += "You're now subscribed to daily updates on RLC! 📊🚀\n\n"
		msg += "I'll send you reports automatically every day. If you ever want to stop receiving them, just type `/unsubscribe`.\n"

		return msg

	case MessageTypeUnsubscribe:
		msg := "👋 *You've Unsubscribed* ❌\n\n"
		msg += "You will no longer receive daily RLC updates. 😔\n\n"
		msg += "If you change your mind, type `/subscribe` anytime to start receiving reports again! 🚀\n"

		return msg

	default:
		msg := "👋 Hi! I'm *RLC Watchdog* 🤖\n\n"
		msg += "This bot keeps you updated on RLC's key metrics 📊—trending status, rank, and how it compares to competitors.\n\n"
		msg += "No need to do anything! I'll send you a daily report automatically 📨. Just sit back and stay informed.\n\n"
		msg += "✅ *Want to receive updates?* Type `/subscribe` to start receiving daily reports.\n"
		msg += "❌ *Want to stop updates?* Type `/unsubscribe` at any time.\n\n"
		msg += "💬 *Need help?* Type `/help` for a list of commands."

		return msg
	}
}
