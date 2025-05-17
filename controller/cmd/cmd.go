// cmd/cmd.go
package cmd

import (
	"nabatdb/controller/internal"
	"nabatdb/controller/api"

	"fmt"
	"os"
	"strconv"
	"net/http"
	"io"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "nabatdb",
	Short: "NabatDB - Controller and DB Management Tool",
	Long:  `NabatDB provides a controller API and DB management utilities.`,
}

var controllerCmd = &cobra.Command{
	Use:   "controller",
	Short: "Start the NabatDB controller",
	Run: func(cmd *cobra.Command, args []string) {
		logrus.SetLevel(logrus.DebugLevel)
		logrus.SetFormatter(&logrus.TextFormatter{
			FullTimestamp: true,
		})
		logrus.Info("Starting: Controller")

		internal.InitState()
		// TODO: Mock new nodes
		// MOCK
		for i := 0; i < 10; i++ {
			port := 8081 + i
			address := "localhost"
			internal.NodeJoin(address, strconv.Itoa(port))
			logrus.Infof("Node joined: %s:%d", address, port)
		}
		// MOCK

		router := gin.Default()
		api.SetupRoutes(router)
		router.Run(":8080")
		fmt.Println("Controller started on :8080")

	},
}

var nabatCmd = &cobra.Command{
	Use:   "nabat",
	Short: "Run Nabat DB",
	Run: func(cmd *cobra.Command, args []string) {
		url := "http://localhost:8080/start-db"
		req, err := http.NewRequest("POST", url, nil)
		if err != nil {
			logrus.Fatalf("Failed to create HTTP request: %v", err)
		}

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			logrus.Fatalf("Failed to send request to /start-db: %v", err)
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)
		if resp.StatusCode != http.StatusOK {
			logrus.Fatalf("Failed to initialize DB via API: %d - %s", resp.StatusCode, body)
		}

		fmt.Printf("DB Started: %s\n", body)
	},
}

func init() {
	rootCmd.AddCommand(controllerCmd)
	rootCmd.AddCommand(nabatCmd)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
