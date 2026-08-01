package keywords

import (
	"bufio"
	"io"
	"strings"
)

var emotionsList = []string{
	"задумчивость",
	"переосмысление жизни",
	"чувство потери (после окончания любимого сериала)",
	"чувство потери",
	"эмоциональная привязанность к персонажам",
	"вдохновение на изменения в жизни",
	"желание творить",
	"удивление",
	"катарсис",
	"незавершённость",
	"фрустрация",
	"надежда",
	"уют",
	"веселье",
	"грусть",
	"тревога",
	"неловкость",
	"интрига",
	"сопереживание",
	"опустошение",
	"безысходность",
	"беспомощность",
	"одиночество",
}

func GetEmotionsWords() []string {
	return emotionsList
}

func GetEmotionsWordsAsMap() map[string]bool {
	result := make(map[string]bool)
	for _, v := range emotionsList {
		result[v] = true
	}

	return result
}

func GetResultFromTxt(txtFile io.Reader, kinopoiskId int64) (Result, error) {
	counter := make(map[string]int)

	scanner := bufio.NewScanner(txtFile)

	emotionsWordsMap := GetEmotionsWordsAsMap()

	inKeywords := false
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		line = strings.TrimPrefix(line, "- ")
		line = strings.TrimSuffix(line, ",")

		if strings.HasPrefix(line, "==========") {
			if line == "==========" {
				inKeywords = false
			} else {
				inKeywords = true
			}
			continue
		}

		if !inKeywords || line == "" {
			continue
		}

		for p := range strings.SplitSeq(line, ",") {
			word := strings.ToLower(strings.TrimSpace(p))
			if word == "" {
				continue
			}

			if !emotionsWordsMap[word] {
				continue
			}

			counter[word]++
		}
	}

	if err := scanner.Err(); err != nil {
		return Result{}, err
	}

	total := 0
	for _, c := range counter {
		total += c
	}

	maxCount := 0
	for _, count := range counter {
		if count > maxCount {
			maxCount = count
		}
	}

	result := Result{
		KinopoiskID: kinopoiskId,
	}

	for emotion, count := range counter {
		result.Keywords = append(result.Keywords, Keyword{
			Emotion: emotion,
			Count:   count,
			Weight:  float64(count) / float64(maxCount),
		})
	}

	return result, nil
}
