package main

import (
	"fatbot/schedule"
	"fatbot/users"
	"os"
	"time"

	"github.com/charmbracelet/log"
	"github.com/getsentry/sentry-go"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type FatBotUpdate struct {
	Bot    *tgbotapi.BotAPI
	Update tgbotapi.Update
}

func main() {
	var bot *tgbotapi.BotAPI
	var err error
	var updates tgbotapi.UpdatesChannel
	if os.Getenv("ENVIRONMENT") != "production" {
		log.SetLevel(log.DebugLevel)
	} else {
		err := sentry.Init(sentry.ClientOptions{
			Dsn:              os.Getenv("SENTRY_DSN"),
			TracesSampleRate: 1.0,
		})
		if err != nil {
			log.Fatalf("sentry.Init: %s", err)
		}
		defer sentry.Flush(2 * time.Second)
	}
	if err := users.InitDB(); err != nil {
		log.Fatal(err)
	}
	if bot, err = tgbotapi.NewBotAPI(os.Getenv("TELEGRAM_APITOKEN")); err != nil {
		log.Fatalf("Issue with token: %s", err)
	} else {
		schedule.Init(bot)
		bot.Debug = false
		log.Infof("Authorized on account %s", bot.Self.UserName)
		u := tgbotapi.NewUpdate(0)
		u.Timeout = 60
		updates = bot.GetUpdatesChan(u)
	}
	fatBotUpdate := FatBotUpdate{Bot: bot}
	for update := range updates {
		fatBotUpdate.Update = update
		if err := handleUpdates(fatBotUpdate); err != nil {
			log.Error(err)
			sentry.CaptureException(err)
		}
	}
}
