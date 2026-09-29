package utils

import (
	"fmt"
	"net/http"
	"node/global"
	"strconv"
	"sync"
	"time"

	"go.uber.org/zap"
	"gopkg.in/telebot.v3"
)

var (
	tgBot        *telebot.Bot
	tgBotOnce    sync.Once
	tgBotErr     error
	messageLimit = 4000
)

func getTelegramBot() (*telebot.Bot, error) {
	tgBotOnce.Do(func() {
		tgBot, tgBotErr = telebot.NewBot(telebot.Settings{
			Token:   global.NODE_CONFIG.Telegram.InformBotToken,
			Offline: true, // 不请求 getMe
			Client:  &http.Client{Timeout: 10 * time.Second},
		})
	})
	return tgBot, tgBotErr
}

// telegram message limit: 4096
func InformToTelegram(message string) bool {
	defer HandlePanic()

	bot, err := getTelegramBot()
	if err != nil {
		global.NODE_LOG.Error("telegram bot init failed")
		return false
	}

	message = TruncateRunes(message, messageLimit)

	_, err = bot.Send(&telebot.Chat{ID: global.NODE_CONFIG.Telegram.InformChannelId}, message)
	if err != nil {
		global.NODE_LOG.Error("telegram send failed", zap.String("err", err.Error()))

	}
	return true
}

func TxInformToTelegram(message string) bool {
	defer HandlePanic()

	bot, err := getTelegramBot()
	if err != nil {
		global.NODE_LOG.Error("telegram bot init failed")
		return false
	}

	message = TruncateRunes(message, messageLimit)

	_, err = bot.Send(&telebot.Chat{ID: global.NODE_CONFIG.Telegram.TxInformChannelId}, message)
	if err != nil {
		global.NODE_LOG.Error("telegram send failed", zap.String("err", err.Error()))
		return false
	}
	return true
}

func NotificationToTelegram(botToken string, tgId string, message string) bool {
	defer HandlePanic()

	tgIdInt, err := strconv.ParseInt(tgId, 10, 64)
	if err != nil {
		global.NODE_LOG.Error(err.Error())
		return false
	}

	botSetting := telebot.Settings{
		Token: botToken,
	}

	bot, err := telebot.NewBot(botSetting)
	if err != nil {
		global.NODE_LOG.Error(err.Error() + fmt.Sprintf(" newbot: %s", botSetting.Token))
		return false
	}

	_, err = bot.Send(&telebot.Chat{ID: tgIdInt}, message)
	if err != nil {
		global.NODE_LOG.Info(err.Error())
		return false
	}

	return true
}
