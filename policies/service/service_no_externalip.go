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

	admissionv1 "k8s.io/api/admission/v1"

	"github.com/cruise-automation/k-rail/v3/policies"
	"github.com/cruise-automation/k-rail/v3/resource"
)

type PolicyServiceNoExternalIP struct{}

func (p PolicyServiceNoExternalIP) Name() string {
	return "service_no_external_ip"
}

func (p PolicyServiceNoExternalIP) Validate(ctx context.Context, config policies.Config, ar *admissionv1.AdmissionRequest) ([]policies.ResourceViolation, []policies.PatchOperation) {

	resourceViolations := []policies.ResourceViolation{}

	serviceResource := resource.GetServiceResource(ctx, ar)
	if serviceResource == nil {
		return resourceViolations, nil
	}

	if len(serviceResource.Service.Spec.ExternalIPs) > 0 {
		resourceViolations = append(resourceViolations, policies.ResourceViolation{
			Namespace:    ar.Namespace,
			ResourceName: serviceResource.ResourceName,
			ResourceKind: serviceResource.ResourceKind,
			Violation:    "Services cannot have External IPs provided due to CVE-2020-8554",
			Policy:       p.Name(),
		})
	}

	return resourceViolations, nil
}
