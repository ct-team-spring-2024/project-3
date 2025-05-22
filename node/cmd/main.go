package main

import (
	"fmt"
	"math/rand"
	"nabatdb/node/api"
	"nabatdb/node/internal"
	"net"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/joho/godotenv"
)

func initConfig() {
	err := godotenv.Load()
	if err != nil {
		logrus.Fatalf("No .env file found or error loading it: %v", err)
	}
	viper.AutomaticEnv()
	logrus.Infof("Log level set to: %s", viper.GetString("LOG_LEVEL"))
}

func getRandomPort() int {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return r.Intn(65535-1024) + 1024
}

func isPortAvailable(port int) bool {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return false
	}
	_ = ln.Close()
	return true
}

func getAvailablePort() int {
	if viper.GetString("PORT") != "" {
		return viper.GetInt("PORT")
	} else {
		for i := 0; i < 100; i++ {
			port := getRandomPort()
			if isPortAvailable(port) {
				return port
			}
		}
	}
	panic("no port found")
}

func main() {
	initConfig()

	levelStr := viper.GetString("LOG_LEVEL")
	level, err := logrus.ParseLevel(levelStr)
	if err != nil {
		level = logrus.InfoLevel
	}
	logrus.SetLevel(level)
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
	logrus.Info("Starting: Node")

	port := getAvailablePort()
	// TODO localhost should be the ip
	address := fmt.Sprintf("%s:%d","localhost", port)
	internal.InitNode(address)

	router := gin.Default()
	api.SetupRoutes(router)
	logrus.Infof("Starting server on port %d", port)

	if err := router.Run(fmt.Sprintf(":%d", port)); err != nil {
		logrus.Fatalf("Failed to start server: %v", err)
	}
}
