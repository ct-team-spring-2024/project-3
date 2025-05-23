package api

import (
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"time"
	"bytes"
	"encoding/json"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"nabatdb/commons"
	"nabatdb/loadbalancer/internal"
)

var client = &http.Client{
	Timeout: 10 * time.Second,
}

// Utility function to forward requests
func forwardRequest(c *gin.Context, url string, bodyReader io.Reader) {
	req, _ := http.NewRequest(c.Request.Method, url, bodyReader)
	logrus.Infof("MM and Body => %+v %+v", c.Request.Method, c.Request.Body)
	resp, err := client.Do(req)

	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	c.Data(resp.StatusCode, resp.Header.Get("Content-Type"), body)
}

func get(c *gin.Context) {
	var requestBody struct {
		Key string `json:"key"`
	}
	logrus.Infof("#1")
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.AbortWithStatusJSON(400, gin.H{"error": "Invalid request"})
		return
	}

	logrus.Infof("#2")

	pid := commons.GetPartitionID(requestBody.Key, internal.AppState.TotalPartitions)
	partition := internal.AppState.RoutingInfo[pid]

	target := partition.NodeAddresses[rand.Intn(len(partition.NodeAddresses))]
	url := fmt.Sprintf("http://%s/get", target)

	bodyBytes, _ := json.Marshal(requestBody)
	bodyReader := bytes.NewReader(bodyBytes)
	forwardRequest(c, url, bodyReader)
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

	pid := commons.GetPartitionID(requestBody.Key, internal.AppState.TotalPartitions)
	partition := internal.AppState.RoutingInfo[pid]

	url := fmt.Sprintf("http://%s/set", partition.LeaderAddress)

	bodyBytes, _ := json.Marshal(requestBody)
	bodyReader := bytes.NewReader(bodyBytes)
	forwardRequest(c, url, bodyReader)
}

func del(c *gin.Context) {
	var requestBody struct {
		Key string `json:"key"`
	}

	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.AbortWithStatusJSON(400, gin.H{"error": "Invalid request"})
		return
	}

	pid := commons.GetPartitionID(requestBody.Key, internal.AppState.TotalPartitions)
	partition := internal.AppState.RoutingInfo[pid]

	url := fmt.Sprintf("http://%s/delete", partition.LeaderAddress)

	bodyBytes, _ := json.Marshal(requestBody)
	bodyReader := bytes.NewReader(bodyBytes)
	forwardRequest(c, url, bodyReader)
}

func SetupRoutes(router *gin.Engine) {
	router.POST("/get", get)
	router.POST("/set", set)
	router.POST("/delete", del)
}
