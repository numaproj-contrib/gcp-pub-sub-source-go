package pubsubsource

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
	exitch := make(chan struct{})
	sendMEssagech := make(chan int)
	go func() {
		for i := 0; i < 20; i++ {
			if err := publish(context.Background(), pubsubClient, TopicID, fmt.Sprintf("Message %d", i)); err != nil {
				log.Fatalf("Failed to publish: %v", err)
			}
			if i%5 == 0 {
				sendMEssagech <- i

			}
		}
		close(exitch)
	}()

	for {
		select {
		case <-sendMEssagech:
			sub := pubsubClient.Subscription(subscriptionID)
			pubsubsource := NewPubSubSource(pubsubClient, sub)
			log.Println("Received called")
			pubsubsource.Read(context.TODO(), mocks.ReadRequest{
				CountValue: 5,
				Timeout:    10 * time.Second,
			}, messageCh)
			assert.Equal(t, 5, len(messageCh))
		case <-exitch:
			break

		}
	}

	//<-doneCh

	/*
		// Try reading 4 more messages
		// Since the previous batch didn't get acked, the data source shouldn't allow us to read more messages
		// We should get 0 messages, meaning the channel only holds the previous 5 messages
		pubsubsource.Read(context.TODO(), mocks.ReadRequest{
			CountValue: 4,
			Timeout:    5 * time.Second,
		}, messageCh)

		assert.Equal(t, 5, len(messageCh))
	*/
	// Ack the first batch

	/*


		msg1 := <-messageCh
			msg2 := <-messageCh
			msg3 := <-messageCh
			msg4 := <-messageCh
			msg5 := <-messageCh

			pubsubsource.Ack(context.TODO(), mocks.TestAckRequest{
				OffsetsValue: []sourcer.Offset{msg1.Offset(), msg2.Offset(), msg3.Offset(), msg4.Offset(), msg5.Offset()},
			})

			sub2 := pubsubClient.Subscription(subscriptionID)

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



			t.Log("Calling ReAD AGAIN---------")

			//time.Sleep(10 * time.Second)
			sendMEssage(topic)



	*/

	//err := sub.Delete(context.Background())
	//assert.Nil(t, err)
	//log.Println("Subscription Deleted--------")

}
