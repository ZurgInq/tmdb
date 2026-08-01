package keywords

type Keyword struct {
	Emotion string  `json:"emotion"`
	Count   int     `json:"count"`
	Weight  float64 `json:"weight"`
}

type Result struct {
	KinopoiskID int64     `json:"kinopoiskId"`
	Keywords    []Keyword `json:"keywords"`
}
