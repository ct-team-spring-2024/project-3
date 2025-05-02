package internal

import (
	"fmt"
	"sync"
)

var (
	DB *InMemorydb
)

type InMemorydb struct {
	Table map[string][]byte
	mu    sync.RWMutex
}

func InitDb() {
	DB = &InMemorydb{
		Table: make(map[string][]byte),
	}
}

func (db *InMemorydb) Get(key string) ([]byte, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	value, ok := db.Table[key]
	if !ok {
		return nil, fmt.Errorf("Error the specified key %v does not exist", key)

	}
	return value, nil
}

func (db *InMemorydb) Set(key string, value []byte) error {
	db.mu.Lock()
	defer db.mu.Unlock()
	db.Table[key] = value
	return nil
}

func (db *InMemorydb) Delete(key string) (bool, error) {
	db.mu.Lock()
	defer db.mu.Unlock()
	delete(db.Table, key)
	return true, nil
}
