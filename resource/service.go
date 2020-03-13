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

// ServiceResource contains the information needed for processing by a Policy
type ServiceResource struct {
	Service      corev1.Service
	ResourceName string
	ResourceKind string
}

// GetServiceResource extracts and ServiceResource from an AdmissionRequest
func GetServiceResource(ctx context.Context, ar *admissionv1beta1.AdmissionRequest) *ServiceResource {
	c := GetResourceCache(ctx)
	return c.getOrSet(cacheKeyService, func() interface{} {
		return decodeServiceResource(ar)
	}).(*ServiceResource)
}

func decodeServiceResource(ar *admissionv1beta1.AdmissionRequest) *ServiceResource {
	switch ar.Resource {
	case metav1.GroupVersionResource{Group: "core", Version: "v1", Resource: "services"}:
		svc := corev1.service{}
		if err := decodeObject(ar.Object.Raw, &svc); err != nil {
			return nil
		}
		return &ServiceResource{
			Service:      svc,
			ResourceName: GetResourceName(svc.ObjectMeta),
			ResourceKind: "Service",
		}
	default:
		return nil
	}
}
