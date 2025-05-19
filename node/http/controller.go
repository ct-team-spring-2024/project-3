package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/sirupsen/logrus"
)

var controllerAddr string = "localhost:8080"

func SendNodeJoin(address string) (string, error) {
	url := fmt.Sprintf("http://%s/node-join", controllerAddr)

	var requestBody struct {
		Address string `json:"address"`
	}
	requestBody.Address = address
	logrus.Infof("requestBody => %+v", requestBody)

	bodyBytes, _ := json.Marshal(requestBody)

	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	var responseBody struct {
		Message string `json:"message"`
		NodeID  string `json:"node_id"`
		Address string `json:"address"`
		Port    string `json:"port"`
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send join request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, body)
	}

	if err := json.NewDecoder(resp.Body).Decode(&responseBody); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	return responseBody.NodeID, nil
}
