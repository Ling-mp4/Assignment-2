package dictionary

import (
	"encoding/json"
	"os"
	"sync"
)

type Dictionary struct {
	words map[string]struct{}
	once  sync.Once
}

var (
	instance *Dictionary
	loadErr  error
)

func Get() (*Dictionary, error) {
	if instance == nil {
		instance = &Dictionary{}
		instance.once.Do(func() {
			loadErr = instance.load("./words_dictionary.json")
		})
	}
	return instance, loadErr
}

func (d *Dictionary) load(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer func() {
		_ = f.Close()
	}()

	var raw map[string]interface{}
	if err := json.NewDecoder(f).Decode(&raw); err != nil {
		return err
	}
	d.words = make(map[string]struct{}, len(raw))
	for k := range raw {
		d.words[k] = struct{}{}
	}
	return nil
}

func (d *Dictionary) Has(word string) bool {
	_, ok := d.words[word]
	return ok
}
