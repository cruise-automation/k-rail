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
	"github.com/prometheus/client_golang/prometheus"
	metrics "github.com/slok/go-http-metrics/metrics/prometheus"
	"github.com/slok/go-http-metrics/middleware"
)

var prometheusMiddleware = middleware.New(middleware.Config{
	Recorder: metrics.NewRecorder(metrics.Config{}),
})

var totalRegisteredPolicies = prometheus.NewCounter(
	prometheus.CounterOpts{
		Namespace: "krail",
		Name:      "total_registered_policies",
		Help:      "Total Policies Registered",
	},
)

var totalLoadedPlugins = prometheus.NewCounter(
	prometheus.CounterOpts{
		Namespace: "krail",
		Name:      "total_loaded_plugins",
		Help:      "Total Plugins Loaded",
	},
)

var policyViolations = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: "krail",
		Name:      "policy_violations",
		Help:      "Count of Violations",
	},
	[]string{"kind", "resource", "namespace", "policy", "user", "enforced", "exempt", "report_only", "global_report_only"},
)
