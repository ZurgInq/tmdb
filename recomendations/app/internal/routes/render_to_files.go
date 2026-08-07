package routes

import (
	"fmt"
	"html/template"
	"log"
	"os"
	"path/filepath"

	getfilms "app/internal/actions/get_films"
	"app/internal/models"
)

const tmplDemoIndex = "demo/index.html"

type indexDemoPageParams struct {
	Title string

	Tags         map[string]*models.Tag
	TagToFilmIds map[string][]int64
	FilmsData    getfilms.Result
	PageType     string
}

func (r *routes) RenderFilmsDemo(dirOutput string) error {
	films := make([]*models.Film, 0, len(r.db.Films))

	for _, f := range r.db.Films {
		if f.Type == models.FilmTypeFilm {
			films = append(films, f)
		}
	}

	return renderToFile(
		r.templates,
		tmplDemoIndex,
		indexDemoPageParams{
			Title:        "Фильмы",
			PageType:     "Films",
			Tags:         r.db.EnabledTags,
			TagToFilmIds: r.db.TagToFilmIds,
			FilmsData: getfilms.Result{
				Films:    films,
				Emotions: r.db.TagsByFilm,
			},
		},
		filepath.Join(dirOutput, "index.html"),
	)
}

func (r *routes) RenderSeriesDemo(dirOutput string) error {
	films := make([]*models.Film, 0, len(r.db.Films))
	for _, f := range r.db.Films {
		if f.Type == models.FilmTypeTvSeries ||
			f.Type == models.FilmTypeMiniSeries {
			films = append(films, f)
		}
	}

	return renderToFile(
		r.templates,
		tmplDemoIndex,
		indexDemoPageParams{
			Title:        "Сериалы",
			PageType:     "Series",
			Tags:         r.db.EnabledTags,
			TagToFilmIds: r.db.TagToFilmIds,
			FilmsData: getfilms.Result{
				Films:    films,
				Emotions: r.db.TagsByFilm,
			},
		},
		filepath.Join(dirOutput, "series.html"),
	)
}

func renderToFile(
	tmpl *template.Template,
	name string,
	data any,
	filename string,
) error {
	f, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("renderToFile %s: %w", filename, err)
	}
	defer func() {
		_ = f.Close()
	}()

	log.Printf("renderToFile: %s\n", filename)
	return tmpl.ExecuteTemplate(f, name, data)
}
