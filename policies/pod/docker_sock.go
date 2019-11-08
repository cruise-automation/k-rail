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

// PolicyDockerSock forbids a hostPath mount for just the Docker socket.
// Note that it does not block mounting '/', '/var', or '/var/run'.
// It is recommended that you use the BindMounts Policy to block all bind mounts.
type PolicyDockerSock struct{}

func (p PolicyDockerSock) Name() string {
	return "pod_no_docker_sock"
}

// Validate is called if the Policy is enabled to detect violations or perform mutations.
// Returning resource violations will cause a resource to be blocked unless there is an exemption for it.
func (p PolicyDockerSock) Apply(ctx context.Context, config policies.Config, ar *admissionv1beta1.AdmissionRequest) ([]policies.ResourceViolation, []policies.PatchOperation) {

	resourceViolations := []policies.ResourceViolation{}

	podResource := resource.GetPodResource(ar)
	if podResource == nil {
		return resourceViolations, nil
	}

	violationText := "Docker Sock Mount: mounting the Docker socket is forbidden"

	for _, volume := range podResource.PodSpec.Volumes {
		if volume.HostPath != nil {
			if volume.HostPath.Path == "/var/run/docker.sock" {
				resourceViolations = append(resourceViolations, policies.ResourceViolation{
					Namespace:    ar.Namespace,
					ResourceName: podResource.ResourceName,
					ResourceKind: podResource.ResourceKind,
					Violation:    violationText,
					Policy:       p.Name(),
				})
			}
		}
	}

	return resourceViolations, nil
}

// Action will be called if the Policy is in violation and not in report-only mode.
func (p PolicyDockerSock) Action(ctx context.Context, exempt bool, config policies.Config, ar *admissionv1beta1.AdmissionRequest) (err error) {
	return
}
