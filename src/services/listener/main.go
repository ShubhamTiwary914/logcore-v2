package main

import (
	"context"
	"log"
	"os"

	"cloud.google.com/go/pubsub/v2"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

const (
	//target= verne container in same pod (hence TCP at localhost:1883)
	broker        string = "localhost"
	port          int    = 1883
	mqttTopicPath string = "mqtt-source"
	projectID     string = "gcplocal-emulator"
	pubsubTopic   string = "source"

	//MQTT QOS=0 (fully async, no ACKs)
	QOS         byte = 0
	PUB_WORKERS int  = 6
	QUEUE_LIM   int  = 100

	LISTENER_HEALTHFILE_PATH = "/tmp/listener.status"
	PUB_HEALTHFILE_PATH      = "/tmp/pub.status"
)

var (
	pubsubPort string = "8085"
	pubsubHost string = "gcp-emulators:8085"
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
	hostIP := string(trimSpace(data))
	pubsubHost = sprintf("%s:%s", hostIP, pubsubPort)

	if err := os.Setenv("PUBSUB_EMULATOR_HOST", pubsubHost); err != nil {
		log.Fatalf("Failed to set emulator host: %v", err)
	}
	log.Printf("\nPUBSUB_HOST: %s", os.Getenv("PUBSUB_EMULATOR_HOST"))
}

func main() {
	localConfigs()

	//connect to pubsub (+channel for publishing)
	pubctx, pubclient = confPubSub(projectID)
	publisher = pubclient.Publisher(sprintf("projects/%s/topics/%s", projectID, pubsubTopic))
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
// region verneMQTT

func connectMQTT() *mqtt.ClientOptions {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(sprintf("tcp://%s:%d", broker, port))
	opts.SetDefaultPublishHandler(mqttMessageHandler)
	opts.OnConnect = mqttConnectHandler
	opts.OnConnectionLost = mqttConnectLostHandler
	return opts
}

func subscribeMQTT(client mqtt.Client) {
	token := client.Subscribe(mqttTopicPath, QOS, mqttMessageHandler)
	token.Wait()
	log.Printf("Subscribed to MQTT topic: %s", mqttTopicPath)
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
// region pubsub

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

// ---------------------
// region helpers
func trimSpace(b []byte) []byte {
	start, end := 0, len(b)
	for start < end && (b[start] == ' ' || b[start] == '\n' || b[start] == '\t' || b[start] == '\r') {
		start++
	}
	for end > start && (b[end-1] == ' ' || b[end-1] == '\n' || b[end-1] == '\t' || b[end-1] == '\r') {
		end--
	}
	return b[start:end]
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	sign := ""
	if n < 0 {
		sign = "-"
		n = -n
	}
	digits := []byte{}
	for n > 0 {
		d := byte(n % 10)
		digits = append([]byte{d + '0'}, digits...)
		n /= 10
	}
	return sign + string(digits)
}

// alternative for fmt.Sprintf() method
func sprintf(format string, args ...interface{}) string {
	out := []byte{}
	argIndex := 0
	for i := 0; i < len(format); i++ {
		if format[i] == '%' && i+1 < len(format) {
			i++
			switch format[i] {
			case 's':
				if v, ok := args[argIndex].(string); ok {
					out = append(out, v...)
				}
				argIndex++
			case 'd':
				if v, ok := args[argIndex].(int); ok {
					out = append(out, itoa(v)...)
				}
				argIndex++
			default:
				out = append(out, '%', format[i])
			}
		} else {
			out = append(out, format[i])
		}
	}
	return string(out)
}
