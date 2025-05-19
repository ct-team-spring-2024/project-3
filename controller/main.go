package main

import (
	"math/rand"
	"time"
	"nabatdb/controller/cmd"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	cmd.Execute()
}
