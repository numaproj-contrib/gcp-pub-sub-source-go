package pubsubsource

/*
import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"cloud.google.com/go/pubsub"
	"github.com/numaproj/numaflow-go/pkg/sourcer"
	"github.com/stretchr/testify/assert"

	"github.com/numaproj-contrib/gcp-pub-sub-source-go/pkg/mocks"
)

const TopicID = "pubsub-test"

var (
	pubsubClient   *pubsub.Client
	subscriptionID = "subscription-09098ui1"
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
	for i := 0; i < 50; i++ {
		if err := publish(ctx, client, topicID, fmt.Sprintf("Message %d", i)); err != nil {
			log.Fatalf("Failed to publish: %v", err)
		}
	}
}

func ensureTopicAndSubscription(ctx context.Context, client *pubsub.Client, topicID, subID string) error {
	// Ensure topic exists.
	topic := client.Topic(topicID)
	exists, err := topic.Exists(ctx)
	if err != nil {
		return err

	}
	if !exists {

		if _, err = client.CreateTopic(ctx, topicID); err != nil {
			return err
		}

	}

	// Ensure subscription exists.
	sub := client.Subscription(subID)
	exists, err = sub.Exists(ctx)
	if err != nil {
		return err

	}
	if !exists {
		log.Println("subscription doesnt exists")
		if _, err = client.CreateSubscription(ctx, subID, pubsub.SubscriptionConfig{Topic: topic}); err != nil {
			return err

		}
	}
	return nil
}

func TestMain(m *testing.M) {
	var err error

	err = os.Setenv("PUBSUB_EMULATOR_HOST", "localhost:8085")
	if err != nil {
		log.Fatalf("error -%s", err)
	}

	ctx := context.Background()

	// Sets your Google Cloud Platform project ID.
	projectID := "just-ratio-366415"

	// Creates a client.
	pubsubClient, err = pubsub.NewClient(ctx, projectID)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer pubsubClient.Close()

	code := m.Run()
	os.Exit(code)
}

// Additional test functions go here.

func TestPubSubSource_Read(t *testing.T) {

	err := ensureTopicAndSubscription(context.TODO(), pubsubClient, TopicID, subscriptionID)
	assert.Nil(t, err)
	messageCh := make(chan sourcer.Message, 20)
	//doneCh := make(chan struct{})
	sub := pubsubClient.Subscription(subscriptionID)
	pubsubsource := NewPubSubSource(pubsubClient, sub)
	ctx := context.Background()
	go func() {
		<-time.After(5 * time.Second)
		sendMessages(context.Background(), pubsubClient, TopicID)

	}()

	pubsubsource.Read(ctx, mocks.ReadRequest{
		CountValue: 5,
		Timeout:    20 * time.Second,
	}, messageCh)

	assert.Equal(t, 5, len(messageCh))

}
*/
