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
// limitations under the License

package virtualservice

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/cruise-automation/k-rail/v3/policies"
	"github.com/cruise-automation/k-rail/v3/resource"
	networkingistiov1beta1 "istio.io/api/networking/v1beta1"
	admissionv1 "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestPolicyRequireVirtualServiceGatewayExemption_Validate(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name           string
		virtualService *resource.VirtualService
		config         *policies.Config
		violations     int
	}{
		{
			name:       "no gateways required, no gateways specified, no violation",
			violations: 0,
			virtualService: &resource.VirtualService{
				Spec: networkingistiov1beta1.VirtualService{
					Gateways: []string{},
				},
			},
			config: &policies.Config{
				PolicyRequireVirtualServiceGateways: &policies.VirtualServiceGatewaysConfig{
					AllowedGateways:    []string{},
					AllowEmptyGateways: true,
				},
			},
		},
		{
			name:       "gateways not required, no gateways specified, no violation",
			violations: 0,
			virtualService: &resource.VirtualService{
				Spec: networkingistiov1beta1.VirtualService{
					Gateways: []string{},
				},
			},
			config: &policies.Config{
				PolicyRequireVirtualServiceGateways: &policies.VirtualServiceGatewaysConfig{
					AllowedGateways:    []string{"istio-system/internal-gateway"},
					AllowEmptyGateways: true,
				},
			},
		},
		{
			name:       "gateways required, no gateways specified, violation",
			violations: 1,
			virtualService: &resource.VirtualService{
				Spec: networkingistiov1beta1.VirtualService{
					Gateways: []string{},
				},
			},
			config: &policies.Config{
				PolicyRequireVirtualServiceGateways: &policies.VirtualServiceGatewaysConfig{
					AllowedGateways:    []string{"istio-system/internal-gateway"},
					AllowEmptyGateways: false,
				},
			},
		},
		{
			name:       "gateways required, correct gateway specified, no violation",
			violations: 0,
			virtualService: &resource.VirtualService{
				Spec: networkingistiov1beta1.VirtualService{
					Gateways: []string{"istio-system/internal-gateway"},
				},
			},
			config: &policies.Config{
				PolicyRequireVirtualServiceGateways: &policies.VirtualServiceGatewaysConfig{
					AllowedGateways:    []string{"istio-system/internal-gateway"},
					AllowEmptyGateways: false,
				},
			},
		},
		{
			name:       "no gateways required, correct gateway specified, no violation",
			violations: 0,
			virtualService: &resource.VirtualService{
				Spec: networkingistiov1beta1.VirtualService{
					Gateways: []string{"istio-system/internal-gateway"},
				},
			},
			config: &policies.Config{
				PolicyRequireVirtualServiceGateways: &policies.VirtualServiceGatewaysConfig{
					AllowedGateways:    []string{"istio-system/internal-gateway"},
					AllowEmptyGateways: true,
				},
			},
		},
		{
			name:       "gateways required, wrong gateways specified, violation",
			violations: 1,
			virtualService: &resource.VirtualService{
				Spec: networkingistiov1beta1.VirtualService{
					Gateways: []string{"istio-system/external-gateway"},
				},
			},
			config: &policies.Config{
				PolicyRequireVirtualServiceGateways: &policies.VirtualServiceGatewaysConfig{
					AllowedGateways:    []string{"istio-system/internal-gateway"},
					AllowEmptyGateways: false,
				},
			},
		},
		{
			name:       "gateways required, one wrong and one correct gateway specified, violation",
			violations: 1,
			virtualService: &resource.VirtualService{
				Spec: networkingistiov1beta1.VirtualService{
					Gateways: []string{"istio-system/internal-gateway", "istio-system/external-gateway"},
				},
			},
			config: &policies.Config{
				PolicyRequireVirtualServiceGateways: &policies.VirtualServiceGatewaysConfig{
					AllowedGateways:    []string{"istio-system/internal-gateway"},
					AllowEmptyGateways: false,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var ar = &admissionv1.AdmissionRequest{}

			raw, _ := json.Marshal(tt.virtualService)
			ar = &admissionv1.AdmissionRequest{
				Namespace: "namespace",
				Name:      "name",
				Object:    runtime.RawExtension{Raw: raw},
				Resource:  metav1.GroupVersionResource{Group: "networking.istio.io", Version: "v1beta1", Resource: "virtualservices"},
			}

			v := PolicyRequireVirtualServiceGatewayExemption{}
			if got, _ := v.Validate(ctx, *tt.config, ar); !reflect.DeepEqual(len(got), tt.violations) {
				t.Errorf("PolicyRequireVirtualServiceGatewayExemption() %s got %v want %v violations", tt.name, len(got), tt.violations)
			}
		})
	}
}
