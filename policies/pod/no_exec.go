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

	admissionv1beta1 "k8s.io/api/admission/v1beta1"

	"github.com/cruise-automation/k-rail/policies"
	"github.com/cruise-automation/k-rail/resource"
)

type PolicyNoExec struct{}

func (p PolicyNoExec) Name() string {
	return "pod_no_exec"
}

func (p PolicyNoExec) Validate(ctx context.Context, config policies.Config, ar *admissionv1beta1.AdmissionRequest) ([]policies.ResourceViolation, []policies.PatchOperation) {

	resourceViolations := []policies.ResourceViolation{}

	podExecResource := resource.GetPodExecResource(ar)
	if podExecResource == nil {
		return resourceViolations, nil
	}

	violationText := "No pod exec: Execing into a Pod is forbidden without an exemption"

	resourceViolations = append(resourceViolations, policies.ResourceViolation{
		Namespace:    ar.Namespace,
		ResourceName: podExecResource.ResourceName,
		ResourceKind: podExecResource.ResourceKind,
		Violation:    violationText,
		Policy:       p.Name(),
	})

	return resourceViolations, nil
}
