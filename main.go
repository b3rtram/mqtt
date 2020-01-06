package main

import (
	broker "github.com/camen6ert/mqtt/mqttbroker"
)

func main() {

	server := broker.Server{}
	server.StartBroker()
	select {}
}
