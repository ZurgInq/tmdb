package main

import (
	"html/template"
	"log"
	"os"
	"slices"
	"strconv"

	getfilms "app/internal/actions/get_films"
	"app/internal/routes"
	"app/internal/storage"
)

const (
	// load tags settings
	minCountForTags  = 2
	minWeightForTags = 0.5
	maxTagsPerFilm   = 20
	minTagsPerFilm   = 5
	//
	maxFilmsForShow = 250
)

func fatalIfErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
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
		maxTagsPerFilm,
		minTagsPerFilm,
	)
	fatalIfErr(err)

	templates := template.Must(template.ParseGlob("templates/index/*"))
	templates = template.Must(templates.ParseGlob("templates/*.html"))

	log.Println("Templates: ", slices.Collect(func(yield func(string) bool) {
		for _, t := range templates.Templates() {
			yield(t.Name())
		}
	}))

	getFilmAction := getfilms.New(db, maxFilmsForShow)

	r := routes.New(templates, db, apiHost)

	if renderStatic != "" {
		if renderStatic == "1" {
			renderStatic = "../static"
		}

		fileInfo, err := os.Stat(renderStatic)
		fatalIfErr(err)

		if !fileInfo.IsDir() {
			log.Fatalf("%s is not dir", fileInfo.Name())
		}

		fatalIfErr(r.RenderIndex(renderStatic))
		fatalIfErr(r.RenderIndexDemo(renderStatic, getFilmAction))

		log.Printf("Files saved to dir: %s\n", renderStatic)

		return
	}

	fatalIfErr(r.Start(serverPort, getFilmAction))
}
