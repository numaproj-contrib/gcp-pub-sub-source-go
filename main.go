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
	"fmt"
	"github.com/numaproj/numaflow-go/pkg/sourcer"
	"log"
	"os"

	"github.com/numaproj-contrib/gcp-pub-sub-source-go/pkg/pubsubsource"
)

// getSubscription checks if the specified topic and subscription exist.
// It returns an error if either the topic or the subscription doesn't exist or the subscription object
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
	ctx := context.Background()
	client, err := pubsub.NewClient(ctx, projectId)
	if err != nil {
		log.Fatalf("error in creating pubsub client: %s", err)
	}
	defer client.Close()
	sub, err := getSubscription(context.Background(), client, topicId, subscriptionId)
	if err != nil {
		log.Fatalf("error in getting subscription : %s", err)
	}
	var monitoringClient *pubsubsource.MonitoringClient
	// monitoring client is available only in production environment
	if len(os.Getenv("PUBSUB_EMULATOR_HOST")) == 0 {
		monitoringClient = pubsubsource.NewMonitoringClient(projectId, subscriptionId)
	}
	googlePubSubSource := pubsubsource.NewPubSubSource(client, sub, monitoringClient)
	err = sourcer.NewServer(googlePubSubSource).Start(context.Background())
	if err != nil {
		log.Panic("failed to start source server : ", err)
	}

}
