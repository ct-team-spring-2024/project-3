package internal
import (
	
)

type nabatNode struct {
	ShardsRole map[int]string
	Shards map[int]InMemorydb
	Rlog []string

}

func (node *nabatNode) GetShardsRoles()(map[int]string , error) {
	return node.ShardsRole , nil
}
func (node *nabatNode) SetShard(index , shardNumber int)(error){
	return nil
}
func (node *nabatNode) SetLeaderForShard(shardNumber int) (bool , error) {
	return true , nil
}
//If it is alive it will send true otherwise the controller will timeout
func (node *nabatNode) IsAlive() bool {
	return true
}




