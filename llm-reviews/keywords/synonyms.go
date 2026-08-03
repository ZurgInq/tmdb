package keywords

import (
	"encoding/json"
	"log"
	"os"
	"regexp"
	"sort"
	"strings"
)

type mapSynonyms struct {
	// prefix -> target word
	prefixToTarget map[string]string
	// regexp для поиска
	re *regexp.Regexp
}

func (m *mapSynonyms) LoadFile(synonymsFilename string) error {
	data, err := os.ReadFile(synonymsFilename)
	if err != nil {
		log.Println("Error reading synonyms:", err)
		os.Exit(1)
	}

	var dict map[string][]string
	if err := json.Unmarshal(data, &dict); err != nil {
		log.Println("Error parsing json:", err)
		os.Exit(1)
	}

	return m.compileDict(dict)
}

func (m *mapSynonyms) LoadDict(dict map[string][]string) error {
	return m.compileDict(dict)
}

func (m *mapSynonyms) compileDict(dict map[string][]string) error {
	// prefix -> target word
	m.prefixToTarget = make(map[string]string)

	// Список префиксов для построения regex
	var prefixes []string

	for target, synonyms := range dict {
		target = strings.ToLower(target)

		for _, synonym := range synonyms {
			synonym = strings.ToLower(strings.TrimSpace(synonym))
			if synonym == "" {
				continue
			}

			m.prefixToTarget[synonym] = target
			prefixes = append(prefixes, regexp.QuoteMeta(synonym))
		}
	}

	// Более длинные префиксы должны идти раньше.
	// Например "радость" раньше "рад".
	sort.Slice(prefixes, func(i, j int) bool {
		return len(prefixes[i]) > len(prefixes[j])
	})

	pattern := `(?i)(` + strings.Join(prefixes, "|") + `)`

	var err error
	m.re, err = regexp.Compile(pattern)

	return err
}

func (m *mapSynonyms) Get(text string) (string, bool) {
	// Находим все совпадения
	matches := m.re.FindAllStringSubmatch(text, -1)

	for _, match := range matches {
		// match[1] содержит совпавший префикс
		prefix := strings.ToLower(match[1])

		if target, ok := m.prefixToTarget[prefix]; ok {
			return target, true
		}
	}

	return "", false
}

func (m *mapSynonyms) GetAll(text string) []string {
	found := make(map[string]struct{})

	// Находим все совпадения
	matches := m.re.FindAllStringSubmatch(text, -1)

	for _, match := range matches {
		// match[1] содержит совпавший префикс
		prefix := strings.ToLower(match[1])

		if target, ok := m.prefixToTarget[prefix]; ok {
			found[target] = struct{}{}
		}
	}

	if len(found) == 0 {
		return nil
	}

	result := make([]string, 0, len(found))
	for word := range found {
		result = append(result, word)
	}

	return result
}
