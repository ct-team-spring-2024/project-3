package main

import (
	"math/rand"
	"time"
	"net"
	"fmt"
	"nabatdb/node/api"
	"nabatdb/node/internal"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

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
	for i := 0; i < 100; i++ {
		port := getRandomPort()
		if isPortAvailable(port) {
			return port
		}
	}
	ln, _ := net.Listen("tcp", ":0")
	port := ln.Addr().(*net.TCPAddr).Port
	_ = ln.Close()
	return port
}

func main() {
	logrus.SetLevel(logrus.InfoLevel)
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
	logrus.Info("Starting: Node")

	// TODO execute node-join
	internal.InitNode()

	router := gin.Default()
	api.SetupRoutes(router)
	port := getAvailablePort()
	logrus.Infof("Starting server on port %d", port)

	if err := router.Run(fmt.Sprintf(":%d", port)); err != nil {
		logrus.Fatalf("Failed to start server: %v", err)
	}
}
