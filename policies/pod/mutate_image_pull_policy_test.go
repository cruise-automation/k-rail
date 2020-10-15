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
	"fmt"
	"reflect"
	"testing"

	"github.com/cruise-automation/k-rail/policies"
	admissionv1beta1 "k8s.io/api/admission/v1beta1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestPolicyImagePullPolicy(t *testing.T) {
	ctx := context.Background()

	tests := map[string]struct {
		config          policies.Config
		podSpec         v1.PodSpec
		expectedPatches map[string]*policies.PatchOperation
	}{
		"ImagePullPolicyNoMatch": {
			config: policies.Config{
				PolicyImagePullPolicy: map[string][]string{
					"IfNotPresent": []string{"^gcr.io/repo1/daytona.*", "^gcr.io/cruise-gcr-dev/daytona.*"},
				},
			},
			podSpec: v1.PodSpec{
				Containers: []v1.Container{
					{
						Image:           "gcr.io/cruise-gcr-abc/daytona@sha256:dad671370a148e9dc2442364406",
						ImagePullPolicy: "Always",
					},
				},
				InitContainers: []v1.Container{
					{
						Image:           "gcr.io/cruise-gcr-dev/daytona@sha256:dad671370a148e9dc2442364406",
						ImagePullPolicy: "IfNotPresent",
					},
				},
			},
			expectedPatches: nil,
		},
		"ImagePullPolicyMatchIfNotPresent1": {
			config: policies.Config{
				PolicyImagePullPolicy: map[string][]string{
					"IfNotPresent": []string{"^gcr.io/repo1/image1.*", "^gcr.io/repo2/image1.*"},
				},
			},
			podSpec: v1.PodSpec{
				Containers: []v1.Container{
					{
						Image:           "gcr.io/repo1/image1@sha256:dad671370a148e9dc2442364406",
						ImagePullPolicy: "Always",
					},
				},
			},
			expectedPatches: map[string]*policies.PatchOperation{
				"/spec/containers/0/imagePullPolicy": &policies.PatchOperation{
					Op:    "replace",
					Path:  "/spec/containers/0/imagePullPolicy",
					Value: "IfNotPresent",
				},
			},
		},

		"ImagePullPolicyMatchIfNotPresent2": {
			config: policies.Config{
				PolicyImagePullPolicy: map[string][]string{
					"IfNotPresent": []string{"^gcr.io/repo1/daytona.*", "^gcr.io/cruise-gcr-dev/daytona.*"},
				},
			},
			podSpec: v1.PodSpec{
				Containers: []v1.Container{
					{
						Image:           "gcr.io/repo1/daytona@sha256:dad671370a148e9dc2442364406",
						ImagePullPolicy: "Always",
					},
				},
				InitContainers: []v1.Container{
					{
						Image:           "gcr.io/cruise-gcr-dev/daytona@sha256:dad671370a148e9dc2442364406",
						ImagePullPolicy: "IfNotPresent",
					},
					{
						Image:           "gcr.io/cruise-gcr-dev/daytona@sha256:dad671370a148e9dc2442364406",
						ImagePullPolicy: "Always",
					},
				},
			},
			expectedPatches: map[string]*policies.PatchOperation{
				"/spec/containers/0/imagePullPolicy": &policies.PatchOperation{
					Op:    "replace",
					Path:  "/spec/containers/0/imagePullPolicy",
					Value: "IfNotPresent",
				},
				"/spec/initContainers/1/imagePullPolicy": &policies.PatchOperation{
					Op:    "replace",
					Path:  "/spec/initContainers/1/imagePullPolicy",
					Value: "IfNotPresent",
				},
			},
		},
		"ImagePullPolicyMatchAlways": {
			config: policies.Config{
				PolicyImagePullPolicy: map[string][]string{
					"IfNotPresent": []string{"^gcr.io/repo1/daytona.*", "^gcr.io/cruise-gcr-dev/daytona.*"},
					"Always":       []string{"gcr.io/repo1/abcd"},
				},
			},
			podSpec: v1.PodSpec{
				Containers: []v1.Container{
					{
						Image:           "gcr.io/repo1/abcd",
						ImagePullPolicy: "IfNotPresent",
					},
				},
				InitContainers: []v1.Container{
					{
						Image:           "gcr.io/cruise-gcr-dev/abcd",
						ImagePullPolicy: "IfNotPresent",
					},
					{
						Image:           "gcr.io/repo1/abcd",
						ImagePullPolicy: "Never",
					},
				},
			},
			expectedPatches: map[string]*policies.PatchOperation{
				"/spec/containers/0/imagePullPolicy": &policies.PatchOperation{
					Op:    "replace",
					Path:  "spec/containers/0/imagePullPolicy",
					Value: "Always",
				},
				"/spec/initContainers/1/imagePullPolicy": &policies.PatchOperation{
					Op:    "replace",
					Path:  "/spec/initContainers/1/imagePullPolicy",
					Value: "Always",
				},
			},
		},
	}
	for k, tt := range tests {
		t.Run(fmt.Sprintf("Testcase:%s", k), func(t *testing.T) {

			raw, _ := json.Marshal(corev1.Pod{Spec: tt.podSpec})
			ar := &admissionv1beta1.AdmissionRequest{
				Namespace: "namespace",
				Name:      "name",
				Object:    runtime.RawExtension{Raw: raw},
				Resource:  metav1.GroupVersionResource{Group: "", Version: "v1", Resource: "pods"},
			}

			v := PolicyImagePullPolicy{}
			_, patches := v.Validate(ctx, tt.config, ar)
			if len(tt.expectedPatches) != len(patches) {
				t.Fatalf("PolicyImagePullPolicy failed, expected number of Patches:%d, returned number of Patches: %d", len(tt.expectedPatches), len(patches))
			}
			for _, patch := range patches {
				p, ok := tt.expectedPatches[patch.Path]
				if !ok {
					t.Fatalf("PolicyImagePullPolicy return unwanted patch: %v", patch)
				}
				if !reflect.DeepEqual(p.Value, patch.Value) || p.Op != patch.Op {
					t.Fatalf("PolicyImagePullPolicy expectedPatch: %v, returned patch: %v", p, patch)
				}
			}
		})
	}
}
