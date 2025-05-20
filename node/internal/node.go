package internal

import (
	"fmt"
	"nabatdb/commons"
	"nabatdb/node/http"

	"github.com/sirupsen/logrus"
)

type nabatNode struct {
	ShardsRole  map[int]string
	Shards      map[int]*InMemorydb
	Rlog        []string
	TotalPartitions int
}

var (
	Node *nabatNode
)

func InitNode(address string) {
	totalPartitions := http.FetchPartitionCount()
	Node = &nabatNode{
		ShardsRole:      make(map[int]string),
		Shards:          make(map[int]*InMemorydb),
		Rlog:            make([]string, 0),
		TotalPartitions: totalPartitions,
	}
	nodeId, _ := http.SendNodeJoin(address)

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

func (node *nabatNode) SetKey(key string, value []byte) {
	sId := commons.GetPartitionID(key, node.TotalPartitions)
	Node.Shards[sId].Set(key, value)
}

func (node *nabatNode) GetKey(key string) ([]byte, error) {
	sId := commons.GetPartitionID(key, node.TotalPartitions)
	value, err := Node.Shards[sId].Get(key)
	return value, err
}

func (node *nabatNode) DeleteKey(key string) error {
	ok, err := Node.Shards[0].Delete(key)
	if !ok {
		return fmt.Errorf("Error occured Deleting the key %v with error  %v \n", key, err)
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
