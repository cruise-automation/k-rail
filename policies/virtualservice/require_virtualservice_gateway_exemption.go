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
	"fmt"
	"strings"

	admissionv1 "k8s.io/api/admission/v1"

	"github.com/cruise-automation/k-rail/v3/policies"
	"github.com/cruise-automation/k-rail/v3/resource"
)

type PolicyRequireVirtualServiceGatewayExemption struct{}

func (p PolicyRequireVirtualServiceGatewayExemption) Name() string {
	return "service_require_virtualservice_gateway_exemption"
}

var allowedGatewaysSet = map[string]bool{}

func (p PolicyRequireVirtualServiceGatewayExemption) Validate(ctx context.Context, config policies.Config, ar *admissionv1.AdmissionRequest) ([]policies.ResourceViolation, []policies.PatchOperation) {

	resourceViolations := []policies.ResourceViolation{}

	virtualServiceResource := resource.GetVirtualServiceResource(ctx, ar)
	if virtualServiceResource == nil {
		return resourceViolations, nil
	}

	if config.PolicyRequireVirtualServiceGateways == nil || len(config.PolicyRequireVirtualServiceGateways.AllowedGateways) == 0 {
		return resourceViolations, nil
	}

	// Memoize the allowedGatewaysSet
	if len(allowedGatewaysSet) == 0 {
		for _, gateway := range config.PolicyRequireVirtualServiceGateways.AllowedGateways {
			allowedGatewaysSet[gateway] = true
		}
	}

	gateways := virtualServiceResource.VirtualService.Spec.GetGateways()
	if len(gateways) == 0 && !config.PolicyRequireVirtualServiceGateways.AllowEmptyGateways {
		resourceViolations = append(resourceViolations, policies.ResourceViolation{
			Namespace:    ar.Namespace,
			ResourceName: virtualServiceResource.ResourceName,
			ResourceKind: virtualServiceResource.ResourceKind,
			Violation: fmt.Sprintf(
				"VirtualService Gateway not specified: Only the following gateways are allowed %s without an exemption",
				strings.Join(config.PolicyRequireVirtualServiceGateways.AllowedGateways, ", ")),
			Policy: p.Name(),
		})

	}

	for _, gateway := range gateways {
		if _, exists := allowedGatewaysSet[gateway]; !exists {
			resourceViolations = append(resourceViolations, policies.ResourceViolation{
				Namespace:    ar.Namespace,
				ResourceName: virtualServiceResource.ResourceName,
				ResourceKind: virtualServiceResource.ResourceKind,
				Violation: fmt.Sprintf(
					"Require VirtualService Gateway: Only the following gateways are allowed %s. Gateway value %s is not allowed without an exemption",
					strings.Join(config.PolicyRequireVirtualServiceGateways.AllowedGateways, ", "),
					gateway),
				Policy: p.Name(),
			})
		}

	}

	return resourceViolations, nil
}
