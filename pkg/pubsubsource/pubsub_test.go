package pubsubsource

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"testing"
	"time"

	"cloud.google.com/go/pubsub"
	"github.com/numaproj/numaflow-go/pkg/sourcer"
	"github.com/stretchr/testify/assert"

	"github.com/numaproj-contrib/gcp-pub-sub-source-go/pkg/mocks"
)

var pubsubclient *pubsub.Client

func publish(client *pubsub.Client, topicID, msg string) error {

	t := client.Topic(topicID)
	result := t.Publish(context.Background(), &pubsub.Message{
		Data: []byte(msg),
	})
	// Block until the result is returned and a server-generated
	// ID is returned for the published message.
	id, err := result.Get(context.Background())
	if err != nil {
		return fmt.Errorf("pubsub: result.Get: %w", err)
	}
	fmt.Printf("Published a message; msg ID: %s\n", id)
	return nil
}

func TestMain(m *testing.M) {

	ctx := context.Background()

	// Sets your Google Cloud Platform project ID.
	projectID := "local-project"

	// Creates a client.
	client, err := pubsub.NewClient(ctx, projectID)
	log.Println(client)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	pubsubclient = client
	defer client.Close()

	// Sets the id for the new topic.
	topicID := "my-topic-5"

	// Creates the new topic.
	//topic := createTopicIfNotExists(pubsubclient, topicID)

	//fmt.Printf("Topic %v created.\n", topic)

	for i := 0; i < 20; i++ {
		err := publish(client, topicID, "hello")
		if err != nil {
			log.Fatalf("Error Publishing message %s", err)
			return
		}

	}

	code := m.Run()
	os.Exit(code)
}

func createTopicIfNotExists(c *pubsub.Client, topic string) *pubsub.Topic {
	ctx := context.Background()
	t := c.Topic(topic)
	ok, err := t.Exists(ctx)
	if err != nil {
		log.Fatal(err)
	}
	if ok {
		return t
	}
	t, err = c.CreateTopic(ctx, topic)
	if err != nil {
		log.Fatalf("Failed to create the topic: %v", err)
	}
	return t
}

func TestPubSubSource_Read(t *testing.T) {
	id := fmt.Sprintf("my-sub-%d", rand.Int())
	t.Log("Executing testst")
	sub_, err := pubsubclient.CreateSubscription(context.Background(), id, pubsub.SubscriptionConfig{
		Topic: createTopicIfNotExists(pubsubclient, "my-topic-5"),
	})
	t.Logf("Created Subscription %s", sub_)

	assert.Nil(t, err)
	sub := pubsubclient.Subscription(id)
	t.Logf("Created Subscription %s", sub)

	messageCh := make(chan sourcer.Message, 20)
	doneCh := make(chan struct{})

	pubsubsource := NewPubSubSource(pubsubclient, sub)
	go func() {
		pubsubsource.Read(context.TODO(), mocks.ReadRequest{
			CountValue: 20,
			Timeout:    50 * time.Second,
		}, messageCh)
		close(doneCh)
	}()
	<-doneCh

	assert.Equal(t, 5, len(messageCh))

	/*
		sub := pubsubclient.Subscription(id)
		messageCh := make(chan sourcer.Message, 20)

		pubsubsource := NewPubSubSource(pubsubclient, sub)
		pubsubsource.Read(context.TODO(), mocks.ReadRequest{
			CountValue: 5,
			Timeout:    5 * time.Second,
		}, messageCh)
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

		pubsubsource.Read(context.TODO(), mocks.ReadRequest{
			CountValue: 4,
			Timeout:    5 * time.Second,
		}, messageCh)

		assert.Equal(t, 50, len(messageCh))

	*/

}
