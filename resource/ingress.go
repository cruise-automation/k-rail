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
	networkingv1 "k8s.io/api/networking/v1"
	networkingv1beta1 "k8s.io/api/networking/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// IngressResource contains the information needed for processing by a Policy
type IngressResource struct {
	IngressExt        extensionsv1beta1.Ingress
	IngressNetV1Beta1 networkingv1beta1.Ingress
	IngressNetV1      networkingv1.Ingress
	ResourceName      string
	ResourceKind      string
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
	case metav1.GroupVersionResource{Group: "networking.k8s.io", Version: "v1beta1", Resource: "ingresses"}:
		ing := networkingv1beta1.Ingress{}
		if err := decodeObject(ar.Object.Raw, &ing); err != nil {
			return nil
		}
		return &IngressResource{
			IngressNetV1Beta1: ing,
			ResourceName:      GetResourceName(ing.ObjectMeta),
			ResourceKind:      "Ingress",
		}
	case metav1.GroupVersionResource{Group: "networking.k8s.io", Version: "v1", Resource: "ingresses"}:
		ing := networkingv1.Ingress{}
		if err := decodeObject(ar.Object.Raw, &ing); err != nil {
			return nil
		}
		return &IngressResource{
			IngressNetV1: ing,
			ResourceName: GetResourceName(ing.ObjectMeta),
			ResourceKind: "Ingress",
		}
	default:
		return nil
	}
}

// GetAnnotations returns ingress annotations, across all available ingress versions
func (ir IngressResource) GetAnnotations() map[string]string {
	if ir.IngressExt.Annotations != nil {
		return ir.IngressExt.Annotations
	} else if ir.IngressNetV1Beta1.Annotations != nil {
		return ir.IngressNetV1Beta1.Annotations
	} else if ir.IngressNetV1.Annotations != nil {
		return ir.IngressNetV1.Annotations
	}
	return nil
}

// GetHosts returns list of all hosts in ingress spec, across all available ingress versions
func (ir IngressResource) GetHosts() []string {
	hosts := []string{}
	for _, rule := range ir.IngressExt.Spec.Rules {
		hosts = append(hosts, rule.Host)
	}
	for _, rule := range ir.IngressNetV1Beta1.Spec.Rules {
		hosts = append(hosts, rule.Host)
	}
	for _, rule := range ir.IngressNetV1.Spec.Rules {
		hosts = append(hosts, rule.Host)
	}
	return hosts
}
