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

package policies

import (
	"testing"

	authenticationv1 "k8s.io/api/authentication/v1"
)

func TestIsExempt(t *testing.T) {
	type args struct {
		resourceName string
		namespace    string
		userInfo     authenticationv1.UserInfo
		policyName   string
		exemption    RawExemption
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "exact match",
			want: true,
			args: args{
				resourceName: "test-pod",
				namespace:    "test-namespace",
				userInfo: authenticationv1.UserInfo{
					Username: "test-user",
					Groups:   []string{"group-a", "group-b"},
				},
				policyName: "test-policy",
				exemption: RawExemption{
					ResourceName:   "test-pod",
					Namespace:      "test-namespace",
					Username:       "test-user",
					Group:          "group-a",
					ExemptPolicies: []string{"test-policy"},
				},
			},
		},
		{
			name: "fail policy",
			want: false,
			args: args{
				resourceName: "test-pod",
				namespace:    "test-namespace",
				userInfo: authenticationv1.UserInfo{
					Username: "test-user",
					Groups:   []string{"group-a", "group-b"},
				},
				policyName: "test-policy",
				exemption: RawExemption{
					ResourceName:   "test-pod",
					Namespace:      "test-namespace",
					Username:       "test-user",
					Group:          "group-a",
					ExemptPolicies: []string{"different-policy"},
				},
			},
		},
		{
			name: "all globs",
			want: true,
			args: args{
				resourceName: "test-pod",
				namespace:    "test-namespace",
				userInfo: authenticationv1.UserInfo{
					Username: "test-user",
					Groups:   []string{"group-a", "group-b"},
				},
				policyName: "test-policy",
				exemption: RawExemption{
					ResourceName:   "test-*",
					Namespace:      "test-*",
					Username:       "test-*",
					Group:          "group-*",
					ExemptPolicies: []string{"test-*"},
				},
			},
		},
		{
			name: "empty field is assumed passing during validation",
			want: true,
			args: args{
				resourceName: "test-pod",
				namespace:    "test-namespace",
				userInfo: authenticationv1.UserInfo{
					Username: "test-user",
					Groups:   []string{"group-a", "group-b"},
				},
				policyName: "test-policy",
				exemption:  RawExemption{},
			},
		},
		{
			name: "fail deepest condition",
			want: false,
			args: args{
				resourceName: "test-pod",
				namespace:    "test-namespace",
				userInfo: authenticationv1.UserInfo{
					Username: "test-user",
					Groups:   []string{"group-a", "group-b"},
				},
				policyName: "test-policy",
				exemption: RawExemption{
					Group: "fail-group",
				},
			},
		},
		{
			name: "implied resource glob",
			want: true,
			args: args{
				resourceName: "test-pod-dsfd32-sduf8ds",
				namespace:    "test-namespace",
				userInfo: authenticationv1.UserInfo{
					Username: "test-user",
					Groups:   []string{"group-a", "group-b"},
				},
				policyName: "test-policy",
				exemption: RawExemption{
					ResourceName: "test-pod",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsExempt(
				tt.args.resourceName,
				tt.args.namespace,
				tt.args.userInfo,
				tt.args.policyName,
				[]CompiledExemption{tt.args.exemption.Compile()},
			); got != tt.want {
				t.Errorf("IsExempt() = %v, want %v", got, tt.want)
			}
		})
	}
}
