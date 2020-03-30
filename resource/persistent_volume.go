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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PersistentVolumeResource contains the information needed for processing by a Policy
type PersistentVolumeResource struct {
	PersistentVolume corev1.PersistentVolume
	ResourceName     string
	ResourceKind     string
}

// GetPersistentVolumeResource extracts and PersistentVolumeResource from an AdmissionRequest
func GetPersistentVolumeResource(ctx context.Context, ar *admissionv1beta1.AdmissionRequest) *PersistentVolumeResource {
	c := GetResourceCache(ctx)
	return c.getOrSet(cacheKeyPersistentVolume, func() interface{} {
		return decodePersistentVolumeResource(ar)
	}).(*PersistentVolumeResource)
}

func decodePersistentVolumeResource(ar *admissionv1beta1.AdmissionRequest) *PersistentVolumeResource {
	switch ar.Kind {
	case metav1.GroupVersionKind{Group: "", Version: "v1", Resource: "PersistentVolume"}:
		pv := corev1.PersistentVolume{}
		if err := decodeObject(ar.Object.Raw, &pv); err != nil {
			return nil
		}
		return &PersistentVolumeResource{
			PersistentVolume: pv,
			ResourceName:     GetResourceName(pv.ObjectMeta),
			ResourceKind:     "PersistentVolume",
		}
	default:
		return nil
	}
}
