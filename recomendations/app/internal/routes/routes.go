package routes

import (
	"app/internal/storage"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"

	getfilms "app/internal/actions/get_films"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

type GetFilmsAction interface {
	Do(filterTags []string) (getfilms.Result, error)
}

type routes struct {
	templates *template.Template
	db        *storage.Storage
	apiHost   string
}

func New(
	templates *template.Template,
	db *storage.Storage,
	apiHost string,
) *routes {
	return &routes{
		templates: templates,
		db:        db,
		apiHost:   apiHost,
	}
}

func (r *routes) Start(
	serverPort int,
	getFilms GetFilmsAction,
) error {
	e := echo.New()
	e.Use(middleware.Recover())

	e.Renderer = &echo.TemplateRenderer{
		Template: r.templates,
	}

	e.GET("/", func(c *echo.Context) error {
		err := c.Render(http.StatusOK, "index.html", map[string]any{
			"Tags":         r.db.EnabledTags,
			"TagToFilmIds": r.db.TagToFilmIds,
			"ApiHost":      r.apiHost,
		})
		if err != nil {
			log.Println(err)
		}

		return err
	})

	e.GET("/films", func(c *echo.Context) error {
		tagParam := strings.Split(c.QueryParam("tag"), ",")

		result, err := getFilms.Do(tagParam)
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

	return e.Start(":" + fmt.Sprintf("%d", serverPort))
}
