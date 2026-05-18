package db

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

func InitDB() error {
	var err error
	DB, err = sql.Open("sqlite3", "movies.db")
	if err != nil {
		return err
	}

	movieTable := `
	CREATE TABLE IF NOT EXISTS movies (
		id INTEGER PRIMARY KEY,
		name TEXT,
		year INTEGER,
		description TEXT,
		poster TEXT,
		rating REAL,
		countries TEXT,
		genres TEXT
	);
	`

	createTable := `
	CREATE TABLE IF NOT EXISTS user_movies (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		chat_id INTEGER,
		movie_id INTEGER,
		UNIQUE(chat_id, movie_id)
	);`

	_, err = DB.Exec(movieTable)
	if err != nil {
		return err
	}
	_, err = DB.Exec(createTable)
	return err
}

func AddMovie(chatID int64, movieID int) error {
	query := `
	INSERT INTO user_movies (chat_id, movie_id)
	VALUES (?, ?)
	`
	_, err := DB.Exec(query, chatID, movieID)
	return err
}

func DeleteMovie(chatID int64, movieID int) error {
	query := `
	DELETE FROM user_movies
	WHERE chat_id = ? AND movie_id = ?
	`
	_, err := DB.Exec(query, chatID, movieID)
	return err
}

func GetUserMovies(chatID int64) ([]int, error) {
	query := `
	SELECT movie_id FROM user_movies
	WHERE chat_id = ?
	`

	rows, err := DB.Query(query, chatID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var movies []int

	for rows.Next() {
		var movieID int

		err := rows.Scan(&movieID)
		if err != nil {
			return nil, err
		}

		movies = append(movies, movieID)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return movies, nil
}
