package internal

import (
	"fmt"
	"nabatdb/node/http"
	"sync"

	"github.com/sirupsen/logrus"
)

type Table map[string][]byte

// The table struct should be sorted all the time
type InMemorydb struct {
	TableSize  int
	Table      *RBTree
	ROTables   [][]Pair // Read-only tables
	Logs       []http.Op
	LogIndex   int
	maximumKey int
	mu         sync.RWMutex
}

func InitDB() *InMemorydb {
	return &InMemorydb{
		Table:    NewRBTree(),
		ROTables: make([][]Pair, 0, 0),
		Logs:     make([]http.Op, 0, 0),
		LogIndex: 0,
	}
}

func (db *InMemorydb) Get(key string) ([]byte, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()
	value := db.Table.searchNode(db.Table.Root, key)
	if value == db.Table.nilNode {
		return nil, fmt.Errorf("Error the specified key %v does not exist", key)

	}
	return value.Pair.Value, nil
}

func (db *InMemorydb) Set(key string, value []byte) error {
	db.mu.Lock()
	defer db.mu.Unlock()
	logrus.Infof("the set request was also added to shard")

	op := http.ConsSetOp(key, value)
	db.Logs = append(db.Logs, op)
	n := db.Table.searchNode(db.Table.Root, key)
	if n == db.Table.nilNode {
		db.Table.Insert(Pair{Key: key, Value: value})
		db.TableSize++
		if db.TableSize > db.maximumKey {
			//Create a new table
		}
		return nil
	}
	db.Table.Update(key, value)

	if db.TableSize > db.maximumKey {
		//Create a new table
	}

	return nil
}

func (db *InMemorydb) Delete(key string) (bool, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	op := http.ConsDelOp(key)
	db.Logs = append(db.Logs, op)
	db.Table.Delete(key)
	//Should be done or not
	//db.TableSize--
	// delete(db.Table, key)

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
func (db *InMemorydb) FlushDB() {

	roTable := db.Table.ToSortedSlice()
	db.ROTables = append(db.ROTables, roTable)
	db.Table = NewRBTree()
}
func (db *InMemorydb) MergeROTables(firstIndex, secondIndex int) {
	compactedTable := merge(db.ROTables[firstIndex], db.ROTables[secondIndex])
	db.ROTables = append(db.ROTables, compactedTable)
	//TODO : Delete the tables that have been merged right now

}
