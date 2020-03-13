package resource

import (
	"context"
	"testing"

	admissionv1beta1 "k8s.io/api/admission/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func BenchmarkDecodePodWithoutCaching(b *testing.B) {
	req := fakeReq([]byte(podExample))
	ctx := context.TODO()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = GetPodResource(ctx, req)
	}
}

func BenchmarkDecodePodCaching(b *testing.B) {
	req := fakeReq([]byte(podExample))
	ctx := WithResourceCache(context.TODO())
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = GetPodResource(ctx, req)
	}
}

func TestWithPodExec(t *testing.T) {
	req := fakePodExecReq(nil)
	ctx := WithResourceCache(context.TODO())
	res := GetPodResource(ctx, req)
	if res != nil {
		t.Fatal("should have gotten nil for pod exec request")
	}
}

func fakePodExecReq(b []byte) *admissionv1beta1.AdmissionRequest {
	return &admissionv1beta1.AdmissionRequest{
		Kind:      metav1.GroupVersionKind{Group: "", Version: "v1", Kind: "PodExecOptions"},
		Name:      "any",
		Namespace: "test",
		Object:    runtime.RawExtension{Raw: b},
	}
}

func fakeReq(b []byte) *admissionv1beta1.AdmissionRequest {
	return &admissionv1beta1.AdmissionRequest{
		Resource:  metav1.GroupVersionResource{Group: "", Version: "v1", Resource: "pods"},
		Name:      "any",
		Namespace: "test",
		Object:    runtime.RawExtension{Raw: b},
	}
}

const podExample = `
kind: Pod
apiVersion: apps/v1
metadata:
  name: foobar
  namespace: testing
  annotations:
    created-by: alpe
  labels:
    app: foo
spec:
  containers:
  - name: foo
    image: v0.1.0
    ports:
    - name: http
      containerPort: 8080
`
