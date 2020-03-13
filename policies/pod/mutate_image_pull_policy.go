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
	"fmt"
	"regexp"

	"github.com/cruise-automation/k-rail/policies"
	"github.com/cruise-automation/k-rail/resource"
	admissionv1beta1 "k8s.io/api/admission/v1beta1"
	corev1 "k8s.io/api/core/v1"
)

// PolicyImagePullPolicy is to enforce the imagePullPolicy
type PolicyImagePullPolicy struct{}

// Name is to return the name of the policy
func (p PolicyImagePullPolicy) Name() string {
	return "pod_image_pull_policy"
}

// Validate is to enforce the imagePullPolicy
func (p PolicyImagePullPolicy) Validate(ctx context.Context, config policies.Config, ar *admissionv1beta1.AdmissionRequest) ([]policies.ResourceViolation, []policies.PatchOperation) {

	podResource := resource.GetPodResource(ctx, ar)
	if podResource == nil {
		return nil, nil
	}

	var patches []policies.PatchOperation

	// if there is nothing configured, directly pass the Validate
	if len(config.PolicyImagePullPolicy) == 0 {
		return nil, nil
	}

	if podResource.ResourceKind == "Pod" {
		for index, container := range podResource.PodSpec.InitContainers {
			patch := checkImagePullPolicy(&container, fmt.Sprintf("/spec/initContainers/%d/imagePullPolicy", index), config.PolicyImagePullPolicy)
			if patch != nil {
				patches = append(patches, *patch)
			}
		}
		for index, container := range podResource.PodSpec.Containers {
			patch := checkImagePullPolicy(&container, fmt.Sprintf("/spec/containers/%d/imagePullPolicy", index), config.PolicyImagePullPolicy)
			if patch != nil {
				patches = append(patches, *patch)
			}
		}
	}

	return nil, patches
}

func checkImagePullPolicy(c *corev1.Container, path string, imagePullPolicyMap map[string][]string) *policies.PatchOperation {
	// loop through pullPolicy enforcement configured
	for enforcedPullPolicy, imageRegexes := range imagePullPolicyMap {
		for _, pattern := range imageRegexes {
			// check if image matches the regex
			matched, _ := regexp.MatchString(pattern, c.Image)
			if !matched {
				continue
			}
			if enforcedPullPolicy != string(c.ImagePullPolicy) {
				return &policies.PatchOperation{
					Op:    "replace",
					Path:  path,
					Value: enforcedPullPolicy,
				}
			} else {
				return nil
			}
		}
	}
	return nil
}
