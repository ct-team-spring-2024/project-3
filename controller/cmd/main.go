package main

import (
	"nabatdb/controller/api"
	"nabatdb/controller/internal"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func main() {
	logrus.SetLevel(logrus.DebugLevel)
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
	logrus.Info("Starting: Controller")

	internal.InitState()

	router := gin.Default()
	api.SetupRoutes(router)
	router.Run(":8080")
}
