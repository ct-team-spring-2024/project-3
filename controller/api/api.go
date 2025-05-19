package api

import (
	"net/http"
	"strings"
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

func fetchPartitionNodes(c *gin.Context) {
	var requestBody struct {
		PartitionId int `json:"partitionId"`
	}
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	nodes := internal.GetPartitionNodes(requestBody.PartitionId)

	type NodeResponse struct {
		ID      int    `json:"id"`
		Host    string `json:"host"`
		Port    string `json:"port"`
	}
	responseBody := make([]NodeResponse, 0, len(nodes))
	for _, node := range nodes {
		responseBody = append(responseBody, NodeResponse{
			ID:   node.NodeId,
			Host: node.Address,
			Port: node.Port,
		})
	}
	c.JSON(http.StatusOK, responseBody)
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
	router.GET("/fetch-partition-nodes", fetchPartitionNodes)
	router.POST("/start-db", startDB)
}
