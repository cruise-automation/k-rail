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
	"strings"

	admissionv1beta1 "k8s.io/api/admission/v1beta1"

	"github.com/cruise-automation/k-rail/policies"
	"github.com/cruise-automation/k-rail/resource"

	digest "github.com/opencontainers/go-digest"
)

type PolicyImageImmutableReference struct{}

func (p PolicyImageImmutableReference) Name() string {
	return "pod_immutable_reference"
}

func (p PolicyImageImmutableReference) Validate(ctx context.Context, config policies.Config, ar *admissionv1beta1.AdmissionRequest) ([]policies.ResourceViolation, []policies.PatchOperation) {

	resourceViolations := []policies.ResourceViolation{}

	podResource := resource.GetPodResource(ar)
	if podResource == nil {
		return resourceViolations, nil
	}

	violationText := "Immutable Image Reference: image tag must include its sha256 digest"

	for _, container := range podResource.PodSpec.Containers {

		// validate that the image name ends with a digest
		refSplit := strings.Split(container.Image, "@")
		if len(refSplit) == 2 {
			d, _ := digest.Parse(refSplit[len(refSplit)-1])
			err := d.Validate()
			if err != nil {
				resourceViolations = append(resourceViolations, policies.ResourceViolation{
					Namespace:    ar.Namespace,
					ResourceName: podResource.ResourceName,
					ResourceKind: podResource.ResourceKind,
					Violation:    violationText,
					Policy:       p.Name(),
					Error:        err,
				})
			}
		} else {
			resourceViolations = append(resourceViolations, policies.ResourceViolation{
				Namespace:    ar.Namespace,
				ResourceName: podResource.ResourceName,
				ResourceKind: podResource.ResourceKind,
				Violation:    violationText,
				Policy:       p.Name(),
				Error:        nil,
			})
		}
	}

	return resourceViolations, nil
}
