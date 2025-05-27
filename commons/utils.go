package commons

import (
	"fmt"
	"hash/fnv"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func GetPartitionID(key string, totalPartitions int) int {
	hash := fnv.New32a()
	hash.Write([]byte(key))
	return int(hash.Sum32()) % totalPartitions
}

type PartitionInfo struct {
	NodeAddresses []string `json:"node_addresses"`
	LeaderAddress string   `json:"leader_address"`
}

type RoutingInfo struct {
	TotalPartitions int                   `json:"total_partitions"`
	RoutingInfo     map[int]PartitionInfo `json:"routing_info"`
}

func InitConfig() {
	err := godotenv.Load()
	if err != nil {
		logrus.Warnf("No .env file found or error loading it: %v", err)
	}
	viper.AutomaticEnv()
	logrus.Infof("Log level set to: %s", viper.GetString("LOG_LEVEL"))
}

func GetAddress(hostname string, port string) string {
	return fmt.Sprintf("%s:%s", hostname, port)
}
