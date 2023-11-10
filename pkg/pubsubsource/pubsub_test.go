package pubsubsource

import (
	"context"
	"fmt"
	"github.com/numaproj-contrib/gcp-pub-sub-source-go/pkg/mocks"
	"github.com/numaproj/numaflow-go/pkg/sourcer"
	"github.com/stretchr/testify/assert"
	"log"
	"os"
	"testing"
	"time"

	"cloud.google.com/go/pubsub"
)

const TopicID = "pubsub-test"

var (
	pubsubClient   *pubsub.Client
	subscriptionID = "subscription-0908"
)

// Simplified publish function.
func publish(ctx context.Context, client *pubsub.Client, topicID, msg string) error {
	t := client.Topic(topicID)
	result := t.Publish(ctx, &pubsub.Message{Data: []byte(msg)})
	// Block until the result is returned.
	id, err := result.Get(ctx)
	if err != nil {
		return fmt.Errorf("pubsub: result.Get: %w", err)
	}
	fmt.Printf("Published a message; msg ID: %s\n", id)
	return nil
}

func sendMessages(ctx context.Context, client *pubsub.Client, topicID string) {
	for i := 0; i < 20; i++ {
		if err := publish(ctx, client, topicID, fmt.Sprintf("Message %d", i)); err != nil {
			log.Fatalf("Failed to publish: %v", err)
		}
	}
}

func ensureTopicAndSubscription(ctx context.Context, client *pubsub.Client, topicID, subID string) {
	// Ensure topic exists.
	topic := client.Topic(topicID)
	exists, err := topic.Exists(ctx)
	if err != nil {
		log.Fatalf("Failed to check if topic exists: %v", err)
	}
	if !exists {
		if _, err = client.CreateTopic(ctx, topicID); err != nil {
			log.Fatalf("Failed to create the topic: %v", err)
		}
	}

	// Ensure subscription exists.
	sub := client.Subscription(subID)
	exists, err = sub.Exists(ctx)
	if err != nil {
		log.Fatalf("Failed to check if the subscription exists: %v", err)
	}
	if !exists {
		if _, err = client.CreateSubscription(ctx, subID, pubsub.SubscriptionConfig{Topic: topic}); err != nil {
			log.Fatalf("Failed to create the subscription: %v", err)
		}
	}
}

func TestMain(m *testing.M) {
	os.Setenv("PUBSUB_EMULATOR_HOST", "localhost:8085")

	ctx := context.Background()

	// Sets your Google Cloud Platform project ID.
	projectID := "just-ratio-366415"

	// Creates a client.
	var err error
	pubsubClient, err = pubsub.NewClient(ctx, projectID)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer pubsubClient.Close()

	ensureTopicAndSubscription(ctx, pubsubClient, TopicID, subscriptionID)

	code := m.Run()
	os.Exit(code)
}

// Additional test functions go here.

func TestPubSubSource_Read(t *testing.T) {

	messageCh := make(chan sourcer.Message, 20)
	//doneCh := make(chan struct{})

	sub := pubsubClient.Subscription(TopicID)
	pubsubsource := NewPubSubSource(pubsubClient, sub)

	sendMEssagech := make(chan struct{})
	go func() {
		sendMessages(context.Background(), pubsubClient, TopicID)
		close(sendMEssagech)
	}()
	<-sendMEssagech

	pubsubsource.Read(context.TODO(), mocks.ReadRequest{
		CountValue: 5,
		Timeout:    10 * time.Second,
	}, messageCh)

	//<-doneCh

	assert.Equal(t, 5, len(messageCh))
	// Try reading 4 more messages
	// Since the previous batch didn't get acked, the data source shouldn't allow us to read more messages
	// We should get 0 messages, meaning the channel only holds the previous 5 messages
	pubsubsource.Read(context.TODO(), mocks.ReadRequest{
		CountValue: 4,
		Timeout:    5 * time.Second,
	}, messageCh)

	assert.Equal(t, 5, len(messageCh))

	// Ack the first batch
	msg1 := <-messageCh
	msg2 := <-messageCh
	msg3 := <-messageCh
	msg4 := <-messageCh
	msg5 := <-messageCh

	pubsubsource.Ack(context.TODO(), mocks.TestAckRequest{
		OffsetsValue: []sourcer.Offset{msg1.Offset(), msg2.Offset(), msg3.Offset(), msg4.Offset(), msg5.Offset()},
	})

	sub2 := pubsubClient.Subscription(TopicID)

	sendMEssagech2 := make(chan struct{})
	go func() {
		sendMessages(context.Background(), pubsubClient, TopicID)
		close(sendMEssagech2)
	}()
	<-sendMEssagech2

	pubsubsource2 := NewPubSubSource(pubsubClient, sub2)

	pubsubsource2.Read(context.TODO(), mocks.ReadRequest{
		CountValue: 4,
		Timeout:    30 * time.Second,
	}, messageCh)

	assert.Equal(t, 4, len(messageCh))

	/*






		t.Log("Calling ReAD AGAIN---------")

		//time.Sleep(10 * time.Second)
		sendMEssage(topic)



	*/

	//err := sub.Delete(context.Background())
	//assert.Nil(t, err)
	log.Println("Subscription Deleted--------")
}
