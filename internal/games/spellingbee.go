package games

import (
	"Assignment-1/internal/dictionary"
	adapter2 "Assignment-1/internal/dictionary/adapter"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"
)

type SpellingBee struct {
	Letters      []string
	CenterLetter string
	Score        int
	//dictionary   *dictionary.Dictionary
	wordSource interface{ HasWord(string) bool }
	usedWords  map[string]struct{}
}

func NewSpellingBee() (*SpellingBee, error) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	file, err := os.Open("./pangrams.json")
	if err != nil {
		return nil, fmt.Errorf("failed to open pangrams.json: %v", err)
	}
	defer func() {
		_ = file.Close()
	}()

	var pangrams map[string]string
	if err := json.NewDecoder(file).Decode(&pangrams); err != nil {
		return nil, fmt.Errorf("failed to decode pangrams.json: %v", err)
	}

	keys := make([]string, 0, len(pangrams))
	for k := range pangrams {
		keys = append(keys, k)
	}

	if len(keys) == 0 {
		return nil, fmt.Errorf("no pangrams found in file")
	}

	word := keys[r.Intn(len(keys))]
	word = strings.ToUpper(word)

	unique := make(map[string]struct{})
	for _, ch := range word {
		if ch >= 'A' && ch <= 'Z' {
			unique[string(ch)] = struct{}{}
		}
	}

	letters := make([]string, 0, len(unique))
	for k := range unique {
		letters = append(letters, k)
	}

	if len(letters) == 0 {
		for len(letters) < 7 {
			letters = append(letters, string(rune('A'+r.Intn(26))))
		}
	}

	center := letters[r.Intn(len(letters))]

	d, err := dictionary.Get()
	if err != nil {
		return nil, err
	}

	adapted := adapter2.MapDictionaryAdapter{Dict: d}

	return &SpellingBee{
		Letters:      letters,
		CenterLetter: center,
		Score:        0,
		//dictionary:   d,
		wordSource: adapted,
		usedWords:  make(map[string]struct{}),
	}, nil
}

func (s *SpellingBee) Name() string {
	return "spellingbee"
}

func (s *SpellingBee) ValidateWord(word string) (bool, int, string) {
	if len(word) == 0 {
		return false, 0, "Enter a 4 letter word to guess!"
	}

	upper := strings.ToUpper(strings.TrimSpace(word))
	lower := strings.ToLower(upper)

	if _, exists := s.usedWords[lower]; exists {
		return false, 0, fmt.Sprintf("Word %s has already been found!", lower)
	}

	if len(upper) < 4 {
		return false, 0, "Too short (min 4 letters)"
	}

	if !strings.Contains(upper, s.CenterLetter) {
		return false, 0, "Missing centre letter!: " + s.CenterLetter
	}

	allowed := make(map[string]struct{}, len(s.Letters))
	for _, l := range s.Letters {
		allowed[l] = struct{}{}
	}

	for _, r := range upper {
		ch := string(r)
		if _, ok := allowed[ch]; !ok {
			return false, 0, fmt.Sprintf("Contains letters not in the word bank!: %s", ch)
		}
	}

	if !s.wordSource.HasWord(lower) {
		return false, 0, fmt.Sprintf("Word %s not found in the dictionary!", lower)
	}

	present := make(map[string]bool, len(s.Letters))
	for _, r := range upper {
		present[string(r)] = true
	}
	isPangram := true
	for _, l := range s.Letters {
		if !present[l] {
			isPangram = false
			break
		}
	}

	var score int
	if len(upper) == 4 {
		score = 1
	} else {
		score = len(upper)
	}
	if isPangram {
		score += 7
	}
	s.Score += score
	s.usedWords[lower] = struct{}{}

	if isPangram {
		return true, score, fmt.Sprintf("Pangram! %d points. Current score: %d", score, s.Score)
	}
	return true, score, fmt.Sprintf("New word scoring %d points. Current score: %d", score, s.Score)
}

func (s *SpellingBee) FormatLetters() string {
	letters := make([]string, len(s.Letters))
	for i, l := range s.Letters {
		if l == s.CenterLetter {
			letters[i] = "[ " + l + " ]"
		} else {
			letters[i] = l
		}
	}
	return strings.Join(letters, " ")
}
