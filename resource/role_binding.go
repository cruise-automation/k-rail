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
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// RoleBindingResource contains the information needed for processing by a Policy
type RoleBindingResource struct {
	RoleBinding  rbacv1.RoleBinding
	ResourceName string
	ResourceKind string
}

// GetRoleBindingResource extracts a RoleBindingResource from an AdmissionRequest
func GetRoleBindingResource(ctx context.Context, ar *admissionv1beta1.AdmissionRequest) *RoleBindingResource {
	c := GetResourceCache(ctx)
	return c.getOrSet(cacheKeyRoleBinding, func() interface{} {
		return decodeRoleBindingResource(ar)
	}).(*RoleBindingResource)
}

func decodeRoleBindingResource(ar *admissionv1beta1.AdmissionRequest) *RoleBindingResource {
	switch ar.Kind {
	case metav1.GroupVersionKind{Group: "rbac.authorization.k8s.io", Version: "v1", Kind: "RoleBinding"}:
		rb := rbacv1.RoleBinding{}
		if err := decodeObject(ar.Object.Raw, &rb); err != nil {
			return nil
		}
		return &RoleBindingResource{
			RoleBinding:  rb,
			ResourceName: GetResourceName(rb.ObjectMeta),
			ResourceKind: "RoleBinding",
		}
	default:
		return nil
	}
}
