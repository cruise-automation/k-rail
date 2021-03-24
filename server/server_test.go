// Copyright 2021 Cruise LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//    https://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package server

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testPrometheusMetrics(t *testing.T, h http.Handler, expMetrics []string) {
	require := require.New(t)
	assert := assert.New(t)

	// Setup server.
	server := httptest.NewServer(h)
	t.Cleanup(func() { server.Close() })

	// Get metrics.
	r, err := http.NewRequest(http.MethodGet, server.URL+"/metrics", nil)
	require.NoError(err)
	resp, err := http.DefaultClient.Do(r)
	require.NoError(err)

	// Check.
	b, err := ioutil.ReadAll(resp.Body)
	require.NoError(err)
	metrics := string(b)

	assert.Contains(metrics, "go_")
	assert.Contains(metrics, "go_gc_")
	assert.Contains(metrics, "http_")
	assert.Contains(metrics, "promhttp_metric_handler")
}
