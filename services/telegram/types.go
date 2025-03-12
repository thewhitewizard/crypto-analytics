package telegram

import (
	telegramRepo "crypto-analytics/repositories/telegram"
	cmcService "crypto-analytics/services/coinmarketcap"
	"crypto-analytics/services/cryptorank"
	twitterService "crypto-analytics/services/twitter"
	"errors"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/patrickmn/go-cache"
)

type MessageType int

const (
	MessageTypeUnknown     MessageType = -1
	MessageTypeStandard    MessageType = 0
	MessageTypeWelcome     MessageType = 1
	MessageTypeHelp        MessageType = 2
	MessageTypeReport      MessageType = 3
	MessageTypeSubscribe   MessageType = 4
	MessageTypeUnsubscribe MessageType = 5
)

var (
	ErrTokenIsMissing         = errors.New("telegram token is missing")
	ErrBotNotInitialized      = errors.New("telegram bot  is not ready yet")
	ErrFailedToStartListening = errors.New("telegram bot can't start to listen command")
)

type Service interface {
	ListenAndDispatch() error
}

type Impl struct {
	bot               *gotgbot.Bot
	updater           *ext.Updater
	telegramRepo      telegramRepo.Repository
	cmcService        cmcService.Service
	twitterService    twitterService.Service
	cryptorankService cryptorank.Service
	cache             *cache.Cache
}
