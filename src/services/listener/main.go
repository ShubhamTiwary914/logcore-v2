package main

import (
	"context"
	utils "listener/utils"
	"log"
	"os"

	"cloud.google.com/go/pubsub/v2"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

var (
	//MQTT QOS=0 (fully async, no ACKs)
	QOS         byte = 0
	PUB_WORKERS int  = 6
	QUEUE_LIM   int  = 100
	//readiness probe file path (container healthcheck)
	LISTENER_HEALTHFILE_PATH = os.Getenv("MQTT_CONNECT_SUCCESS_PATH")
	PUB_HEALTHFILE_PATH      = os.Getenv("PUBSUB_CONNECT_SUCCESS_PATH")
)

var (
	//target= verne container in same pod (hence TCP at localhost:1883)
	broker      string = "localhost"
	port        int    = 1883
	projectID   string = "gcplocal-emulator"
	pubsubTopic string = "source"

	mqttTopicPath string = os.Getenv("MQTT_TOPIC")
	pubsubPort    string = "8085"
	pubsubHost    string = "gcp-emulators:8085"
)

var (
	pubctx    context.Context
	pubclient *pubsub.Client
	pubJobs   chan PublishJob
	publisher *pubsub.Publisher
)

// worker pool data packet (for publishing -> pubsub)
type PublishJob struct {
	message []byte
}

// publish pubsub concurrent worker pool
// recommended: set ~CPU cores in node
func startWorkers() {
	for i := 0; i < PUB_WORKERS; i++ {
		go func(id int) {
			for job := range pubJobs {
				publishTopic(job.message)
			}
		}(i)
	}
}

// only local mode(remove in prod)
func localConfigs() {
	//read host_ip (for k3s local) -> node where gcp emulator runs
	data, err := os.ReadFile("/envs/host_ip")
	if err != nil {
		panic(err)
	}
	hostIP := string(utils.TrimSpace(data))
	pubsubHost = utils.Sprintf("%s:%s", hostIP, pubsubPort)

	if err := os.Setenv("PUBSUB_EMULATOR_HOST", pubsubHost); err != nil {
		utils.Log(utils.LOG_ERROR, utils.Sprintf("Failed to set emulator HOST: %v", err))
	}
	utils.Log(utils.LOG_INFO, utils.Sprintf("Set PubSub Emulator Host: %s", os.Getenv("PUBSUB_EMULATOR_HOST")))
}

func main() {
	localConfigs()

	//connect to pubsub (+channel for publishing)
	pubctx, pubclient = confPubSub(projectID)
	publisher = pubclient.Publisher(utils.Sprintf("projects/%s/topics/%s", projectID, pubsubTopic))
	defer pubclient.Close()
	defer publisher.Stop()
	pubJobs = make(chan PublishJob, QUEUE_LIM)
	startWorkers()

	//connect to verneMQTT
	opts := connectMQTT()
	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}
	subscribeMQTT(client)

	//keep alive
	select {}
}

// ----------------------
// region verneMQTT

func connectMQTT() *mqtt.ClientOptions {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(utils.Sprintf("tcp://%s:%d", broker, port))
	opts.SetDefaultPublishHandler(mqttMessageHandler)
	opts.OnConnect = mqttConnectHandler
	opts.OnConnectionLost = mqttConnectLostHandler
	return opts
}

func subscribeMQTT(client mqtt.Client) {
	token := client.Subscribe(mqttTopicPath, QOS, mqttMessageHandler)
	token.Wait()
	log.Printf("Subscribed to MQTT topic: %s", mqttTopicPath)
	utils.Log(utils.LOG_INFO, utils.Sprintf("subscribed to MQTT topic: %s", mqttTopicPath))
}

var mqttConnectHandler mqtt.OnConnectHandler = func(mqtt.Client) {
	utils.Log(utils.LOG_INFO, utils.Sprintf("connected to MQTT host: %s:%d", broker, port))
	logSuccess(LISTENER_HEALTHFILE_PATH)
}

// log use for healthcheck in listener container
func logSuccess(statusFile string) {
	f, err := os.OpenFile(statusFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		utils.Log(utils.LOG_ERROR, utils.Sprintf("failed to write status at: %s  error: %v", statusFile, err))
		return
	}
	defer f.Close()
	if _, err := f.WriteString("SUCCESS\n"); err != nil {
		utils.Log(utils.LOG_ERROR, utils.Sprintf("failed to write status at: %s  error: %v", statusFile, err))
	}
}

var mqttMessageHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	log.Printf("Received Message in topic %s: \n%s\n", msg.Topic(), msg.Payload())
	//offload push to publish channel
	pubJobs <- PublishJob{message: msg.Payload()}
}

var mqttConnectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	utils.Log(utils.LOG_ERROR, utils.Sprintf("connection lost: %v", err))
}

// ----------------------
// region pubsub
func confPubSub(projectID string) (context.Context, *pubsub.Client) {
	ctx := context.Background()
	client, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		utils.Log(utils.LOG_ERROR, utils.Sprintf("failed to create pubsub client: %v", err))
	}
	utils.Log(utils.LOG_INFO, utils.Sprintf("connected to pubsub host: %s", os.Getenv("PUBSUB_EMULATOR_HOST")))
	logSuccess(PUB_HEALTHFILE_PATH)
	return ctx, client
}

func publishTopic(msg []byte) {
	publisher.Publish(pubctx, &pubsub.Message{Data: msg})
	log.Printf("Queued Message for publishing: %s", msg)
	utils.Log(utils.LOG_INFO, utils.Sprintf("queued messages"))
}
