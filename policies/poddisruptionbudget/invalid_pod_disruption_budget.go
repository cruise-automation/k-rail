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

package poddisruptionbudget

import (
	"context"
	"fmt"

	log "github.com/sirupsen/logrus"
	admissionv1 "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/cruise-automation/k-rail/policies"
	"github.com/cruise-automation/k-rail/resource"
)

// NewPolicyInvalidPodDisruptionBudget new
func NewPolicyInvalidPodDisruptionBudget() (PolicyInvalidPodDisruptionBudget, error) {
	p := PolicyInvalidPodDisruptionBudget{}

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

// PolicyInvalidPodDisruptionBudget type
type PolicyInvalidPodDisruptionBudget struct {
	client *kubernetes.Clientset
}

// Item item
type Item struct {
	Name     string
	Kind     string
	Replicas int
}

// Name name
func (p PolicyInvalidPodDisruptionBudget) Name() string {
	return "invalid_pod_disruption_budget"
}

// GetMatchingItems find all deployments, replicasets, and statefulsets which match the pdb labelselector
func (p PolicyInvalidPodDisruptionBudget) GetMatchingItems(namespace string, labelSelector *metav1.LabelSelector) ([]Item, error) {
	match := make([]Item, 0)

	labelMap, err := metav1.LabelSelectorAsMap(labelSelector)
	if err != nil {
		return match, err
	}

	options := metav1.ListOptions{
		LabelSelector: labels.SelectorFromSet(labelMap).String(),
	}

	// Check all types & versions, fail silently when a an api version does not exist
	// extensionsv1beta1/deployments
	if extsv1b1Deployments, err := p.client.ExtensionsV1beta1().Deployments(namespace).List(options); err == nil {
		for _, item := range extsv1b1Deployments.Items {
			do := Item{
				Name:     item.Name,
				Kind:     "deployment",
				Replicas: int(*item.Spec.Replicas),
			}
			match = append(match, do)
		}
	}

	// appsv1beta1/deployments
	if appsv1b1Deployments, err := p.client.AppsV1beta1().Deployments(namespace).List(options); err == nil {
		for _, item := range appsv1b1Deployments.Items {
			do := Item{
				Name:     item.Name,
				Kind:     "deployment",
				Replicas: int(*item.Spec.Replicas),
			}
			match = append(match, do)
		}
	}

	// appsv1beta2/deployments
	if appsv1b2Deployments, err := p.client.AppsV1beta2().Deployments(namespace).List(options); err == nil {
		for _, item := range appsv1b2Deployments.Items {
			do := Item{
				Name:     item.Name,
				Kind:     "deployment",
				Replicas: int(*item.Spec.Replicas),
			}
			match = append(match, do)
		}
	}

	// appsv1/deployments
	if appsV1Deployments, err := p.client.AppsV1().Deployments(namespace).List(options); err == nil {
		for _, item := range appsV1Deployments.Items {
			do := Item{
				Name:     item.Name,
				Kind:     "deployment",
				Replicas: int(*item.Spec.Replicas),
			}
			match = append(match, do)
		}
	}

	// extensionsv1beta1/replicasets
	if extsv1b1ReplicaSets, err := p.client.ExtensionsV1beta1().ReplicaSets(namespace).List(options); err == nil {
		for _, item := range extsv1b1ReplicaSets.Items {
			do := Item{
				Name:     item.Name,
				Kind:     "replicaset",
				Replicas: int(*item.Spec.Replicas),
			}
			match = append(match, do)
		}
	}

	// appsv1beta2/replicasets
	if appsv1b2ReplicaSets, err := p.client.AppsV1beta2().ReplicaSets(namespace).List(options); err == nil {
		for _, item := range appsv1b2ReplicaSets.Items {
			do := Item{
				Name:     item.Name,
				Kind:     "replicaset",
				Replicas: int(*item.Spec.Replicas),
			}
			match = append(match, do)
		}
	}

	// appsv1/replicasets
	if appsv1ReplicaSets, err := p.client.AppsV1().ReplicaSets(namespace).List(options); err == nil {
		for _, item := range appsv1ReplicaSets.Items {
			do := Item{
				Name:     item.Name,
				Kind:     "replicaset",
				Replicas: int(*item.Spec.Replicas),
			}
			match = append(match, do)
		}
	}

	// appsv1beta1/statefulsets
	if appsv1b1StatefulSets, err := p.client.AppsV1beta1().StatefulSets(namespace).List(options); err == nil {
		for _, item := range appsv1b1StatefulSets.Items {
			do := Item{
				Name:     item.Name,
				Kind:     "statefulset",
				Replicas: int(*item.Spec.Replicas),
			}
			match = append(match, do)
		}
	}

	// appsv1beta2/statefulsets
	if appsv1b2StatefulSets, err := p.client.AppsV1beta2().StatefulSets(namespace).List(options); err == nil {
		for _, item := range appsv1b2StatefulSets.Items {
			do := Item{
				Name:     item.Name,
				Kind:     "statefulset",
				Replicas: int(*item.Spec.Replicas),
			}
			match = append(match, do)
		}
	}

	// appsv1/statefulsets
	if appsv1StatefulSets, err := p.client.AppsV1().StatefulSets(namespace).List(options); err == nil {
		for _, item := range appsv1StatefulSets.Items {
			do := Item{
				Name:     item.Name,
				Kind:     "statefulset",
				Replicas: int(*item.Spec.Replicas),
			}
			match = append(match, do)
		}
	}

	return match, nil
}

// Validate poddisruptionbudget checks
func (p PolicyInvalidPodDisruptionBudget) Validate(ctx context.Context, config policies.Config, ar *admissionv1.AdmissionRequest) ([]policies.ResourceViolation, []policies.PatchOperation) {

	resourceViolations := []policies.ResourceViolation{}

	pdbResource := resource.GetPodDisruptionBudgetResource(ctx, ar)
	if pdbResource == nil {
		return resourceViolations, nil
	}

	minAvailable := pdbResource.PodDisruptionBudget.Spec.MinAvailable
	maxUnavailable := pdbResource.PodDisruptionBudget.Spec.MaxUnavailable

	// nothing to do
	if minAvailable == nil && maxUnavailable == nil {
		return resourceViolations, nil
	}

	// get a slice of all items matching the pdb labelselector
	items, err := p.GetMatchingItems(pdbResource.PodDisruptionBudget.Namespace, pdbResource.PodDisruptionBudget.Spec.Selector)
	if err != nil {
		log.Error(err)
		return nil, nil
	}
	// label selector doesn't match any items
	if len(items) == 0 {
		return resourceViolations, nil
	}

	// minAvailable set
	if minAvailable != nil {
		for _, item := range items {
			min, err := intstr.GetValueFromIntOrPercent(minAvailable, item.Replicas, true)
			if err != nil {
				log.Error(err)
				return nil, nil
			}
			// pass if pdb minAvailable == 0, valid though effectively useless configuration
			// pass if item replicas == 0, possible on replicasets
			// fail if minAvailable >= item replicas
			if min != 0 && item.Replicas != 0 && min >= item.Replicas {
				violationText := "Invalid Pod Disruption Budget: Minimum available pods must be less than the target replicas -"
				resourceViolations = append(resourceViolations, policies.ResourceViolation{
					Namespace:    ar.Namespace,
					ResourceName: pdbResource.PodDisruptionBudget.Name,
					ResourceKind: pdbResource.PodDisruptionBudget.Kind,
					Violation: fmt.Sprintf("%s PodDisruptionBudget minAvailable: %d, target %s/%s replicas: %d",
						violationText,
						min,
						item.Kind,
						item.Name,
						item.Replicas,
					),
					Policy: p.Name(),
				})
			}
		}
	}

	// maxUnavailable set
	if maxUnavailable != nil {
		for _, item := range items {
			max, err := intstr.GetValueFromIntOrPercent(maxUnavailable, item.Replicas, false)
			if err != nil {
				log.Error(err)
				return nil, nil
			}
			// fail if maxUnavailable < 1
			if max < 1 {
				violationText := "Invalid Pod Disruption Budget: Maximum unavailable pods must be greater than 1 -"
				resourceViolations = append(resourceViolations, policies.ResourceViolation{
					Namespace:    ar.Namespace,
					ResourceName: pdbResource.PodDisruptionBudget.Name,
					ResourceKind: pdbResource.PodDisruptionBudget.Kind,
					Violation: fmt.Sprintf("%s PodDisruptionBudget maxUnavailable: %d, target %s/%s replicas: %d",
						violationText,
						max,
						item.Kind,
						item.Name,
						item.Replicas,
					),
					Policy: p.Name(),
				})
			}
		}
	}

	return resourceViolations, nil
}
