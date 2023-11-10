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
	ctx, cancelFunc := context.WithTimeout(context.Background(), readRequest.TimeOut())
	defer cancelFunc()
	cctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	p.subscription.ReceiveSettings.MaxOutstandingMessages = int(readRequest.Count())
	p.subscription.ReceiveSettings.MaxExtensionPeriod = 4 * time.Minute
	go func() {
		err := p.subscription.Receive(cctx, func(ctx context.Context, msg *pubsub.Message) {
			p.lock.Lock()
			messageCh <- sourcesdk.NewMessage(
				msg.Data,
				sourcesdk.NewOffset([]byte(msg.ID), "0"),
				msg.PublishTime,
			)
			p.messages[msg.ID] = msg
			p.lock.Unlock()
		})
		if err != nil {
			log.Fatalf("Error--%s", err)
			return
		}

	}()

	for {
		select {
		case <-ctx.Done():
			cancel()
			return

		}

	}
}

func (p *PubSubSource) Pending(_ context.Context) int64 {
	return -1
}
func (p *PubSubSource) Ack(_ context.Context, request sourcesdk.AckRequest) {
	for _, offset := range request.Offsets() {
		p.messages[string(offset.Value())].Ack()
		delete(p.messages, string(offset.Value()))
	}

}
