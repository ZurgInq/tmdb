package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"

	getfilms "app/internal/actions/get_films"
	"app/internal/storage"
)

const (
	minCountForTags    = 2
	minWeightForTags   = 0.5
	maxShowTagsForFilm = 5
	maxFilmsForShow    = 20
	minTagsPerFilm     = 5
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

	serverPort := 8080
	serverPortEnv := os.Getenv("SERVER_PORT")
	if serverPortEnv != "" {
		var err error
		serverPort, err = strconv.Atoi(serverPortEnv)
		if err != nil {
			log.Fatal("Invalid SERVER_PORT", err)
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

	e := echo.New()

	e.Use(middleware.Recover())

	e.Renderer = &echo.TemplateRenderer{
		Template: template.Must(template.ParseGlob("templates/*.html")),
	}

	actions := getfilms.New(db, maxFilmsForShow)

	e.GET("/", func(c *echo.Context) error {
		err := c.Render(http.StatusOK, "index.html", map[string]any{
			"Tags":         db.EnabledTags,
			"TagToFilmIds": db.TagToFilmIds,
		})
		if err != nil {
			log.Println(err)
		}

		return err
	})

	e.GET("/films", func(c *echo.Context) error {
		tagParam := strings.Split(c.QueryParam("tag"), ",")

		result, err := actions.Do(tagParam)
		if err != nil {
			log.Println(err)
		}

		if err := c.Render(
			http.StatusOK,
			"films.html",
			result,
		); err != nil {
			log.Println(err)
		}

		return err
	})

	err = e.Start(":" + fmt.Sprintf("%d", serverPort))
	if err != nil {
		e.Logger.Error("start error:" + err.Error())
	}
}
