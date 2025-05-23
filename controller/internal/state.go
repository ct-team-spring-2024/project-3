package internal

import (
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/sirupsen/logrus"

	"nabatdb/commons"
	"nabatdb/controller/http"
)

type Node struct {
	NodeId  int
	Address string
	Port    string
	Status  NodeStatus
}

type NodeStatus int

const (
	Healthy NodeStatus = iota
	Unhealthy
)

type DBStatus int

const (
	Init DBStatus = iota
	Running
	Updating
)

var AppState *State

type State struct {
	DBStatus                 DBStatus
	Nodes                    []Node
	NextNodeId               int
	PartitionCount           int
	ReplicationCount         int
	ClusterSize              int
	PartitionNodes           map[int][]int
	PartitionLeaderNodes     map[int]int
	NextPartitionNodes       map[int][]int // the topology that the system must diverg to
	NextPartitionLeaderNodes map[int]int
	NextActionTrigger        chan struct{}
}

func InitState() {
	AppState = &State{
		DBStatus:                 Init,
		Nodes:                    make([]Node, 0, 0),
		NextNodeId:               0,
		PartitionCount:           1,
		ReplicationCount:         3,
		ClusterSize:              0,
		PartitionNodes:           make(map[int][]int),
		PartitionLeaderNodes:     make(map[int]int),
		NextPartitionNodes:       make(map[int][]int),
		NextPartitionLeaderNodes: make(map[int]int),
		NextActionTrigger:        make(chan struct{}),
	}
}

// Handlers
// TODO Before the start, send a fallback to disble previous operations.
func InitDB() {
	calculateNext()
	nodePartitions := getNodePartitions(AppState.NextPartitionNodes)
	for _, node := range AppState.Nodes {
		for _, p := range nodePartitions[node.NodeId] {
			address := fmt.Sprintf("%s:%s", node.Address, node.Port)
			logrus.Infof("Assigning %d to %d (%s)", p, node.NodeId, address)
			http.AssignPartitionToNode(address, p)
		}
	}
	for pId, nodeId := range AppState.NextPartitionLeaderNodes {
		node, _ := getNode(nodeId)
		address := fmt.Sprintf("%s:%s", node.Address, node.Port)
		logrus.Infof("Assigning Leader of %d to %d (%s)", pId, nodeId, address)
		http.AssignPartitionLeaderToNode(address, pId)
	}
	AppState.DBStatus = Running
	AppState.PartitionNodes = AppState.NextPartitionNodes
	AppState.PartitionLeaderNodes = AppState.NextPartitionLeaderNodes
	AppState.NextPartitionNodes = make(map[int][]int)
	AppState.NextPartitionLeaderNodes = make(map[int]int)
	// Start the async handler for topology change
	go NextAction()
}

func hasPartition(partitionNodesMap map[int][]int, nodeId int, partitionId int) bool {
	for _, p := range partitionNodesMap[nodeId] {
		if p == partitionId {
			return true
		}
	}
	return false
}

func getAddedPartitions(current map[int][]int, next map[int][]int) map[int][]int {
	added := make(map[int][]int)

	for nodeId, partitions := range next {
		for _, p := range partitions {
			if !hasPartition(current, nodeId, p) {
				added[nodeId] = append(added[nodeId], p)
			}
		}
	}

	return added
}

func getRemovedPartitions(current map[int][]int, next map[int][]int) map[int][]int {
	removed := make(map[int][]int)

	for nodeId, partitions := range current {
		for _, p := range partitions {
			if !hasPartition(next, nodeId, p) {
				removed[nodeId] = append(removed[nodeId], p)
			}
		}
	}

	return removed
}

func NextAction() {
	for {
		select {
		case <-AppState.NextActionTrigger:
			logrus.Info("Received signal to trigger next action")

			// added := getAddedPartitions(AppState.PartitionNodes, AppState.NextPartitionNodes)
			// removed := getRemovedPartitions(AppState.PartitionNodes, AppState.NextPartitionNodes)

			// Update application state after processing the signal.
			AppState.PartitionNodes = AppState.NextPartitionNodes
			AppState.PartitionLeaderNodes = AppState.NextPartitionLeaderNodes
			AppState.NextPartitionNodes = make(map[int][]int)
			AppState.NextPartitionLeaderNodes = make(map[int]int)

		default:
			// Optional: Add a small delay to avoid busy waiting.
			time.Sleep(100 * time.Millisecond)
		}
	}
}

func NodeDisconnect(nodeId int) {
	// TODO if dbstatus is updating, do a Rollback on all nodes!
	AppState.DBStatus = Updating
	node, _ := getNode(nodeId)
	if node.Status == Unhealthy {
		return
	}
	node.Status = Unhealthy
	setNode(nodeId, node)
	calculateNext()

	logrus.Infof("After Disconnect => \n %+v \n %+v", AppState.NextPartitionNodes, AppState.NextPartitionLeaderNodes)
	AppState.NextActionTrigger <- struct{}{}
	// TODO Given the next, and the current, we will initiate a number of tasks:
	// - Migrate: Telling the nodes to stop using the previous version
	// - Rollback: Forget about the copy and everything. Stick to old data.
	// - GetCopy(node1, node2): Copy the data from node1 to node2.
	// - AssignNextPartitionToNode: In the NextVersion, assign partition to node.
	// - AssignNextPartitionLeaderToNode
}

// TODO The reliving of a unhealthy node
func NodeJoin(address string, port string) string {
	nodeId := AppState.NextNodeId
	AppState.Nodes = append(AppState.Nodes, Node{
		NodeId:  AppState.NextNodeId,
		Address: address,
		Port:    port,
		Status:  Healthy,
	})
	AppState.NextNodeId++

	http.StartHealthCheck(nodeId, address, port, NodeDisconnect)

	return strconv.Itoa(nodeId)
}

func FetchRoutingInfo() commons.RoutingInfo {
	partitionNodes := getPartitionNodes()
	partitionLeaderNodes := getPartitionLeaderNodes()

	routing := make(map[int]commons.PartitionInfo)

	for pID, nodes := range partitionNodes {
		leader, _ := getNode(partitionLeaderNodes[pID])
		leaderAddress := fmt.Sprintf("%s:%s", leader.Address, leader.Port)

		var addresses []string
		for _, nodeID := range nodes {
			node, _ := getNode(nodeID)
			if node.Status == Unhealthy {
				continue
			}
			addresses = append(addresses, fmt.Sprintf("%s:%s", node.Address, node.Port))
		}

		routing[pID] = commons.PartitionInfo{
			NodeAddresses: addresses,
			LeaderAddress: leaderAddress,
		}
	}
	return commons.RoutingInfo{
		TotalPartitions: AppState.PartitionCount,
		RoutingInfo:     routing,
	}
}

// Utils
// TODO code smell! outer code should not work with node
func getNode(id int) (Node, error) {
	for _, node := range AppState.Nodes {
		if node.NodeId == id {
			return node, nil
		}
	}
	return Node{}, fmt.Errorf("Node id not found")
}

func setNode(id int, newNode Node) error {
	for i, node := range AppState.Nodes {
		if node.NodeId == id {
			AppState.Nodes[i] = newNode
			return nil
		}
	}
	return fmt.Errorf("Node id not found")
}

// TODO code smell! outer code should not work with PartitionNodes
func getPartitionNodes() map[int][]int {
	return AppState.PartitionNodes
}

func getPartitionLeaderNodes() map[int]int {
	return AppState.PartitionLeaderNodes
}

// Delete the node from the NextPartitionNodes and NextPartitionLeaderNodes
func deleteFromNextNodes(nodeId int) {
	for pId, nodes := range AppState.NextPartitionNodes {
		newNodes := make([]int, 0, 0)
		for _, n := range nodes {
			if n != nodeId {
				newNodes = append(newNodes, n)
			}
		}
		AppState.NextPartitionNodes[pId] = newNodes
	}
	newNextPartitionLeaderNodes := make(map[int]int)
	for pId, node := range AppState.NextPartitionLeaderNodes {
		if node != nodeId {
			newNextPartitionLeaderNodes[pId] = node
		}
	}
	AppState.NextPartitionLeaderNodes = newNextPartitionLeaderNodes
}

func calculateNext() {
	// 0. First copy the current partitions to the new one
	AppState.NextPartitionNodes = make(map[int][]int)
	for k, v := range AppState.PartitionNodes {
		AppState.NextPartitionNodes[k] = v
	}

	// 1. TODO Delete extra replicas (replicas from dead nodes,
	//                                replicas which are extra,
	//                                replicas from overloaded nodes)
	// Dead nodes
	nodePartitions := getNodePartitions(AppState.NextPartitionNodes)
	for _, node := range AppState.Nodes {
		if len(nodePartitions[node.NodeId]) == 0 {
			continue
		}
		if node.Status == Unhealthy {
			deleteFromNextNodes(node.NodeId)
		}
	}

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
				// Skip if node is unhealthy
				if node.Status == Unhealthy {
					continue
				}
				// Skip if node already has this partition
				hasPartition := false
				for _, existingPid := range AppState.NextPartitionNodes[pId] {
					if existingPid == node.NodeId {
						hasPartition = true
						break
					}
				}
				if hasPartition {
					continue
				}

				nodeCurrentCount := len(getNodePartitions(AppState.NextPartitionNodes)[node.NodeId])
				if mnCnt > nodeCurrentCount || mnCntId == -1 {
					mnCnt = nodeCurrentCount
					mnCntId = node.NodeId
				}
			}
			if mnCntId == -1 {
				logrus.Error("No node found for assiging the replica!!")
				if currentCount == 0 {
					logrus.Error("There is no replica. Possible loss of data.")
				}
				break
			}
			logrus.Infof("mntCntId => %d", mnCntId)
			AppState.NextPartitionNodes[pId] = append(AppState.NextPartitionNodes[pId], mnCntId)
		}
	}
	// 3. Assign Leaders (In each partition, assign one random replica as the leader)
	for pId := 0; pId < AppState.PartitionCount; pId++ {
		nodes := AppState.NextPartitionNodes[pId]
		if len(nodes) == 0 {
			fmt.Printf("Partition %d has no nodes assigned\n", pId)
			continue
		}
		// Pick a random node as leader
		leader := nodes[rand.Intn(len(nodes))]
		logrus.Infof("Partition %d: Random leader selected -> %d\n", pId, leader)
		AppState.NextPartitionLeaderNodes[pId] = leader
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
