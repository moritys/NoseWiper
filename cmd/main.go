// 5902172202:AAHewEIuOCNWpnyfWyVkiY1o7i8RlJrXFKU
package main

import (
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func main() {
	bot, err := tgbotapi.NewBotAPI("5902172202:AAHewEIuOCNWpnyfWyVkiY1o7i8RlJrXFKU")
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
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Привет!")
			bot.Send(msg)
		}
	}
}