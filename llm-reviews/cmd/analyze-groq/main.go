package main

import (
	"fmt"
	"llm-reviews/api"
	"llm-reviews/processor"
	"llm-reviews/source/kinopoisk"
	"log"
	"log/slog"
	"net"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"slices"
	"time"
)

// Настройки API и запросов
const (
	ApiURL     = "https://api.groq.com/openai/v1/chat/completions"
	MaxRetries = 5 // количество повторных запросов в случае ошибки
)

// Лимиты на обработку по количеству.
// Лимиты на длину рецензий в processor/review-emotions.go
const (
	MaxReviews = 5  // сколько ревью обрабатываем на один фильм
	MaxFilms   = 20 // сколько фильмов обрабатываем на один запуск программы
)

const (
	// groq/compound
	modelGroq_compound = "groq/compound" // не форматирует вывод как надо
	// llama
	modelLlama3_1_8b_Instant    = "llama-3.1-8b-instant" // не форматирует вывод как надо
	modelLlama3_3_70b_Versatile = "llama-3.3-70b-versatile"
	// openai
	modelOpenAI_Gpt_Oss_20b  = "openai/gpt-oss-20b"
	modelOpenAI_Gpt_Oss_120b = "openai/gpt-oss-120b"
	// qwen
	modelQwen_Qwen3_6_27b = "qwen/qwen3.6-27b"

	groqDefaultModel = modelLlama3_3_70b_Versatile
)

var (
	validGroqModels = []string{
		modelLlama3_3_70b_Versatile,
		modelOpenAI_Gpt_Oss_20b,
		modelOpenAI_Gpt_Oss_120b,
		modelQwen_Qwen3_6_27b,
	}

	choosenModel = groqDefaultModel
)

func panicIfErr(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	apiKey := os.Getenv("GROQ_API_KEY")
	if apiKey == "" {
		log.Fatalln("GROQ_API_KEY is not set")
	}

	groqModel := os.Getenv("GROQ_MODEL")
	if groqModel != "" {
		if idx := slices.Index(validGroqModels, groqModel); idx > -1 {
			choosenModel = validGroqModels[idx]
		} else {
			log.Fatalln("Invalid GROQ_MODEL")
		}
	}

	inputDir := os.Getenv("INPUT_DIR")
	if inputDir == "" {
		log.Fatalln("INPUT_DIR is not set")
	}

	outputDir := os.Getenv("OUTPUT_DIR")
	if outputDir == "" {
		log.Fatalln("OUTPUT_DIR is not set")
	}

	logDir := os.Getenv("LOG_DIR")
	if logDir == "" {
		log.Fatalln("LOG_DIR is not set")
	}

	logDirStat, err := os.Stat(logDir)
	if err != nil {
		log.Fatalf("Dir %s not exists: %s\n", logDir, err.Error())
	}
	if !logDirStat.IsDir() {
		log.Fatalf("Path %s is not dir", logDir)
	}

	startMessage, err := os.ReadFile("startMessage.txt")
	panicIfErr(err)

	emotionsMapping, err := os.ReadFile("emotionsMapping.txt")
	panicIfErr(err)

	files, err := filepath.Glob(inputDir + "/*.json")
	panicIfErr(err)

	client := &api.Client{
		Client: &http.Client{
			Transport: &http.Transport{
				Proxy: http.ProxyFromEnvironment,
				DialContext: (&net.Dialer{
					Timeout:   5 * time.Second,
					KeepAlive: 30 * time.Second,
				}).DialContext,
				ForceAttemptHTTP2:     true,
				MaxIdleConns:          100,
				IdleConnTimeout:       90 * time.Second,
				TLSHandshakeTimeout:   10 * time.Second,
				ExpectContinueTimeout: 1 * time.Second,
			},
		},
		ApiURL:     ApiURL,
		ApiKey:     apiKey,
		MaxRetries: MaxRetries,
	}

	client.ChatRequestModifier = func(cr *api.ChatRequest) any {
		cr.Model = choosenModel
		if cr.Model == modelQwen_Qwen3_6_27b {
			cr.ReasoningFormat = "parsed"
		} else {
			cr.ReasoningFormat = ""
		}
		return cr
	}

	timestamp := time.Now().Format("2006_01_02_15_04_05")
	logFilename := path.Join(logDir, fmt.Sprint("log-", timestamp, ".log"))
	logFile, err := os.Create(logFilename)
	if err != nil {
		log.Fatal("Failed to create log file:", err)
	}

	processorLogger := slog.New(slog.NewTextHandler(logFile, nil))
	processor := processor.NewProcessor(
		client,
		processorLogger.With("model", choosenModel),
	)

	log.Printf("Start with model=%s log-file=%s", choosenModel, logFilename)

	processed, err := kinopoisk.StartProcessing(
		client,
		files,
		outputDir,
		func(filmID string, content string, id int64) (string, error) {
			return processor.ProcessReview(string(startMessage), string(emotionsMapping), content, filmID, id)
		},
		MaxReviews,
		MaxFilms,
	)
	panicIfErr(err)
	fmt.Printf("Обработано %d рецензий\n", processed)
}
