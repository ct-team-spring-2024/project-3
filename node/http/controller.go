package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"nabatdb/commons"
	"net/http"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)


func SendNodeJoin(address string) (string, error) {
	url := fmt.Sprintf("http://%s/node-join", viper.GetString("CONTROLLER_ADDRESS"))

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

func FetchPartitionCount() int {
	url := fmt.Sprintf("http://%s/fetch-routing-info", viper.GetString("CONTROLLER_ADDRESS"))
	resp, _ := http.Get(url)
	defer resp.Body.Close()
	var result struct {
		TotalPartitions int                           `json:"total_partitions"`
		RoutingInfo     map[int]commons.PartitionInfo `json:"routing_info"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		panic(err)
	}
	logrus.Infof("rrrr => %+v", result)
	return result.TotalPartitions
}
