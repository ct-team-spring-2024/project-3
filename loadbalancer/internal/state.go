package internal

type PartitionInfo struct {
	NodeAddresses []string `json:"node_addresses"`
	LeaderAddress string   `json:"leader_address"`
}

type RoutingInfo struct {
	TotalPartitions int                   `json:"total_partitions"`
	RoutingInfo     map[int]PartitionInfo `json:"routing_info"`
}

var AppState RoutingInfo
