package api

import (
	"nabatdb/node/internal"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func get(c *gin.Context) {
	var requestBody struct {
		Key string `json:"key"`
	}

	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.AbortWithStatusJSON(400, gin.H{"error": "Invalid request"})
		return
	}
	value, err := internal.Node.GetKey(requestBody.Key)
	if err != nil {
		c.AbortWithStatusJSON(404, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"key": requestBody.Key, "value": string(value)})
}

func set(c *gin.Context) {
	var requestBody struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	}

	logrus.Infof("#1")
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.AbortWithStatusJSON(400, gin.H{"error": "Invalid request"})
		return
	}
	logrus.Infof("#2")

	// TODO process error
	internal.Node.SetKey(requestBody.Key, []byte(requestBody.Value))
	c.JSON(200, gin.H{"message": "Key-Value pair saved", "key": requestBody.Key})
}

func del(c *gin.Context) {
	var requestBody struct {
		Key string `json:"key"`
	}

	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.AbortWithStatusJSON(400, gin.H{"error": "Invalid request"})
		return
	}

	err := internal.Node.DeleteKey(requestBody.Key)
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": "Failed to delete key"})
		return
	}

	c.JSON(200, gin.H{"message": "Key deleted successfully", "key": requestBody.Key})
}

func setShard(c *gin.Context) {
	var requestBody struct {
		ShardID int `json:"shard_id"`
	}

	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request"})
		return
	}

	if err := internal.Node.SetShard(requestBody.ShardID); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"status": "Shard set successfully", "shard_id": requestBody.ShardID})
}

func setShardLeader(c *gin.Context) {
	var requestBody struct {
		ShardID int `json:"shard_id"`
	}

	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request"})
		return
	}

	if _, err := internal.Node.SetShardLeader(requestBody.ShardID); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"status": "Shard Leader set successfully", "shard_id": requestBody.ShardID})
}

func migrate(c *gin.Context) {
	err := internal.Node.Migrate()
	if err != nil {
		c.JSON(500 , err.Error())
		return

	}
	c.JSON(200 , gin.H{"status" : "Migration succesful."})
}

func copyShard(c *gin.Context) {
	var requestBody struct {
		PartitionID    int    `json:"partition_id"`
		SourceAddress string `json:"source_address"`
	}

	// Bind and validate JSON input
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request format", "details": err.Error()})
		return
	}

	// Call internal Node method to perform shard copy
	// if err := internal.Node.CopyShard(requestBody.PartitionID, requestBody.SourceAddress); err != nil {
	//	c.JSON(500, gin.H{"error": err.Error()})
	//	return
	// }

	// Return success response
	c.JSON(200, gin.H{
		"status":         "Shard copied successfully",
		"partition_id":   requestBody.PartitionID,
		"source_address": requestBody.SourceAddress,
	})
}

func getDB(c *gin.Context) {
	var requestBody struct {
		ShardID int `json:"shard_id"`
	}

	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request format", "details": err.Error()})
		return
	}

	dbData := internal.Node.GetDB(requestBody.ShardID)
	logrus.Infof("dbData => %+v", dbData)

	c.JSON(200, gin.H{
		"status":   "Database retrieved successfully",
		"shard_id": requestBody.ShardID,
		"data":     dbData,
	})
}

func getLogs(c *gin.Context) {
	var requestBody struct {
		ShardID int `json:"shard_id"`
	}

	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request format", "details": err.Error()})
		return
	}

	logs := internal.Node.GetLogs(requestBody.ShardID)

	c.JSON(200, gin.H{
		"status":   "Logs retrieved successfully",
		"shard_id": requestBody.ShardID,
		"logs":     logs,
	})
}

func checkHealth(c *gin.Context) {
	logrus.Debugf("Health check requested")
	c.JSON(200, gin.H{"status": "OK"})
}

func SetupRoutes(router *gin.Engine) {
	router.POST("/get", get)
	router.POST("/set", set)
	router.POST("/delete", del)

	router.POST("/set-shard", setShard)
	router.POST("/set-shard-leader", setShardLeader)
	router.GET("/health", checkHealth)
	router.POST("/migrate", migrate)
	router.POST("/copy-shard", copyShard)
	router.GET("/get-db", getDB)
	router.GET("/get-logs", getLogs)
}
