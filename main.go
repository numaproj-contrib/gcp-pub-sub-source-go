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
	"cloud.google.com/go/pubsub"
	"context"
	"github.com/numaproj-contrib/gcp-pub-sub-source-go/pkg/pubsubsource"
	"github.com/numaproj/numaflow-go/pkg/sourcer"
	"log"
	"os"
)

func ensureTopicAndSubscription(ctx context.Context, client *pubsub.Client, topicID, subID string) (error, *pubsub.Subscription) {
	topic := client.Topic(topicID)
	exists, err := topic.Exists(ctx)
	if err != nil {
		return err, nil
	}
	if !exists {
		if _, err = client.CreateTopic(ctx, topicID); err != nil {
			return err, nil
		}
	}

	topic = client.Topic(topicID)
	sub := client.Subscription(subID)

	// Check if the subscription exists, create it if it doesn't.
	exists, err = sub.Exists(ctx)
	if err != nil {
		return err, nil
	}
	if !exists {
		if sub, err = client.CreateSubscription(ctx, subID, pubsub.SubscriptionConfig{Topic: topic}); err != nil {
			return err, nil
		}
	}

	return nil, sub
}

func main() {
	ctx := context.Background()
	client, err := pubsub.NewClient(ctx, os.Getenv("PROJECT_ID"))
	defer client.Close()

	if err != nil {
		log.Fatalf("error in creating pubsub client: %s", err)
	}

	err, sub := ensureTopicAndSubscription(context.Background(), client, os.Getenv("TOPIC_ID"), os.Getenv("SUBSCRIPTION_ID"))
	if err != nil {
		log.Fatalf("error in ensuring topic and subscription : %s", err)
	}
	googlePubSubSource := pubsubsource.NewPubSubSource(client, sub)
	err = sourcer.NewServer(googlePubSubSource).Start(context.Background())
	if err != nil {
		log.Panic("Failed to start source server : ", err)
	}

}
