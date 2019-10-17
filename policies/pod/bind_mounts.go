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

package pod

import (
	"context"

	"github.com/cruise-automation/k-rail/policies"
	"github.com/cruise-automation/k-rail/resource"
	admissionv1beta1 "k8s.io/api/admission/v1beta1"
)

type PolicyBindMounts struct{}

func (p PolicyBindMounts) Name() string {
	return "pod_no_bind_mounts"
}

func (p PolicyBindMounts) Validate(ctx context.Context, config policies.Config, ar *admissionv1beta1.AdmissionRequest) []policies.ResourceViolation {

	resourceViolations := []policies.ResourceViolation{}

	podResource := resource.GetPodResource(ar)
	if podResource == nil {
		return resourceViolations
	}

	violationText := "Host Bind Mounts: host bind mounts are forbidden"

	for _, volume := range podResource.PodSpec.Volumes {
		if volume.HostPath != nil {
			resourceViolations = append(resourceViolations, policies.ResourceViolation{
				Namespace:    ar.Namespace,
				ResourceName: podResource.ResourceName,
				ResourceKind: podResource.ResourceKind,
				Violation:    violationText,
				Policy:       p.Name(),
			})
		}
	}

	return resourceViolations
}
