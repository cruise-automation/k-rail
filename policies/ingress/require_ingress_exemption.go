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

	admissionv1beta1 "k8s.io/api/admission/v1beta1"

	"github.com/cruise-automation/k-rail/policies"
	"github.com/cruise-automation/k-rail/resource"
)

type PolicyRequireIngressExemption struct{}

func (p PolicyRequireIngressExemption) Name() string {
	return "ingress_require_ingress_exemption"
}

// Apply is called if the Policy is enabled to detect violations or perform mutations.
// Returning resource violations will cause a resource to be blocked unless there is an exemption for it.
func (p PolicyRequireIngressExemption) Apply(ctx context.Context, config policies.Config, ar *admissionv1beta1.AdmissionRequest) ([]policies.ResourceViolation, []policies.PatchOperation) {

	resourceViolations := []policies.ResourceViolation{}

	ingressResource := resource.GetIngressResource(ar)
	if ingressResource == nil {
		return resourceViolations, nil
	}

	violationText := "Require Ingress Exemption: Using certain Ingress classes requires an exemption"

	for _, ingressClass := range config.PolicyRequireIngressExemptionClasses {
		for annotation, value := range ingressResource.IngressExt.ObjectMeta.GetAnnotations() {
			if annotation == "kubernetes.io/ingress.class" {
				if value == ingressClass {
					resourceViolations = append(resourceViolations, policies.ResourceViolation{
						Namespace:    ar.Namespace,
						ResourceName: ingressResource.ResourceName,
						ResourceKind: ingressResource.ResourceKind,
						Violation:    violationText,
						Policy:       p.Name(),
					})
				}
			}
		}
		for annotation, value := range ingressResource.IngressNet.ObjectMeta.GetAnnotations() {
			if annotation == "kubernetes.io/ingress.class" {
				if value == ingressClass {
					resourceViolations = append(resourceViolations, policies.ResourceViolation{
						Namespace:    ar.Namespace,
						ResourceName: ingressResource.ResourceName,
						ResourceKind: ingressResource.ResourceKind,
						Violation:    violationText,
						Policy:       p.Name(),
					})
				}
			}
		}
	}
	return resourceViolations, nil
}

// Action will be called if the Policy is in violation and not in report-only mode.
func (p PolicyRequireIngressExemption) Action(ctx context.Context, exempt bool, config policies.Config, ar *admissionv1beta1.AdmissionRequest) (err error) {
	return
}
