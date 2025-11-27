package games

import (
	"fmt"
)

func New(kind string) (Game, error) {
	switch kind {
	case "spellingbee":
		return NewSpellingBee()
	default:
		return nil, fmt.Errorf("unknown game kind: %s", kind)
	}

}
