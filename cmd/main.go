package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	KP_TOKEN = "66XDW7D-CT0MF9Z-GZB7XYF-65WAXFN"
	TG_BOT_TOKEN = "8664604930:AAHIlnf3I1wsJzEfSAM3V0d9w0H6cBjVVJY"
	URL_RANDOM = "https://api.poiskkino.dev/v1.4/movie/random"
)

type UserData struct {
	State string
	Name  string
}

type Movie struct {
	Name        string
	Rating      float64
	Year        int
	Country     string
	Genre       string
	Description string
	Poster      string
	ID          int
	Type        string
}

func GetMovie(url string) (*Movie, error) {
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("X-API-KEY", KP_TOKEN)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var raw map[string]interface{}
	json.Unmarshal(body, &raw)

	movie := &Movie{
		Name: raw["name"].(string),
		Year: int(raw["year"].(float64)),
		Description: raw["description"].(string),
	}

	return movie, nil
}

func main() {
	bot, err := tgbotapi.NewBotAPI(TG_BOT_TOKEN)
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
			chatID := update.Message.Chat.ID
			text := update.Message.Text

			if text == "/random" {
				movie, err := GetMovie(URL_RANDOM)
				if err != nil {
					bot.Send(tgbotapi.NewMessage(chatID, "Ошибка API"))
					continue
				}

				msgText := fmt.Sprintf(
					"%s\nГод: %d\n%s",
					movie.Name,
					movie.Year,
					movie.Description,
				)

				msg := tgbotapi.NewMessage(chatID, msgText)
				bot.Send(msg)
			}
		}
		//if update.CallbackQuery != nil {
		//}
	}
}
