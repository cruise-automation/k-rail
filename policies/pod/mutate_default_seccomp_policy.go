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

type PolicyDefaultSeccompPolicy struct{}

func (p PolicyDefaultSeccompPolicy) Name() string {
	return "pod_default_seccomp_policy"
}

func (p PolicyDefaultSeccompPolicy) Validate(ctx context.Context, config policies.Config, ar *admissionv1beta1.AdmissionRequest) ([]policies.ResourceViolation, []policies.PatchOperation) {

	podResource := resource.GetPodResource(ar)
	if podResource == nil {
		return nil, nil
	}

	var patches []policies.PatchOperation

	seccompPolicy := config.PolicyDefaultSeccompPolicy
	if len(seccompPolicy) == 0 {
		seccompPolicy = "runtime/default"
	}

	if podResource.ResourceKind == "Pod" {
		apply := true
		for name := range podResource.PodAnnotations {
			if name == "seccomp.security.alpha.kubernetes.io/pod" {
				apply = false
			}
		}

		if apply {
			if podResource.PodAnnotations == nil {
				patches = append(patches, policies.PatchOperation{
					Op:   "add",
					Path: "/metadata/annotations",
					Value: map[string]string{
						"seccomp.security.alpha.kubernetes.io/pod": seccompPolicy,
					},
				})
			} else {
				patches = append(patches, policies.PatchOperation{
					Op: "replace",
					// escape `/` in a path with ~1 in JSONPatch format 🤮
					Path:  "/metadata/annotations/seccomp.security.alpha.kubernetes.io~1pod",
					Value: seccompPolicy,
				})
			}
		}
	}

	return nil, patches
}
