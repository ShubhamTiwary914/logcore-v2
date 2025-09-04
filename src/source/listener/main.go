package main

import (
	"fmt"
	"log"
	"os"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

const (
	broker string = "127.0.0.1"
	port   int    = 1883
	topic  string = "dev"
	QOS    byte   = 0
)

var messagePubHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	log.Printf("Received message: %s from topic: %s\n", msg.Payload(), msg.Topic())
}

var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
	log.Printf("Connected to MQTT host: %s:%d", broker, port)
	logSuccess()
}

func logSuccess() {
	f, err := os.OpenFile("/tmp/listener.status", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("Failed to write status: %v", err)
		return
	}
	defer f.Close()

	if _, err := f.WriteString("SUCCESS\n"); err != nil {
		log.Printf("Failed to write status: %v", err)
	}
}

var connectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	log.Printf("Connection lost: %v", err)
}

func main() {
	//connect
	opts := connectMQTT()
	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}
	subscribeMQTT(client)

	select {}
}

func connectMQTT() *mqtt.ClientOptions {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tcp://%s:%d", broker, port))
	opts.SetDefaultPublishHandler(messagePubHandler)
	opts.OnConnect = connectHandler
	opts.OnConnectionLost = connectLostHandler

	return opts
}

func subscribeMQTT(client mqtt.Client) {
	token := client.Subscribe(topic, QOS, messagePubHandler)
	token.Wait()
	log.Printf("Subscribed to topic: %s", topic)
}
