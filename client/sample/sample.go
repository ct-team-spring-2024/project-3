package main

import (
	"nabatdbclient"
	"github.com/sirupsen/logrus"
)

func main(){
	client, err := nabatdbclient.Connect("http://localhost:80")
	if err != nil {
		logrus.Fatal(err)
	}

	key := "mykey"
	//value := "myvalue"

	// // Set
	// if err := client.Set(key, value); err != nil {
	// 	logrus.Println("Set failed:", err)
	// }

	// Get
	val, err := client.Get(key)
	if err != nil {
		logrus.Println("Get failed:", err)
	} else {
		logrus.Println("Got value:", val)
	}

	// // Delete
	// if err := client.Delete(key); err != nil {
	// 	logrus.Println("Delete failed:", err)
	// }
}
