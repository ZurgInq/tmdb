package storage

import (
	"cmp"
	"encoding/json"
	"fmt"
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
	Films map[int64]*models.Film // FilmId => Film
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

type LoadParams struct {
	FilmsDir string
	TagsDir  string

	FilmTypes []models.FilmType

	MinCountForTags  int
	MinWeightForTags float64
	MaxTagsPerFilm   int
	MinTagsPerFilm   int
}

func Load(params LoadParams) (*Storage, error) {
	log.Println("Load films from ", params.FilmsDir)
	filmFiles, err := filepath.Glob(filepath.Join(params.FilmsDir, "*.json"))
	if err != nil {
		return nil, fmt.Errorf("get file list from %s:%w", params.FilmsDir, err)
	}

	films, err := loadFilms(filmFiles, params.FilmTypes)
	if err != nil {
		return nil, err
	}

	log.Println("Films count", len(films))

	log.Println("Load tags from ", params.TagsDir)
	tagsFiles, err := filepath.Glob(filepath.Join(params.TagsDir, "*.json"))
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

		var filmKW models.FilmKeywords

		if err := json.Unmarshal(data, &filmKW); err != nil {
			return nil, err
		}

		if len(filmKW.Keywords) == 0 {
			continue
		}

		kw := applyParamsToEmotionData(films, filmKW, params)

		for _, kw := range kw {
			uniqTags[kw.Emotion] = struct{}{}

			tagToFilms[kw.Emotion] = append(
				tagToFilms[kw.Emotion],
				filmKW.KinopoiskID,
			)

			tag := &models.Tag{
				Id:      kw.Emotion,
				Name:    capitalizeFirst(kw.Emotion),
				Enabled: true,
			}
			tags[kw.Emotion] = tag
			tagsByFilm[filmKW.KinopoiskID] = append(tagsByFilm[filmKW.KinopoiskID], tag)
		}
	}

	enabledTags := make(map[string]*models.Tag)
	for t := range uniqTags {
		enabled := true
		if len(tagToFilms[t]) < params.MinTagsPerFilm {
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

func applyParamsToEmotionData(
	films map[int64]*models.Film,
	review models.FilmKeywords,
	params LoadParams,
) []models.Keyword {
	filtered := slices.Collect(func(yield func(models.Keyword) bool) {
		for _, kw := range review.Keywords {
			if kw.Count < params.MinCountForTags {
				continue
			}

			if kw.Weight < params.MinWeightForTags {
				continue
			}

			if _, ok := films[review.KinopoiskID]; !ok {
				continue
			}

			yield(kw)
		}
	})

	slices.SortFunc(filtered, func(i, j models.Keyword) int {
		return cmp.Compare(j.Count, i.Count)
	})

	if len(filtered) > params.MaxTagsPerFilm {
		filtered = filtered[:params.MaxTagsPerFilm]
	}

	return filtered
}

func loadFilms(filmFiles []string, filmTypes []models.FilmType) (map[int64]*models.Film, error) {
	films := make(map[int64]*models.Film)

	for _, file := range filmFiles {
		data, err := os.ReadFile(file)
		if err != nil {
			return nil, err
		}

		var film models.Film

		if err := json.Unmarshal(data, &film); err != nil {
			return nil, err
		}

		if filmTypes == nil ||
			slices.Contains(filmTypes, film.Type) {
			films[film.KinopoiskId] = &film
		}
	}

	return films, nil
}
