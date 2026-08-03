package routes

import (
	"fmt"
	"html/template"
	"os"
	"path/filepath"
)

func (r *routes) RenderIndex(
	dirOutput string,
	getFilms GetFilmsAction,
) error {
	fileInfo, err := os.Stat(dirOutput)
	if err != nil {
		return err
	}

	if !fileInfo.IsDir() {
		return fmt.Errorf("%s is not dir", fileInfo.Name())
	}

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

func renderToFile(
	tmpl *template.Template,
	name string,
	data any,
	filename string,
) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	return tmpl.ExecuteTemplate(f, name, data)
}
