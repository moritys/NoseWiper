package api

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/moritys/nosewiper/internal/models"
)

func GetMovieFromURL(url string, token string) (*models.Movie, error) {
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("X-API-KEY", token)

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

	movie := &models.Movie{}

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
