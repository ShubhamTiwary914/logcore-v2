package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"cloud.google.com/go/pubsub/v2"
	pb "cloud.google.com/go/pubsub/v2/apiv1/pubsubpb"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	broker        string = "verne-test"
	port          int    = 1883
	mqttTopicPath string = "mqtt-source"
	projectID     string = "gcplocal-emulator"
	pubsubTopic   string = "source"
	pubsubHost    string = "gcp-emulators:8085"

	QOS         byte = 0
	PUB_WORKERS int  = 6
	QUEUE_LIM   int  = 100

	LISTENER_HEALTHFILE_PATH = "/tmp/listener.status"
	PUB_HEALTHFILE_PATH      = "/tmp/pub.status"
)

var (
	pubctx    context.Context
	pubclient *pubsub.Client
	pubJobs   chan PublishJob
	publisher *pubsub.Publisher
)

// Worker pool for publishing
type PublishJob struct {
	message []byte
}

func startWorkers() {
	for i := 0; i < PUB_WORKERS; i++ {
		go func(id int) {
			for job := range pubJobs {
				publishTopic(job.message)
			}
		}(i)
	}
}

func main() {
	//only local mode(remove in prod)
	if err := os.Setenv("PUBSUB_EMULATOR_HOST", pubsubHost); err != nil {
		log.Fatalf("Failed to set emulator host: %v", err)
	}
	//connect to pubsub (+channel for publishing)
	pubctx, pubclient = confPubSub(projectID)
	publisher = pubclient.Publisher(fmt.Sprintf("projects/%s/topics/%s", projectID, pubsubTopic))
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
	token := client.Subscribe(mqttTopicPath, QOS, mqttMessageHandler)
	token.Wait()
	log.Printf("Subscribed to topic: %s", mqttTopicPath)
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
	log.Printf("Received Message in topic %s: \n%s\n", msg.Topic(), msg.Payload())
	//offload push to publish channel
	pubJobs <- PublishJob{message: msg.Payload()}
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

func publishTopic(msg []byte) {
	publisher.Publish(pubctx, &pubsub.Message{Data: msg})
	log.Printf("Queued Message for publishing: %s", msg)
}
