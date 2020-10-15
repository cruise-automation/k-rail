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

package persistentVolume

import (
	"context"

	admissionv1beta1 "k8s.io/api/admission/v1beta1"

	"github.com/cruise-automation/k-rail/policies"
	"github.com/cruise-automation/k-rail/resource"
)

type PolicyNoPersistentVolumeHost struct{}

func (p PolicyNoPersistentVolumeHost) Name() string {
	return "persistent_volume_no_host_path"
}

func (p PolicyNoPersistentVolumeHost) Validate(ctx context.Context, config policies.Config, ar *admissionv1beta1.AdmissionRequest) ([]policies.ResourceViolation, []policies.PatchOperation) {

	resourceViolations := []policies.ResourceViolation{}

	pvResource := resource.GetPersistentVolumeResource(ctx, ar)
	if pvResource == nil {
		return resourceViolations, nil
	}

	violationText := "No Persistent Volume Host Path: Using the host path is forbidden"

	if pvResource.PersistentVolume.Spec.PersistentVolumeSource.HostPath != nil {
		resourceViolations = append(resourceViolations, policies.ResourceViolation{
			Namespace:    ar.Namespace,
			ResourceName: pvResource.ResourceName,
			ResourceKind: pvResource.ResourceKind,
			Violation:    violationText,
			Policy:       p.Name(),
		})
	}

	return resourceViolations, nil
}
