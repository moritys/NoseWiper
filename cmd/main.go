package main

import (
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type UserData struct {
	State string
	Name  string
}

func main() {
	users := make(map[int64]*UserData)

	bot, err := tgbotapi.NewBotAPI("8664604930:AAHIlnf3I1wsJzEfSAM3V0d9w0H6cBjVVJY")
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true

	log.Printf("Авторизован как %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil {
			text := update.Message.Text
			chatID := update.Message.Chat.ID
			user, exists := users[chatID]
			if !exists {
				user = &UserData{}
				users[chatID] = user
			}

			if text == "/start" {
				user.State = "waiting_name"

				msg := tgbotapi.NewMessage(chatID, "Как тебя зовут?")
				bot.Send(msg)
				continue
			}

			if user.State == "waiting_name" {
				user.Name = text
				user.State = "waiting_age"

				msg := tgbotapi.NewMessage(chatID, "Сколько тебе лет?")
				bot.Send(msg)
				continue
			}

			if user.State == "waiting_age" {
				msg := tgbotapi.NewMessage(chatID, "Тебя зовут "+user.Name+", тебе "+text)
				bot.Send(msg)

				delete(users, chatID)
				continue
			}

			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Ты написал: "+update.Message.Text)
			bot.Send(msg)
		}
		if update.CallbackQuery != nil {
			data := update.CallbackQuery.Data

			msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "Ты нажал: "+data)
			bot.Send(msg)

			callback := tgbotapi.NewCallback(update.CallbackQuery.ID, "")
			bot.Request(callback)
		}
	}
}
