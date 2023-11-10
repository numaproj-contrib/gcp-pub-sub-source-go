package pubsubsource

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
import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"testing"
	"time"

	"cloud.google.com/go/pubsub"
	"github.com/numaproj/numaflow-go/pkg/sourcer"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/stretchr/testify/assert"

	"github.com/numaproj-contrib/gcp-pub-sub-source-go/pkg/mocks"
)

const TopicID = "pubsub-test"
const Project = "pubsub-local-test"

var (
	pubsubClient   *pubsub.Client
	subscriptionID = "subscription-09098ui1"
	resource       *dockertest.Resource
	pool           *dockertest.Pool
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

	// connect to docker
	p, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("could not connect to docker ;is it running ? %s", err)
	}
	pool = p

	// Check if pubsub container is already running
	containers, err := pool.Client.ListContainers(docker.ListContainersOptions{})
	if err != nil {
		log.Fatalf("could not list containers %s", err)
	}
	pubSubRunning := false
	for _, container := range containers {
		for _, name := range container.Names {
			if strings.Contains(name, "google-cloud-cli") {
				pubSubRunning = true
				break
			}
		}
		if pubSubRunning {
			break
		}
	}

	if !pubSubRunning {
		// Start goaws container if not already running
		opts := dockertest.RunOptions{
			Repository:   "gcr.io/google.com/cloudsdktool/google-cloud-cli",
			Env:          []string{fmt.Sprintf("project=%s", Project)},
			Tag:          "latest",
			ExposedPorts: []string{"8085"},
			PortBindings: map[docker.Port][]docker.PortBinding{
				"8085": {
					{HostIP: "127.0.0.1", HostPort: "8085"},
				},
			},
		}
		resource, err = pool.RunWithOptions(&opts)
		if err != nil {
			_ = pool.Purge(resource)
			log.Fatalf("could not start resource %s", err)
		}
	}

	err = os.Setenv("PUBSUB_EMULATOR_HOST", "localhost:8085")
	if err != nil {
		log.Fatalf("error -%s", err)
	}
	ctx := context.Background()

	if err := pool.Retry(func() error {
		// Creates a client.
		pubsubClient, err = pubsub.NewClient(ctx, Project)
		if err != nil {
			log.Fatalf("Failed to create client: %v", err)
		}
		return nil
	}); err != nil {
		if resource != nil {
			_ = pool.Purge(resource)
		}
		log.Fatalf("could not connect to gcloud pubsub sqs %s", err)
	}
	code := m.Run()
	if resource != nil {
		if err := pool.Purge(resource); err != nil {
			log.Fatalf("Couln't purge resource %s", err)
		}
	}
	defer pubsubClient.Close()
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
