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
// limitations under the License.

package pod

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/cruise-automation/k-rail/v3/policies"
	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestPolicyNoTiller(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name       string
		podSpec    v1.PodSpec
		violations int
	}{
		{
			name:       "allow",
			violations: 0,
			podSpec: v1.PodSpec{
				Containers: []v1.Container{
					{
						Image: "alpine@sha256:dad671370a148e9d9573e3e10a9f8cc26ce937bea78f3da80b570c2442364406",
					},
				},
			},
		},
		{
			name:       "deny",
			violations: 1,
			podSpec: v1.PodSpec{
				Containers: []v1.Container{
					{
						Image: "repo/tiller",
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			raw, _ := json.Marshal(corev1.Pod{Spec: tt.podSpec})
			ar := &admissionv1.AdmissionRequest{
				Namespace: "namespace",
				Name:      "name",
				Object:    runtime.RawExtension{Raw: raw},
				Resource:  metav1.GroupVersionResource{Group: "", Version: "v1", Resource: "pods"},
			}

			v := PolicyNoTiller{}
			if got, _ := v.Validate(ctx, policies.Config{}, ar); !reflect.DeepEqual(len(got), tt.violations) {
				t.Errorf("PolicyNoTiller() %s got %v want %v violations", tt.name, len(got), tt.violations)
			}
		})
	}
}
