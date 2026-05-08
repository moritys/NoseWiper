package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/moritys/nosewiper/internal/db"
	"github.com/moritys/nosewiper/internal/handlers"
)

var (
	TGToken string
	KPToken string
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Ошибка загрузки .env")
	}

	TGToken = os.Getenv("TG_BOT_TOKEN")
	KPToken = os.Getenv("KP_TOKEN")

	err = db.InitDB()
	if err != nil {
		log.Panic(err)
	}

	bot, err := tgbotapi.NewBotAPI(TGToken)
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
			handlers.HandleMessage(bot, update, KPToken)
		}

		if update.CallbackQuery != nil {
			handlers.HandleCallback(bot, update, KPToken)
		}
	}
}
