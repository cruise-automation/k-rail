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
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ClusterRoleBindingResource contains the information needed for processing by a Policy
type ClusterRoleBindingResource struct {
	ClusterRoleBinding rbacv1.ClusterRoleBinding
	ResourceName       string
	ResourceKind       string
}

// GetClusterRoleBindingResource extracts a ClusterRoleBindingResource from an AdmissionRequest
func GetClusterRoleBindingResource(ctx context.Context, ar *admissionv1.AdmissionRequest) *ClusterRoleBindingResource {
	c := GetResourceCache(ctx)
	return c.getOrSet(cacheKeyClusterRoleBinding, func() interface{} {
		return decodeClusterRoleBindingResource(ar)
	}).(*ClusterRoleBindingResource)
}

func decodeClusterRoleBindingResource(ar *admissionv1.AdmissionRequest) *ClusterRoleBindingResource {
	switch ar.Kind {
	case metav1.GroupVersionKind{Group: "rbac.authorization.k8s.io", Version: "v1", Kind: "ClusterRoleBinding"}:
		crb := rbacv1.ClusterRoleBinding{}
		if err := decodeObject(ar.Object.Raw, &crb); err != nil {
			return nil
		}
		return &ClusterRoleBindingResource{
			ClusterRoleBinding: crb,
			ResourceName:       GetResourceName(crb.ObjectMeta),
			ResourceKind:       "ClusterRoleBinding",
		}
	default:
		return nil
	}
}
