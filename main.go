package main

import (
	"fmt"

	"github.com/OSSystems/hulk/mqtt"
	MQTT "github.com/eclipse/paho.mqtt.golang"
)

func main() {
	opts := MQTT.NewClientOptions()
	opts.AddBroker("tcp://localhost:1883")

	client := mqtt.NewPahoClient(opts)
	client.Connect()

	h := NewHulk(client, "/tmp/")

	done := make(chan bool)

	err := h.LoadSubscribers()
	if err != nil {
		fmt.Println(err)
	}

	<-done
}
