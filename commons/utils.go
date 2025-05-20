package commons

import (
	"hash/fnv"
)

func GetPartitionID(key string, totalPartitions int) int {
	hash := fnv.New32a()
	hash.Write([]byte(key))
	return int(hash.Sum32()) % totalPartitions
}

type PartitionInfo struct {
	NodeAddresses []string `json:"node_addresses"`
	LeaderAddress string   `json:"leader_address"`
}

type RoutingInfo struct {
	TotalPartitions int                   `json:"total_partitions"`
	RoutingInfo     map[int]PartitionInfo `json:"routing_info"`
}
