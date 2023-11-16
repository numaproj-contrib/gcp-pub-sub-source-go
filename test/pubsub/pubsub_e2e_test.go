//go:build test

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

package pubsub

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"cloud.google.com/go/pubsub"
	"github.com/numaproj-contrib/numaflow-utils-go/testing/fixtures"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

const (
	SUBSCRIPTION_ID = "subscription-09098ui1"
	PROJECT_ID      = "pubsub-test"
	TOPIC_ID        = "pubsub-test-topic"
	PUB_SUB_PORT    = 8681
)

type GCPPubSubSourceSuite struct {
	fixtures.E2ESuite
}

func publish(ctx context.Context, client *pubsub.Client, topicID, msg string) error {
	t := client.Topic(topicID)
	result := t.Publish(ctx, &pubsub.Message{Data: []byte(msg)})
	_, err := result.Get(ctx)
	if err != nil {
		return fmt.Errorf("pubsub: result.Get: %w", err)
	}
	return nil
}

func sendMessages(ctx context.Context, client *pubsub.Client, topicID string, count int, message string) {
	for i := 0; i < count; i++ {
		if err := publish(ctx, client, topicID, message); err != nil {
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

func createPubSubClient() *pubsub.Client {
	pubsubClient, err := pubsub.NewClient(context.Background(), PROJECT_ID)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	return pubsubClient
}
func (suite *GCPPubSubSourceSuite) SetupTest() {

	suite.T().Log("e2e Api resources are ready")

	suite.StartPortForward("e2e-api-pod", 8378)

	// Create Redis Resource
	redisDeleteCmd := fmt.Sprintf("kubectl delete -k ../../config/apps/redis -n %s --ignore-not-found=true", fixtures.Namespace)
	suite.Given().When().Exec("sh", []string{"-c", redisDeleteCmd}, fixtures.OutputRegexp(""))
	redisCreateCmd := fmt.Sprintf("kubectl apply -k ../../config/apps/redis -n %s", fixtures.Namespace)
	suite.Given().When().Exec("sh", []string{"-c", redisCreateCmd}, fixtures.OutputRegexp("service/redis created"))

	suite.T().Log("Redis resources are ready")

	// Create gcloud-pubsub resources.
	gcloudPubSubDeleteCmd := fmt.Sprintf("kubectl delete -k ../../config/apps/gcloud-pubsub -n %s --ignore-not-found=true", fixtures.Namespace)
	suite.Given().When().Exec("sh", []string{"-c", gcloudPubSubDeleteCmd}, fixtures.OutputRegexp(""))
	gcloudPubSubCreateCmd := fmt.Sprintf("kubectl apply -k ../../config/apps/gcloud-pubsub -n %s", fixtures.Namespace)
	suite.Given().When().Exec("sh", []string{"-c", gcloudPubSubCreateCmd}, fixtures.OutputRegexp("service/gcloud-pubsub created"))
	gcloudPubSubLabelSelector := fmt.Sprintf("app=%s", "gcloud-pubsub")
	suite.Given().When().WaitForStatefulSetReady(gcloudPubSubLabelSelector)
	suite.T().Log("gcloud-pubsub resources are ready")
	//delay to make system ready in CI
	time.Sleep(time.Minute)

	suite.T().Log("port forwarding gcloud-pubsub service")
	suite.StartPortForward("gcloud-pubsub-0", PUB_SUB_PORT)
}

func (suite *GCPPubSubSourceSuite) TestPubSubSource() {
	err := os.Setenv("PUBSUB_EMULATOR_HOST", fmt.Sprintf("localhost:%d", PUB_SUB_PORT))
	assert.Nil(suite.T(), err)
	var testMessage = "pubsub-go"
	pubsubclient := createPubSubClient()
	assert.NotNil(suite.T(), pubsubclient)
	err = ensureTopicAndSubscription(context.Background(), pubsubclient, TOPIC_ID, SUBSCRIPTION_ID)
	assert.Nil(suite.T(), err)

	workflow := suite.Given().Pipeline("@testdata/pubsub_source.yaml").When().CreatePipelineAndWait()
	defer workflow.DeletePipelineAndWait()
	workflow.Expect().VertexPodsRunning()

	sendMessages(context.Background(), pubsubclient, TOPIC_ID, 40, testMessage)

	workflow.Expect().SinkContains("redis-sink", testMessage, fixtures.WithTimeout(2*time.Minute))
}

func TestGCPPubSubSourceSuite(t *testing.T) {
	suite.Run(t, new(GCPPubSubSourceSuite))
}
