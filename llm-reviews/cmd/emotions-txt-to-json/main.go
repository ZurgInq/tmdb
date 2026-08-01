package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"llm-reviews/keywords"
)

func main() {
	workingDir := os.Getenv("WORKING_DIR")
	if workingDir == "" {
		workingDir = "../emotions/kinopoisk"
	}

	files, err := filepath.Glob(filepath.Join(workingDir, "*.txt"))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Read files from ", workingDir, " count files = ", len(files))

	for _, file := range files {
		if err := processFile(file); err != nil {
			log.Printf("error processing %s: %v", file, err)
		}
	}
}

func processFile(path string) error {
	kinopoiskIdStr := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	kinopoiskId, err := strconv.ParseInt(kinopoiskIdStr, 10, 64)
	if err != nil {
		return err
	}

	outputFilename := filepath.Join(
		filepath.Dir(path),
		kinopoiskIdStr+".json",
	)

	if _, err := os.Stat(outputFilename); err == nil {
		fmt.Printf("Файл %s уже существует, пропускаю.\n", outputFilename)
		return nil
	}

	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	result, err := keywords.GetResultFromTxt(file, kinopoiskId)
	if err != nil {
		return err
	}

	jsonData, err := json.MarshalIndent(result, "", "  ")

	if err := os.WriteFile(outputFilename, jsonData, 0644); err != nil {
		return err
	}

	fmt.Printf("Processed %d (%d unique keywords)\n", result.KinopoiskID, len(result.Keywords))

	return nil
}
