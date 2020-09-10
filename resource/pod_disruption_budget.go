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

package resource

import (
	"context"

	admissionv1beta1 "k8s.io/api/admission/v1beta1"
	policyv1beta1 "k8s.io/api/policy/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PodDisruptionBudgetResource contains the information needed for processing by a Policy
type PodDisruptionBudgetResource struct {
	PodDisruptionBudget policyv1beta1.PodDisruptionBudget
	ResourceName        string
	ResourceKind        string
}

// GetPodDisruptionBudgetResource extracts an PodDisruptionBudgetResource from an AdmissionRequest
func GetPodDisruptionBudgetResource(ctx context.Context, ar *admissionv1beta1.AdmissionRequest) *PodDisruptionBudgetResource {
	c := GetResourceCache(ctx)
	return c.getOrSet(cacheKeyPodDisruptionBudget, func() interface{} {
		return decodePodDisruptionBudgetResource(ar)
	}).(*PodDisruptionBudgetResource)
}

func decodePodDisruptionBudgetResource(ar *admissionv1beta1.AdmissionRequest) *PodDisruptionBudgetResource {
	switch ar.Resource {
	case metav1.GroupVersionResource{Group: "policy", Version: "v1beta1", Resource: "poddisruptionbudgets"}:
		pdb := policyv1beta1.PodDisruptionBudget{}
		if err := decodeObject(ar.Object.Raw, &pdb); err != nil {
			return nil
		}
		return &PodDisruptionBudgetResource{
			PodDisruptionBudget: pdb,
			ResourceName:        GetResourceName(pdb.ObjectMeta),
			ResourceKind:        "PodDisruptionBudget",
		}
	default:
		return nil
	}
}
