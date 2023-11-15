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
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestMonitoringClient_GetPendingMessageCount(t *testing.T) {
	err := os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", "/Users/shubhamdixit/gcp-pub-sub-source-go/pkg/pubsubsource/client.json")
	assert.Nil(t, err)
	monitoringClient := NewMonitoringClient("my-project-1500816644689", "numaflowtest")
	assert.NotNil(t, monitoringClient)
	count, err := monitoringClient.GetPendingMessageCount(context.Background())
	assert.Nil(t, err)
	assert.GreaterOrEqual(t, count, 0)

}
