package commons

import (
	"hash/fnv"
)

func GetPartitionID(key string, totalPartitions int) int {
	hash := fnv.New32a()
	hash.Write([]byte(key))
	return int(hash.Sum32()) % totalPartitions
}
