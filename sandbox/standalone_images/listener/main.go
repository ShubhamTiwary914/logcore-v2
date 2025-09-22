package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"encoding/json"

	"cloud.google.com/go/pubsub/v2"
	pb "cloud.google.com/go/pubsub/v2/apiv1/pubsubpb"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	broker    string = "verne-test"
	port      int    = 1883
	mqttTopic string = "dev"
	projectID string = "gcplocal-emulator"

	QOS         byte = 0
	PUB_WORKERS int  = 6
	QUEUE_LIM   int  = 100

	LISTENER_HEALTHFILE_PATH = "/tmp/listener.status"
	PUB_HEALTHFILE_PATH      = "/tmp/pub.status"
)

const ()

var (
	pubctx    context.Context
	pubclient *pubsub.Client
	pubJobs   chan PublishJob
)

// Worker pool for publishing
type PublishJob struct {
	TopicID string
	Message string
}

func startWorkers() {
	for i := 0; i < PUB_WORKERS; i++ {
		go func(id int) {
			for job := range pubJobs {
				publishTopic(pubctx, pubclient, job.TopicID, job.Message)
			}
		}(i)
	}
}

func main() {
	//only local mode(remove in prod)
	if err := os.Setenv("PUBSUB_EMULATOR_HOST", "localhost:8085"); err != nil {
		log.Fatalf("Failed to set emulator host: %v", err)
	}
	//connect to pubsub (+channel for pub)
	pubctx, pubclient = confPubSub(projectID)
	pubJobs = make(chan PublishJob, QUEUE_LIM)
	startWorkers()

	//connect to verneMQTT
	opts := connectMQTT()
	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}
	subscribeMQTT(client)

	select {}
}

// ----------------------
// region MQTT (verne)
func connectMQTT() *mqtt.ClientOptions {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tcp://%s:%d", broker, port))
	opts.SetDefaultPublishHandler(mqttMessageHandler)
	opts.OnConnect = mqttConnectHandler
	opts.OnConnectionLost = mqttConnectLostHandler
	return opts
}

func subscribeMQTT(client mqtt.Client) {
	token := client.Subscribe(mqttTopic, QOS, mqttMessageHandler)
	token.Wait()
	log.Printf("Subscribed to topic: %s", mqttTopic)
}

var mqttConnectHandler mqtt.OnConnectHandler = func(mqtt.Client) {
	log.Printf("Connected to MQTT host: %s:%d", broker, port)
	logSuccess(LISTENER_HEALTHFILE_PATH)
}

// log use for healthcheck in listener container
func logSuccess(statusFile string) {
	f, err := os.OpenFile(statusFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("Failed to write status: %v", err)
		return
	}
	defer f.Close()
	if _, err := f.WriteString("SUCCESS\n"); err != nil {
		log.Printf("Failed to write status: %v", err)
	}
}

var mqttMessageHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	log.Printf("Received message: %s from topic: %s\n", msg.Payload(), msg.Topic())
	var payload map[string]interface{}
	if err := json.Unmarshal(msg.Payload(), &payload); err != nil {
		log.Printf("Failed to parse payload: %v", err)
		return
	}
	pubtopicID := payload["tag"].(string)
	log.Printf("Now pushing to pubsub topic: %s", pubtopicID)
	//offload push to publish channel
	pubJobs <- PublishJob{TopicID: pubtopicID, Message: string(msg.Payload())}
}

var mqttConnectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	log.Printf("Connection lost: %v", err)
}

// ----------------------
// region Pubsub
func createPubSubTopic(ctx *context.Context, client *pubsub.Client, topicID string) {
	topicName := fmt.Sprintf("projects/%s/topics/%s", projectID, topicID)
	_, err := client.TopicAdminClient.CreateTopic(*ctx, &pb.Topic{Name: topicName})
	if err != nil {
		if st, ok := status.FromError(err); !ok || st.Code() != codes.AlreadyExists {
			log.Fatalf("create topic: %v", err)
		}
	}
}

func confPubSub(projectID string) (context.Context, *pubsub.Client) {
	ctx := context.Background()
	client, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	log.Printf("Connected to PubSub host: %s", os.Getenv("PUBSUB_EMULATOR_HOST"))
	logSuccess(PUB_HEALTHFILE_PATH)
	return ctx, client
}

func publishTopic(ctx context.Context, client *pubsub.Client, topicID string, msg string) {
	topic := client.Publisher(fmt.Sprintf("projects/%s/topics/%s", projectID, topicID))
	res := topic.Publish(ctx, &pubsub.Message{Data: []byte(msg)})
	_, err := res.Get(ctx)
	if err != nil {
		log.Printf("Failed to publish message: %v", err)
	}
	fmt.Println("Published message:", msg)
}
