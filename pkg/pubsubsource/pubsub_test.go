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

const Topic = "pubsub-test"

var id = "subscription-0908"

var pubsubclient *pubsub.Client

func publish(client *pubsub.Client, topicID, msg string) error {

	t := client.Topic(topicID)
	result := t.Publish(context.Background(), &pubsub.Message{
		Data: []byte(msg),
	})
	// Block until the result is returned and a server-generated
	// ID is returned for the published message.
	id, err := result.Get(context.TODO())
	if err != nil {
		log.Fatalf("pubsub: result.Get: %s", err)
	}
	fmt.Printf("Published a message; msg ID: %s\n", id)
	return nil
}

func sendMEssage() {

	for i := 0; i < 20; i++ {
		publish(pubsubclient, Topic, "Some Random MEssage")
	}

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

func createSubscription(client *pubsub.Client, id string) {
	//check if the subscription already exists
	log.Println("Creating subscription----------------*****")
	sub := client.Subscription(id)
	exists, err := sub.Exists(context.Background())
	if err != nil {
		log.Fatalf("failed to check if  the susbcription exists: %v", err)

	}
	log.Println(exists)
	if !exists {
		sub_, err := client.CreateSubscription(context.Background(), id, pubsub.SubscriptionConfig{
			Topic: createTopicIfNotExists(client, "my-topic-5"),
		})
		if err != nil {
			log.Fatalf("Failed to create the susbcription: %v", err)

		}
		log.Println("New Subscription  Created ---------", sub_)
		// Call send message to send the message to new subscription

	}
	log.Println("New Subscription  Created ---------")

}

func TestMain(m *testing.M) {
	err := os.Setenv("PUBSUB_EMULATOR_HOST", "localhost:8085")
	ctx := context.Background()
	log.Println("Main executed")
	// Sets your Google Cloud Platform project ID.
	projectID := "just-ratio-366415"

	// Creates a client.
	client, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()
	pubsubclient = client
	createTopicIfNotExists(pubsubclient, Topic)
	createSubscription(pubsubclient, id)

	code := m.Run()
	os.Exit(code)

}
func TestPubSubSource_Read(t *testing.T) {

	messageCh := make(chan sourcer.Message, 20)
	//doneCh := make(chan struct{})

	sub := pubsubclient.Subscription(id)
	pubsubsource := NewPubSubSource(pubsubclient, sub)

	sendMEssagech := make(chan struct{})
	go func() {
		sendMEssage()
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

	sub2 := pubsubclient.Subscription(id)

	sendMEssagech2 := make(chan struct{})
	go func() {
		sendMEssage()
		close(sendMEssagech2)
	}()
	<-sendMEssagech2

	pubsubsource2 := NewPubSubSource(pubsubclient, sub2)

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
