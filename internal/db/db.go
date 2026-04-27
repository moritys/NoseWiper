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

	createTable := `
	CREATE TABLE IF NOT EXISTS user_movies (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		chat_id INTEGER,
		movie_id INTEGER
	);`

	_, err = DB.Exec(createTable)
	return err
}
