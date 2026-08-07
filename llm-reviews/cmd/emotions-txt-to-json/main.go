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

var synonyms = map[string][]string{
	"страдания": {"депрессия", "мучительн", "раскаяние", "эмоциональная боль", "унижен", "недооценк", "отчаянн"},
	"смятение":  {"задумчив", "нервозн", "недоум"},
	"бессилие":  {"фрустрация", "беспомощн", "бессил", "опустошен", "отчаяние"},
	"трагизм":   {"смерт", "судьба", "безнадежн", "безнадёжн", "трагед", "безысходн", "горе", "потеря надежды", "рок"},
	"грусть":    {"печал", "груст", "скорб", "пессим", "расстройство", "потеря надежды", "унын", "тоска", "потеря", "огорчен", "слёз", "меланхол", "боль"},
	"страх":     {"жуткост", "паник", "страшн", "ужас"},

	"злость":  {"ярость", "гнев", "агресс", "ненавист", "призрен"},
	"тревога": {"тревожность"},

	"веселье":       {"рад", "смешн", "захватывающ", "восхищен", "смех", "восторг"},
	"любовь":        {"влюблённ", "обожан", "ревност", "любов", "романтик", "забота", "близость", "гармония"},
	"вдохновение":   {"желание творить", "мотивация", "вдохновение на изменения в жизни"},
	"сопереживание": {"сострадание", "сочувствие", "эмоциональная привязанность к персонажам", "чувство потери", "чувство потери (после окончания любимого сериала)"},
	"надежда":       {"вера"},
	"уют":           {"беззабот", "лёгкость", "расслабление", "утешение", "теплота", "тёплота", "душевность"},

	"интрига": {"интриг", "интерес", "непредсказуем", "запутанн", "удивлен", "сюрприз", "обман", "тайна", "раздум", "напряжен", "любопыт"},
}

func main() {
	workingDir := os.Getenv("WORKING_DIR")
	if workingDir == "" {
		workingDir = "../reviews/emotions/kinopoisk"
	}

	force := os.Getenv("FORCE")

	files, err := filepath.Glob(filepath.Join(workingDir, "*.txt"))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Read files from ", workingDir, " count files = ", len(files))

	for _, file := range files {
		if err := processFile(file, force == "1"); err != nil {
			log.Printf("error processing %s: %v", file, err)
		}
	}
}

func processFile(path string, force bool) error {
	kinopoiskIdStr := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	kinopoiskId, err := strconv.ParseInt(kinopoiskIdStr, 10, 64)
	if err != nil {
		return err
	}

	outputFilename := filepath.Join(
		filepath.Dir(path),
		kinopoiskIdStr+".json",
	)

	if !force {
		if _, err := os.Stat(outputFilename); err == nil {
			fmt.Printf("Файл %s уже существует, пропускаю.\n", outputFilename)
			return nil
		}
	}

	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	result, err := keywords.GetResultFromTxt(synonyms, file, kinopoiskId)
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
