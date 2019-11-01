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

package resource

import (
	"encoding/json"
	"strings"

	admissionv1beta1 "k8s.io/api/admission/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PodExecResource contains the information needed for processing by a Policy
type PodExecResource struct {
	Command      string
	ResourceName string
	ResourceKind string
}

// GetPodExecResource extracts and PodExecResource from an AdmissionRequest
func GetPodExecResource(ar *admissionv1beta1.AdmissionRequest) *PodExecResource {
	switch ar.Kind {
	case metav1.GroupVersionKind{Group: "", Version: "v1", Kind: "PodExecOptions"}:
		podExecOptions := corev1.PodExecOptions{}
		if err := json.Unmarshal(ar.Object.Raw, &podExecOptions); err != nil {
			return nil
		}
		return &PodExecResource{
			Command:      strings.Join(podExecOptions.Command, " "),
			ResourceName: ar.Name,
			ResourceKind: "PodExec",
		}
	default:
		return nil
	}
}
