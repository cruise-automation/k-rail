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

	"github.com/cruise-automation/k-rail/policies"
	admissionv1beta1 "k8s.io/api/admission/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestPolicyDefaultSeccompPolicy(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name            string
		podSpec         corev1.PodSpec
		annotations     map[string]string
		expectedPatches map[string]*policies.PatchOperation
	}{
		{
			name:    "with annotation",
			podSpec: corev1.PodSpec{},
			annotations: map[string]string{
				"cluster-autoscaler.kubernetes.io/safe-to-evict": "true",
			},
			expectedPatches: map[string]*policies.PatchOperation{
				"/metadata/annotations/seccomp.security.alpha.kubernetes.io~1pod": &policies.PatchOperation{
					Op:    "replace",
					Path:  "/metadata/annotations/seccomp.security.alpha.kubernetes.io~1pod",
					Value: "runtime/default",
				},
			},
		},
		{
			name:        "without annotation",
			podSpec:     corev1.PodSpec{},
			annotations: nil,
			expectedPatches: map[string]*policies.PatchOperation{
				"/metadata/annotations": &policies.PatchOperation{
					Op:   "add",
					Path: "/metadata/annotations",
					Value: map[string]string{
						"seccomp.security.alpha.kubernetes.io/pod": "runtime/default",
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			raw, _ := json.Marshal(corev1.Pod{Spec: tt.podSpec, ObjectMeta: metav1.ObjectMeta{Annotations: tt.annotations}})
			ar := &admissionv1beta1.AdmissionRequest{
				Namespace: "namespace",
				Name:      "name",
				Object:    runtime.RawExtension{Raw: raw},
				Resource:  metav1.GroupVersionResource{Group: "", Version: "v1", Resource: "pods"},
			}

			v := PolicyDefaultSeccompPolicy{}
			conf := policies.Config{}
			_, patches := v.Validate(ctx, conf, ar)
			if len(tt.expectedPatches) != len(patches) {
				t.Fatalf("PolicyDefaultSeccompPolicy failed, expected number of Patches:%d, returned number of Patches: %d", len(tt.expectedPatches), len(patches))
			}
			for _, patch := range patches {
				p, ok := tt.expectedPatches[patch.Path]
				if !ok {
					t.Fatalf("PolicyDefaultSeccompPolicy return unwanted patch: %v", patch)
				}
				if !reflect.DeepEqual(p.Value, patch.Value) || p.Op != patch.Op {
					t.Fatalf("PolicyDefaultSeccompPolicy expectedPatch: %v, returned patch: %v", p, patch)
				}
			}
		})
	}
}
