package games

type Game interface {
	Name() string
	ValidateWord(word string) (bool, int, string)
}
