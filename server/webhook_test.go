// Copyright 2019 Cruise LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//    https://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.package ingress

package server

import (
	"encoding/json"
	"testing"

	log "github.com/sirupsen/logrus"

	"github.com/cruise-automation/k-rail/policies"
	admissionv1beta1 "k8s.io/api/admission/v1beta1"
	authenticationv1 "k8s.io/api/authentication/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type test struct {
	name              string
	resourceName      string
	resourceNamespace string
	podSpec           corev1.PodSpec
	allow             bool
}

func test_setup() (Server, []test) {
	rawExemptions := []policies.RawExemption{
		{
			ResourceName:   "test-resource",
			Namespace:      "test-namespace",
			ExemptPolicies: []string{"*"},
		},
	}
	var compiledExemptions []policies.CompiledExemption
	for _, e := range rawExemptions {
		compiledExemptions = append(compiledExemptions, e.Compile())
	}

	testSrv := Server{
		Config: Config{
			Policies: []PolicySettings{
				{
					Name:       "pod_no_host_network",
					Enabled:    true,
					ReportOnly: false,
				},
			},
		},
		Exemptions: compiledExemptions,
	}

	log.SetLevel(log.ErrorLevel)

	testSrv.registerPolicies()

	tests := []test{
		{
			name:  "deny by policy",
			allow: false,
			podSpec: corev1.PodSpec{
				HostNetwork: true,
			},
		},
		{
			name:  "allow by policy",
			allow: true,
			podSpec: corev1.PodSpec{
				HostNetwork: false,
			},
		},
		{
			name:              "allow by name exemption",
			allow:             true,
			resourceNamespace: "test-namespace",
			resourceName:      "test-resource-lol",
			podSpec: corev1.PodSpec{
				HostNetwork: true,
			},
		},
	}

	return testSrv, tests
}

func runBenchmark(b *testing.B, testSrv Server, tt test) {
	pod := corev1.Pod{Spec: tt.podSpec}
	pod.Name = tt.resourceName
	raw, _ := json.Marshal(pod)
	ar := admissionv1beta1.AdmissionReview{
		Request: &admissionv1beta1.AdmissionRequest{
			Namespace: tt.resourceNamespace,
			Object:    runtime.RawExtension{Raw: raw},
			Resource:  metav1.GroupVersionResource{Group: "", Version: "v1", Resource: "pods"},
			UserInfo:  authenticationv1.UserInfo{Groups: []string{"group1"}},
		},
	}

	if got := testSrv.validateResources(ar); got.Response.Allowed != tt.allow {
		b.Errorf("Server.validateResources() = %v, want %v", got.Response.Allowed, tt.allow)
	}
}

func BenchmarkServer_DenyPolicy(b *testing.B) {
	testSrv, tests := test_setup()
	tt := tests[0]
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			runBenchmark(b, testSrv, tt)
		}
	})
}

func BenchmarkServer_AllowPolicy(b *testing.B) {
	testSrv, tests := test_setup()
	tt := tests[1]
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			runBenchmark(b, testSrv, tt)
		}
	})
}

func BenchmarkServer_AllowExemption(b *testing.B) {
	testSrv, tests := test_setup()
	tt := tests[2]
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			runBenchmark(b, testSrv, tt)
		}
	})
}

func TestServer_validateResources(t *testing.T) {
	testSrv, tests := test_setup()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pod := corev1.Pod{Spec: tt.podSpec}
			pod.Name = tt.resourceName
			raw, _ := json.Marshal(pod)
			ar := admissionv1beta1.AdmissionReview{
				Request: &admissionv1beta1.AdmissionRequest{
					Namespace: tt.resourceNamespace,
					Object:    runtime.RawExtension{Raw: raw},
					Resource:  metav1.GroupVersionResource{Group: "", Version: "v1", Resource: "pods"},
					UserInfo:  authenticationv1.UserInfo{Groups: []string{"group1"}},
				},
			}

			if got := testSrv.validateResources(ar); got.Response.Allowed != tt.allow {
				t.Errorf("Server.validateResources() = %v, want %v", got.Response.Allowed, tt.allow)
			}
		})
	}
}
