package utils

import (
	"fmt"
	"node/global"
	"strconv"

	"gopkg.in/telebot.v3"
)

func InformToTelegram(message string) bool {
	defer HandlePanic()

	botSetting := telebot.Settings{
		Token: global.NODE_CONFIG.Telegram.InformBotToken,
	}

	bot, err := telebot.NewBot(botSetting)
	if err != nil {
		global.NODE_LOG.Error(err.Error() + fmt.Sprintf(" newbot: %s", botSetting.Token))
		return false
	}

	_, err = bot.Send(&telebot.Chat{ID: global.NODE_CONFIG.Telegram.InformChannelId}, message)
	if err != nil {
		global.NODE_LOG.Info(err.Error())
		return false
	}

	return true
}

func TxInformToTelegram(message string) bool {
	defer HandlePanic()

	botSetting := telebot.Settings{
		Token: global.NODE_CONFIG.Telegram.InformBotToken,
	}

	bot, err := telebot.NewBot(botSetting)
	if err != nil {
		global.NODE_LOG.Error(err.Error() + fmt.Sprintf(" newbot: %s", botSetting.Token))
		return false
	}

	_, err = bot.Send(&telebot.Chat{ID: global.NODE_CONFIG.Telegram.TxInformChannelId}, message)
	if err != nil {
		global.NODE_LOG.Info(err.Error())
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
