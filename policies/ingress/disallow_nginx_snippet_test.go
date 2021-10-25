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

package ingress

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"

	admissionv1 "k8s.io/api/admission/v1"
	extensionsv1beta1 "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/cruise-automation/k-rail/v3/policies"
)

func TestPolicyDisallowNGINXSnippet(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name       string
		ingressExt *extensionsv1beta1.Ingress
		violations int
	}{
		{
			name:       "deny 1",
			violations: 1,
			ingressExt: &extensionsv1beta1.Ingress{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"nginx.ingress.kubernetes.io/server-snippet": "i'm malicious",
					},
				},
			},
		},
		{
			name:       "deny 2",
			violations: 2,
			ingressExt: &extensionsv1beta1.Ingress{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"nginx.ingress.kubernetes.io/server-snippet": "i'm malicious",
						"nginx.ingress.kubernetes.io/auth-snippet":   "me too",
					},
				},
			},
		},
		{
			name:       "allow",
			violations: 0,
			ingressExt: &extensionsv1beta1.Ingress{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"foo": "bar",
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var ar = &admissionv1.AdmissionRequest{}

			if tt.ingressExt != nil {
				raw, _ := json.Marshal(tt.ingressExt)
				ar = &admissionv1.AdmissionRequest{
					Namespace: "namespace",
					Name:      "name",
					Object:    runtime.RawExtension{Raw: raw},
					Resource:  metav1.GroupVersionResource{Group: "extensions", Version: "v1beta1", Resource: "ingresses"},
				}
			}

			v := PolicyDisallowNGINXSnippet{}
			if got, _ := v.Validate(ctx, policies.Config{}, ar); !reflect.DeepEqual(len(got), tt.violations) {
				t.Errorf("PolicyDisallowNGINXSnippet() %s got %v want %v violations", tt.name, len(got), tt.violations)
			}
		})
	}
}
