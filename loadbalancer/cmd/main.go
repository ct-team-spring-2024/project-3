package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	"nabatdb/commons"
	"nabatdb/loadbalancer/api"
	"nabatdb/loadbalancer/internal"
)



func fetchRoutingData(client *http.Client, url string, routingData *internal.RoutingInfo) error {
	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var result struct {
		TotalPartitions int                   `json:"total_partitions"`
		RoutingInfo     map[int]internal.PartitionInfo `json:"routing_info"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}

	routingData.RoutingInfo = result.RoutingInfo
	routingData.TotalPartitions = result.TotalPartitions

	return nil
}

func main() {
	commons.InitConfig()

	client := &http.Client{
		Timeout: time.Second * 5,
	}

	fetchRoutingDataURL := fmt.Sprintf("http://%s/fetch-routing-info", viper.GetString("CONTROLLER_ADDRESS"))
	err := fetchRoutingData(client, fetchRoutingDataURL, &internal.AppState)
	if err != nil {
		logrus.Fatalf("Failed to fetch initial routing data: %v", err)
	}

	// Periodically update every 500 ms
	go func() {
		ticker := time.NewTicker(500 * time.Millisecond)
		defer ticker.Stop()

		for {
			<-ticker.C
			err := fetchRoutingData(client, fetchRoutingDataURL, &internal.AppState)
			if err != nil {
				logrus.Warnf("Failed to update routing data: %v", err)
			} else {
				logrus.Info("Successfully updated routing data")
				logrus.Infof("routingInfo => %+v", internal.AppState)
			}
		}
	}()

	router := gin.Default()
	api.SetupRoutes(router)
	address := fmt.Sprintf(":%s", viper.GetString("PORT"))
	router.Run(address)

	select {} // Keep the main function running
}
