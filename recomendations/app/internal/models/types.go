package models

type Tag struct {
	Id      string
	Name    string
	Enabled bool
}

type Keyword struct {
	Emotion string  `json:"emotion"`
	Count   int     `json:"count"`
	Weight  float64 `json:"weight"`
}

type Review struct {
	KinopoiskID int64     `json:"kinopoiskId"`
	Keywords    []Keyword `json:"keywords"`
}

type Film struct {
	KinopoiskId     int64   `json:"kinopoiskId"`
	WebUrl          string  `json:"webUrl"`
	NameRu          string  `json:"nameRu"`
	Year            int     `json:"year"`
	RatingKinopoisk float64 `json:"ratingKinopoisk"`
	Emotions        []string
}
