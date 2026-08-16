package main

import (
	"html/template"
	"log"
	"os"
	"path/filepath"
	"slices"
	"strconv"

	getfilms "app/internal/actions/get_films"
	"app/internal/models"
	"app/internal/routes"
	"app/internal/storage"
)

const (
	// load tags settings
	minCountForTags  = 2
	minWeightForTags = 0.4
	maxTagsPerFilm   = 20
	minTagsPerFilm   = 5
	//
	maxFilmsForShow = 250
)

func main() {
	tagsDir := os.Getenv("EMOTIONS_DIR")
	if tagsDir == "" {
		tagsDir = "../../reviews/emotions/kinopoisk"
	}

	filmsDir := os.Getenv("FILMS_DIR")
	if filmsDir == "" {
		filmsDir = "../../reviews/kinopoisk/films"
	}

	renderDemo := os.Getenv("RENDER_DEMO")
	apiHost := os.Getenv("API_HOST")

	serverPort := 8080
	if renderDemo == "" {
		serverPortEnv := os.Getenv("SERVER_PORT")
		if serverPortEnv != "" {
			var err error
			serverPort, err = strconv.Atoi(serverPortEnv)
			if err != nil {
				log.Fatal("Invalid SERVER_PORT", err)
			}
		}
	}

	if renderDemo != "" {
		if renderDemo == "1" {
			renderDemo = filepath.Join("..", "static", "demo")
		}

		fileInfo, err := os.Stat(renderDemo)
		fatalIfErr(err)

		if !fileInfo.IsDir() {
			log.Fatalf("%s is not dir", fileInfo.Name())
		}

		buildDemoPages(filmsDir, tagsDir, renderDemo)

		return
	}

	db := mustGetStorage(
		filmsDir,
		tagsDir,
		nil,
	)

	templates := template.Must(template.ParseGlob("templates/partials/*"))
	templates = template.Must(templates.ParseGlob("templates/*.html"))

	log.Println("Templates: ", slices.Collect(func(yield func(string) bool) {
		for _, t := range templates.Templates() {
			yield(t.Name())
		}
	}))

	r := routes.New(templates, db, apiHost)

	getFilsmAction := getfilms.New(db, maxFilmsForShow)
	fatalIfErr(r.Start(serverPort, getFilsmAction))
}

func fatalIfErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func mustGetStorage(
	filmsDir string,
	tagsDir string,
	filmTypes []models.FilmType,
) *storage.Storage {
	db, err := storage.Load(
		storage.LoadParams{
			FilmsDir:         filmsDir,
			TagsDir:          tagsDir,
			FilmTypes:        filmTypes,
			MinCountForTags:  minCountForTags,
			MinWeightForTags: minWeightForTags,
			MaxTagsPerFilm:   maxTagsPerFilm,
			MinTagsPerFilm:   minTagsPerFilm,
		},
	)
	fatalIfErr(err)

	return db
}

func buildDemoPages(filmsDir string, tagsDir string, outputDir string) {
	templates := template.Must(template.ParseGlob("templates/partials/*"))
	templates = template.Must(templates.ParseGlob("templates/demo/*.html"))

	log.Println("Templates: ", slices.Collect(func(yield func(string) bool) {
		for _, t := range templates.Templates() {
			yield(t.Name())
		}
	}))

	dbFilms := mustGetStorage(
		filmsDir,
		tagsDir,
		[]models.FilmType{models.FilmTypeFilm},
	)

	dbSeries := mustGetStorage(
		filmsDir,
		tagsDir,
		[]models.FilmType{
			models.FilmTypeTvSeries,
			models.FilmTypeMiniSeries,
		},
	)

	r := routes.New(templates, nil, "")

	fatalIfErr(r.WithDB(dbFilms).RenderFilmsDemo(outputDir))
	fatalIfErr(r.WithDB(dbSeries).RenderSeriesDemo(outputDir))

	log.Printf("Files saved to dir: %s\n", outputDir)
}
