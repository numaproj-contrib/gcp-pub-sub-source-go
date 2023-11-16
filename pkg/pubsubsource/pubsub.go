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

const MAX_OUT_STANDING_MESSAGES = 1000 //  maximum number of unacknowledged messages that can be held at a given time in buffer

// PubSubSource represents a source of messages in a publish-subscribe system.
type PubSubSource struct {
	client       *pubsub.Client
	subscription *pubsub.Subscription
	receiveCh    chan *pubsub.Message       //buffered channel for buffering messages received from the subscription.
	messages     map[string]*pubsub.Message //messages field  serves as a cache or storage for unacknowledged pubsub.Message objects that have been received but not yet processed.
	lock         *sync.Mutex
}

func NewPubSubSource(client *pubsub.Client, subscription *pubsub.Subscription, maxExtensionPeriod time.Duration) *PubSubSource {
	subscription.ReceiveSettings.MaxExtension = maxExtensionPeriod
	receiveCh := make(chan *pubsub.Message, MAX_OUT_STANDING_MESSAGES)
	pubSubSource := &PubSubSource{client: client, subscription: subscription, receiveCh: receiveCh, messages: make(map[string]*pubsub.Message), lock: new(sync.Mutex)}
	return pubSubSource
}

func (p *PubSubSource) Read(_ context.Context, readRequest sourcesdk.ReadRequest, messageCh chan<- sourcesdk.Message) {

	// this condition will happen only if there are unacked messages.
	// Read can happen only after all the messages are acked.
	if len(p.messages) > 0 {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), readRequest.TimeOut())
	defer cancel()
	for i := 0; i < int(readRequest.Count()) && ctx.Err() == nil; i++ {
		select {
		case msg := <-p.receiveCh:
			p.lock.Lock()
			messageCh <- sourcesdk.NewMessage(
				msg.Data,
				sourcesdk.NewOffset([]byte(msg.ID), "0"),
				msg.PublishTime,
			)
			p.messages[msg.ID] = msg
			p.lock.Unlock()
		case <-ctx.Done():
			return
		}
	}
}

func (p *PubSubSource) Pending(_ context.Context) int64 {
	return int64(len(p.receiveCh))
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

func (p *PubSubSource) StartReceiving(ctx context.Context) {
	go func() {
		if err := p.subscription.Receive(ctx, func(ctx context.Context, msg *pubsub.Message) {
			select {
			case p.receiveCh <- msg:
			case <-ctx.Done():
				return
			}
		}); err != nil {
			log.Fatalf("error receiving messages: %v", err)
		}
	}()
}
