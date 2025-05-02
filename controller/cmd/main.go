package main

import (
	"nabatdb/controller/api"

	"github.com/sirupsen/logrus"
	"github.com/gin-gonic/gin"
)

func main() {
	logrus.SetLevel(logrus.DebugLevel)
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
	logrus.Info("Starting: Controller")

	router := gin.Default()
	api.SetupRoutes(router)
	router.Run(":8080")
}
