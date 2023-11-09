/*
Copyright 2022 The Numaproj Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"context"
	"log"
	"time"

	"cloud.google.com/go/pubsub"
	"github.com/numaproj/numaflow-go/pkg/sourcer"

	"github.com/numaproj-contrib/gcp-pub-sub-source-go/pkg/mocks"
	"github.com/numaproj-contrib/gcp-pub-sub-source-go/pkg/pubsubsource"
)

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

func main() {
	ctx := context.Background()
	client, err := pubsub.NewClient(ctx, "local-project")
	defer client.Close()

	if err != nil {
		log.Fatalf("pubsub.NewClient: %s", err)
	}
	//id := fmt.Sprintf("my-sub-%d", rand.Int())
	/*
		sub_, err := client.CreateSubscription(ctx, id, pubsub.SubscriptionConfig{
			Topic: createTopicIfNotExists(client, "my-topic-5"),
		})

		if err != nil {
			log.Fatalf("error creating subscription: %s", err)
		}
		fmt.Printf("Created subscription: %v\n", sub_)

	*/
	sub := client.Subscription("my-sub-4879843889150859861")

	readRequest := &mocks.ReadRequest{
		CountValue: 50,
		Timeout:    10 * time.Second,
	}
	messageChan := make(chan sourcer.Message, 20)
	googlePubSubSource := pubsubsource.NewPubSubSource(client, sub)
	googlePubSubSource.Read(context.Background(), readRequest, messageChan)

	googlePubSubSource.Ack(context.Background(), mocks.TestAckRequest{OffsetsValue: make([]sourcer.Offset, 0)})

	log.Println("Acknowledge Done----------&&&&&&&&&&&&&&&&&&&&&&&&&")
	googlePubSubSource.Read(context.Background(), readRequest, messageChan)

	/*
		err = sourcer.NewServer(googlePubSubSource).Start(context.Background())
		if err != nil {
			log.Panic("Failed to start source server : ", err)
		}
	*/

}
