package internal

import (
	"net/http"
	"time"

	"github.com/sirupsen/logrus"

	"nabatdb/commons"
	nodehttp "nabatdb/node/http"
)

type nabatNode struct {
	ShardsRole       map[int]string
	Shards           map[int]*InMemorydb
	TotalPartitions  int
	NodeAddress      string
	RoutingInfo      *commons.RoutingInfo
	ControllerClient *http.Client
}

var (
	Node *nabatNode
)

func InitNode(nodeAddress string) {
	client := &http.Client{
		Timeout: time.Second * 5,
	}
	totalPartitions := nodehttp.FetchPartitionCount(client)
	nodeId, _ := nodehttp.SendNodeJoin(nodeAddress)
	Node = &nabatNode{
		ShardsRole:       make(map[int]string),
		Shards:           make(map[int]*InMemorydb),
		TotalPartitions:  totalPartitions,
		NodeAddress:      nodeAddress,
		RoutingInfo:      &commons.RoutingInfo{},
		ControllerClient: client,
	}
	nodehttp.RoutingInfoUpdater(client, Node.RoutingInfo)

	logrus.Infof("nodeId => %s", nodeId)
}

func (node *nabatNode) GetShardsRoles() (map[int]string, error) {
	return node.ShardsRole, nil
}

func (node *nabatNode) SetShard(shardNumber int) error {
	Node.ShardsRole[shardNumber] = "follower"
	Node.Shards[shardNumber] = InitDB()
	// TODO get snapshots + WAL from leader
	return nil
}

func (node *nabatNode) SetShardLeader(shardNumber int) (bool, error) {
	Node.ShardsRole[shardNumber] = "leader"
	return true, nil
}

// 1. (DB) Write in WAL
// 2. (DB) Write in Table
// 3. (DB) if MaxSize Reached, create new Table and add to ROTables
// 4. (Async) Send the change to other nodes
func (node *nabatNode) SetKey(key string, value []byte) {
	sId := commons.GetPartitionID(key, node.TotalPartitions)
	node.Shards[sId].Set(key, value)
	ops := node.Shards[sId].GetRemainingLogs()
	logrus.Infof("ops => %+v", ops)
	if node.ShardsRole[sId] == "leader" {
		nodehttp.BroadcastOp(node.ControllerClient, node.RoutingInfo, node.NodeAddress, ops)
	}
}

func (node *nabatNode) GetKey(key string) ([]byte, error) {
	sId := commons.GetPartitionID(key, node.TotalPartitions)
	value, err := Node.Shards[sId].Get(key)
	return value, err
}

func (node *nabatNode) DeleteKey(key string) error {
	sId := commons.GetPartitionID(key, node.TotalPartitions)
	node.Shards[sId].Delete(key)
	ops := node.Shards[sId].GetRemainingLogs()
	logrus.Infof("ops => %+v", ops)
	if node.ShardsRole[sId] == "leader" {
		nodehttp.BroadcastOp(node.ControllerClient, node.RoutingInfo, node.NodeAddress, ops)
	}
	return nil
}

// If it is alive it will send true otherwise the controller will timeout
func (node *nabatNode) IsAlive() bool {
	return true
}
func updateState() {
	//Things this should do
	//Read the logs and update the shards
	//Send is Alive events back to controller
	//Answer the set and get from controller
	for {

	}

}
