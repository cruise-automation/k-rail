package pod

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/cruise-automation/k-rail/v3/policies"
	admissionv1 "k8s.io/api/admission/v1"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	apiresource "k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestEmptyDirSizeLimit(t *testing.T) {
	config := policies.Config{
		MutateEmptyDirSizeLimit: policies.MutateEmptyDirSizeLimit{
			DefaultSizeLimit: *apiresource.NewQuantity(1, apiresource.DecimalSI),
			MaximumSizeLimit: *apiresource.NewQuantity(10, apiresource.DecimalSI),
		},
	}

	specs := map[string]struct {
		src           v1.PodSpec
		expViolations []policies.ResourceViolation
		expPatches    []policies.PatchOperation
	}{
		"limit set within range": {
			src: v1.PodSpec{
				Volumes: []v1.Volume{{
					VolumeSource: v1.VolumeSource{
						EmptyDir: &v1.EmptyDirVolumeSource{
							SizeLimit: apiresource.NewQuantity(2, apiresource.DecimalSI)},
					},
				}},
			},
		},
		"limit set within range with multiple volumes": {
			src: v1.PodSpec{
				Volumes: []v1.Volume{{
					VolumeSource: v1.VolumeSource{
						EmptyDir: &v1.EmptyDirVolumeSource{
							SizeLimit: apiresource.NewQuantity(2, apiresource.DecimalSI)},
					},
				}, {
					VolumeSource: v1.VolumeSource{
						EmptyDir: &v1.EmptyDirVolumeSource{
							SizeLimit: apiresource.NewQuantity(3, apiresource.DecimalSI)},
					},
				}},
			},
		},
		"set default value when 0": {
			src: v1.PodSpec{
				Volumes: []v1.Volume{{
					VolumeSource: v1.VolumeSource{
						EmptyDir: &v1.EmptyDirVolumeSource{
							SizeLimit: apiresource.NewQuantity(0, apiresource.DecimalExponent),
						},
					}},
				},
			},
			expPatches: []policies.PatchOperation{
				{
					Path:  "/spec/template/spec/volumes/0/emptyDir/sizeLimit",
					Op:    "replace",
					Value: "1",
				},
			},
		}, "set default value when empty": {
			src: v1.PodSpec{
				Volumes: []v1.Volume{{
					VolumeSource: v1.VolumeSource{
						EmptyDir: &v1.EmptyDirVolumeSource{},
					}},
				},
			},
			expPatches: []policies.PatchOperation{
				{
					Path:  "/spec/template/spec/volumes/0/emptyDir/sizeLimit",
					Op:    "replace",
					Value: "1",
				},
			},
		},
		"set default value when empty with multiple": {
			src: v1.PodSpec{
				Volumes: []v1.Volume{{
					VolumeSource: v1.VolumeSource{
						EmptyDir: &v1.EmptyDirVolumeSource{},
					}}, {
					VolumeSource: v1.VolumeSource{
						EmptyDir: &v1.EmptyDirVolumeSource{},
					}},
				},
			},
			expPatches: []policies.PatchOperation{
				{
					Path:  "/spec/template/spec/volumes/0/emptyDir/sizeLimit",
					Op:    "replace",
					Value: "1",
				},
				{
					Path:  "/spec/template/spec/volumes/1/emptyDir/sizeLimit",
					Op:    "replace",
					Value: "1",
				},
			},
		},
		"allow max limit size": {
			src: v1.PodSpec{
				Volumes: []v1.Volume{{
					VolumeSource: v1.VolumeSource{
						EmptyDir: &v1.EmptyDirVolumeSource{
							SizeLimit: apiresource.NewQuantity(10, apiresource.DecimalSI),
						},
					}},
				},
			},
		},
		"prevent greater than max limit size": {
			src: v1.PodSpec{
				Volumes: []v1.Volume{{
					VolumeSource: v1.VolumeSource{
						EmptyDir: &v1.EmptyDirVolumeSource{
							SizeLimit: apiresource.NewQuantity(11, apiresource.DecimalSI),
						},
					}},
				},
			},
			expViolations: []policies.ResourceViolation{
				{
					ResourceName: "test",
					ResourceKind: "Deployment",
					Namespace:    "test",
					Violation:    "Empty dir size limit: size limit exceeds the max value",
					Policy:       "pod_empty_dir_size_limit",
				},
			},
		},
		"skip non empty dir volume": {
			src: v1.PodSpec{
				Volumes: []v1.Volume{{
					VolumeSource: v1.VolumeSource{
						HostPath: &v1.HostPathVolumeSource{},
					}},
				},
			},
		},
	}
	for msg, spec := range specs {
		t.Run(msg, func(t *testing.T) {
			policy := PolicyEmptyDirSizeLimit{}
			v, p := policy.Validate(context.TODO(), config, asFakeAdmissionRequest(spec.src))
			if exp, got := spec.expViolations, v; !reflect.DeepEqual(exp, got) {
				t.Errorf("expected %#v but got %#v", exp, got)
			}
			if exp, got := spec.expPatches, p; !reflect.DeepEqual(exp, got) {
				t.Errorf("expected %#v but got %#v", exp, got)
			}

		})
	}
}

func TestResourceVolumePatchPath(t *testing.T) {
	specs := map[string]string{
		"Pod":                   "/spec/volumes",
		"ReplicationController": "/spec/template/spec/volumes",
		"Deployment":            "/spec/template/spec/volumes",
		"ReplicaSet":            "/spec/template/spec/volumes",
		"DaemonSet":             "/spec/template/spec/volumes",
		"StatefulSet":           "/spec/template/spec/volumes",
		"Job":                   "/spec/template/spec/volumes",
		"CronJob":               "/spec/jobTemplate/spec/template/spec/volumes",
	}
	for kind, exp := range specs {
		t.Run(kind, func(t *testing.T) {
			got := volumePatchPath(kind)
			if exp != got {
				t.Errorf("expected %q but got %q", exp, got)
			}
		})
	}

}

func asFakeAdmissionRequest(src v1.PodSpec) *admissionv1.AdmissionRequest {
	xxx := appsv1.Deployment{
		TypeMeta:   metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{Name: "test"},
		Spec: appsv1.DeploymentSpec{
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{},
				Spec:       src,
			},
		},
		Status: appsv1.DeploymentStatus{},
	}
	b, err := json.Marshal(&xxx)
	if err != nil {
		panic(err)
	}
	return &admissionv1.AdmissionRequest{
		Resource:  metav1.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"},
		Name:      "any",
		Namespace: "test",
		Object:    runtime.RawExtension{Raw: b},
	}
}
