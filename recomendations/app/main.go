package main

import (
	"html/template"
	"log"
	"os"
	"strconv"

	getfilms "app/internal/actions/get_films"
	"app/internal/routes"
	"app/internal/storage"
)

const (
	// tags settings
	minCountForTags    = 2
	minWeightForTags   = 0.5
	maxShowTagsForFilm = 5
	minTagsPerFilm     = 5
	//
	maxFilmsForShow = 20
)

type Template struct {
	templates *template.Template
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

	renderStatic := os.Getenv("RENDER_STATIC")
	apiHost := os.Getenv("API_HOST")

	serverPort := 8080
	if renderStatic == "" {
		serverPortEnv := os.Getenv("SERVER_PORT")
		if serverPortEnv != "" {
			var err error
			serverPort, err = strconv.Atoi(serverPortEnv)
			if err != nil {
				log.Fatal("Invalid SERVER_PORT", err)
			}
		}
	}

	db, err := storage.Load(
		filmsDir,
		tagsDir,
		minCountForTags,
		minWeightForTags,
		maxShowTagsForFilm,
		minTagsPerFilm,
	)
	if err != nil {
		log.Fatal(err)
	}

	templates := template.Must(template.ParseGlob("templates/*.html"))
	actions := getfilms.New(db, maxFilmsForShow)

	r := routes.New(templates, db, apiHost)

	if renderStatic != "" {
		err = r.RenderIndex(renderStatic, actions)
	} else {
		err = r.Start(serverPort, actions)
	}

	if err != nil {
		log.Fatal(err)
	}
}
