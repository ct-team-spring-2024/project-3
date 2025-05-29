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
	NextShardRole    map[int]string
	NextShards       map[int]*InMemorydb
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
		NextShardRole:    make(map[int]string),
		NextShards:       make(map[int]*InMemorydb),
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
	//This will wait until it is ready to be set by the migrate command
	Node.NextShardRole[shardNumber] = "follower"
	Node.NextShards[shardNumber] = InitDB()


	// TODO get snapshots + WAL from leader

	return nil
}

func (node *nabatNode) SetShardLeader(shardNumber int) (bool, error) {
	Node.NextShardRole[shardNumber] = "leader"
	return true, nil
}

// 1. (DB) Write in WAL
// 2. (DB) Write in Table
// 3. (DB) if MaxSize Reached, create new Table and add to ROTables
// 4. (Async) Send the change to other nodes
func (node *nabatNode) SetKey(key string, value []byte) {
	logrus.Infof("the Set request was sent")
	sId := commons.GetPartitionID(key, node.TotalPartitions)
	node.Shards[sId].Set(key, value)
	ops := node.Shards[sId].GetRemainingLogs()
	logrus.Infof("ops => %+v", ops)
	if node.ShardsRole[sId] == "leader" {
		logrus.Infof("the Set request was sent to follower replicas")
		nodehttp.BroadcastOp(node.ControllerClient, node.RoutingInfo, node.NodeAddress, ops)
	}
}

func (node *nabatNode) GetAllLogsFrom(partitionId int, lastLogIndex int) []nodehttp.Op {
	shard := node.Shards[partitionId]
	logs := shard.GetLogs(lastLogIndex)
	return logs
}

func (node *nabatNode) GetLogsFromLeader() error {
	for k := range node.Shards {
		leaderAddress := node.RoutingInfo.RoutingInfo[k].LeaderAddress
		//Check if itself is not the leader
		url := leaderAddress + "/getlogs"

		go exectuteLogsForShards(url, node, k)

	}

	return nil
}

func exectuteLogsForShards(url string, node *nabatNode, shardId int) error {
	logs, err := nodehttp.GetLogsFromLeaderByIndex(url, 0 , shardId)

	for _, v := range logs {
		log := nodehttp.Op{
			OpId:    v.OpId,
			OpType:  v.OpType,
			OpValue: v.OpValue,
		}

		node.Shards[shardId].ExecuteLog(log)

	}
	logrus.Errorf("Error executing logs for shard number %v", shardId)
	return err
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

func (node *nabatNode) Migrate() error {
	node.Shards = node.NextShards
	node.NextShards = make(map[int]*InMemorydb)
	node.ShardsRole = node.NextShardRole
	node.NextShardRole = make(map[int]string)
	return nil
}
func (node *nabatNode) RollBack(shardId int) {
	//It is not needed now
}

// If it is alive it will send true otherwise the controller will timeout
func (node *nabatNode) IsAlive() bool {
	return true
}
