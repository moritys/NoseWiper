package handlers

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/moritys/nosewiper/internal/api"
	"github.com/moritys/nosewiper/internal/db"
	"github.com/moritys/nosewiper/internal/models"
	"github.com/moritys/nosewiper/internal/state"
)

const (
	URL_RANDOM    = "https://api.poiskkino.dev/v1.4/movie/random"
	URL_GET_MOVIE = "https://api.poiskkino.dev/v1.4/movie/"
)

func SendMovie(
	bot *tgbotapi.BotAPI,
	chatID int64,
	movie *models.Movie,
	mode string,
) {
	var keyboard tgbotapi.InlineKeyboardMarkup

	if mode == "random" {
		keyboard = tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("💿 Добавить", fmt.Sprintf("add_random %d", movie.ID)),
			),
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("🎲 Ещё фильм", "more_random"),
			),
		)
	} else {
		keyboard = tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("🗑 Удалить из коллекции", fmt.Sprintf("del_choiced %d", movie.ID)),
			),
		)
	}
	text := fmt.Sprintf(
		"*%s*\n`Рейтинг: %.1f`\n\n*Жанр:* %s\n_(%s, %d)_\n----\n%s",
		movie.Name,
		movie.Rating,
		movie.Genres,
		movie.Countries,
		movie.Year,
		movie.Description,
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

func HandleMessage(
	bot *tgbotapi.BotAPI,
	update tgbotapi.Update,
	kpToken string,
) {
	if update.Message != nil {
		chatID := update.Message.Chat.ID
		text := update.Message.Text

		user, exists := state.Users[chatID]

		if !exists {
			user = &state.UserState{}
			state.Users[chatID] = user
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

		if text == "/cancel" {
			user.State = ""
			bot.Send(tgbotapi.NewMessage(chatID, "Ок, отменили 👌"))
			return
		}

		if text == "/add" {
			user.State = "waiting_add"

			msg := tgbotapi.NewMessage(chatID, "Отправь ссылку на фильм с Кинопоиска:")
			bot.Send(msg)
			return
		}

		if text == "/del" {
			user.State = "waiting_del"

			msg := tgbotapi.NewMessage(chatID, "Отправь ссылку для удаления:")
			bot.Send(msg)
			return
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
				return
			}

			movieIDStr := parts[len(parts)-2]
			movieID, err := strconv.Atoi(movieIDStr)
			if err != nil {
				bot.Send(tgbotapi.NewMessage(
					chatID,
					"❌ Не удалось извлечь ID\nПопробуй ещё раз или /cancel",
				))
				return
			}

			err = db.AddMovie(chatID, movieID)
			if err != nil {
				bot.Send(tgbotapi.NewMessage(chatID, "Фильм уже добавлен"))
			} else {
				bot.Send(tgbotapi.NewMessage(chatID, "Фильм добавлен 🎬"))
			}

			user.State = ""
			return
		}

		if user.State == "waiting_del" {
			link := text

			parts := strings.Split(link, "/")
			if len(parts) < 2 {
				bot.Send(tgbotapi.NewMessage(chatID, "Неверная ссылка"))
				return
			}

			movieIDStr := parts[len(parts)-2]
			movieID, err := strconv.Atoi(movieIDStr)
			if err != nil {
				bot.Send(tgbotapi.NewMessage(chatID, "Ошибка ID"))
				return
			}

			err = db.DeleteMovie(chatID, movieID)
			if err != nil {
				bot.Send(tgbotapi.NewMessage(chatID, "Ошибка удаления"))
				return
			} else {
				bot.Send(tgbotapi.NewMessage(chatID, "Фильм удален 🎬"))
			}

			user.State = ""
			return
		}

		if text == "/random" {
			bot.Send(tgbotapi.NewMessage(chatID, "Ищу 🔍"))

			movie, err := api.GetMovieFromURL(URL_RANDOM, kpToken)
			if err != nil {
				bot.Send(tgbotapi.NewMessage(chatID, "Ошибка API"))
				return
			}

			SendMovie(bot, chatID, movie, "random")
		}

		if text == "/wisechoice" {
			bot.Send(tgbotapi.NewMessage(chatID, "🛋️ Твой фильм на сегодня:"))
			movies, err := db.GetUserMovies(chatID)
			if err != nil {
				bot.Send(tgbotapi.NewMessage(chatID, "Ошибка БД"))
				return
			}

			if len(movies) == 0 {
				bot.Send(tgbotapi.NewMessage(chatID, "В коллекции нет фильмов 😔"))
				return
			}

			randomIndex := rand.Intn(len(movies))
			movieID := movies[randomIndex]

			movie, err := api.GetMovieFromURL(fmt.Sprintf(URL_GET_MOVIE+"%d", movieID), kpToken)
			if err != nil {
				bot.Send(tgbotapi.NewMessage(chatID, "Ошибка API"))
				return
			}

			SendMovie(bot, chatID, movie, "collection")
		}
	}
}

func HandleCallback(
	bot *tgbotapi.BotAPI,
	update tgbotapi.Update,
	kpToken string,
) {
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
			err := db.AddMovie(chatID, movieID)
			if err != nil {
				bot.Send(tgbotapi.NewMessage(chatID, "Фильм уже есть в коллекции"))
				return
			}

			bot.Send(tgbotapi.NewMessage(chatID, "Фильм добавлен в коллекцию 🎬"))
		}

		if action == "del_choiced" {
			err := db.DeleteMovie(chatID, movieID)
			if err != nil {
				bot.Send(tgbotapi.NewMessage(chatID, "Ошибка удаления"))
				return
			}

			bot.Send(tgbotapi.NewMessage(chatID, "Фильм удалён из коллекции 🎬"))
		}

		if action == "more_random" {
			bot.Send(tgbotapi.NewMessage(chatID, "Ищу 🔍"))

			movie, err := api.GetMovieFromURL(URL_RANDOM, kpToken)
			if err != nil {
				bot.Send(tgbotapi.NewMessage(chatID, "Не удалось найти фильм 😢"))
				return
			}

			SendMovie(bot, chatID, movie, "random")
		}

		callback := tgbotapi.NewCallback(update.CallbackQuery.ID, "")
		bot.Request(callback)
	}
}
