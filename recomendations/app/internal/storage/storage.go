package storage

import (
	"cmp"
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"slices"
	"unicode"
	"unicode/utf8"

	"app/internal/models"
)

type Storage struct {
	Tags  map[string]*models.Tag // TagId => Tag
	Films map[int64]models.Film  // FilmId => Film
	// index
	EnabledTags  map[string]*models.Tag
	TagsByFilm   map[int64][]*models.Tag // filmId => Tag
	TagToFilmIds map[string][]int64      // TagId => filmdId[]
}

func capitalizeFirst(s string) string {
	r, size := utf8.DecodeRuneInString(s)
	if r == utf8.RuneError {
		return s
	}
	return string(unicode.ToUpper(r)) + s[size:]
}

func Load(
	filmsDir string,
	tagsDir string,
	minCountForTags int,
	minWeightForTags float64,
	maxShowTagsForFilm int,
	minTagsPerFilm int,
) (*Storage, error) {
	log.Println("Load films from ", filmsDir)
	filmFiles, err := filepath.Glob(filepath.Join(filmsDir, "*.json"))

	films := make(map[int64]models.Film)

	for _, file := range filmFiles {
		data, err := os.ReadFile(file)
		if err != nil {
			return nil, err
		}

		var film models.Film

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
	tagsByFilm := make(map[int64][]*models.Tag)
	uniqTags := make(map[string]struct{})

	tags := make(map[string]*models.Tag)
	for _, file := range tagsFiles {

		data, err := os.ReadFile(file)
		if err != nil {
			return nil, err
		}

		var review models.Review

		if err := json.Unmarshal(data, &review); err != nil {
			return nil, err
		}

		if len(review.Keywords) == 0 {
			continue
		}

		filtered := slices.Collect(func(yield func(models.Keyword) bool) {
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

		slices.SortFunc(filtered, func(i, j models.Keyword) int {
			return cmp.Compare(j.Count, i.Count)
		})

		if len(filtered) > maxShowTagsForFilm {
			// filtered = filtered[:maxShowTagsForFilm]
		}

		for _, kw := range filtered {
			uniqTags[kw.Emotion] = struct{}{}

			tagToFilms[kw.Emotion] = append(
				tagToFilms[kw.Emotion],
				review.KinopoiskID,
			)

			tag := &models.Tag{
				Id:      kw.Emotion,
				Name:    capitalizeFirst(kw.Emotion),
				Enabled: true,
			}
			tags[kw.Emotion] = tag
			tagsByFilm[review.KinopoiskID] = append(tagsByFilm[review.KinopoiskID], tag)
		}
	}

	enabledTags := make(map[string]*models.Tag)
	for t := range uniqTags {
		enabled := true
		if len(tagToFilms[t]) < minTagsPerFilm {
			enabled = false
		}

		if enabled {
			enabledTags[t] = tags[t]
		}
	}

	return &Storage{
		Tags:         tags,
		TagsByFilm:   tagsByFilm,
		TagToFilmIds: tagToFilms,
		Films:        films,
		EnabledTags:  enabledTags,
	}, nil
}
