package main

import (
	"pub-sub/broker"
	"pub-sub/handlers"
)

func main() {
	NewBroker := broker.NewBroker()
	go NewBroker.Start()
	defer NewBroker.Stop()
	route := handlers.InitializeRouters(NewBroker)
	route.Run(":8080")
}
