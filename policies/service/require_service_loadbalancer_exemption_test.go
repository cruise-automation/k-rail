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
// limitations under the License

package service

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/cruise-automation/k-rail/v3/policies"
	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestPolicyRequireServiceLoadbalancerExemption_Validate(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name       string
		service    *corev1.Service
		config     *policies.Config
		violations int
	}{
		{
			name:       "annotation present, no violation",
			violations: 0,
			service: &corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"cloud.google.com/load-balancer-type": "internal",
					},
				},
				Spec: corev1.ServiceSpec{
					Type: corev1.ServiceTypeLoadBalancer,
				},
			},
			config: &policies.Config{
				PolicyRequireServiceLoadBalancerAnnotations: []*policies.AnnotationConfig{
					&policies.AnnotationConfig{
						Annotation:    "cloud.google.com/load-balancer-type",
						AllowedValues: []string{"internal"},
						AllowMissing:  false,
					},
				},
			},
		},
		{
			name:       "annotation present but bad annotation, violation",
			violations: 1,
			service: &corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"cloud.google.com/load-balancer-type": "whatever123",
					},
				},
				Spec: corev1.ServiceSpec{
					Type: corev1.ServiceTypeLoadBalancer,
				},
			},
			config: &policies.Config{
				PolicyRequireServiceLoadBalancerAnnotations: []*policies.AnnotationConfig{
					&policies.AnnotationConfig{
						Annotation:    "cloud.google.com/load-balancer-type",
						AllowedValues: []string{"internal"},
						AllowMissing:  false,
					},
				},
			},
		},
		{
			name:       "no annotation present, violation",
			violations: 1,
			service: &corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{},
				},
				Spec: corev1.ServiceSpec{
					Type: corev1.ServiceTypeLoadBalancer,
				},
			},
			config: &policies.Config{
				PolicyRequireServiceLoadBalancerAnnotations: []*policies.AnnotationConfig{
					&policies.AnnotationConfig{
						Annotation:    "cloud.google.com/load-balancer-type",
						AllowedValues: []string{"internal"},
						AllowMissing:  false,
					},
				},
			},
		},
		{
			name:       "no annotation present, but empty ok, no violation",
			violations: 0,
			service: &corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{},
				},
				Spec: corev1.ServiceSpec{
					Type: corev1.ServiceTypeLoadBalancer,
				},
			},
			config: &policies.Config{
				PolicyRequireServiceLoadBalancerAnnotations: []*policies.AnnotationConfig{
					&policies.AnnotationConfig{
						Annotation:    "cloud.google.com/load-balancer-type",
						AllowedValues: []string{"internal"},
						AllowMissing:  true,
					},
				},
			},
		},
		{
			name:       "annotation present, multiple possibilities, no violation",
			violations: 0,
			service: &corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"cloud.google.com/load-balancer-type": "internal",
					},
				},
				Spec: corev1.ServiceSpec{
					Type: corev1.ServiceTypeLoadBalancer,
				},
			},
			config: &policies.Config{
				PolicyRequireServiceLoadBalancerAnnotations: []*policies.AnnotationConfig{
					&policies.AnnotationConfig{
						Annotation:    "cloud.google.com/load-balancer-type",
						AllowedValues: []string{"external", "internal"},
						AllowMissing:  false,
					},
				},
			},
		},
		{
			name:       "multiple annotations present, multiple possibilities, no violation",
			violations: 0,
			service: &corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"cloud.google.com/load-balancer-type": "internal",
						"myannotation":                        "yes",
						"anotherannotation":                   "yup",
					},
				},
				Spec: corev1.ServiceSpec{
					Type: corev1.ServiceTypeLoadBalancer,
				},
			},
			config: &policies.Config{
				PolicyRequireServiceLoadBalancerAnnotations: []*policies.AnnotationConfig{
					&policies.AnnotationConfig{
						Annotation:    "cloud.google.com/load-balancer-type",
						AllowedValues: []string{"external", "internal"},
						AllowMissing:  false,
					},
					&policies.AnnotationConfig{
						Annotation:    "myannotation",
						AllowedValues: []string{"yes", "internal"},
						AllowMissing:  false,
					},
				},
			},
		},
		{
			name:       "multiple annotations allowed, annotation present, no violation",
			violations: 0,
			service: &corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"networking.gke.io/load-balancer-type": "internal",
					},
				},
				Spec: corev1.ServiceSpec{
					Type: corev1.ServiceTypeLoadBalancer,
				},
			},
			config: &policies.Config{
				PolicyRequireServiceLoadBalancerAnnotations: []*policies.AnnotationConfig{
					&policies.AnnotationConfig{
						Annotations:   []string{"cloud.google.com/load-balancer-type", "networking.gke.io/load-balancer-type"},
						AllowedValues: []string{"internal"},
						AllowMissing:  false,
					},
				},
			},
		},
		{
			name:       "multiple annotations allowed, annotation present but bad annotation, violation",
			violations: 1,
			service: &corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"networking.gke.io/load-balancer-type": "whatever123",
					},
				},
				Spec: corev1.ServiceSpec{
					Type: corev1.ServiceTypeLoadBalancer,
				},
			},
			config: &policies.Config{
				PolicyRequireServiceLoadBalancerAnnotations: []*policies.AnnotationConfig{
					&policies.AnnotationConfig{
						Annotations:   []string{"cloud.google.com/load-balancer-type", "networking.gke.io/load-balancer-type"},
						AllowedValues: []string{"internal"},
						AllowMissing:  false,
					},
				},
			},
		},
		{
			name:       "multiple annotations allowed, no annotation present, violation",
			violations: 1,
			service: &corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{},
				},
				Spec: corev1.ServiceSpec{
					Type: corev1.ServiceTypeLoadBalancer,
				},
			},
			config: &policies.Config{
				PolicyRequireServiceLoadBalancerAnnotations: []*policies.AnnotationConfig{
					&policies.AnnotationConfig{
						Annotations:   []string{"cloud.google.com/load-balancer-type", "networking.gke.io/load-balancer-type"},
						AllowedValues: []string{"internal"},
						AllowMissing:  false,
					},
				},
			},
		},
		{
			name:       "multiple annotations allowed, no annotation present, but empty ok, no violation",
			violations: 0,
			service: &corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{},
				},
				Spec: corev1.ServiceSpec{
					Type: corev1.ServiceTypeLoadBalancer,
				},
			},
			config: &policies.Config{
				PolicyRequireServiceLoadBalancerAnnotations: []*policies.AnnotationConfig{
					&policies.AnnotationConfig{
						Annotations:   []string{"cloud.google.com/load-balancer-type", "networking.gke.io/load-balancer-type"},
						AllowedValues: []string{"internal"},
						AllowMissing:  true,
					},
				},
			},
		},
		{
			name:       "multiple annotations allowed, annotation present, multiple possibilities, no violation",
			violations: 0,
			service: &corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"networking.gke.io/load-balancer-type": "internal",
					},
				},
				Spec: corev1.ServiceSpec{
					Type: corev1.ServiceTypeLoadBalancer,
				},
			},
			config: &policies.Config{
				PolicyRequireServiceLoadBalancerAnnotations: []*policies.AnnotationConfig{
					&policies.AnnotationConfig{
						Annotations:   []string{"cloud.google.com/load-balancer-type", "networking.gke.io/load-balancer-type"},
						AllowedValues: []string{"external", "internal"},
						AllowMissing:  false,
					},
				},
			},
		},
		{
			name:       "multiple annotations allowed, multiple annotations present, multiple possibilities, no violation",
			violations: 0,
			service: &corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"networking.gke.io/load-balancer-type": "internal",
						"myannotation":                         "yes",
						"anotherannotation":                    "yup",
					},
				},
				Spec: corev1.ServiceSpec{
					Type: corev1.ServiceTypeLoadBalancer,
				},
			},
			config: &policies.Config{
				PolicyRequireServiceLoadBalancerAnnotations: []*policies.AnnotationConfig{
					&policies.AnnotationConfig{
						Annotations:   []string{"cloud.google.com/load-balancer-type", "networking.gke.io/load-balancer-type"},
						AllowedValues: []string{"external", "internal"},
						AllowMissing:  false,
					},
					&policies.AnnotationConfig{
						Annotation:    "myannotation",
						AllowedValues: []string{"yes", "internal"},
						AllowMissing:  false,
					},
				},
			},
		},
		{
			name:       "multiple annotations allowed, multiple annotations present, one possibilities, violation",
			violations: 1,
			service: &corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"networking.gke.io/load-balancer-type": "internal",
						"cloud.google.com/load-balancer-type":  "external",
					},
				},
				Spec: corev1.ServiceSpec{
					Type: corev1.ServiceTypeLoadBalancer,
				},
			},
			config: &policies.Config{
				PolicyRequireServiceLoadBalancerAnnotations: []*policies.AnnotationConfig{
					&policies.AnnotationConfig{
						Annotations:   []string{"cloud.google.com/load-balancer-type", "networking.gke.io/load-balancer-type"},
						AllowedValues: []string{"internal"},
						AllowMissing:  false,
					},
				},
			},
		},
		{
			name:       "multiple annotations allowed, multiple annotations present, one possibility, no violation",
			violations: 0,
			service: &corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"networking.gke.io/load-balancer-type": "internal",
						"cloud.google.com/load-balancer-type":  "internal",
					},
				},
				Spec: corev1.ServiceSpec{
					Type: corev1.ServiceTypeLoadBalancer,
				},
			},
			config: &policies.Config{
				PolicyRequireServiceLoadBalancerAnnotations: []*policies.AnnotationConfig{
					&policies.AnnotationConfig{
						Annotations:   []string{"cloud.google.com/load-balancer-type", "networking.gke.io/load-balancer-type"},
						AllowedValues: []string{"internal"},
						AllowMissing:  false,
					},
				},
			},
		},

		{
			name:       "no annotation present, but not type LB, no violation",
			violations: 0,
			service: &corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{},
				},
				Spec: corev1.ServiceSpec{
					Type: corev1.ServiceTypeClusterIP,
				},
			},
			config: &policies.Config{
				PolicyRequireServiceLoadBalancerAnnotations: []*policies.AnnotationConfig{
					&policies.AnnotationConfig{
						Annotation:    "cloud.google.com/load-balancer-type",
						AllowedValues: []string{"internal"},
						AllowMissing:  false,
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var ar = &admissionv1.AdmissionRequest{}

			raw, _ := json.Marshal(tt.service)
			ar = &admissionv1.AdmissionRequest{
				Namespace: "namespace",
				Name:      "name",
				Object:    runtime.RawExtension{Raw: raw},
				Resource:  metav1.GroupVersionResource{Group: "", Version: "v1", Resource: "services"},
			}

			v := PolicyRequireServiceLoadbalancerExemption{}
			if got, _ := v.Validate(ctx, *tt.config, ar); !reflect.DeepEqual(len(got), tt.violations) {
				t.Errorf("PolicyRequireServiceLoadbalancerExemption() %s got %v want %v violations", tt.name, len(got), tt.violations)
			}
		})
	}
}
