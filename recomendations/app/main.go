package main

import (
	"cmp"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

type Template struct {
	templates *template.Template
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

type Storage struct {
	Tags         map[string]string
	TagsByFilm   map[int64][]string
	TagToFilmIds map[string][]int64
	Films        map[int64]Film
}

func capitalizeFirst(s string) string {
	r, size := utf8.DecodeRuneInString(s)
	if r == utf8.RuneError {
		return s
	}
	return string(unicode.ToUpper(r)) + s[size:]
}

func intersect(a, b []string) []string {
	set := make(map[string]struct{}, len(a))
	for _, s := range a {
		set[s] = struct{}{}
	}

	seen := make(map[string]struct{})
	var result []string

	for _, s := range b {
		if _, ok := set[s]; ok {
			if _, exists := seen[s]; !exists {
				result = append(result, s)
				seen[s] = struct{}{}
			}
		}
	}

	return result
}

func ContainsAll(a []string, b []string) bool {
	set := make(map[string]struct{}, len(a))

	for _, v := range a {
		set[v] = struct{}{}
	}

	for _, v := range b {
		if _, ok := set[v]; !ok {
			return false
		}
	}

	return true
}

func loadStorage(
	filmsDir string,
	tagsDir string,
	minCountForTags int,
	minWeightForTags float64,
	maxTags int,
) (*Storage, error) {
	log.Println("Load films from ", filmsDir)
	filmFiles, err := filepath.Glob(filepath.Join(filmsDir, "*.json"))

	films := make(map[int64]Film)

	for _, file := range filmFiles {
		data, err := os.ReadFile(file)
		if err != nil {
			return nil, err
		}

		var film Film

		if err := json.Unmarshal(data, &film); err != nil {
			return nil, err
		}
		films[film.KinopoiskId] = film
	}

	log.Println("Films count ", len(films))

	log.Println("Load tags from ", tagsDir)
	tagsFiles, err := filepath.Glob(filepath.Join(tagsDir, "*.json"))
	if err != nil {
		return nil, err
	}

	tagToFilms := make(map[string][]int64)
	tagsByFilm := make(map[int64][]string)
	uniqTags := make(map[string]struct{})

	for _, file := range tagsFiles {

		data, err := os.ReadFile(file)
		if err != nil {
			return nil, err
		}

		var review Review

		if err := json.Unmarshal(data, &review); err != nil {
			return nil, err
		}

		if len(review.Keywords) == 0 {
			continue
		}

		filtered := slices.Collect(func(yield func(Keyword) bool) {
			for _, kw := range review.Keywords {
				if kw.Count < minCountForTags {
					continue
				}

				if kw.Weight < minWeightForTags {
					continue
				}

				yield(kw)
			}
		})

		slices.SortFunc(filtered, func(i, j Keyword) int {
			return cmp.Compare(j.Count, i.Count)
		})

		if len(filtered) > maxTags {
			filtered = filtered[:maxTags]
		}

		for _, kw := range filtered {
			uniqTags[kw.Emotion] = struct{}{}

			tagToFilms[kw.Emotion] = append(
				tagToFilms[kw.Emotion],
				review.KinopoiskID,
			)

			tagsByFilm[review.KinopoiskID] = append(tagsByFilm[review.KinopoiskID], kw.Emotion)
		}
	}

	tags := make(map[string]string)
	for t := range uniqTags {
		tags[t] = capitalizeFirst(t)
	}

	return &Storage{
		Tags:         tags,
		TagsByFilm:   tagsByFilm,
		TagToFilmIds: tagToFilms,
		Films:        films,
	}, nil
}

func main() {
	tagsDir := os.Getenv("EMOTIONS_DIR")
	if tagsDir == "" {
		tagsDir = "../../reviews/emotions/kinopoisk"
	}

	filmsDir := os.Getenv("FILMS_DIR")
	if filmsDir == "" {
		filmsDir = "../../reviews/kinopoisk/films"
	}

	serverPort := 8080
	serverPortEnv := os.Getenv("SERVER_PORT")
	if serverPortEnv != "" {
		var err error
		serverPort, err = strconv.Atoi(serverPortEnv)
		if err != nil {
			log.Fatal("Invalid SERVER_PORT", err)
		}
	}

	const minCountForTags = 2
	const minWeightForTags = 0.5
	const maxTags = 5
	const maxFilmsForShow = 20

	storage, err := loadStorage(
		filmsDir,
		tagsDir,
		minCountForTags,
		minWeightForTags,
		maxTags,
	)
	if err != nil {
		log.Fatal(err)
	}

	e := echo.New()

	e.Use(middleware.Recover())

	e.Renderer = &echo.TemplateRenderer{
		Template: template.Must(template.ParseGlob("templates/*.html")),
	}

	e.GET("/", func(c *echo.Context) error {
		return c.Render(http.StatusOK, "index.html", map[string]any{
			"Tags": storage.Tags,
		})
	})

	e.GET("/films", func(c *echo.Context) error {
		tagParam := strings.Split(c.QueryParam("tag"), ",")
		tags := make([]string, 0, len(tagParam))
		for _, tag := range tagParam {
			if _, exists := storage.Tags[tag]; exists {
				tags = append(tags, tag)
			}
		}

		films := make([]Film, 0)
		filmsSet := make(map[int64]struct{})

		for _, tag := range tags {
			breakByMax := false

			ids := storage.TagToFilmIds[tag]
			for _, id := range ids {
				if !ContainsAll(storage.TagsByFilm[id], tags) {
					continue
				}

				if _, exists := filmsSet[id]; exists {
					continue
				}

				if film, ok := storage.Films[id]; ok {
					films = append(films, film)
					filmsSet[id] = struct{}{}
				}

				if len(films) >= maxFilmsForShow {
					breakByMax = true
					break
				}
			}

			if breakByMax {
				break
			}
		}

		err := c.Render(
			http.StatusOK,
			"films.html",
			map[string]any{
				"Films":    films,
				"Emotions": storage.TagsByFilm,
			},
		)
		if err != nil {
			fmt.Println(err)
		}

		return err
	})

	err = e.Start(":" + fmt.Sprintf("%d", serverPort))
	if err != nil {
		e.Logger.Error("start error:" + err.Error())
	}
}
