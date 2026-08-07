package getfilms

import (
	"slices"

	"app/internal/models"
	"app/internal/storage"
)

type action struct {
	db              *storage.Storage
	maxFilmsForShow int
}

type Result struct {
	Films    []*models.Film
	Emotions map[int64][]*models.Tag
}

func New(
	db *storage.Storage,
	maxFilmsForShow int,
) *action {
	return &action{
		db,
		maxFilmsForShow,
	}
}

func ContainsAll(a []*models.Tag, b []string) bool {
	set := make(map[string]struct{}, len(a))

	for _, v := range a {
		set[v.Id] = struct{}{}
	}

	for _, v := range b {
		if _, ok := set[v]; !ok {
			return false
		}
	}

	return true
}

func (act *action) Do(
	filterTags []string,
	filmTypes []models.FilmType,
) (Result, error) {
	tagIds := make([]string, 0, len(filterTags))
	for _, tagId := range filterTags {
		if tag, exists := act.db.Tags[tagId]; exists {
			if tag.Enabled {
				tagIds = append(tagIds, tagId)
			}
		}
	}

	films := make([]*models.Film, 0)
	filmsSet := make(map[int64]struct{}) // skip duplicate

	for _, tag := range tagIds {
		breakByMax := false

		ids := act.db.TagToFilmIds[tag]
		for _, id := range ids {

			if !ContainsAll(act.db.TagsByFilm[id], tagIds) {
				continue
			}

			if _, exists := filmsSet[id]; exists {
				continue
			}

			if film, ok := act.db.Films[id]; ok {
				if filmTypes == nil ||
					slices.Contains(filmTypes, film.Type) {
					films = append(films, film)
					filmsSet[id] = struct{}{}
				}
			}

			if len(films) >= act.maxFilmsForShow {
				breakByMax = true
				break
			}
		}

		if breakByMax {
			break
		}
	}

	emotions := make(map[int64][]*models.Tag)
	for _, film := range films {
		for _, tag := range act.db.TagsByFilm[film.KinopoiskId] {
			if tag.Enabled {
				emotions[film.KinopoiskId] = append(emotions[film.KinopoiskId], tag)
			}
		}
	}

	return Result{
		Films:    films,
		Emotions: emotions,
	}, nil
}
