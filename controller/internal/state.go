package internal

type Node struct {
	NodeId  int
	Address string
	Port    string
}

var AppState *State

type State struct {
	Nodes                []Node
	PartitionCount       int
	ReplicationCount     int
	ClusterSize          int
	PartitionNodes       map[int][]int
	PartitionLeaderNodes map[int]int
}

func InitState() {
	AppState = &State{
		Nodes: make([]Node, 0, 0),
		PartitionCount: 1,
		ReplicationCount: 3,
		ClusterSize: 0,
		PartitionNodes: make(map[int][]int),
		PartitionLeaderNodes: make(map[int]int),
	}
}

func NodeJoin(address string, port string) {
	AppState.Nodes = append(AppState.Nodes, Node{
		Address: address,
		Port: port,
	})
}
