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
	"fmt"
	"log"
	"os"
	"time"

	"cloud.google.com/go/pubsub"
	"github.com/numaproj/numaflow-go/pkg/sourcer"

	"github.com/numaproj-contrib/gcp-pub-sub-source-go/pkg/pubsubsource"
)

// getSubscription checks if the specified topic and subscription exist.
// It returns an error if either the topic or the subscription doesn't exist
func getSubscription(ctx context.Context, client *pubsub.Client, topicID, subID string) (*pubsub.Subscription, error) {
	topic := client.Topic(topicID)
	exists, err := topic.Exists(ctx)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, fmt.Errorf("topic does not exist: %s", topicID)
	}
	sub := client.Subscription(subID)
	exists, err = sub.Exists(ctx)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, fmt.Errorf("subscription does not exist: %s", subID)
	}
	return sub, nil
}

func main() {
	subscriptionId := os.Getenv("SUBSCRIPTION_ID")
	topicId := os.Getenv("TOPIC_ID")
	projectId := os.Getenv("PROJECT_ID")
	maxExtensionPeriod, err := time.ParseDuration(os.Getenv("MAX_EXTENSION_PERIOD"))
	if err != nil {
		log.Fatalf("error creating source, max extension period is invalid %s", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	client, err := pubsub.NewClient(ctx, projectId)
	if err != nil {
		log.Fatalf("error in creating pubsub client: %s", err)
	}
	defer client.Close()
	sub, err := getSubscription(context.Background(), client, topicId, subscriptionId)
	if err != nil {
		log.Fatalf("error in getting subscription : %s", err)
	}
	googlePubSubSource := pubsubsource.NewPubSubSource(client, sub, maxExtensionPeriod)
	googlePubSubSource.StartReceiving(ctx)
	err = sourcer.NewServer(googlePubSubSource).Start(context.Background())
	if err != nil {
		log.Panic("failed to start source server : ", err)
	}
}
