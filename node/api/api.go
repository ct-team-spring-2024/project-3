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
func getLogs(c *gin.Context) {
	var requestBody struct {
		Id      int `json:"Id"`
		ShardId int `json:"Shard_Id"`
	}
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request"})
		return
	}
	logs, err := internal.Node.GetAllLogsFrom(requestBody.ShardId, requestBody.Id)
	if err != nil {
		logrus.Error(err)
		c.JSON(500, gin.H{"error": err})
	}
	 //Todo : send logs to the node
	 c.JSON(200, gin.H{"logs": logs})

}
func migrate(c *gin.Context) {
	err := internal.Node.Migrate()
	if err != nil {
		c.JSON(500, err.Error())
		return

	}
	c.JSON(200, gin.H{"status": "Migration succesful."})
}

func checkHealth(c *gin.Context) {
	logrus.Debugf("Health check requested")
	c.JSON(200, gin.H{"status": "OK"})
}
func testMergeSortedSlices(c *gin.Context) {

}

func SetupRoutes(router *gin.Engine) {
	router.POST("/get", get)
	router.POST("/set", set)
	router.POST("/delete", del)

	router.POST("/set-shard", setShard)
	router.POST("/set-shard-leader", setShardLeader)
	router.GET("/health", checkHealth)
	router.GET("/getlogs" , getLogs)
	router.POST("/migrate", migrate)
}
