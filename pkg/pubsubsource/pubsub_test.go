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
const subscriptionID = "subscription-09098"
const MAX_EXTENSION_PERIOD = 240 * time.Second
const PUB_SUB_EMULATOR_HOST = "localhost:8681"
const MAX_OUT_STANDING_MESSAGES = 20

var (
	pubsubClient *pubsub.Client
	resource     *dockertest.Resource
	pool         *dockertest.Pool
)

func publish(ctx context.Context, client *pubsub.Client, topicID, msg string) error {
	t := client.Topic(topicID)
	result := t.Publish(ctx, &pubsub.Message{Data: []byte(msg)})
	_, err := result.Get(ctx)
	if err != nil {
		return fmt.Errorf("pubsub: result.Get: %w", err)
	}
	return nil
}

func sendMessages(ctx context.Context, client *pubsub.Client, topicID string, count int) {
	for i := 0; i < count; i++ {
		if err := publish(ctx, client, topicID, fmt.Sprintf("Message %d", i)); err != nil {
			log.Fatalf("Failed to publish: %v", err)
		}
	}
}

func ensureTopicAndSubscription(ctx context.Context, client *pubsub.Client, topicID, subID string) error {
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

	sub := client.Subscription(subID)
	exists, err = sub.Exists(ctx)
	if err != nil {
		return err

	}
	if !exists {
		if _, err = client.CreateSubscription(ctx, subID, pubsub.SubscriptionConfig{Topic: topic}); err != nil {
			return err

		}
	}
	return nil
}

func TestMain(m *testing.M) {
	var err error
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
		fmt.Println(container)
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
			Repository:   "thekevjames/gcloud-pubsub-emulator",
			Tag:          "latest",
			ExposedPorts: []string{"8681"},
			PortBindings: map[docker.Port][]docker.PortBinding{
				"8681": {
					{HostIP: "127.0.0.1", HostPort: "8681"},
				},
			},
		}
		resource, err = pool.RunWithOptions(&opts)
		if err != nil {
			_ = pool.Purge(resource)
			log.Fatalf("could not start resource %s", err)
		}
	}

	err = os.Setenv("PUBSUB_EMULATOR_HOST", PUB_SUB_EMULATOR_HOST)
	if err != nil {
		log.Fatalf("error -%s", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	if err := pool.Retry(func() error {
		pubsubClient, err = pubsub.NewClient(ctx, Project)
		if err != nil {
			log.Fatalf("Failed to create client: %v", err)
		}
		return nil
	}); err != nil {
		if resource != nil {
			_ = pool.Purge(resource)
		}
		log.Fatalf("could not connect to gcloud pubsub %s", err)
	}
	defer pubsubClient.Close()
	defer cancel()
	code := m.Run()
	if resource != nil {
		if err := pool.Purge(resource); err != nil {
			log.Fatalf("Couln't purge resource %s", err)
		}
	}
	os.Exit(code)
}

func TestPubSubSource_Read(t *testing.T) {
	setupCtx, cancelSetup := context.WithCancel(context.Background())
	err := ensureTopicAndSubscription(setupCtx, pubsubClient, TopicID, subscriptionID)
	assert.Nil(t, err)
	cancelSetup()
	messageCh := make(chan sourcer.Message, 20)
	subscription := pubsubClient.Subscription(subscriptionID)
	pubsubSource := NewPubSubSource(pubsubClient, subscription, MAX_EXTENSION_PERIOD, MAX_OUT_STANDING_MESSAGES)
	ctx, cancel := context.WithCancel(context.Background())
	publishChan := make(chan struct{})
	defer cancel()
	pubsubSource.StartReceiving(ctx)
	go func() {
		sendMessages(ctx, pubsubClient, TopicID, 100)
		close(publishChan)

	}()

	pubsubSource.Read(ctx, mocks.ReadRequest{
		CountValue: 5,
		Timeout:    20 * time.Second,
	}, messageCh)
	assert.Equal(t, 5, len(messageCh))

	// Try reading 4 more messages
	// Since the previous batch didn't get acked, the data source shouldn't allow us to read more messages
	// We should get 0 messages, meaning the channel only holds the previous 5 messages
	pubsubSource.Read(ctx, mocks.ReadRequest{
		CountValue: 4,
		Timeout:    10 * time.Second,
	}, messageCh)
	assert.Equal(t, 5, len(messageCh)) // Expecting no new messages

	// acknowledging the first batch
	msg1 := <-messageCh
	msg2 := <-messageCh
	msg3 := <-messageCh
	msg4 := <-messageCh
	msg5 := <-messageCh
	pubsubSource.Ack(ctx, mocks.TestAckRequest{
		OffsetsValue: []sourcer.Offset{msg1.Offset(), msg2.Offset(), msg3.Offset(), msg4.Offset(), msg5.Offset()},
	})

	pubsubSource.Read(ctx, mocks.ReadRequest{
		CountValue: 6,
		Timeout:    10 * time.Second,
	}, messageCh)
	assert.Equal(t, 6, len(messageCh))

	<-publishChan
}
