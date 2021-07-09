package pod

import (
	"context"
	"encoding/json"
	"github.com/cruise-automation/k-rail/v3/policies"
	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/pointer"
	"reflect"
	"testing"
)

func TestPolicyNoRoot(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name       string
		podSpec    v1.PodSpec
		violations int
	}{
		{
			name:       "empty securityContext",
			violations: 3,
			podSpec: v1.PodSpec{
				SecurityContext: &v1.PodSecurityContext{},
				Containers: []v1.Container{
					{
						SecurityContext: &v1.SecurityContext{},
					},
				},
				InitContainers: []v1.Container{
					{
						SecurityContext: &v1.SecurityContext{},
					},
				},
			},
		},
		{
			name:       "runAsRoot",
			violations: 1,
			podSpec: v1.PodSpec{
				SecurityContext: &v1.PodSecurityContext{
					RunAsUser: pointer.Int64Ptr(0),
				},
			},
		},
		{
			name:       "runAsUser",
			violations: 0,
			podSpec: v1.PodSpec{
				SecurityContext: &v1.PodSecurityContext{
					RunAsUser: pointer.Int64Ptr(1001),
				},
			},
		},
		{
			name:       "runAsNonRoot",
			violations: 0,
			podSpec: v1.PodSpec{
				SecurityContext: &v1.PodSecurityContext{
					RunAsNonRoot: pointer.BoolPtr(true),
				},
			},
		},
		{
			name:       "runAsUser container",
			violations: 0,
			podSpec: v1.PodSpec{
				SecurityContext: &v1.PodSecurityContext{
					RunAsNonRoot: pointer.BoolPtr(true),
				},
				Containers: []v1.Container{
					{
						SecurityContext: &v1.SecurityContext{
							RunAsUser: pointer.Int64Ptr(1001),
						},
					},
				},
			},
		},
		{
			name:       "no pod context, but containers are set",
			violations: 0,
			podSpec: v1.PodSpec{
				SecurityContext: nil,
				Containers: []v1.Container{
					{
						SecurityContext: &v1.SecurityContext{
							RunAsNonRoot: pointer.BoolPtr(true),
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			raw, _ := json.Marshal(corev1.Pod{Spec: tt.podSpec})
			ar := &admissionv1.AdmissionRequest{
				Namespace: "namespace",
				Name:      "name",
				Object:    runtime.RawExtension{Raw: raw},
				Resource:  metav1.GroupVersionResource{Group: "", Version: "v1", Resource: "pods"},
			}
			v := PolicyNoRootUser{}
			if got, _ := v.Validate(ctx, policies.Config{}, ar); !reflect.DeepEqual(len(got), tt.violations) {
				t.Errorf("PolicyNoRootUser() %s got %v want %v violations", tt.name, len(got), tt.violations)
			}
		})
	}
}
