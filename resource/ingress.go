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

	admissionv1 "k8s.io/api/admission/v1"
	extensionsv1beta1 "k8s.io/api/extensions/v1beta1"
	networkingv1beta1 "k8s.io/api/networking/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// IngressResource contains the information needed for processing by a Policy
type IngressResource struct {
	IngressExt   extensionsv1beta1.Ingress
	IngressNet   networkingv1beta1.Ingress
	ResourceName string
	ResourceKind string
}

// GetIngressResource extracts and IngressResource from an AdmissionRequest
func GetIngressResource(ctx context.Context, ar *admissionv1.AdmissionRequest) *IngressResource {
	c := GetResourceCache(ctx)
	return c.getOrSet(cacheKeyIngress, func() interface{} {
		return decodeIngressResource(ar)
	}).(*IngressResource)
}

func decodeIngressResource(ar *admissionv1.AdmissionRequest) *IngressResource {
	switch ar.Resource {
	case metav1.GroupVersionResource{Group: "extensions", Version: "v1beta1", Resource: "ingresses"}:
		ing := extensionsv1beta1.Ingress{}
		if err := decodeObject(ar.Object.Raw, &ing); err != nil {
			return nil
		}
		return &IngressResource{
			IngressExt:   ing,
			ResourceName: GetResourceName(ing.ObjectMeta),
			ResourceKind: "Ingress",
		}
	case metav1.GroupVersionResource{Group: "networking", Version: "v1beta1", Resource: "ingresses"}:
		ing := networkingv1beta1.Ingress{}
		if err := decodeObject(ar.Object.Raw, &ing); err != nil {
			return nil
		}
		return &IngressResource{
			IngressNet:   ing,
			ResourceName: GetResourceName(ing.ObjectMeta),
			ResourceKind: "Ingress",
		}
	default:
		return nil
	}
}
