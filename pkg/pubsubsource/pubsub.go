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

package pubsubsource

import (
	"cloud.google.com/go/pubsub"
	"context"
	sourcesdk "github.com/numaproj/numaflow-go/pkg/sourcer"
	"log"
	"os"
	"sync"
	"time"
)

// PubSubSource represents a source of messages in a publish-subscribe system.
type PubSubSource struct {
	client       *pubsub.Client
	subscription *pubsub.Subscription
	messages     map[string]*pubsub.Message //messages field  serves as a cache or storage for unacknowledged pubsub.Message objects that have been received but not yet processed.
	lock         *sync.Mutex
}

func NewPubSubSource(client *pubsub.Client, subscription *pubsub.Subscription) *PubSubSource {
	maxExtensionPeriod, err := time.ParseDuration(os.Getenv("MAX_EXTENSION_PERIOD"))
	if err != nil {
		log.Fatalf("error creating source, max extension period is invalid %s", err)
	}
	subscription.ReceiveSettings.MaxExtensionPeriod = maxExtensionPeriod
	return &PubSubSource{client: client, subscription: subscription, messages: make(map[string]*pubsub.Message), lock: new(sync.Mutex)}
}

func (p *PubSubSource) Read(ctx context.Context, readRequest sourcesdk.ReadRequest, messageCh chan<- sourcesdk.Message) {
	// this condition will happen only if there are unacked messages.
	// Read can happen only after all the messages are acked.
	if len(p.messages) > 0 {
		return
	}

	p.subscription.ReceiveSettings.MaxOutstandingMessages = int(readRequest.Count())

	ctx, cancel := context.WithTimeout(ctx, readRequest.TimeOut())
	defer cancel()

	errCh := make(chan error, 1)
	go func() {
		err := p.subscription.Receive(ctx, func(msgCtx context.Context, msg *pubsub.Message) {
			p.lock.Lock()
			defer p.lock.Unlock()
			messageCh <- sourcesdk.NewMessage(
				msg.Data,
				sourcesdk.NewOffset([]byte(msg.ID), "0"),
				msg.PublishTime,
			)
			p.messages[msg.ID] = msg
		})
		if err != nil {
			errCh <- err
		}
	}()
	select {
	case <-ctx.Done():
		return
	case err := <-errCh:
		log.Printf("error occurred while receiving messsages %s", err)
		return
	}
}

func (p *PubSubSource) Pending(_ context.Context) int64 {
	// gcp pubsub doesn't provide any api to get the number of available messages
	return -1
}
func (p *PubSubSource) Ack(ctx context.Context, request sourcesdk.AckRequest) {
	for _, offset := range request.Offsets() {
		select {
		case <-ctx.Done():
			return
		default:
			p.lock.Lock()
			p.messages[string(offset.Value())].Ack()
			delete(p.messages, string(offset.Value()))
			p.lock.Unlock()
		}
	}
}

// SetSubscription  allows the PubSubSource to switch to a different subscriber at runtime without needing to create a new PubSubSource instance.
func (p *PubSubSource) SetSubscription(sub *pubsub.Subscription) {
	p.subscription = sub
}
