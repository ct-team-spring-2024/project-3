package api

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
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

	host := parts[0]
	port := parts[1]

	logrus.Infof("Address: %s, Port: %s", host, port)

	c.JSON(http.StatusOK, gin.H{
		"message": "Received and processed successfully",
		"address": host,
		"port":    port,
	})
}

func SetupRoutes(router *gin.Engine) {
	// router.GET("/node-join", func(c *gin.Context) {
	//	c.JSON(200, gin.H{"message": "Hello from another file!"})
	// })
	router.POST("/node-join", nodeJoin)
}
