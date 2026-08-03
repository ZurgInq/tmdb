package main

import (
	"fmt"
	"llm-reviews/keywords"
	"log"
	"path/filepath"
)

func main() {
	files, err := filepath.Glob("../reviews/emotions/kinopoisk/*.json")
	if err != nil {
		log.Fatal(err)
	}

	result, err := keywords.GetStats(files, 0, 6)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("total=", len(result))

	for k := range result {
		fmt.Println(k)
	}
}
