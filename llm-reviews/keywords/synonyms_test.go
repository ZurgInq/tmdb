package keywords

import (
	"fmt"
	"reflect"
	"testing"
)

func TestGetAll(t *testing.T) {
	tests := []struct {
		name string
		dict map[string][]string
		text string
		want []string
	}{
		{
			name: "find by exact word",
			dict: map[string][]string{
				"радость": {"радость"},
			},
			text: "Сегодня была радость.",
			want: []string{"радость"},
		},
		{
			name: "find by prefix",
			dict: map[string][]string{
				"радость": {"весел"},
			},
			text: "Мы веселились весь день.",
			want: []string{"радость"},
		},
		{
			name: "multiple targets",
			dict: map[string][]string{
				"радость": {"рад", "весел"},
				"грусть":  {"печал", "груст"},
			},
			text: "Он был печальным, а потом все радовались.",
			want: []string{"грусть", "радость"},
		},
		{
			name: "duplicates removed",
			dict: map[string][]string{
				"радость": {"рад"},
			},
			text: "Рад, радость, радовался.",
			want: []string{"радость"},
		},
		{
			name: "case insensitive",
			dict: map[string][]string{
				"радость": {"весел"},
			},
			text: "ВЕСЕЛИЛИСЬ все.",
			want: []string{"радость"},
		},
		{
			name: "nothing found",
			dict: map[string][]string{
				"радость": {"рад"},
			},
			text: "Сегодня идет дождь.",
			want: nil,
		},
		{
			name: "longest prefix wins",
			dict: map[string][]string{
				"радость": {"радость"},
				"рад":     {"рад"},
			},
			text: "Радость переполняла всех.",
			want: []string{"радость"},
		},
		{
			name: "ignore empty synonym",
			dict: map[string][]string{
				"радость": {"", "рад"},
			},
			text: "Он рад.",
			want: []string{"радость"},
		},
	}

	synonyms := &mapSynonyms{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			synonyms.LoadDict(tt.dict)
			got := synonyms.GetAll(tt.text)

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGet(t *testing.T) {
	dict := map[string][]string{
		"страдания": {"депрессия", "мучительн", "раскаяние", "эмоциональная боль", "унижен", "недооценк", "отчаянн"},
		"смятение":  {"задумчив", "нервозн", "недоум"},
	}
	tests := []struct {
		dict map[string][]string
		text string
		want string
	}{
		{
			dict: dict,
			text: "мучительное",
			want: "страдания",
		},
		{
			dict: dict,
			text: "отчаянный",
			want: "страдания",
		},
		{
			dict: dict,
			text: "нервозный",
			want: "смятение",
		},
	}

	synonyms := &mapSynonyms{}

	for k, tt := range tests {
		t.Run(fmt.Sprintf("case %d", k), func(t *testing.T) {
			synonyms.LoadDict(tt.dict)
			got, _ := synonyms.Get(tt.text)

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}
