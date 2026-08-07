package models

type FilmType string

const (
	FilmTypeFilm       FilmType = "FILM"
	FilmTypeVideo      FilmType = "VIDEO"
	FilmTypeTvSeries   FilmType = "TV_SERIES"
	FilmTypeMiniSeries FilmType = "MINI_SERIES"
	FilmTypeTvShow     FilmType = "TV_SHOW "
)

type Tag struct {
	Id      string
	Name    string
	Enabled bool
}

// see llm-reviews/keywords/types.go
type Keyword struct {
	Emotion string  `json:"emotion"`
	Count   int     `json:"count"`
	Weight  float64 `json:"weight"`
}

// see llm-reviews/keywords/types.go
type FilmKeywords struct {
	KinopoiskID int64     `json:"kinopoiskId"`
	Keywords    []Keyword `json:"keywords"`
}

// Film, see https://kinopoiskapiunofficial.tech/documentation/api/#/films/get_api_v2_2_films__id_
type Film struct {
	KinopoiskId     int64    `json:"kinopoiskId"`
	WebUrl          string   `json:"webUrl"`
	NameRu          string   `json:"nameRu"`
	Year            int      `json:"year"`
	RatingKinopoisk float64  `json:"ratingKinopoisk"`
	Type            FilmType `json:"type"`
}
