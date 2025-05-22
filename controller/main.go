package main

import (
	"nabatdb/commons"
	"nabatdb/controller/cmd"
)

func main() {
	commons.InitConfig()

	cmd.Execute()
}
