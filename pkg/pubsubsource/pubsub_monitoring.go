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
	monitoring "cloud.google.com/go/monitoring/apiv3/v2"
	"cloud.google.com/go/monitoring/apiv3/v2/monitoringpb"
	"context"
	"errors"
	"fmt"
	"github.com/golang/protobuf/ptypes/timestamp"
	"google.golang.org/api/iterator"
	"log"
	"os"
	"time"
)

type MonitoringClient struct {
	metricClient   *monitoring.MetricClient
	projectID      string
	subscriptionID string
}

func NewMonitoringClient(projectID, subscriptionID string) *MonitoringClient {
	ctx := context.Background()
	fmt.Println(os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"))
	client, err := monitoring.NewMetricClient(ctx)
	log.Println("Metric client", err)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	return &MonitoringClient{metricClient: client, projectID: projectID, subscriptionID: subscriptionID}
}

// GetPendingMessageCount retrieves the current count of undelivered messages in a specified Pub/Sub subscription.
// This method provides a workaround for Google Cloud Pub/Sub's lack of a direct message counting feature.
// It utilizes the Cloud Monitoring API to fetch the most recent metric data for the 'num_undelivered_messages' metric,
// which reflects the number of messages that have yet to be delivered to subscribers.
func (m *MonitoringClient) GetPendingMessageCount(ctx context.Context) (int64, error) {
	req := &monitoringpb.ListTimeSeriesRequest{
		Name:   "projects/" + m.projectID,
		Filter: fmt.Sprintf(`metric.type="pubsub.googleapis.com/subscription/num_undelivered_messages" AND resource.labels.subscription_id="%s"`, m.subscriptionID),
		Interval: &monitoringpb.TimeInterval{
			StartTime: &timestamp.Timestamp{Seconds: time.Now().Add(-10 * time.Minute).Unix()},
			EndTime:   &timestamp.Timestamp{Seconds: time.Now().Unix()},
		},
		View: monitoringpb.ListTimeSeriesRequest_FULL,
	}
	it := m.metricClient.ListTimeSeries(ctx, req)
	var messageCount int64
	for {
		resp, err := it.Next()
		if errors.Is(err, iterator.Done) {
			log.Println(err)
			break
		}
		if err != nil {
			return 0, fmt.Errorf("Error retrieving time series data: %v\n", err)
		}
		for _, point := range resp.Points {
			messageCount = point.Value.GetInt64Value() // returning the last value seen

		}
	}
	return messageCount, nil
}
