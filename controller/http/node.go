package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"bytes"
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
