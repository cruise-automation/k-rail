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

package resource

import (
	admissionv1beta1 "k8s.io/api/admission/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	appsv1beta1 "k8s.io/api/apps/v1beta1"
	appsv1beta2 "k8s.io/api/apps/v1beta2"
	batchv1 "k8s.io/api/batch/v1"
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	batchv2alpha1 "k8s.io/api/batch/v2alpha1"
	corev1 "k8s.io/api/core/v1"
	extensionsv1beta1 "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PodResource contains the information needed for processing by a Policy
type PodResource struct {
	PodSpec        corev1.PodSpec
	PodAnnotations map[string]string
	ResourceName   string
	ResourceKind   string
}

// GetPodResource extracts a PodResource from an AdmissionRequest
func GetPodResource(ar *admissionv1beta1.AdmissionRequest) *PodResource {

	switch ar.Resource {
	case metav1.GroupVersionResource{Group: "", Version: "v1", Resource: "pods"}:
		pod := corev1.Pod{}
		if err := decodeObject(ar.Object.Raw, &pod); err != nil {
			return nil
		}
		return &PodResource{
			PodSpec:        pod.Spec,
			PodAnnotations: pod.ObjectMeta.Annotations,
			ResourceName:   GetResourceName(pod.ObjectMeta),
			ResourceKind:   "Pod",
		}
	case metav1.GroupVersionResource{Group: "", Version: "v1", Resource: "replicationcontrollers"}:
		rc := corev1.ReplicationController{}
		if err := decodeObject(ar.Object.Raw, &rc); err != nil {
			return nil
		}
		return &PodResource{
			PodSpec:        rc.Spec.Template.Spec,
			PodAnnotations: rc.Spec.Template.ObjectMeta.Annotations,
			ResourceName:   GetResourceName(rc.ObjectMeta),
			ResourceKind:   "ReplicationController",
		}
	case metav1.GroupVersionResource{Group: "extensions", Version: "v1beta1", Resource: "deployments"}:
		dep := extensionsv1beta1.Deployment{}
		if err := decodeObject(ar.Object.Raw, &dep); err != nil {
			return nil
		}
		return &PodResource{
			PodSpec:        dep.Spec.Template.Spec,
			PodAnnotations: dep.Spec.Template.ObjectMeta.Annotations,
			ResourceName:   GetResourceName(dep.ObjectMeta),
			ResourceKind:   "Deployment",
		}
	case metav1.GroupVersionResource{Group: "apps", Version: "v1beta1", Resource: "deployments"}:
		dep := appsv1beta1.Deployment{}
		if err := decodeObject(ar.Object.Raw, &dep); err != nil {
			return nil
		}
		return &PodResource{
			PodSpec:        dep.Spec.Template.Spec,
			PodAnnotations: dep.Spec.Template.ObjectMeta.Annotations,
			ResourceName:   GetResourceName(dep.ObjectMeta),
			ResourceKind:   "Deployment",
		}
	case metav1.GroupVersionResource{Group: "apps", Version: "v1beta2", Resource: "deployments"}:
		dep := appsv1beta2.Deployment{}
		if err := decodeObject(ar.Object.Raw, &dep); err != nil {
			return nil
		}
		return &PodResource{
			PodSpec:        dep.Spec.Template.Spec,
			PodAnnotations: dep.Spec.Template.ObjectMeta.Annotations,
			ResourceName:   GetResourceName(dep.ObjectMeta),
			ResourceKind:   "Deployment",
		}
	case metav1.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"}:
		dep := appsv1.Deployment{}
		if err := decodeObject(ar.Object.Raw, &dep); err != nil {
			return nil
		}
		return &PodResource{
			PodSpec:        dep.Spec.Template.Spec,
			PodAnnotations: dep.Spec.Template.ObjectMeta.Annotations,
			ResourceName:   GetResourceName(dep.ObjectMeta),
			ResourceKind:   "Deployment",
		}
	case metav1.GroupVersionResource{Group: "apps", Version: "v1", Resource: "replicasets"}:
		rs := appsv1.ReplicaSet{}
		if err := decodeObject(ar.Object.Raw, &rs); err != nil {
			return nil
		}
		return &PodResource{
			PodSpec:        rs.Spec.Template.Spec,
			PodAnnotations: rs.Spec.Template.ObjectMeta.Annotations,
			ResourceName:   GetResourceName(rs.ObjectMeta),
			ResourceKind:   "ReplicaSet",
		}
	case metav1.GroupVersionResource{Group: "extensions", Version: "v1beta1", Resource: "replicasets"}:
		rs := extensionsv1beta1.ReplicaSet{}
		if err := decodeObject(ar.Object.Raw, &rs); err != nil {
			return nil
		}
		return &PodResource{
			PodSpec:        rs.Spec.Template.Spec,
			PodAnnotations: rs.Spec.Template.ObjectMeta.Annotations,
			ResourceName:   GetResourceName(rs.ObjectMeta),
			ResourceKind:   "ReplicaSet",
		}
	case metav1.GroupVersionResource{Group: "apps", Version: "v1beta2", Resource: "replicasets"}:
		rs := appsv1beta2.ReplicaSet{}
		if err := decodeObject(ar.Object.Raw, &rs); err != nil {
			return nil
		}
		return &PodResource{
			PodSpec:        rs.Spec.Template.Spec,
			PodAnnotations: rs.Spec.Template.ObjectMeta.Annotations,
			ResourceName:   GetResourceName(rs.ObjectMeta),
			ResourceKind:   "ReplicaSet",
		}
	case metav1.GroupVersionResource{Group: "apps", Version: "v1", Resource: "daemonsets"}:
		ds := appsv1.DaemonSet{}
		if err := decodeObject(ar.Object.Raw, &ds); err != nil {
			return nil
		}
		return &PodResource{
			PodSpec:        ds.Spec.Template.Spec,
			PodAnnotations: ds.Spec.Template.ObjectMeta.Annotations,
			ResourceName:   GetResourceName(ds.ObjectMeta),
			ResourceKind:   "DaemonSet",
		}
	case metav1.GroupVersionResource{Group: "extensions", Version: "v1beta1", Resource: "daemonsets"}:
		ds := extensionsv1beta1.DaemonSet{}
		if err := decodeObject(ar.Object.Raw, &ds); err != nil {
			return nil
		}
		return &PodResource{
			PodSpec:        ds.Spec.Template.Spec,
			PodAnnotations: ds.Spec.Template.ObjectMeta.Annotations,
			ResourceName:   GetResourceName(ds.ObjectMeta),
			ResourceKind:   "DaemonSet",
		}
	case metav1.GroupVersionResource{Group: "apps", Version: "v1beta2", Resource: "daemonsets"}:
		ds := appsv1beta2.DaemonSet{}
		if err := decodeObject(ar.Object.Raw, &ds); err != nil {
			return nil
		}
		return &PodResource{
			PodSpec:        ds.Spec.Template.Spec,
			PodAnnotations: ds.Spec.Template.ObjectMeta.Annotations,
			ResourceName:   GetResourceName(ds.ObjectMeta),
			ResourceKind:   "DaemonSet",
		}
	case metav1.GroupVersionResource{Group: "apps", Version: "v1", Resource: "statefulsets"}:
		ss := appsv1.StatefulSet{}
		if err := decodeObject(ar.Object.Raw, &ss); err != nil {
			return nil
		}
		return &PodResource{
			PodSpec:        ss.Spec.Template.Spec,
			PodAnnotations: ss.Spec.Template.ObjectMeta.Annotations,
			ResourceName:   GetResourceName(ss.ObjectMeta),
			ResourceKind:   "StatefulSet",
		}
	case metav1.GroupVersionResource{Group: "apps", Version: "v1beta1", Resource: "statefulsets"}:
		ss := appsv1beta1.StatefulSet{}
		if err := decodeObject(ar.Object.Raw, &ss); err != nil {
			return nil
		}
		return &PodResource{
			PodSpec:        ss.Spec.Template.Spec,
			PodAnnotations: ss.Spec.Template.ObjectMeta.Annotations,
			ResourceName:   GetResourceName(ss.ObjectMeta),
			ResourceKind:   "StatefulSet",
		}
	case metav1.GroupVersionResource{Group: "apps", Version: "v1beta2", Resource: "statefulsets"}:
		ss := appsv1beta2.StatefulSet{}
		if err := decodeObject(ar.Object.Raw, &ss); err != nil {
			return nil
		}
		return &PodResource{
			PodSpec:        ss.Spec.Template.Spec,
			PodAnnotations: ss.Spec.Template.ObjectMeta.Annotations,
			ResourceName:   GetResourceName(ss.ObjectMeta),
			ResourceKind:   "StatefulSet",
		}
	case metav1.GroupVersionResource{Group: "batch", Version: "v1", Resource: "jobs"}:
		job := batchv1.Job{}
		if err := decodeObject(ar.Object.Raw, &job); err != nil {
			return nil
		}
		return &PodResource{
			PodSpec:        job.Spec.Template.Spec,
			PodAnnotations: job.Spec.Template.ObjectMeta.Annotations,
			ResourceName:   GetResourceName(job.ObjectMeta),
			ResourceKind:   "Job",
		}
	case metav1.GroupVersionResource{Group: "batch", Version: "v1beta1", Resource: "cronjobs"}:
		job := batchv1beta1.CronJob{}
		if err := decodeObject(ar.Object.Raw, &job); err != nil {
			return nil
		}
		return &PodResource{
			PodSpec:        job.Spec.JobTemplate.Spec.Template.Spec,
			PodAnnotations: job.Spec.JobTemplate.ObjectMeta.Annotations,
			ResourceName:   GetResourceName(job.ObjectMeta),
			ResourceKind:   "CronJob",
		}
	case metav1.GroupVersionResource{Group: "batch", Version: "v2alpha1", Resource: "cronjobs"}:
		job := batchv2alpha1.CronJob{}
		if err := decodeObject(ar.Object.Raw, &job); err != nil {
			return nil
		}
		return &PodResource{
			PodSpec:        job.Spec.JobTemplate.Spec.Template.Spec,
			PodAnnotations: job.Spec.JobTemplate.ObjectMeta.Annotations,
			ResourceName:   GetResourceName(job.ObjectMeta),
			ResourceKind:   "CronJob",
		}
	default:
		return nil
	}
}
