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

func TestPolicySafeToEvict(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name        string
		podSpec     corev1.PodSpec
		annotations map[string]string
		violations  int
	}{
		{
			name:       "allow with annotation and volume",
			violations: 0,
			podSpec: corev1.PodSpec{
				Volumes: []corev1.Volume{
					{
						VolumeSource: corev1.VolumeSource{
							HostPath: &corev1.HostPathVolumeSource{
								Path: "/host-path",
							},
						},
					},
				},
			},
			annotations: map[string]string{
				"cluster-autoscaler.kubernetes.io/safe-to-evict": "true",
			},
		},
		{
			name:       "disallow with annotation and volume",
			violations: 1,
			podSpec: corev1.PodSpec{
				Volumes: []corev1.Volume{
					{
						VolumeSource: corev1.VolumeSource{
							HostPath: &corev1.HostPathVolumeSource{
								Path: "/host-path",
							},
						},
					},
				},
			},
			annotations: map[string]string{
				"cluster-autoscaler.kubernetes.io/safe-to-evict": "false",
			},
		},
		{
			name:       "disallow with volume and without annotation",
			violations: 1,
			podSpec: corev1.PodSpec{
				Volumes: []corev1.Volume{
					{
						VolumeSource: corev1.VolumeSource{
							HostPath: &corev1.HostPathVolumeSource{
								Path: "/host-path",
							},
						},
					},
				},
			},
			annotations: map[string]string{},
		},
		{
			name:       "allow with annotation and no volume",
			violations: 0,
			podSpec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Image: "gcr.io/some-gcr-repo/security/k-rail",
					},
				},
			},
			annotations: map[string]string{
				"cluster-autoscaler.kubernetes.io/safe-to-evict": "false",
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

			v := PolicySafeToEvict{}
			conf := policies.Config{}
			if got := v.Validate(ctx, conf, ar); !reflect.DeepEqual(len(got), tt.violations) {
				t.Errorf("PolicySafeToEvict() %s got %v want %v violations", tt.name, len(got), tt.violations)
			}
		})
	}
}
