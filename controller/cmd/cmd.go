// cmd/cmd.go
package cmd

import (
	"nabatdb/controller/internal"
	"nabatdb/controller/api"

	"fmt"
	"os"
	"net/http"
	"io"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
	Use:   "nabatdb",
	Short: "NabatDB - Controller and DB Management Tool",
	Long:  `NabatDB provides a controller API and DB management utilities.`,
}


func controllerFunc(cmd *cobra.Command, args []string) {
	levelStr := viper.GetString("LOG_LEVEL")
	level, err := logrus.ParseLevel(levelStr)
	if err != nil {
		level = logrus.InfoLevel
	}
	logrus.SetLevel(level)
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
	logrus.Info("Starting: Controller")

	internal.InitState()
	// TODO: Mock new nodes
	// MOCK
	// for i := 0; i < 10; i++ {
	//	port := 8081 + i
	//	address := "localhost"
	//	internal.NodeJoin(address, strconv.Itoa(port))
	//	logrus.Infof("Node joined: %s:%d", address, port)
	// }
	// MOCK
	router := gin.New()
	router.Use(func(c *gin.Context) {
		if c.Request.URL.Path == "/fetch-routing-info" {
			c.Next()
			return
		}
		gin.Logger()(c)
	})

	api.SetupRoutes(router)
	addr := fmt.Sprintf(":%s", viper.GetString("PORT"))
	router.Run(addr)
	fmt.Println("Controller started on :8080")
}

func nabatFunc(cmd *cobra.Command, args []string) {
	url := fmt.Sprintf("http://localhost:%s/start-db", viper.GetString("PORT"))
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
}

func init() {
	rootCmd.AddCommand(&cobra.Command{
		Use:   "controller",
		Short: "Start the NabatDB controller",
		Run: controllerFunc,
	})
	rootCmd.AddCommand(&cobra.Command{
		Use:   "nabat",
		Short: "Run Nabat DB",
		Run: nabatFunc,
	})
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
