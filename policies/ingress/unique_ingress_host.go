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

package ingress

import (
	"context"

	"github.com/cruise-automation/k-rail/v3/policies"
	"github.com/cruise-automation/k-rail/v3/resource"
	log "github.com/sirupsen/logrus"
	admissionv1 "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func NewPolicyRequireUniqueHost() (PolicyRequireUniqueHost, error) {
	p := PolicyRequireUniqueHost{}

	config, err := rest.InClusterConfig()
	if err != nil {
		return p, err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return p, err
	}
	p.client = clientset
	return p, err
}

type PolicyRequireUniqueHost struct {
	client *kubernetes.Clientset
}

func (p PolicyRequireUniqueHost) Name() string {
	return "ingress_unique_ingress_host"
}

func (p PolicyRequireUniqueHost) CheckIngressNamespaces(ctx context.Context, host string) (map[string]struct{}, error) {
	ingressNamespacesMap := make(map[string]struct{})
	ingresses, err := p.client.ExtensionsV1beta1().Ingresses("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return ingressNamespacesMap, err
	}
	for _, ingress := range ingresses.Items {
		rules := ingress.Spec.Rules
		for _, rule := range rules {
			if rule.Host == host {
				ingressNamespace := ingress.ObjectMeta.Namespace
				ingressNamespacesMap[ingressNamespace] = struct{}{}
			}
		}
	}
	return ingressNamespacesMap, nil
}

func Find(slice []string, val string) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}

func (p PolicyRequireUniqueHost) Validate(ctx context.Context, config policies.Config, ar *admissionv1.AdmissionRequest) ([]policies.ResourceViolation, []policies.PatchOperation) {

	resourceViolations := []policies.ResourceViolation{}

	ingressResource := resource.GetIngressResource(ctx, ar)
	if ingressResource == nil {
		return resourceViolations, nil
	}

	violationText := "Requires Unique Ingress Host: Ingress Host should not point to multiple namespaces. Host already in:"

	for _, rule := range ingressResource.IngressExt.Spec.Rules {
		ingressNamespacesMap, err := p.CheckIngressNamespaces(ctx, rule.Host)
		if err != nil {
			log.Error(err)
			return nil, nil
		}
		foundNamespace := false
		_, ok := ingressNamespacesMap[ar.Namespace]
		if ok {
			foundNamespace = true
		}
		if (len(ingressNamespacesMap) == 0) || (len(ingressNamespacesMap) == 1 && foundNamespace) {
			return resourceViolations, nil
		} else {
			namespacesStr := ""
			for k := range ingressNamespacesMap {
				namespacesStr = namespacesStr + " " + k
			}
			resourceViolations = append(resourceViolations, policies.ResourceViolation{
				Namespace:    ar.Namespace,
				ResourceName: ingressResource.ResourceName,
				ResourceKind: ingressResource.ResourceKind,
				Violation:    violationText + ": " + namespacesStr,
				Policy:       p.Name(),
			})
		}
	}

	for _, rule := range ingressResource.IngressNet.Spec.Rules {
		ingressNamespacesMap, err := p.CheckIngressNamespaces(ctx, rule.Host)
		if err != nil {
			log.Error(err)
			return nil, nil
		}
		foundNamespace := false
		_, ok := ingressNamespacesMap[ar.Namespace]
		if ok {
			foundNamespace = true
		}
		if (len(ingressNamespacesMap) == 0) || (len(ingressNamespacesMap) == 1 && foundNamespace) {
			return resourceViolations, nil
		} else {
			namespacesStr := ""
			for k := range ingressNamespacesMap {
				namespacesStr = namespacesStr + " " + k
			}
			resourceViolations = append(resourceViolations, policies.ResourceViolation{
				Namespace:    ar.Namespace,
				ResourceName: ingressResource.ResourceName,
				ResourceKind: ingressResource.ResourceKind,
				Violation:    violationText + ": " + namespacesStr,
				Policy:       p.Name(),
			})
		}
	}
	return resourceViolations, nil
}
