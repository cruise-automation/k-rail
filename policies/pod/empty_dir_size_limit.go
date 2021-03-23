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
	"fmt"

	"github.com/cruise-automation/k-rail/v3/policies"
	"github.com/cruise-automation/k-rail/v3/resource"
	admissionv1 "k8s.io/api/admission/v1"
)

type PolicyEmptyDirSizeLimit struct {
}

func (p PolicyEmptyDirSizeLimit) Name() string {
	return "pod_empty_dir_size_limit"
}

const violationText = "Empty dir size limit: size limit exceeds the max value"

func (p PolicyEmptyDirSizeLimit) Validate(ctx context.Context, config policies.Config, ar *admissionv1.AdmissionRequest) ([]policies.ResourceViolation, []policies.PatchOperation) {
	var resourceViolations []policies.ResourceViolation

	podResource := resource.GetPodResource(ctx, ar)
	if podResource == nil {
		return resourceViolations, nil
	}

	cfg := config.MutateEmptyDirSizeLimit
	var patches []policies.PatchOperation

	for i, volume := range podResource.PodSpec.Volumes {
		if volume.EmptyDir == nil {
			continue
		}
		if volume.EmptyDir.SizeLimit == nil || volume.EmptyDir.SizeLimit.IsZero() {
			patches = append(patches, policies.PatchOperation{
				Op:    "replace",
				Path:  fmt.Sprintf(volumePatchPath(podResource.ResourceKind)+"/%d/emptyDir/sizeLimit", i),
				Value: cfg.DefaultSizeLimit.String(),
			})
			continue
		}

		if volume.EmptyDir.SizeLimit.Cmp(cfg.MaximumSizeLimit) > 0 {
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

const templateVolumePath = "/spec/template/spec/volumes"

func volumePatchPath(podKind string) string {
	nonTemplateKinds := map[string]string{
		"Pod":     "/spec/volumes",
		"CronJob": "/spec/jobTemplate/spec/template/spec/volumes",
	}
	if pathPath, ok := nonTemplateKinds[podKind]; ok {
		return pathPath
	}
	return templateVolumePath
}
