package utils

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"node/global"
	"strconv"
	"sync"
	"time"

	"go.uber.org/zap"
	"golang.org/x/time/rate"
	"gopkg.in/telebot.v3"
)

/*
如何防止无限请求

分四层,由内到外:

- 令牌桶限流:控制发送速率,超出的直接丢弃或排队;
- 相同内容去重(冷却):同一条告警 N 分钟内只发一次,并统计被抑制的次数;
- 有界队列 + 单 worker 异步发送:队列满了就丢弃,业务线程永不阻塞;
- 429 处理:遵守 retry_after,而不是立刻重试。
*/
var (
	tgBot        *telebot.Bot
	tgBotMu      *sync.Mutex
	messageLimit = 4000

	// 全局限流:每 3 秒 1 条,突发最多 5 条(约 20 条/分钟,贴合单频道限制)
	tgLimiter = rate.NewLimiter(rate.Every(3*time.Second), 5)

	// 内容去重冷却
	dedup = newDedup(5 * time.Minute)

	tgQueue      = make(chan string, 200)
	tgWorkerOnce sync.Once
)

// 初始化失败不缓存,下次调用会重试
func getTelegramBot() (*telebot.Bot, error) {
	tgBotMu.Lock()
	defer tgBotMu.Unlock()
	if tgBot != nil {
		return tgBot, nil
	}
	bot, err := telebot.NewBot(telebot.Settings{
		Token:   global.NODE_CONFIG.Telegram.InformBotToken,
		Offline: true, // 不请求 getMe
		Client:  &http.Client{Timeout: 10 * time.Second},
	})
	if err != nil {
		return nil, err
	}
	tgBot = bot
	return tgBot, nil
}

func tgWorker() {
	defer HandlePanic()
	for msg := range tgQueue {
		// 按速率等待,而不是丢弃
		if err := tgLimiter.Wait(context.Background()); err != nil {
			continue
		}
		sendWithRetry(msg)
	}
}

func sendWithRetry(message string) {
	bot, err := getTelegramBot()
	if err != nil {
		global.NODE_LOG.Error("telegram bot init failed", zap.Error(err))
		return
	}

	message = TruncateRunes(message, messageLimit)
	chat := &telebot.Chat{ID: global.NODE_CONFIG.Telegram.InformChannelId}

	const maxRetry = 3
	for i := 0; i < maxRetry; i++ {
		_, err = bot.Send(chat, message)
		if err == nil {
			return
		}

		// 429:按服务端要求等待后重试
		var fe telebot.FloodError
		if errors.As(err, &fe) {
			time.Sleep(time.Duration(fe.RetryAfter) * time.Second)
			continue
		}
		// 其他错误:指数退避
		time.Sleep(time.Duration(1<<i) * time.Second)
	}
	global.NODE_LOG.Error("telegram send failed after retries", zap.Error(err))
}

// telegram message limit: 4096
// InformToTelegram 异步入队,不阻塞业务。返回 false 表示被去重或队列已满而丢弃。
func InformToTelegram(message string) bool {
	tgWorkerOnce.Do(func() { go tgWorker() })

	if !dedup.Allow(message) {
		return false // 冷却期内的重复消息
	}

	select {
	case tgQueue <- message:
		return true
	default:
		global.NODE_LOG.Warn("telegram queue full, drop message")
		return false
	}
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
