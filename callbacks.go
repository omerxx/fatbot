package main

import (
	"fatbot/users"
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/log"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func handleCallbacks(fatBotUpdate FatBotUpdate) error {
	update := fatBotUpdate.Update
	bot := fatBotUpdate.Bot
	msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "")
	switch update.CallbackQuery.Message.Text {
	case "Pick a user to delete last workout for":
		callback := tgbotapi.NewCallback(update.CallbackQuery.ID, update.CallbackQuery.Data)
		if _, err := bot.Request(callback); err != nil {
			panic(err)
		}
		userId, _ := strconv.ParseInt(update.CallbackQuery.Data, 10, 64)
		user, err := users.GetUserById(userId)
		if err != nil {
			return err
		}
		if newLastWorkout, err := user.RollbackLastWorkout(); err != nil {
			return err
		} else {
			message := fmt.Sprintf("Deleted last workout for user %s\nRolledback to: %s",
				user.GetName(), newLastWorkout.CreatedAt.Format("2006-01-02 15:04:05"))
			msg.Text = message
			if _, err := bot.Send(msg); err != nil {
				return err
			}
			messageToUser := tgbotapi.NewMessage(0,
				fmt.Sprintf("Your last workout was cancelled by the admin.\nUpdated workout: %s",
					newLastWorkout.CreatedAt))
			user.SendPrivateMessage(bot, messageToUser)
		}
	case "Pick a user to rename":
		callback := tgbotapi.NewCallback(update.CallbackQuery.ID, update.CallbackQuery.Data)
		if _, err := bot.Request(callback); err != nil {
			panic(err)
		}
		msg.Text = fmt.Sprintf("/admin_rename %s newname", update.CallbackData())
		if _, err := bot.Send(msg); err != nil {
			return err
		}
	case "Pick a user to change workout for":
		callback := tgbotapi.NewCallback(update.CallbackQuery.ID, update.CallbackQuery.Data)
		if _, err := bot.Request(callback); err != nil {
			panic(err)
		}
		msg.Text = fmt.Sprintf("/admin_push_workout %s days", update.CallbackData())
		if _, err := bot.Send(msg); err != nil {
			return err
		}
	}
	if strings.Contains(update.CallbackQuery.Message.Text, "rejoin his group do you approve") {
		if update.CallbackQuery.Data == "false" {
			msg.Text = "Declined the request"
		} else {
			userId, _ := strconv.ParseInt(update.CallbackData(), 10, 64)
			user, err := users.GetUser(uint(userId))
			if err != nil {
				return err
			}
			if err := user.UnBan(fatBotUpdate.Bot); err != nil {
				return fmt.Errorf("Issue with unbanning %s: %s", user.GetName(), err)
			}
			if err := user.InviteExistingUser(fatBotUpdate.Bot); err != nil {
				return fmt.Errorf("Issue with inviting %s: %s", user.GetName(), err)
			}
			if err := user.UpdateActive(true); err != nil {
				return fmt.Errorf("Issue updating active %s: %s", user.GetName(), err)
			}
			if err := user.UpdateOnProbation(true); err != nil {
				return fmt.Errorf("Issue updating probation %s: %s", user.GetName(), err)
			}
			msg.Text = "Ok, approved"
		}
		if _, err := bot.Send(msg); err != nil {
			return err
		}
	} else if strings.Contains(update.CallbackQuery.Message.Text, "new and wants to join a group") {
		dataSlice := strings.Split(update.CallbackData(), " ")
		userId, _ := strconv.ParseInt(dataSlice[1], 10, 64)
		if dataSlice[0] == "block" {
			msg.Text = "Blocked"
			if err := users.BlockUserId(userId); err != nil {
				log.Error(err)
			}
		} else {
			chatId, _ := strconv.ParseInt(dataSlice[0], 10, 64)
			name := dataSlice[2]
			username := dataSlice[3]
			user := users.User{
				Username:       username,
				Name:           name,
				ChatID:         chatId,
				TelegramUserID: userId,
				Active:         true,
			}
			if err := user.InviteNewUser(fatBotUpdate.Bot); err != nil {
				log.Error(fmt.Errorf("Issue with inviting: %s", err))
			}
			msg.Text = "Invitation sent"
		}
		if _, err := bot.Send(msg); err != nil {
			return err
		}
	}
	return nil
}
