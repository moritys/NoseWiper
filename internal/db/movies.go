package db

import "github.com/moritys/nosewiper/internal/models"

func SaveMovie(movie *models.Movie) error {
	query := `
	INSERT OR IGNORE INTO movies (
		id,
		name,
		year,
		description,
		poster,
		rating,
		countries,
		genres
	)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := DB.Exec(
		query,
		movie.ID,
		movie.Name,
		movie.Year,
		movie.Description,
		movie.Poster,
		movie.Rating,
		movie.Countries,
		movie.Genres,
	)

	return err
}

func GetMovieByID(movieID int) (*models.Movie, error) {
	query := `
	SELECT
		id,
		name,
		year,
		description,
		poster,
		rating,
		countries,
		genres
	FROM movies
	WHERE id = ?
	`

	row := DB.QueryRow(query, movieID)

	movie := &models.Movie{}

	err := row.Scan(
		&movie.ID,
		&movie.Name,
		&movie.Year,
		&movie.Description,
		&movie.Poster,
		&movie.Rating,
		&movie.Countries,
		&movie.Genres,
	)

	if err != nil {
		return nil, err
	}

	return movie, nil
}
