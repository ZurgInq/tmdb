package keywords

import (
	"bufio"
	"encoding/json"
	"io"
	"os"
	"strings"
)

func GetResultFromTxt(
	synonyms map[string][]string,
	txtFile io.Reader,
	kinopoiskId int64,
) (Result, error) {
	counter := make(map[string]int)

	scanner := bufio.NewScanner(txtFile)

	synonymsMatch := &mapSynonyms{}
	synonymsMatch.compileDict(synonyms)

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

			if found, ok := synonymsMatch.Get(word); ok {
				word = found
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

// calc stats
func GetStats(files []string, minCount int, maxCount int) (map[string]int, error) {
	type FilmData struct {
		Keywords []Keyword `json:"keywords"`
	}

	counts := make(map[string]int)

	for _, fileName := range files {
		data, err := os.ReadFile(fileName)
		if err != nil {
			return nil, err
		}

		var filmData FilmData
		if err := json.Unmarshal(data, &filmData); err != nil {
			return nil, err
		}

		for _, kw := range filmData.Keywords {
			if kw.Emotion == "" {
				continue
			}
			counts[kw.Emotion]++
		}
	}

	result := make(map[string]int, 0)
	for emotion, count := range counts {
		if count > minCount &&
			count < maxCount {
			result[emotion] = count
		}
	}

	return result, nil
}
