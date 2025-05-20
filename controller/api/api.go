package api

import (
	"net/http"
	"strings"
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/gin-gonic/gin"

	"nabatdb/controller/internal"
)

func nodeJoin(c *gin.Context) {
	var requestBody struct {
		Address string `json:"address"`
	}

	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	parts := strings.Split(requestBody.Address, ":")
	if len(parts) < 2 {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid address format"})
		return
	}

	address := parts[0]
	port := parts[1]

	logrus.Infof("Address: %s, Port: %s", address, port)
	nodeId := internal.NodeJoin(address, port)

	logrus.Info("#3")
	c.JSON(http.StatusOK, gin.H{
		"node_id": nodeId,
		"message": "Received and processed successfully",
		"address": address,
		"port":    port,
	})
}

func fetchRoutingInfo(c *gin.Context) {
	// Fetch the partition nodes
	partitionNodes := internal.GetPartitionNodes()
	partitionLeaderNodes := internal.GetPartitionLeaderNodes()

	type partitionInfo struct {
		NodeAddresses []string `json:"node_addresses"`
		LeaderAddress string   `json:"leader_address"`
	}

	routing := make(map[int]partitionInfo)

	for pID, nodes := range partitionNodes {
		leader, _ := internal.GetNode(partitionLeaderNodes[pID])
		leaderAddress := fmt.Sprintf("%s:%s", leader.Address, leader.Port)

		var addresses []string
		for _, nodeID := range nodes {
			node, _ := internal.GetNode(nodeID)
			addresses = append(addresses, fmt.Sprintf("%s:%s", node.Address, node.Port))
		}

		routing[pID] = partitionInfo{
			NodeAddresses: addresses,
			LeaderAddress: leaderAddress,
		}
	}


	var response struct {
		TotalPartitions int                   `json:"total_partitions"`
		RoutingInfo     map[int]partitionInfo `json:"routing_info"`
	}

	response.TotalPartitions = internal.AppState.PartitionCount
	response.RoutingInfo = routing


	c.JSON(http.StatusOK, response)
}

func startDB(c *gin.Context) {
	logrus.Info("Starting: NabatDB")
	internal.InitDB()

	c.JSON(http.StatusOK, gin.H{
		"message": "Database initialized successfully",
	})
}

func SetupRoutes(router *gin.Engine) {
	// router.GET("/node-join", func(c *gin.Context) {
	//	c.JSON(200, gin.H{"message": "Hello from another file!"})
	// })
	router.POST("/node-join", nodeJoin)
	router.GET("/fetch-routing-info", fetchRoutingInfo)
	router.POST("/start-db", startDB)
}
