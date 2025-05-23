package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

func AssignPartitionToNode(address string, pId int) error {
	var requestBody struct {
		ShardID int `json:"shard_id"` // Must match exactly what the server expects
	}
	requestBody.ShardID = pId

	// Marshal the body into JSON
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request body: %w", err)
	}

	// Create a new POST request
	url := fmt.Sprintf("http://%s/set-shard", address)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set the Content-Type header
	req.Header.Set("Content-Type", "application/json")

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send HTTP request: %w", err)
	}
	defer resp.Body.Close()

	// Check response status code
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

func AssignPartitionLeaderToNode(address string, pId int) error {
	var requestBody struct {
		ShardID int `json:"shard_id"` // Must match exactly what the server expects
	}
	requestBody.ShardID = pId

	// Marshal the body into JSON
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request body: %w", err)
	}

	// Create a new POST request
	url := fmt.Sprintf("http://%s/set-shard-leader", address)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set the Content-Type header
	req.Header.Set("Content-Type", "application/json")

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send HTTP request: %w", err)
	}
	defer resp.Body.Close()

	// Check response status code
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

func CopyPartition(partitionId int, sourceAddress string, destinationAddress string) error {
	return nil

	url := fmt.Sprintf("http://%s/copy-shard", destinationAddress)

	var requestBody struct {
		PartitionID   int    `json:"partition_id"`
		SourceAddress string `json:"source_address"`
	}
	requestBody.PartitionID = partitionId
	requestBody.SourceAddress = sourceAddress

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request body: %v", err)
	}

	// Send the POST request
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Check the response status code
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

func NodeMigrate(address string) error {
	url := fmt.Sprintf("http://%s/migrate", address)

	resp, err := http.Post(url, "application/json", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

func NodeRollback(address string) {
}

func StartHealthCheck(nodeId int, address string, port string, disconnetcHandler func(int)) {
	go func() {
		ticker := time.NewTicker(5 * time.Second) // Check every 5 seconds
		defer ticker.Stop()

		healthURL := fmt.Sprintf("http://%s:%s/health", address, port)

		for {
			select {
			case <-ticker.C:
				resp, err := http.Get(healthURL)
				if err != nil || resp.StatusCode != http.StatusOK {
					logrus.Warnf("Node %d at %s:%s is unhealthy", nodeId, address, port)
					disconnetcHandler(nodeId)
				} else {
					logrus.Debugf("Node %d at %s:%s is healthy", nodeId, address, port)
				}
			}
		}
	}()
}
