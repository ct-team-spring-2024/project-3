package internal

import (
	"fmt"
	"nabatdb/node/http"

	"github.com/sirupsen/logrus"
)

var (
	Node *nabatNode
)

type nabatNode struct {
	ShardsRole map[int]string
	Shards     map[int]*InMemorydb
	Rlog       []string
}

func InitNode(address string) {
	Node = &nabatNode{ShardsRole: make(map[int]string), Shards: make(map[int]*InMemorydb), Rlog: make([]string, 0)}
	Node.Shards[0] = &InMemorydb{Table: make(map[string][]byte, 0)}

	// Send a join to the controller
	nodeId, _ := http.SendNodeJoin(address)
	logrus.Infof("nodeId => %s", nodeId)
}

func (node *nabatNode) GetShardsRoles() (map[int]string, error) {
	return node.ShardsRole, nil
}
func (node *nabatNode) SetShard(index, shardNumber int) error {
	return nil
}
func (node *nabatNode) SetLeaderForShard(shardNumber int) (bool, error) {
	return true, nil
}
func (node *nabatNode) SetKey(key string, value []byte) {
	//See if shard is leader first
	Node.Shards[0].Set(key, value)

}
func (node *nabatNode) GetKey(key string) ([]byte, error) {
	value, err := Node.Shards[0].Get(key)
	return value, err
}
func (node *nabatNode) DeleteKey(key string) error {
	ok, err := Node.Shards[0].Delete(key)
	if !ok{
		return fmt.Errorf("Error occured Deleting the key %v with error  %v \n" , key , err)
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
