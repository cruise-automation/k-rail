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

package ingress

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/cruise-automation/k-rail/policies"
	admissionv1beta1 "k8s.io/api/admission/v1beta1"
	extensionsv1beta1 "k8s.io/api/extensions/v1beta1"
	networkingv1beta1 "k8s.io/api/networking/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestValidatePodDockerSock(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name       string
		ingressExt *extensionsv1beta1.Ingress
		ingressNet *networkingv1beta1.Ingress
		violations int
	}{
		{
			name:       "deny ext ing",
			violations: 1,
			ingressExt: &extensionsv1beta1.Ingress{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"kubernetes.io/ingress.class": "public",
					},
				},
			},
		},
		{
			name:       "deny net ing",
			violations: 1,
			ingressNet: &networkingv1beta1.Ingress{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"kubernetes.io/ingress.class": "public",
					},
				},
			},
		},
		{
			name:       "allow ext ing",
			violations: 0,
			ingressExt: &extensionsv1beta1.Ingress{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"kubernetes.io/ingress.class": "private",
					},
				},
			},
		},
		{
			name:       "allow net ing",
			violations: 0,
			ingressNet: &networkingv1beta1.Ingress{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"kubernetes.io/ingress.class": "private",
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var ar = &admissionv1beta1.AdmissionRequest{}

			if tt.ingressExt != nil {
				raw, _ := json.Marshal(tt.ingressExt)
				ar = &admissionv1beta1.AdmissionRequest{
					Namespace: "namespace",
					Name:      "name",
					Object:    runtime.RawExtension{Raw: raw},
					Resource:  metav1.GroupVersionResource{Group: "extensions", Version: "v1beta1", Resource: "ingresses"},
				}
			}

			if tt.ingressNet != nil {
				raw, _ := json.Marshal(tt.ingressNet)
				ar = &admissionv1beta1.AdmissionRequest{
					Namespace: "namespace",
					Name:      "name",
					Object:    runtime.RawExtension{Raw: raw},
					Resource:  metav1.GroupVersionResource{Group: "networking", Version: "v1beta1", Resource: "ingresses"},
				}
			}

			v := PolicyRequireIngressExemption{}
			if got := v.Validate(ctx, policies.Config{PolicyRequireIngressExemptionClasses: []string{"public"}}, ar); !reflect.DeepEqual(len(got), tt.violations) {
				t.Errorf("PolicyRequireIngressExemption() %s got %v want %v violations", tt.name, len(got), tt.violations)
			}
		})
	}
}
