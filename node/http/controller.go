package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	"nabatdb/commons"
)

type OpType int

const (
	Set OpType = iota
	Del
)

type OpValue interface{}

type Op struct {
	OpId    int
	OpType  OpType
	OpValue OpValue
}

type SetOpValue struct {
	Key   string
	Value []byte
}

type DelOpValue struct {
	Key string
}

func getKeyFromOp(op Op) string {
	switch v := op.OpValue.(type) {
	case SetOpValue:
		return v.Key
	case DelOpValue:
		return v.Key
	default:
		return ""
	}
}

func ConsSetOp(key string, value []byte, id int) Op {
	return Op{
		OpId:   id,
		OpType: Set,
		OpValue: SetOpValue{
			Key:   key,
			Value: value,
		},
	}
}

func ConsDelOp(key string) Op {
	return Op{
		OpType: Del,
		OpValue: DelOpValue{
			Key: key,
		},
	}
}

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

// fetches
func FetchRoutingInfo(client *http.Client, routingInfo *commons.RoutingInfo) error {
	url := fmt.Sprintf("http://%s/fetch-routing-info", viper.GetString("CONTROLLER_ADDRESS"))
	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var result struct {
		TotalPartitions int                           `json:"total_partitions"`
		RoutingInfo     map[int]commons.PartitionInfo `json:"routing_info"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}

	routingInfo.RoutingInfo = result.RoutingInfo
	routingInfo.TotalPartitions = result.TotalPartitions

	return nil
}

func FetchPartitionCount(client *http.Client) int {
	routingInfo := &commons.RoutingInfo{}
	FetchRoutingInfo(client, routingInfo)
	return routingInfo.TotalPartitions
}

func RoutingInfoUpdater(client *http.Client, routingInfo *commons.RoutingInfo) {
	go func() {
		ticker := time.NewTicker(2000 * time.Millisecond)
		defer ticker.Stop()

		for {
			<-ticker.C
			err := FetchRoutingInfo(client, routingInfo)
			if err != nil {
				logrus.Warnf("Failed to update routing data: %v", err)
			} else {
				logrus.Debug("Successfully updated routing data")
				logrus.Debugf("routingInfo => %+v", *routingInfo)
			}
		}
	}()
}

func sendOpToNode(client *http.Client, nodeAddr string, op Op) error {
	var endpoint string
	var bodyData []byte
	var err error

	switch op.OpType {
	case Set:
		endpoint = "/set"
		setVal, ok := op.OpValue.(SetOpValue)
		if !ok {
			return fmt.Errorf("invalid value for Set operation")
		}
		bodyData, err = json.Marshal(struct {
			Key   string `json:"key"`
			Value string `json:"value"`
			Id    int    `json:"id"`
		}{
			Key:   setVal.Key,
			Value: string(setVal.Value),
			Id:    op.OpId,
		})
		//TODO : Deletes must be handled differently than sets .
	case Del:
		endpoint = "/delete"
		delVal, ok := op.OpValue.(DelOpValue)
		if !ok {
			return fmt.Errorf("invalid value for Del operation")
		}
		bodyData, err = json.Marshal(struct {
			Key string `json:"key"`
		}{
			Key: delVal.Key,
		})
	default:
		return fmt.Errorf("unknown operation type")
	}

	if err != nil {
		return err
	}

	url := fmt.Sprintf("http://%s%s", nodeAddr, endpoint)

	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(bodyData))
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("received non-OK status: %d", resp.StatusCode)
	}

	return nil
}

func BroadcastOp(client *http.Client, routingInfo *commons.RoutingInfo, nodeAddress string, ops []Op) {
	go func() {
		// return
		for _, op := range ops {
			key := getKeyFromOp(op)
			if key == "" {
				continue
			}

			partition := commons.GetPartitionID(key, routingInfo.TotalPartitions)
			partitionInfo, exists := routingInfo.RoutingInfo[partition]
			if !exists {
				continue
			}

			for _, addr := range partitionInfo.NodeAddresses {
				if addr == nodeAddress {
					continue // skip self
				}

				// Send the op to the remote node
				err := sendOpToNode(client, addr, op)
				logrus.Infof("Sent the write operation to the other nodes")
				if err != nil {
					logrus.WithError(err).Warnf("Failed to propagate op to node %s", addr)
					continue
				}
			}
		}
	}()
}

func UpdatePartitionLogIndex(client *http.Client, shardId int, logIndex int) {

}

func GetLogsFromLeaderByIndex(leaderUrl string, index, Shard_Id int) ([]Op, error) {
	bodyData, _ := json.Marshal(struct {
		Id      int `json:"Id"`
		ShardId int `json:"Shard_Id"`
	}{
		Id:      index,
		ShardId: Shard_Id,
	})
	req, err := http.NewRequest("GET", leaderUrl, bytes.NewBuffer(bodyData))
	if err != nil {
		logrus.Errorf("Error creating a new http request %v", err)
	}
	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		logrus.Errorf("Error sending the http request to the leader : %v", err)
	}
	defer resp.Body.Close()
	var result struct {
		Ops []Op `json:"logs"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return result.Ops, nil

}
