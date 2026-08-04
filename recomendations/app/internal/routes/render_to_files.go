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

func (r *routes) RenderIndex(
	dirOutput string,
) error {
	return renderToFile(
		r.templates,
		"index.html",
		map[string]any{
			"Tags":         r.db.EnabledTags,
			"TagToFilmIds": r.db.TagToFilmIds,
			"ApiHost":      r.apiHost,
		},
		filepath.Join(dirOutput, "index.html"),
	)
}

func (r *routes) RenderIndexDemo(
	dirOutput string,
	getFilms GetFilmsAction,
) error {
	films := make([]*models.Film, 0, len(r.db.Films))
	for _, f := range r.db.Films {
		films = append(films, f)
	}

	return renderToFile(
		r.templates,
		"index_demo.html",
		map[string]any{
			"Tags":         r.db.EnabledTags,
			"TagToFilmIds": r.db.TagToFilmIds,
			"FilmsData": getfilms.Result{
				Films:    films,
				Emotions: r.db.TagsByFilm,
			},
		},
		filepath.Join(dirOutput, "index_demo.html"),
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
