package internal

import "nabatdb/node/http"

type db interface {
	Delete(key string) (bool, error)
	Set(Key string, value []byte) error
	Get(key string) ([]byte, error)
	GetRemainingLogs() []http.Op
}

type node interface {
	GetShardsRoles() (map[int]string, error)
	SetShards(shardNumber int) ([]int, error)
	SetShardLeader(int) (bool, error)
	IsAlive() bool
}
