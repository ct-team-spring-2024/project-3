package internal

import (
	"fmt"

	"github.com/sirupsen/logrus"
)

type Node struct {
	NodeId  int
	Address string
	Port    string
}

type DBStatus int

const (
	Init     DBStatus = iota
	Running
	Updating
)

var AppState *State

type State struct {
	Nodes                []Node
	NextNodeId           int
	PartitionCount       int
	ReplicationCount     int
	ClusterSize          int
	PartitionNodes       map[int][]int
	PartitionLeaderNodes map[int]int
	NodesPartitions      map[int][]int
	DBStatus             DBStatus
	NextPartitionNodes   map[int][]int // the topology that the system must diverg to
}

func InitState() {
	AppState = &State{
		Nodes: make([]Node, 0, 0),
		NextNodeId: 0,
		PartitionCount: 1,
		ReplicationCount: 3,
		ClusterSize: 0,
		PartitionNodes: make(map[int][]int),
		PartitionLeaderNodes: make(map[int]int),
		NodesPartitions: make(map[int][]int),
		DBStatus: Init,
		NextPartitionNodes: make(map[int][]int),
	}
}

func InitDB() {
	logrus.Infof("#1 Init AppState %+v", AppState)
	AppState.DBStatus = Updating
	AppState.NextPartitionNodes = AppState.PartitionNodes
	CalculateNext()
	logrus.Infof("#2 Init AppState %+v", AppState)
}

func NodeJoin(address string, port string) {
	logrus.Info("ADDED")
	AppState.Nodes = append(AppState.Nodes, Node{
		NodeId: AppState.NextNodeId,
		Address: address,
		Port: port,
	})
	AppState.NextNodeId++
	// TODO recalculate the partitions + select leader
}

func GetNode(id int) (Node, error) {
	for _, node := range AppState.Nodes {
		if node.NodeId == id {
			return node, nil
		}
	}
	return Node{}, fmt.Errorf("Node id not found")
}

func GetPartitionNodes(partitionId int) []Node {
	nodeIds := AppState.PartitionNodes[partitionId]
	result := make([]Node, 0, 0)
	for _, id := range nodeIds {
		node, _ := GetNode(id)
		result = append(result, node)
	}
	return result
}

func CalculateNext() {
	// First copy the current partitions to the new one


	// 1. TODO Delete extra replicas (replicas from dead nodes, replicas from bigger replica count,
	//                                replicas from overloaded nodes)

	// 2. Add replicas
	for pId := 0; pId < AppState.PartitionCount; pId++ {
		// calculate current count
		currentCount := len(AppState.PartitionNodes[pId])
		toBeAddedCount := AppState.ReplicationCount - currentCount
		// Find nodes will lowest number of partition
		for t := 0; t < toBeAddedCount; t++ {
			mnCnt := 0
			mnCntId := -1
			for _, node := range AppState.Nodes {
				nodeCurrentCount := len(AppState.NodesPartitions[node.NodeId])
				if mnCnt < nodeCurrentCount || mnCntId == -1 {
					mnCnt = nodeCurrentCount
					mnCntId = node.NodeId
				}
			}
			if mnCntId == -1 {
				logrus.Error("No node found for assiging the replica!!")
				panic("No node found for assiging the replica")
			}
			AppState.NextPartitionNodes[pId] = append(AppState.NextPartitionNodes[pId], mnCntId)
		}
	}
}

// func CalHash(k string) uint64 {
//	h := fnv.New64a()
//	data := []byte(k)
//	h.Write(data)
//	hashValue := h.Sum64()
//	return hashValue
// }
