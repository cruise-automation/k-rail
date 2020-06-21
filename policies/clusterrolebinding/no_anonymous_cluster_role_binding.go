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

package clusterrolebinding

import (
	"context"
	"strings"

	admissionv1beta1 "k8s.io/api/admission/v1beta1"

	"github.com/cruise-automation/k-rail/policies"
	"github.com/cruise-automation/k-rail/resource"
)

type PolicyNoAnonymousClusterRoleBinding struct{}

func (p PolicyNoAnonymousClusterRoleBinding) Name() string {
	return "cluster_role_binding_no_anonymous_subject"
}

func (p PolicyNoAnonymousClusterRoleBinding) Validate(ctx context.Context, config policies.Config, ar *admissionv1beta1.AdmissionRequest) ([]policies.ResourceViolation, []policies.PatchOperation) {

	resourceViolations := []policies.ResourceViolation{}
	crbResource := resource.GetClusterRoleBindingResource(ctx, ar)
	if crbResource == nil {
		return resourceViolations, nil
	}

	violationText := "No Anonymous Cluster Role Binding: Granting permissions to anonymous subject is forbidden"
	for _, subject := range crbResource.ClusterRoleBinding.Subjects {
		if (strings.ToLower(subject.Name) == "system:anonymous") || (strings.ToLower(subject.Name) == "system:unauthenticated") {
			resourceViolations = append(resourceViolations, policies.ResourceViolation{
				Namespace:    ar.Namespace,
				ResourceName: crbResource.ResourceName,
				ResourceKind: crbResource.ResourceKind,
				Violation:    violationText,
				Policy:       p.Name(),
			})
		}
	}

	return resourceViolations, nil
}
