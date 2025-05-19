package internal

import (
	"fmt"
	"nabatdb/controller/http"
	"strconv"

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
		DBStatus: Init,
		NextPartitionNodes: make(map[int][]int),
	}
}

func InitDB() {
	// Calculate the cluster topology
	AppState.DBStatus = Updating
	AppState.NextPartitionNodes = AppState.PartitionNodes
	CalculateNext()
	// Call nodes for announcing their roles
	nodePartitions := getNodePartitions(AppState.NextPartitionNodes)
	for _, node := range AppState.Nodes {
		for _, p := range nodePartitions[node.NodeId] {
			address := fmt.Sprintf("%s:%s", node.Address, node.Port)
			logrus.Infof("Assigning %d to %s", p, address)
			http.AssignPartitionToNode(address, p)
		}
	}
}

func NodeJoin(address string, port string) string {
	nodeId := AppState.NextNodeId
	AppState.Nodes = append(AppState.Nodes, Node{
		NodeId: AppState.NextNodeId,
		Address: address,
		Port: port,
	})
	AppState.NextNodeId++
	return strconv.Itoa(nodeId)
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
	AppState.NextPartitionNodes = make(map[int][]int)
	for k, v := range AppState.PartitionNodes {
		AppState.NextPartitionNodes[k] = v
	}

	// 1. TODO Delete extra replicas (replicas from dead nodes, replicas from bigger replica count,
	//                                replicas from overloaded nodes)

	// 2. Add replicas
	for pId := 0; pId < AppState.PartitionCount; pId++ {
		// calculate current count
		currentCount := len(AppState.NextPartitionNodes[pId])
		toBeAddedCount := AppState.ReplicationCount - currentCount
		// Find nodes will lowest number of partition
		for t := 0; t < toBeAddedCount; t++ {
			mnCnt := 0
			mnCntId := -1
			for _, node := range AppState.Nodes {
				nodeCurrentCount := len(getNodePartitions(AppState.NextPartitionNodes)[node.NodeId])
				if mnCnt > nodeCurrentCount || mnCntId == -1 {
					mnCnt = nodeCurrentCount
					mnCntId = node.NodeId
				}
			}
			if mnCntId == -1 {
				logrus.Error("No node found for assiging the replica!!")
				panic("No node found for assiging the replica")
			}
			logrus.Infof("mntCntId => %d", mnCntId)
			AppState.NextPartitionNodes[pId] = append(AppState.NextPartitionNodes[pId], mnCntId)
		}
	}
}

func getNodePartitions(partitionNodes map[int][]int) map[int][]int {
	nodePartitions := make(map[int][]int)
	for partitionID, nodeIDs := range partitionNodes {
		for _, nodeID := range nodeIDs {
			nodePartitions[nodeID] = append(nodePartitions[nodeID], partitionID)
		}
	}
	return nodePartitions
}



// func CalHash(k string) uint64 {
//	h := fnv.New64a()
//	data := []byte(k)
//	h.Write(data)
//	hashValue := h.Sum64()
//	return hashValue
// }
