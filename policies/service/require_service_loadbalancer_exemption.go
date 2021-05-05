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
	"fmt"
	"strings"

	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"

	"github.com/cruise-automation/k-rail/v3/policies"
	"github.com/cruise-automation/k-rail/v3/resource"
)

type PolicyRequireServiceLoadbalancerExemption struct{}

func (p PolicyRequireServiceLoadbalancerExemption) Name() string {
	return "service_require_loadbalancer_exemption"
}

func (p PolicyRequireServiceLoadbalancerExemption) Validate(ctx context.Context, config policies.Config, ar *admissionv1.AdmissionRequest) ([]policies.ResourceViolation, []policies.PatchOperation) {

	resourceViolations := []policies.ResourceViolation{}

	serviceResource := resource.GetServiceResource(ctx, ar)
	if serviceResource == nil {
		return resourceViolations, nil
	}

	// Only Servcices of type LoadBalancers are policed.
	if serviceResource.Service.Spec.Type != corev1.ServiceTypeLoadBalancer {
		return resourceViolations, nil
	}

	annotations := serviceResource.Service.ObjectMeta.GetAnnotations()

	// Each annotation entry in the config is tested sequentially
	for _, annotationConfig := range config.PolicyRequireServiceLoadBalancerAnnotations {
		annotationsToCheck := annotationConfig.Annotations
		if annotationConfig.Annotation != "" {
			annotationsToCheck = append(annotationsToCheck, annotationConfig.Annotation)
		}

		atLeastOneAnnotationExists := false

		for _, annotation := range annotationsToCheck {
			value, exists := annotations[annotation]
			atLeastOneAnnotationExists = atLeastOneAnnotationExists || exists
			// The annotation is only tested against possible values if it exists
			if exists {
				valueAllowed := false
				for _, allowedValue := range annotationConfig.AllowedValues {
					if value == allowedValue {
						valueAllowed = true
						break
					}
				}

				if !valueAllowed {
					resourceViolations = append(resourceViolations, policies.ResourceViolation{
						Namespace:    ar.Namespace,
						ResourceName: serviceResource.ResourceName,
						ResourceKind: serviceResource.ResourceKind,
						Violation:    fmt.Sprintf("Require Service LoadBalancer annotations: Annotation %s value %s is not allowed", annotationConfig.Annotation, value),
						Policy:       p.Name(),
					})
				}
			}
		}

		// If the annotation(s) is not present, the policy config defines if this is acceptable.
		if !atLeastOneAnnotationExists && !annotationConfig.AllowMissing {
			requiredAnnotations := strings.Join(annotationConfig.Annotations, " or ")
			resourceViolations = append(resourceViolations, policies.ResourceViolation{
				Namespace:    ar.Namespace,
				ResourceName: serviceResource.ResourceName,
				ResourceKind: serviceResource.ResourceKind,
				Violation:    fmt.Sprintf("Require Service LoadBalancer annotations: Annotation %s cannot be empty", requiredAnnotations),
				Policy:       p.Name(),
			})
		}
	}

	return resourceViolations, nil
}
