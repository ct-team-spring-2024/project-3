package api

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"nabatdb/controller/internal"
)

// Add a node. Also start a goroutine to check the health.
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
	response :=  internal.FetchRoutingInfo()
	c.JSON(http.StatusOK, response)
}

func startDB(c *gin.Context) {
	logrus.Info("Starting: NabatDB")

	// 1- calculate the cluster topology + inform the nodes
	internal.InitDB()

	c.JSON(http.StatusOK, gin.H{
		"message": "Database initialized successfully",
	})
}

func SetupRoutes(router *gin.Engine) {
	router.POST("/node-join", nodeJoin)
	router.GET("/fetch-routing-info", fetchRoutingInfo)
	router.POST("/start-db", startDB)
}
