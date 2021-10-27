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
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	admissionv1 "k8s.io/api/admission/v1"
	extensionsv1beta1 "k8s.io/api/extensions/v1beta1"
	networkingv1 "k8s.io/api/networking/v1"
	networkingv1beta1 "k8s.io/api/networking/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func TestGetAnnotations(t *testing.T) {
	for _, tt := range []struct {
		name    string
		ingress interface {
			GetObjectKind() schema.ObjectKind
		}
		wantAnnotations map[string]string
	}{
		{
			ingress: &extensionsv1beta1.Ingress{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "extensions/v1beta1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"foo": "bar",
					},
				},
			},
			wantAnnotations: map[string]string{
				"foo": "bar",
			},
		},
		{
			ingress: &networkingv1beta1.Ingress{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "networking/v1beta1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"foo": "bar",
					},
				},
			},
			wantAnnotations: map[string]string{
				"foo": "bar",
			},
		},
		{
			ingress: &networkingv1.Ingress{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "networking/v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"foo": "bar",
					},
				},
			},
			wantAnnotations: map[string]string{
				"foo": "bar",
			},
		},
	} {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			raw, _ := json.Marshal(tt.ingress)
			ar := &admissionv1.AdmissionRequest{
				Namespace: "namespace",
				Name:      "name",
				Object:    runtime.RawExtension{Raw: raw},
				Resource: metav1.GroupVersionResource{
					Group:    tt.ingress.GetObjectKind().GroupVersionKind().Group,
					Version:  tt.ingress.GetObjectKind().GroupVersionKind().Version,
					Resource: "ingresses",
				},
			}
			ingressResource := decodeIngressResource(ar)
			assert.Equal(t, tt.wantAnnotations, ingressResource.GetAnnotations())
		})
	}
}

func TestGetHosts(t *testing.T) {
	for _, tt := range []struct {
		name    string
		ingress interface {
			GetObjectKind() schema.ObjectKind
		}
		wantHosts []string
	}{
		{
			ingress: &extensionsv1beta1.Ingress{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "extensions/v1beta1",
				},
				Spec: extensionsv1beta1.IngressSpec{
					Rules: []extensionsv1beta1.IngressRule{
						{Host: "foo"},
						{Host: "bar"},
					},
				},
			},
			wantHosts: []string{"foo", "bar"},
		},
		{
			ingress: &networkingv1beta1.Ingress{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "networking/v1beta1",
				},
				Spec: networkingv1beta1.IngressSpec{
					Rules: []networkingv1beta1.IngressRule{
						{Host: "foo"},
						{Host: "bar"},
					},
				},
			},
			wantHosts: []string{"foo", "bar"},
		},
		{
			ingress: &networkingv1.Ingress{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "networking/v1",
				},
				Spec: networkingv1.IngressSpec{
					Rules: []networkingv1.IngressRule{
						{Host: "foo"},
						{Host: "bar"},
					},
				},
			},
			wantHosts: []string{"foo", "bar"},
		},
	} {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			raw, _ := json.Marshal(tt.ingress)
			ar := &admissionv1.AdmissionRequest{
				Namespace: "namespace",
				Name:      "name",
				Object:    runtime.RawExtension{Raw: raw},
				Resource: metav1.GroupVersionResource{
					Group:    tt.ingress.GetObjectKind().GroupVersionKind().Group,
					Version:  tt.ingress.GetObjectKind().GroupVersionKind().Version,
					Resource: "ingresses",
				},
			}
			ingressResource := GetIngressResource(context.Background(), ar)
			assert.Equal(t, tt.wantHosts, ingressResource.GetHosts())
		})
	}
}
