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
// limitations under the License.

package pod

import (
	"context"

	"github.com/cruise-automation/k-rail/policies"
	"github.com/cruise-automation/k-rail/resource"
	admissionv1 "k8s.io/api/admission/v1"
)

type PolicyNoShareProcessNamespace struct{}

func (p PolicyNoShareProcessNamespace) Name() string {
	return "pod_no_shareprocessnamespace"
}

func (p PolicyNoShareProcessNamespace) Validate(ctx context.Context, config policies.Config, ar *admissionv1.AdmissionRequest) ([]policies.ResourceViolation, []policies.PatchOperation) {

	resourceViolations := []policies.ResourceViolation{}

	podResource := resource.GetPodResource(ctx, ar)
	if podResource == nil {
		return resourceViolations, nil
	}

	violationText := "No ShareProcessNamespace: sharing the process namespace among containers in a Pod is forbidden."

	if podResource.PodSpec.ShareProcessNamespace != nil && *podResource.PodSpec.ShareProcessNamespace {
		resourceViolations = append(resourceViolations, policies.ResourceViolation{
			Namespace:    ar.Namespace,
			ResourceName: podResource.ResourceName,
			ResourceKind: podResource.ResourceKind,
			Violation:    violationText,
			Policy:       p.Name(),
		})
	}

	return resourceViolations, nil
}
