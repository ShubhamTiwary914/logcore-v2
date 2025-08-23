package gcplocal

import (
	"context"
	"fmt"
	"log"
	"os"

	"cloud.google.com/go/pubsub/v2"
	pb "cloud.google.com/go/pubsub/v2/apiv1/pubsubpb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	projectID = "gcplocal-emulator"
	topicID   = "glocaltest"
	subID     = topicID
)

func confPubSub() (context.Context, *pubsub.Client) {
	if err := os.Setenv("PUBSUB_EMULATOR_HOST", "localhost:8085"); err != nil {
		log.Fatalf("Failed to set emulator host: %v", err)
	}
	ctx := context.Background()
	client, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	topicName := fmt.Sprintf("projects/%s/topics/%s", projectID, topicID)
	_, err = client.TopicAdminClient.CreateTopic(ctx, &pb.Topic{Name: topicName})
	if err != nil {
		if st, ok := status.FromError(err); !ok || st.Code() != codes.AlreadyExists {
			log.Fatalf("create topic: %v", err)
		}
	}
	return ctx, client
}

func SubTopic(stopAfter int, out chan<- string) {
	ctx, client := confPubSub()
	defer client.Close()

	cctx, cancel := context.WithCancel(ctx)
	defer cancel()

	//check topic & subscription (already exists?)
	topicName := fmt.Sprintf("projects/%s/topics/%s", projectID, topicID)
	subName := fmt.Sprintf("projects/%s/subscriptions/%s", projectID, subID)
	_, err := client.SubscriptionAdminClient.CreateSubscription(ctx, &pb.Subscription{
		Name:  subName,
		Topic: topicName,
	})
	if err != nil {
		if st, ok := status.FromError(err); !ok || st.Code() != codes.AlreadyExists {
			log.Fatalf("create subscription: %v", err)
		}
	}

	sub := client.Subscriber(fmt.Sprintf("projects/%s/subscriptions/%s", projectID, topicID))

	count := 0
	err = sub.Receive(cctx, func(ctx context.Context, m *pubsub.Message) {
		out <- string(m.Data)
		m.Ack()
		count++
		if count >= stopAfter {
			cancel()
		}
	})
	if err != nil {
		log.Fatalf("sub receive error: %v", err)
	}
}

func PubTopic(msg string) {
	ctx, client := confPubSub()
	defer client.Close()

	topic := client.Publisher(fmt.Sprintf("projects/%s/topics/%s", projectID, topicID))
	res := topic.Publish(ctx, &pubsub.Message{Data: []byte(msg)})
	_, err := res.Get(ctx)
	if err != nil {
		log.Fatalf("Failed to publish message: %v", err)
	}
	fmt.Println("Published message:", msg)
}
