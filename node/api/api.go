package api

import (
	"nabatdb/node/internal"
	"github.com/gin-gonic/gin"
)

func get(c *gin.Context) {
	var requestBody struct {
		Key string `json:"key"`
	}

	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.AbortWithStatusJSON(400, gin.H{"error": "Invalid request"})
		return
	}

	value, err := internal.DB.Get(requestBody.Key)
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

	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.AbortWithStatusJSON(400, gin.H{"error": "Invalid request"})
		return
	}

	// TODO process error
	internal.DB.Set(requestBody.Key, []byte(requestBody.Value))

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

	_, err := internal.DB.Delete(requestBody.Key)
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": "Failed to delete key"})
		return
	}

	c.JSON(200, gin.H{"message": "Key deleted successfully", "key": requestBody.Key})
}

func SetupRoutes(router *gin.Engine) {
	router.GET("/get", get)
	router.POST("/set", set)
	router.POST("/delete", del)
}
