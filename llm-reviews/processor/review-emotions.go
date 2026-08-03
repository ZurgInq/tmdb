package processor

import (
	"llm-reviews/api"
	"log/slog"
	"strings"
)

// Лимиты на длину рецензий. Обрезаем длинные рецензии по параграфам
const (
	maxTextLen   = 4000 // Обрезаем рецензию до этого количества символов (на самом деле байт)
	maxParagraph = 2000 // Обрезаем параграфы
	minParagraph = 40   // Исключаем слишком короткие параграфы как вероятно мусорные
)

type processor struct {
	client *api.Client
	Logger *slog.Logger
}

func NewProcessor(
	client *api.Client,
	logger *slog.Logger,
) *processor {
	return &processor{
		client: client,
		Logger: logger,
	}
}

func (p *processor) ProcessReview(
	startMessage string,
	emotionsMapping string,
	content string,
	filmId string,
	reviewId int64,
) (string, error) {
	logger := p.Logger.With("filmId", filmId, "review_id", reviewId)
	logger.Info("Start processing review")

	contentLen := len(content)
	if contentLen > maxTextLen {
		logger.Info("Cropping the text", "maxTextLen", maxTextLen, "contentLen", len(content))
		content = ShortenText(content)
	}

	logger.Info("Send first instruct...")
	result, err := p.client.SendInstruct(startMessage + content)
	if err != nil {
		return "", err
	}

	logger.Info("Response from first instruct", "result", result)
	if result == "" {
		logger.Info("Empty response, skip")
		return "", nil
	}

	// logger.Info("Send second instruct...")
	// result, err = p.client.SendInstruct(emotionsMapping + result)
	// if err != nil {
	// 	return "", err
	// }

	// logger.Info("Response from second instruct", "result", result)
	// if result == "" {
	// 	fmt.Println("empty response, skip")
	// 	return "", nil
	// }

	return result, err
}

func ShortenText(text string) string {
	if len(text) <= maxTextLen {
		return text
	}

	// Функция обрезки до ближайшей точки.
	trimToSentence := func(s string, limit int) string {
		if len(s) <= limit {
			return s
		}

		end := limit
		if end > len(s) {
			end = len(s)
		}

		// Ищем конец предложения до лимита
		for i := end - 1; i >= 0; i-- {
			switch s[i] {
			case '.', '!', '?':
				return strings.TrimSpace(s[:i+1])
			}
		}

		// или просто обрезаем.
		return strings.TrimSpace(s[:limit])
	}

	// Разделяем по пустым строкам (абзацам).
	paragraphs := strings.Split(text, "\n\n")

	// Если абзац один.
	if len(paragraphs) == 1 {
		return trimToSentence(text, maxTextLen)
	}

	// Подготавливаем абзацы.
	filtered := make([]string, 0, len(paragraphs))
	for _, p := range paragraphs {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}

		if len(p) > maxParagraph {
			p = trimToSentence(p, maxParagraph)
		}

		if len(p) >= minParagraph {
			filtered = append(filtered, p)
		}
	}

	if len(filtered) == 0 {
		return trimToSentence(text, maxTextLen)
	}

	if len(filtered) == 1 {
		return filtered[0]
	}

	result := []string{filtered[0]}
	totalLen := len(filtered[0])

	last := filtered[len(filtered)-1]
	lastLen := len(last)

	// Добавляем средние абзацы, оставляя место под последний.
	for i := 1; i < len(filtered)-1; i++ {
		p := filtered[i]
		need := len(p) + 2 // "\n\n"

		// +2 перед последним абзацем.
		if totalLen+need+2+lastLen > maxTextLen {
			break
		}

		result = append(result, p)
		totalLen += need
	}

	// Добавляем последний абзац.
	result = append(result, last)

	return strings.Join(result, "\n\n")
}
