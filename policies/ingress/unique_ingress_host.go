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

package ingress

import (
	"context"
	"flag"
	"log"

	"github.com/cruise-automation/k-rail/policies"
	"github.com/cruise-automation/k-rail/resource"
	admissionv1beta1 "k8s.io/api/admission/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

//Helper initialization function (create object, create clientset, and return new struct PolicyRequireUniqueHost)
func NewPolicyRequireUniqueHost() PolicyRequireUniqueHost {

	p := PolicyRequireUniqueHost{}

	//initialize clientset as *Clientset
	var kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file: `<home>/.kube/config`")
	flag.Parse()
	flag.Set("logtostderr", "true") // glog: no disk log

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig) //builds configs from a master url or a kubeconfig filepath. These are passed in as command line flags for cluster components
	if err != nil {
		log.Fatal(err)
	}

	clientset, err := kubernetes.NewForConfig(config) //Creates a new *Clientset for the given config.
	if err != nil {
		log.Fatal(err)
	}
	p.client = clientset
	return p
	//TODO: will need to query K8 with K8 SA and create K8 role-binding
}

type PolicyRequireUniqueHost struct {
	client *kubernetes.Clientset //clientset with pointer
}

//Method that implements Policy interface's Name() method
func (p PolicyRequireUniqueHost) Name() string {
	return "ingress_require_unique_host"
}

//Method that queries K8s cluster for all Ingress configs with given host, returns namespaces for each ingress host
func (p PolicyRequireUniqueHost) CheckIngresses(ctx context.Context, host string) ([]string, error) {
	ingressNamespaces := []string{}
	ingresses, err := p.client.ExtensionsV1beta1().Ingresses("").List(metav1.ListOptions{}) //Returns *v1beta1.IngressList, and catches error find out what package its from and import; https://godoc.org/k8s.io/client-go/kubernetes/typed/extensions/v1beta1#ExtensionsV1beta1Interface
	if err != nil {
		return ingressNamespaces, err
	}
	for _, ingress := range ingresses.Items {
		rules := ingress.Spec.Rules
		for _, rule := range rules {
			if rule.Host == host {
				ingressNamespace := ingress.ObjectMeta.Namespace
				ingressNamespaces = append(ingressNamespaces, ingressNamespace)
			}
		}
	}
	return ingressNamespaces, nil
}

//Helper function to find if string in a slice
func Find(slice []string, val string) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}

//Method that implements Policy interface's Validate() method; given context, config policies, and admission request, return ResourceViolation slice and PatchOperation slice
func (p PolicyRequireUniqueHost) Validate(ctx context.Context, config policies.Config, ar *admissionv1beta1.AdmissionRequest) ([]policies.ResourceViolation, []policies.PatchOperation) {

	resourceViolations := []policies.ResourceViolation{}

	ingressResource := resource.GetIngressResource(ctx, ar)
	if ingressResource == nil {
		return resourceViolations, nil
	}

	violationText := "Requires Unique Ingress Host: Ingress Host should not point to multiple namespaces"

	//--------------------------------INGRESSEXT----------------------------------
	for _, rule := range ingressResource.IngressExt.Spec.Rules {
		namespaces, err := p.CheckIngresses(ctx, rule.Host)
		if err != nil {
			panic(err)
		}
		if len(namespaces) != 0 {
			foundNamespace := Find(namespaces, ar.Namespace)
			if !foundNamespace {
				resourceViolations = append(resourceViolations, policies.ResourceViolation{
					Namespace:    ar.Namespace,
					ResourceName: ingressResource.ResourceName,
					ResourceKind: ingressResource.ResourceKind,
					Violation:    violationText,
					Policy:       p.Name(),
				})
			}
		}
	}

	//--------------------------------INGRESSNET----------------------------------
	for _, rule := range ingressResource.IngressNet.Spec.Rules {
		namespaces, err := p.CheckIngresses(ctx, rule.Host)
		if err != nil {
			panic(err)
		}
		if len(namespaces) != 0 {
			foundNamespace := Find(namespaces, ar.Namespace)
			if !foundNamespace {
				resourceViolations = append(resourceViolations, policies.ResourceViolation{
					Namespace:    ar.Namespace,
					ResourceName: ingressResource.ResourceName,
					ResourceKind: ingressResource.ResourceKind,
					Violation:    violationText,
					Policy:       p.Name(),
				})
			}
		}
	}
	return resourceViolations, nil
}
