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
	"context"
	"log"
	"sync"
	"time"

	"cloud.google.com/go/pubsub"
	sourcesdk "github.com/numaproj/numaflow-go/pkg/sourcer"
)

// MAX_EXTENSION_PERIOD The MaxExtensionPeriod in the context of Google Cloud Pub/Sub subscription
// settings is a parameter that specifies the maximum period for which
// the deadline for message acknowledgment may be extended.
const MAX_EXTENSION_PERIOD = 4 * time.Minute

type PubSubSource struct {
	client       *pubsub.Client
	subscription *pubsub.Subscription
	messages     map[string]*pubsub.Message
	lock         *sync.Mutex
}

func NewPubSubSource(client *pubsub.Client, subscription *pubsub.Subscription) *PubSubSource {
	return &PubSubSource{client: client, subscription: subscription, messages: make(map[string]*pubsub.Message), lock: new(sync.Mutex)}
}

func (p *PubSubSource) Read(_ context.Context, readRequest sourcesdk.ReadRequest, messageCh chan<- sourcesdk.Message) {
	if len(p.messages) > 0 {
		return
	}
	p.subscription.ReceiveSettings.MaxOutstandingMessages = int(readRequest.Count())
	p.subscription.ReceiveSettings.MaxExtensionPeriod = MAX_EXTENSION_PERIOD

	ctx, cancel := context.WithTimeout(context.Background(), readRequest.TimeOut())
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
func (p *PubSubSource) Ack(_ context.Context, request sourcesdk.AckRequest) {
	for _, offset := range request.Offsets() {
		p.messages[string(offset.Value())].Ack()
		delete(p.messages, string(offset.Value()))
	}
}
