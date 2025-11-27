package adapter

import (
	"Assignment-1/internal/dictionary"
)

type WordSource interface {
	HasWord(word string) bool
}

type MapDictionaryAdapter struct {
	Dict *dictionary.Dictionary
}

func (a MapDictionaryAdapter) HasWord(word string) bool {
	return a.Dict.Has(word)
}
