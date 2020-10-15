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
	"testing"

	"github.com/cruise-automation/k-rail/policies"
	admissionv1beta1 "k8s.io/api/admission/v1beta1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/pointer"
)

func TestPolicyNoShareProcessNamespace_Validate(t *testing.T) {
	type args struct {
		ctx    context.Context
		config policies.Config
		ar     *admissionv1beta1.AdmissionRequest
	}
	tests := []struct {
		name       string
		podSpec    v1.PodSpec
		violations int
	}{
		{
			name: "nil, no violation",
			podSpec: v1.PodSpec{
				ShareProcessNamespace: nil,
			},
			violations: 0,
		},
		{
			name: "false, no violation",
			podSpec: v1.PodSpec{
				ShareProcessNamespace: pointer.BoolPtr(false),
			},
			violations: 0,
		},
		{
			name: "true, violation",
			podSpec: v1.PodSpec{
				ShareProcessNamespace: pointer.BoolPtr(true),
			},
			violations: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := PolicyNoShareProcessNamespace{}
			raw, _ := json.Marshal(corev1.Pod{Spec: tt.podSpec})
			ar := &admissionv1beta1.AdmissionRequest{
				Namespace: "namespace",
				Name:      "name",
				Object:    runtime.RawExtension{Raw: raw},
				Resource:  metav1.GroupVersionResource{Group: "", Version: "v1", Resource: "pods"},
			}
			got, _ := p.Validate(context.Background(), policies.Config{}, ar)
			if len(got) != tt.violations {
				t.Errorf("PolicyNoShareProcessNamespace.Validate() got = %v, want %v", len(got), tt.violations)
			}
		})
	}
}
