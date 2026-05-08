package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/moritys/nosewiper/internal/db"
)

const (
	URL_RANDOM    = "https://api.poiskkino.dev/v1.4/movie/random"
	URL_GET_MOVIE = "https://api.poiskkino.dev/v1.4/movie/"
)

var (
	TGToken string
	KPToken string
)

type Movie struct {
	ID          int
	Name        string
	Year        int
	Description string
	Poster      string
	Rating      float64
	Countries   string
	Genres      string
	Type        string
}

type UserState struct {
	State string
}

func GetMovieFromURL(url string) (*Movie, error) {
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("X-API-KEY", KPToken)

	q := req.URL.Query()
	q.Add("notNullFields", "name")
	q.Add("notNullFields", "poster.url")
	q.Add("notNullFields", "description")
	q.Add("notNullFields", "rating.kp")
	q.Add("rating.kp", "6-10")
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
		return nil, err
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

	if rating, ok := raw["rating"].(map[string]interface{}); ok {
		if kp, ok := rating["kp"].(float64); ok {
			movie.Rating = kp
		}
	}

	if genres, ok := raw["genres"].([]interface{}); ok {
		var list []string
		for _, g := range genres {
			if genreMap, ok := g.(map[string]interface{}); ok {
				if name, ok := genreMap["name"].(string); ok {
					list = append(list, name)
				}
			}
		}
		movie.Genres = strings.Join(list, ", ")
	}

	if countries, ok := raw["countries"].([]interface{}); ok {
		var list []string
		for _, c := range countries {
			if countryMap, ok := c.(map[string]interface{}); ok {
				if name, ok := countryMap["name"].(string); ok {
					list = append(list, name)
				}
			}
		}
		movie.Countries = strings.Join(list, ", ")
	}

	if poster, ok := raw["poster"].(map[string]interface{}); ok {
		if url, ok := poster["url"].(string); ok {
			movie.Poster = url
		}
	}

	return movie, nil
}

func AddMovie(chatID int64, movieID int) error {
	query := `
	INSERT INTO user_movies (chat_id, movie_id)
	VALUES (?, ?)
	`
	_, err := db.DB.Exec(query, chatID, movieID)
	return err
}

func DeleteMovie(chatID int64, movieID int) error {
	query := `
	DELETE FROM user_movies
	WHERE chat_id = ? AND movie_id = ?
	`
	_, err := db.DB.Exec(query, chatID, movieID)
	return err
}

func GetUserMovies(chatID int64) ([]int, error) {
	query := `
	SELECT movie_id FROM user_movies
	WHERE chat_id = ?
	`

	rows, err := db.DB.Query(query, chatID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var movies []int

	for rows.Next() {
		var movieID int
		rows.Scan(&movieID)
		movies = append(movies, movieID)
	}

	return movies, nil
}

func sendMovie(bot *tgbotapi.BotAPI, chatID int64, movie *Movie) {
	text := fmt.Sprintf(
		"*%s*\n`Рейтинг: %.1f`\n\n*Жанр:* %s\n_(%s, %d)_\n----\n%s",
		movie.Name,
		movie.Rating,
		movie.Genres,
		movie.Countries,
		movie.Year,
		movie.Description,
	)

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("💿 Добавить", fmt.Sprintf("add_random %d", movie.ID)),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🎲 Ещё фильм", "more_random"),
		),
	)

	if len(text) > 1024 {
		bot.Send(tgbotapi.NewPhoto(chatID, tgbotapi.FileURL(movie.Poster)))

		msg := tgbotapi.NewMessage(chatID, text)
		msg.ParseMode = "Markdown"
		msg.ReplyMarkup = keyboard
		bot.Send(msg)
	} else {
		msg := tgbotapi.NewPhoto(chatID, tgbotapi.FileURL(movie.Poster))
		msg.Caption = text
		msg.ParseMode = "Markdown"
		msg.ReplyMarkup = keyboard
		bot.Send(msg)
	}
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Ошибка загрузки .env")
	}

	TGToken = os.Getenv("TG_BOT_TOKEN")
	KPToken = os.Getenv("KP_TOKEN")
	users := make(map[int64]*UserState)

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
			chatID := update.Message.Chat.ID
			text := update.Message.Text

			user, exists := users[chatID]

			if !exists {
				user = &UserState{}
				users[chatID] = user
			}

			keyboard := tgbotapi.NewReplyKeyboard(
				tgbotapi.NewKeyboardButtonRow(
					tgbotapi.NewKeyboardButton("/random"),
					tgbotapi.NewKeyboardButton("/wisechoice"),
				),
				tgbotapi.NewKeyboardButtonRow(
					tgbotapi.NewKeyboardButton("/add"),
					tgbotapi.NewKeyboardButton("/del"),
				),
			)

			if text == "/add" {
				user.State = "waiting_add"

				msg := tgbotapi.NewMessage(chatID, "Отправь ссылку на фильм с Кинопоиска:")
				bot.Send(msg)
				continue
			}

			if text == "/del" {
				user.State = "waiting_del"

				msg := tgbotapi.NewMessage(chatID, "Отправь ссылку для удаления:")
				bot.Send(msg)
				continue
			}

			if text == "/start" {
				msg := tgbotapi.NewMessage(chatID,
					"Привет!\n"+
						"Я помогу выбрать фильм 🎬\n\n"+
						"Вот что я умею:\n"+
						"/random — случайные фильмы с Кинопоиска\n"+
						"/wisechoice — фильм из твоей коллекции\n"+
						"/add — добавить фильм\n"+
						"/del — удалить фильм\n"+
						"Или просто нажми кнопку ниже 👇\n",
				)
				msg.ReplyMarkup = keyboard

				bot.Send(msg)
			}

			if text == "/about_random" {
				msg := tgbotapi.NewMessage(chatID,
					"Команда /random выбирает фильмы:\n"+
						"▸ рейтинг выше 6\n"+
						"▸ любые жанры",
				)
				bot.Send(msg)
			}

			if user.State == "waiting_add" {
				link := text

				parts := strings.Split(link, "/")
				if len(parts) < 2 {
					bot.Send(tgbotapi.NewMessage(
						chatID,
						"❌ Это не похоже на ссылку с Кинопоиска\n\nПопробуй ещё раз или нажми /cancel",
					))
					continue
				}

				movieIDStr := parts[len(parts)-2]
				movieID, err := strconv.Atoi(movieIDStr)
				if err != nil {
					bot.Send(tgbotapi.NewMessage(
						chatID,
						"❌ Не удалось извлечь ID\nПопробуй ещё раз или /cancel",
					))
					continue
				}

				err = AddMovie(chatID, movieID)
				if err != nil {
					bot.Send(tgbotapi.NewMessage(chatID, "Фильм уже добавлен"))
				} else {
					bot.Send(tgbotapi.NewMessage(chatID, "Фильм добавлен 🎬"))
				}

				user.State = ""
				continue
			}

			if text == "/cancel" {
				user.State = ""
				bot.Send(tgbotapi.NewMessage(chatID, "Ок, отменили 👌"))
				continue
			}

			if user.State == "waiting_del" {
				link := text

				parts := strings.Split(link, "/")
				if len(parts) < 2 {
					bot.Send(tgbotapi.NewMessage(chatID, "Неверная ссылка"))
					continue
				}

				movieIDStr := parts[len(parts)-2]
				movieID, err := strconv.Atoi(movieIDStr)
				if err != nil {
					bot.Send(tgbotapi.NewMessage(chatID, "Ошибка ID"))
					continue
				}

				err = DeleteMovie(chatID, movieID)
				if err != nil {
					bot.Send(tgbotapi.NewMessage(chatID, "Ошибка удаления"))
					continue
				} else {
					bot.Send(tgbotapi.NewMessage(chatID, "Фильм удален 🎬"))
				}

				user.State = ""
				continue
			}

			if text == "/random" {
				bot.Send(tgbotapi.NewMessage(chatID, "Ищу 🔍"))

				movie, err := GetMovieFromURL(URL_RANDOM)
				if err != nil {
					bot.Send(tgbotapi.NewMessage(chatID, "Ошибка API"))
					continue
				}

				sendMovie(bot, chatID, movie)
			}

			if text == "/wisechoice" {
				movies, err := GetUserMovies(chatID)
				if err != nil {
					bot.Send(tgbotapi.NewMessage(chatID, "Ошибка БД"))
					continue
				}

				if len(movies) == 0 {
					bot.Send(tgbotapi.NewMessage(chatID, "В коллекции нет фильмов 😔"))
					continue
				}

				randomIndex := rand.Intn(len(movies))
				movieID := movies[randomIndex]

				movie, err := GetMovieFromURL(fmt.Sprintf(URL_GET_MOVIE+"%d", movieID))
				if err != nil {
					bot.Send(tgbotapi.NewMessage(chatID, "Ошибка API"))
					continue
				}

				button := tgbotapi.NewInlineKeyboardButtonData(
					"🛋️ Буду смотреть это, удалить из коллекции",
					fmt.Sprintf("del_choiced %d", movie.ID),
				)

				keyboard := tgbotapi.NewInlineKeyboardMarkup(
					tgbotapi.NewInlineKeyboardRow(button),
				)

				msg := tgbotapi.NewMessage(chatID, "Твой фильм на сегодня: "+movie.Name)
				msg.ReplyMarkup = keyboard
				bot.Send(msg)
			}
		}

		if update.CallbackQuery != nil {
			data := update.CallbackQuery.Data

			parts := strings.Split(data, " ")

			action := parts[0]

			var movieID int

			if len(parts) > 1 {
				movieID, _ = strconv.Atoi(parts[1])
			}

			chatID := update.CallbackQuery.Message.Chat.ID

			if action == "add_random" {
				err := AddMovie(chatID, movieID)
				if err != nil {
					bot.Send(tgbotapi.NewMessage(chatID, "Фильм уже есть в коллекции"))
					continue
				}

				bot.Send(tgbotapi.NewMessage(chatID, "Фильм добавлен в коллекцию 🎬"))
			}

			if action == "del_choiced" {
				err := DeleteMovie(chatID, movieID)
				if err != nil {
					bot.Send(tgbotapi.NewMessage(chatID, "Ошибка удаления"))
					continue
				}

				bot.Send(tgbotapi.NewMessage(chatID, "Фильм удалён из коллекции 🎬"))
			}

			if action == "more_random" {
				movie, err := GetMovieFromURL(URL_RANDOM)
				if err != nil {
					bot.Send(tgbotapi.NewMessage(chatID, "Не удалось найти фильм 😢"))
					continue
				}

				sendMovie(bot, chatID, movie)
			}

			callback := tgbotapi.NewCallback(update.CallbackQuery.ID, "")
			bot.Request(callback)
		}
	}
}
