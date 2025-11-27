package stats

import (
	"encoding/json"
	"os"
	"sync"
)

type Stats struct {
	TotalGames     int      `json:"total_games"`
	TotalPangrams  int      `json:"total_pangrams"`
	HighScore      int      `json:"high_score"`
	TotalScore     int      `json:"total_score"`
	GamesWithScore int      `json:"games_with_score"`
	PangramList    []string `json:"pangram_list"`
}

type Store struct {
	path string
	mu   sync.Mutex
	S    Stats
}

func NewStore(path string) (*Store, error) {
	s := &Store{path: path}
	if err := s.load(); err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	return s, nil
}

func (s *Store) load() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	f, err := os.Open(s.path)
	if err != nil {
		return err
	}
	defer f.Close()
	return json.NewDecoder(f).Decode(&s.S)
}

func (s *Store) Save() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	tmp := s.path + ".tmp"
	f, err := os.Create(tmp)
	if err != nil {
		return err
	}
	if err := json.NewEncoder(f).Encode(&s.S); err != nil {
		_ = f.Close()
		return err
	}
	_ = f.Close()
	return os.Rename(tmp, s.path)
}

func (s *Store) AddGame() error {
	s.mu.Lock()
	s.S.TotalGames++
	defer s.mu.Unlock()
	return s.Save()
}

func (s *Store) AddPangram(word string) error {
	s.mu.Lock()
	s.S.TotalPangrams++
	if word != "" {
		s.S.PangramList = append([]string{word}, s.S.PangramList...)
		if len(s.S.PangramList) > 50 {
			s.S.PangramList = s.S.PangramList[:50]
		}
	}
	s.mu.Unlock()
	return s.Save()
}

func (s *Store) AddScore(score int) error {
	s.mu.Lock()
	s.S.TotalScore += score
	s.S.GamesWithScore++
	if score > s.S.HighScore {
		s.S.HighScore = score
	}
	s.mu.Unlock()
	return s.Save()
}
func (s *Store) Snapshot() Stats {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.S
}
