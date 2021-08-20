// Copyright 2021 Cruise LLC
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

	networkingistiov1beta1 "istio.io/api/networking/v1beta1"
	admissionv1 "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

type VirtualService struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object's metadata.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// Spec defines the behavior of a service.
	// https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#spec-and-status
	// +optional
	Spec networkingistiov1beta1.VirtualService `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`
}

// DeepCopyInto is a deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VirtualService) DeepCopyInto(out *VirtualService) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	return
}

// DeepCopy is a deepcopy function, copying the receiver, creating a new Service.
func (in *VirtualService) DeepCopy() *VirtualService {
	if in == nil {
		return nil
	}
	out := new(VirtualService)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is a deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *VirtualService) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// ServiceResource contains the information needed for processing by a Policy
type VirtualServiceResource struct {
	VirtualService VirtualService
	ResourceName   string
	ResourceKind   string
}

// GetVirtualServiceResource extracts a VirtualServiceResource from an AdmissionRequest
func GetVirtualServiceResource(ctx context.Context, ar *admissionv1.AdmissionRequest) *VirtualServiceResource {
	c := GetResourceCache(ctx)
	return c.getOrSet(cacheKeyVirtualService, func() interface{} {
		return decodeVirtualServiceResource(ar)
	}).(*VirtualServiceResource)
}

func decodeVirtualServiceResource(ar *admissionv1.AdmissionRequest) *VirtualServiceResource {
	switch ar.Resource {
	case metav1.GroupVersionResource{Group: "networking.istio.io", Version: "v1beta1", Resource: "virtualservices"}, metav1.GroupVersionResource{Group: "networking.istio.io", Version: "v1alpha3", Resource: "virtualservices"}:
		vsvc := VirtualService{}
		if err := decodeObject(ar.Object.Raw, &vsvc); err != nil {
			return nil
		}
		return &VirtualServiceResource{
			VirtualService: vsvc,
			ResourceName:   GetResourceName(vsvc.ObjectMeta),
			ResourceKind:   "VirtualService",
		}
	default:
		return nil
	}
}
