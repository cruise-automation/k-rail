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

package rolebinding

import (
	"context"
	"strings"

	admissionv1beta1 "k8s.io/api/admission/v1beta1"

	"github.com/cruise-automation/k-rail/policies"
	"github.com/cruise-automation/k-rail/resource"
)

type PolicyNoAnonymousRoleBinding struct{}

func (p PolicyNoAnonymousRoleBinding) Name() string {
	return "role_binding_no_anonymous_subject"
}

func (p PolicyNoAnonymousRoleBinding) Validate(ctx context.Context, config policies.Config, ar *admissionv1beta1.AdmissionRequest) ([]policies.ResourceViolation, []policies.PatchOperation) {

	resourceViolations := []policies.ResourceViolation{}
	rbResource := resource.GetRoleBindingResource(ctx, ar)
	if rbResource == nil {
		return resourceViolations, nil
	}

	violationText := "No Anonymous Role Binding: Granting permissions to anonymous or unauthenticated subject is forbidden"
	for _, subject := range rbResource.RoleBinding.Subjects {
		if (strings.ToLower(subject.Name) == "system:anonymous") || (strings.ToLower(subject.Name) == "system:unauthenticated") {
			resourceViolations = append(resourceViolations, policies.ResourceViolation{
				Namespace:    ar.Namespace,
				ResourceName: rbResource.ResourceName,
				ResourceKind: rbResource.ResourceKind,
				Violation:    violationText,
				Policy:       p.Name(),
			})
		}
	}

	return resourceViolations, nil
}
