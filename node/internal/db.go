package internal

import (
	"fmt"
	"sync"
)

var (
	d *InMemorydb
)

type InMemorydb struct {
	Table map[string][]byte
	mu    sync.RWMutex
}

func InitDb() {
	d = &InMemorydb{
		Table: make(map[string][]byte),
	}

}

func (db *InMemorydb) Get(key string) ([]byte, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	value, ok := d.Table[key]
	if !ok {
		return nil, fmt.Errorf("Error the specified key %v does not exist", key)

	}
	return value, nil
}
func (db *InMemorydb) Set(key string, value []byte) error {

	d.mu.Lock()
	defer d.mu.Unlock()
	d.Table[key] = value
	return nil
}
func (db *InMemorydb) Delete(key string) (bool, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	delete(d.Table, key)
	return true, nil

}
