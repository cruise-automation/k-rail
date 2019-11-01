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
	"regexp"

	admissionv1beta1 "k8s.io/api/admission/v1beta1"
	corev1 "k8s.io/api/core/v1"

	"github.com/cruise-automation/k-rail/policies"
	"github.com/cruise-automation/k-rail/resource"
)

type PolicyTrustedRepository struct{}

func (p PolicyTrustedRepository) Name() string {
	return "pod_trusted_repository"
}

func (p PolicyTrustedRepository) Validate(ctx context.Context, config policies.Config, ar *admissionv1beta1.AdmissionRequest) ([]policies.ResourceViolation, []policies.PatchOperation) {

	resourceViolations := []policies.ResourceViolation{}

	podResource := resource.GetPodResource(ar)
	if podResource == nil {
		return resourceViolations, nil
	}

	violationText := "Trusted Image Repository: image must be sourced from a trusted repository"

	validateContainer := func(container corev1.Container) {
		matches := 0
		for _, pattern := range config.PolicyTrustedRepositoryRegexes {
			if matched, _ := regexp.MatchString(pattern, container.Image); matched {
				matches++
			}
		}

		if matches == 0 {
			resourceViolations = append(resourceViolations, policies.ResourceViolation{
				Namespace:    ar.Namespace,
				ResourceName: podResource.ResourceName,
				ResourceKind: podResource.ResourceKind,
				Violation:    violationText,
				Policy:       p.Name(),
			})
		}
	}

	for _, container := range podResource.PodSpec.Containers {
		validateContainer(container)
	}

	for _, container := range podResource.PodSpec.InitContainers {
		validateContainer(container)
	}

	return resourceViolations, nil
}
