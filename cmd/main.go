package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

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

	q := req.URL.Query()
	q.Add("notNullFields", "name")
	q.Add("notNullFields", "poster.url")
	q.Add("notNullFields", "description")
	q.Add("notNullFields", "rating.kp")
	req.URL.RawQuery = q.Encode()

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var raw map[string]interface{}
	if err := json.Unmarshal(body, &raw); err != nil {
		log.Fatal(err)
	}
	
	movie := &Movie{}

	name := "-"

	if val, ok := raw["name"].(string); ok && val != "" {
		name = val
	} else if val, ok := raw["alternativeName"].(string); ok && val != "" {
		name = val
	}
	movie.Name = name

	if year, ok := raw["year"].(float64); ok {
		movie.Year = int(year)
	}

	if desc, ok := raw["description"].(string); ok {
		movie.Description = desc
	}

	if id, ok := raw["id"].(float64); ok {
		movie.ID = int(id)
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

				button := tgbotapi.NewInlineKeyboardButtonData(
					"💿 Добавить в коллекцию", 
					fmt.Sprintf("add_random %d", movie.ID),
				)

				keyboard := tgbotapi.NewInlineKeyboardMarkup(
					tgbotapi.NewInlineKeyboardRow(button),
				)

				msg := tgbotapi.NewMessage(chatID, msgText)
				msg.ReplyMarkup = keyboard

				bot.Send(msg)
			}
		}
		if update.CallbackQuery != nil {
			data := update.CallbackQuery.Data

			parts := strings.Split(data, " ")
			if len(parts) < 2 {
				return
			}
			action := parts[0]

			if action == "add_random" {
				movieID := parts[1]

				msg := tgbotapi.NewMessage(
					update.CallbackQuery.Message.Chat.ID,
					"Фильм добавлен: "+movieID,
				)
				bot.Send(msg)
			}

			callback := tgbotapi.NewCallback(update.CallbackQuery.ID, "")
			bot.Request(callback)
		}
	}
}
