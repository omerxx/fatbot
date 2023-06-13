package schedule

import (
	"fatbot/users"
	"fmt"
	"time"

	"github.com/charmbracelet/log"
	"github.com/getsentry/sentry-go"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func scanUsers(bot *tgbotapi.BotAPI) error {
	groups := users.GetGroupsWithUsers()
	const totalDays = 5.0
	for _, group := range groups {
		for _, user := range group.Users {
			if !user.Active {
				continue
			}
			if user.OnProbation {
				handleProbation(bot, user, group, totalDays)
				continue
			}
			if isNew, err := user.IsNewToday(group.ChatID); err != nil {
				return err
			} else if isNew {
				continue
			}
			lastWorkout, err := user.GetLastXWorkout(1, group.ChatID)
			if err != nil {
				log.Errorf("Err getting last workout for user %s: %s", user.GetName(), err)
				sentry.CaptureException(err)
			}
			diffHours := int(totalDays*24 - time.Now().Sub(lastWorkout.CreatedAt).Hours())
			if diffHours == 24 {
				msg := tgbotapi.NewMessage(group.ChatID, fmt.Sprintf("[%s](tg://user?id=%d) you have 24 hours left",
					user.GetName(),
					user.TelegramUserID))
				msg.ParseMode = "MarkdownV2"
				bot.Send(msg)
				if err := user.RegisterLastDayNotificationEvent(); err != nil {
					log.Errorf("Error while registering ban event: %s", err)
					sentry.CaptureException(err)
				}
			} else if diffHours < 0 {
				if err := user.Ban(bot, group.ChatID); err != nil {
					log.Errorf("Issue banning %s from %d: %s", user.GetName(), group.ChatID, err)
					continue
				}
			}
		}
	}
	return nil
}

func handleProbation(bot *tgbotapi.BotAPI, user users.User, group users.Group, totalDays float64) {
	lastWorkout, err := user.GetLastXWorkout(2, group.ChatID)
	if err != nil {
		log.Errorf("Err getting last 2 workout for user %s: %s", user.GetName(), err)
		sentry.CaptureException(err)
	}
	diffHours := int(totalDays*24 - time.Now().Sub(lastWorkout.CreatedAt).Hours())
	rejoinedLastHour := time.Now().Sub(user.UpdatedAt).Minutes() <= 60
	lastTwoWorkoutsOk := diffHours > 0
	if !lastTwoWorkoutsOk && !rejoinedLastHour {
		if errors := user.Ban(bot, group.ChatID); errors != nil {
			log.Errorf("Issue banning %s from %d: %s", user.GetName(), group.ChatID, errors)
			sentry.CaptureException(err)
		}
	} else if lastTwoWorkoutsOk {
		if err := user.UpdateOnProbation(false); err != nil {
			log.Errorf("Issue updating unprobation %s from %d: %s", user.GetName(), group.ChatID, err)
			sentry.CaptureException(err)
		}
	}
}
