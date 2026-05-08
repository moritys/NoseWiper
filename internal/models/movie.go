package models

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
