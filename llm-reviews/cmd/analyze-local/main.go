package main

import (
	"fmt"
	"log"
	"log/slog"
	"net"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"time"

	"llm-reviews/api"
	"llm-reviews/processor"
	"llm-reviews/source/kinopoisk"
)

const (
	ApiURL     = "/v1/chat/completions"
	MaxRetries = 2
	MaxReviews = 5
	MaxFilms   = 10
)

var emotionsMapping []byte

func panicIfErr(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	apiHost := os.Getenv("API_HOST")
	if apiHost == "" {
		log.Fatalln("API_HOST is not set")
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

	emotionsMapping, err = os.ReadFile("emotionsMapping.txt")
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
		ApiURL:     apiHost + ApiURL,
		MaxRetries: MaxRetries,
	}

	client.ChatRequestModifier = func(cr *api.ChatRequest) any {
		cr.Mode = "instruct"
		return cr
	}

	timestamp := time.Now().Format("2006_01_02_15_04_05")
	logFilename := path.Join(logDir, fmt.Sprint("log-", timestamp, ".log"))
	logFile, err := os.Create(logFilename)
	if err != nil {
		log.Fatal("Failed to create log file:", err)
	}

	processor := processor.NewProcessor(
		client,
		slog.New(slog.NewTextHandler(logFile, nil)),
	)

	processed, err := kinopoisk.StartProcessing(
		client,
		files,
		outputDir,
		func(filmID string, content string, id int64) (string, error) {
			return processor.ProcessReview(string(startMessage), string(emotionsMapping), content, id)
		},
		MaxReviews,
		MaxFilms,
	)
	panicIfErr(err)
	fmt.Printf("Обработано %d рецензий\n", processed)
}
