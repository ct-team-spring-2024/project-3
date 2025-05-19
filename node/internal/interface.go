package internal

// TODO : set which shards to keep , either in this or somewhere else
type db interface {
	Delete(key string) (bool, error)
	Set(Key string, value []byte) error
	Get(key string) ([]byte, error)
}

type node interface {
	GetShardsRoles() (map[int]string, error)
	SetShards(shardNumber int) ([]int, error)
	SetLeaderForShard(int) (bool, error)
	IsAlive() bool
}
