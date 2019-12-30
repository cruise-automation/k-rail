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

	"github.com/cruise-automation/k-rail/policies"
	"github.com/cruise-automation/k-rail/resource"
	admissionv1beta1 "k8s.io/api/admission/v1beta1"
	apiresource "k8s.io/apimachinery/pkg/api/resource"
)

type PolicyEmptyDirSizeLimit struct {
	MaxSize, DefaultSize apiresource.Quantity
}

func (p PolicyEmptyDirSizeLimit) Name() string {
	return "pod_empty_dir_size_limit"
}

const violationText = "Empty dir size limit: size limit is required for Pods that use emptyDir"

func (p PolicyEmptyDirSizeLimit) Validate(ctx context.Context, config policies.Config, ar *admissionv1beta1.AdmissionRequest) ([]policies.ResourceViolation, []policies.PatchOperation) {
	var resourceViolations []policies.ResourceViolation

	podResource := resource.GetPodResource(ar)
	if podResource == nil {
		return resourceViolations, nil
	}

	var patches []policies.PatchOperation

	for i, volume := range podResource.PodSpec.Volumes {
		if volume.EmptyDir == nil {
			continue
		}
		if volume.EmptyDir.SizeLimit == nil || volume.EmptyDir.SizeLimit.IsZero() {
			patches = append(patches, policies.PatchOperation{
				Op:    "replace",
				Path:  fmt.Sprintf("/spec/volumes/%d/emptyDir/sizeLimit", i),
				Value: p.DefaultSize.String(),
			})
			continue
		}

		if volume.EmptyDir.SizeLimit.Cmp(p.MaxSize) > 0 {
			resourceViolations = append(resourceViolations, policies.ResourceViolation{
				Namespace:    ar.Namespace,
				ResourceName: podResource.ResourceName,
				ResourceKind: podResource.ResourceKind,
				Violation:    violationText,
				Policy:       p.Name(),
			})
		}
	}
	return resourceViolations, patches
}
