package manager

import (
	"Assignment-1/internal/games"
	"fmt"
	"sync"
)

type Manager interface {
	Create(kind string) (string, games.Game, error)
	Get(id string) (games.Game, bool)
}

type mgr struct {
	mu     sync.RWMutex
	nextID int
	items  map[string]games.Game
}

func (m *mgr) Create(kind string) (string, games.Game, error) {
	g, err := games.New(kind)
	if err != nil {
		return "", nil, err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.nextID++
	id := fmt.Sprintf("g-%d", m.nextID)
	m.items[id] = g
	return id, g, nil
}

func (m *mgr) Get(id string) (games.Game, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	g, ok := m.items[id]
	return g, ok
}

var (
	once     sync.Once
	instance Manager
)

func Get() Manager {
	once.Do(func() {
		instance = &mgr{items: make(map[string]games.Game)}
	})
	return instance
}
