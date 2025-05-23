package internal

import (
	"fmt"
	"nabatdb/node/http"
	"sync"
)

type Table map[string][]byte

type InMemorydb struct {
	Table    Table
	ROTables []Table // Read-only tables
	Logs     []http.Op
	LogIndex int
	mu       sync.RWMutex
}

func InitDB() *InMemorydb{
	return &InMemorydb{
		Table: make(map[string][]byte),
		ROTables: make([]Table, 0, 0),
		Logs: make([]http.Op, 0, 0),
		LogIndex: 0,
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

	op := http.ConsSetOp(key, value)
	db.Logs = append(db.Logs, op)
	db.Table[key] = value

	return nil
}

func (db *InMemorydb) Delete(key string) (bool, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	op := http.ConsDelOp(key)
	db.Logs = append(db.Logs, op)
	delete(db.Table, key)

	return true, nil
}

func (db *InMemorydb) GetRemainingLogs() []http.Op {
	db.mu.RLock()
	defer db.mu.RUnlock()

	if db.LogIndex == len(db.Logs) {
		return nil
	}
	result := db.Logs[db.LogIndex:]
	db.LogIndex = len(db.Logs)
	return result
}
