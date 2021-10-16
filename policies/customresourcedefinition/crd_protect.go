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

package customresourcedefinition

import (
	"context"

	admissionv1 "k8s.io/api/admission/v1"
	apiextensionv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"

	"github.com/cruise-automation/k-rail/v3/policies"
)

// NewPolicyCRDProtect new
func NewPolicyCRDProtect() (PolicyCRDProtect, error) {
	p := PolicyCRDProtect{}

	restConfig, err := rest.InClusterConfig()
	if err != nil {
		return p, err
	}
	config := dynamic.ConfigFor(restConfig)
	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return p, err
	}
	p.client = dynamicClient

	return p, nil
}

// PolicyCRDProtect type
type PolicyCRDProtect struct {
	client dynamic.Interface
}

// Name name
func (p PolicyCRDProtect) Name() string {
	return "crd_protect"
}

// Validate CRD Resources
func (p PolicyCRDProtect) Validate(ctx context.Context, config policies.Config, ar *admissionv1.AdmissionRequest) ([]policies.ResourceViolation, []policies.PatchOperation) {

	resourceViolations := []policies.ResourceViolation{}

	// check if the target resource is a CRD
	if ar.Kind.Kind != "CustomResourceDefinition" && ar.Kind.Group != "apiextensions.k8s.io" {
		return resourceViolations, nil
	}

	// check if this is a delete operation
	if ar.Operation != "DELETE" {
		return resourceViolations, nil
	}

	// check if protect annotation is set on the CRD
	crd := &apiextensionv1beta1.CustomResourceDefinition{}

	scheme := runtime.NewScheme()
	codecFactory := serializer.NewCodecFactory(scheme)
	deserializer := codecFactory.UniversalDeserializer()
	if _, _, err := deserializer.Decode(ar.OldObject.Raw, nil, crd); err != nil {
		return resourceViolations, nil
	}

	if v, ok := crd.ObjectMeta.Annotations["k-rail.crd.protect"]; ok {
		if v != "enabled" {
			return resourceViolations, nil
		}
	} else {
		return resourceViolations, nil
	}

	// check if any CRs exist for CRD
	for _, v := range crd.Spec.Versions {
		var cr *unstructured.UnstructuredList
		var err error
		crgvr := schema.GroupVersionResource{
			Group:    crd.Spec.Group,
			Version:  v.Name,
			Resource: crd.Spec.Names.Plural,
		}

		if crd.Spec.Scope == "Cluster" {
			cr, err = p.client.Resource(crgvr).List(ctx, metav1.ListOptions{})
		} else {
			cr, err = p.client.Resource(crgvr).Namespace("").List(ctx, metav1.ListOptions{})
		}
		if err != nil {
			return resourceViolations, nil
		}
		if len(cr.Items) > 0 {
			resourceViolations = append(
				resourceViolations, policies.ResourceViolation{
					Namespace:    ar.Namespace,
					ResourceName: crd.Name,
					ResourceKind: crd.Spec.Names.Kind,
					Violation:    "Can not delete custom resource definition (CRD) while custom resources (CRs) exist",
					Policy:       p.Name(),
				},
			)
			return resourceViolations, nil
		}
	}

	return resourceViolations, nil
}
