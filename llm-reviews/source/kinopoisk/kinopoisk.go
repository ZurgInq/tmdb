package kinopoisk

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"llm-reviews/api"
)

type ReviewResp struct {
	Items []Item `json:"items"`
}

type Item struct {
	Type        string `json:"type"`
	Description string `json:"description"`
	KinopoiskId int64  `json:"kinopoiskId"`
}

func extractIDs(filename string) ([]int64, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var ids []int64
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if idStr, ok := strings.CutPrefix(line, "=========="); ok {
			if idStr == "" {
				// Это строка-закрытие "=========="
				continue
			}

			id, err := strconv.ParseInt(idStr, 10, 64)
			if err == nil {
				ids = append(ids, id)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return ids, nil
}

func int64ToMap(values []int64) map[int64]struct{} {
	result := make(map[int64]struct{}, len(values))
	for _, v := range values {
		result[v] = struct{}{}
	}

	return result
}

func writeReviewToTxt(output io.Writer, reviewId int64, content string) error {
	_, err := io.WriteString(
		output,
		fmt.Sprint("==========", reviewId, "\n"),
	)

	if err != nil {
		return err
	}

	_, err = io.WriteString(
		output,
		fmt.Sprint(content, "\n==========\n\n"),
	)

	return err
}

func StartProcessing(
	client *api.Client,
	files []string,
	outputDir string,
	processingReviewFn func(filmID string, content string, id int64) (string, error),
	maxReviews int,
	maxFilms int,
) (int, error) {
	totalProcessed := 0
	filmsProcessed := 0
	for _, inputFilename := range files {
		processedReviews := 0
		skippedReviews := 0

		log.Printf("Обрабатываем файл с рецензиями %s...\n", inputFilename)

		filmID := filepath.Base(inputFilename)
		filmID = filmID[:len(filmID)-len(filepath.Ext(filmID))]

		var reviewsIdsInFile map[int64]struct{}

		outputFilename := filepath.Join(outputDir, filmID+".txt")
		var outputFile *os.File

		if _, err := os.Stat(outputFilename); err == nil {
			log.Printf("Итоговый файл %s уже существует, извлекаем ИД рецензий\n", outputFilename)

			ids, err := extractIDs(outputFilename)
			if err != nil {
				return totalProcessed, fmt.Errorf("extractIDs from %s: %w", outputFilename, err)
			}
			reviewsIdsInFile = int64ToMap(ids)
			log.Println("Найдены обработанные рецензии:", ids)

			log.Printf("Открываем файл %s на запись\n", outputFilename)
			outputFile, err = os.OpenFile(outputFilename, os.O_WRONLY|os.O_APPEND, 0644)
			if err != nil {
				return totalProcessed, fmt.Errorf("open file %s: %w", outputFilename, err)
			}
		} else {
			outputFile, err = os.OpenFile(outputFilename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
			if err != nil {
				return totalProcessed, fmt.Errorf("create file %s: %w", outputFilename, err)
			}
		}

		reviewResponseRaw, err := os.ReadFile(inputFilename)
		if err != nil {
			return totalProcessed, err
		}

		var reviewResponse ReviewResp
		if err := json.Unmarshal(reviewResponseRaw, &reviewResponse); err != nil {
			return totalProcessed, err
		}

		log.Printf("Начинается обработка рецензий для filmID=%s...\n", filmID)
		for _, item := range reviewResponse.Items {
			if processedReviews+skippedReviews >= maxReviews {
				log.Printf("Достигнуто максимальное количество ревью (%d) на фильм ИД=%s\n", maxReviews, filmID)
				break
			}

			if item.Type != "POSITIVE" {
				continue
			}
			if reviewsIdsInFile != nil {
				if _, exists := reviewsIdsInFile[item.KinopoiskId]; exists {
					skippedReviews++
					log.Printf("Рецензия %d для фильма %s уже обработана, пропускаем\n", item.KinopoiskId, filmID)
					continue
				}
			}

			log.Printf("Начинается обработка рецензии %d для filmID=%s...\n", item.KinopoiskId, filmID)
			if reviewResult, err := processingReviewFn(filmID, item.Description, item.KinopoiskId); err == nil {
				log.Printf("Записываем результат в файл %s...\n", outputFile.Name())
				if err = writeReviewToTxt(outputFile, item.KinopoiskId, reviewResult); err != nil {
					return totalProcessed, fmt.Errorf("Write to file %s: %w", outputFile.Name(), err)
				}
			} else {
				log.Printf("Ошибка при обработке ревью %d:%s\n", item.KinopoiskId, err.Error())
				outputFile.Close()
				return totalProcessed, fmt.Errorf("Processing review: %w", err)
			}

			// stats
			totalProcessed++
			processedReviews++
		}

		if processedReviews > 0 {
			filmsProcessed++
		}

		log.Printf("Обработано рецензий %d для фильма %s\n", processedReviews, filmID)
		log.Printf("Всего обработано фильмов %d, рецензий %d\n", filmsProcessed, totalProcessed)

		outputFile.Close()
		if filmsProcessed >= maxFilms {
			log.Printf("Достигнуто максимальное количество фильмов на запуск %d\n", filmsProcessed)
			break
		}
	}

	return totalProcessed, nil
}
