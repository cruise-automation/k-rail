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
	"testing"

	corev1 "k8s.io/api/core/v1"
)

func TestGetResourceName(t *testing.T) {
	tests := []struct {
		name     string
		json     string
		wantName string
	}{
		{
			name: "owner ref",
			json: `
			{
				"kind": "Pod",
				"apiVersion": "v1",
				"metadata": {
					"generateName": "istio-cni-node-",
					"creationTimestamp": null,
					"labels": {
						"controller-revision-hash": "76db549475",
						"k8s-app": "istio-cni-node",
						"pod-template-generation": "1"
					},
					"annotations": {
						"kubernetes.io/psp": "istio-cni-node",
						"scheduler.alpha.kubernetes.io/critical-pod": "",
						"seccomp.security.alpha.kubernetes.io/pod": "runtime/default",
						"sidecar.istio.io/inject": "false"
					},
					"ownerReferences": [
						{
							"apiVersion": "apps/v1",
							"kind": "DaemonSet",
							"name": "istio-cni-node",
							"uid": "d6c11e2a-c5eb-4569-bb77-8ee35a83b169",
							"controller": true,
							"blockOwnerDeletion": true
						}
					]
				},
				"spec": {
					"volumes": [
						{
							"name": "cni-bin-dir",
							"hostPath": {
								"path": "/opt/cni/bin",
								"type": ""
							}
						},
						{
							"name": "cni-net-dir",
							"hostPath": {
								"path": "/etc/cni/net.d",
								"type": ""
							}
						},
						{
							"name": "istio-cni-token-s6fzj",
							"secret": {
								"secretName": "istio-cni-token-s6fzj"
							}
						}
					],
					"containers": [
						{
							"name": "install-cni",
							"image": "registry-internal.elpenguino.net/library/istio-install-cni:1.3.3",
							"command": [
								"/install-cni.sh"
							],
							"env": [
								{
									"name": "CNI_NETWORK_CONFIG",
									"valueFrom": {
										"configMapKeyRef": {
											"name": "istio-cni-config",
											"key": "cni_network_config"
										}
									}
								}
							],
							"resources": {},
							"volumeMounts": [
								{
									"name": "cni-bin-dir",
									"mountPath": "/host/opt/cni/bin"
								},
								{
									"name": "cni-net-dir",
									"mountPath": "/host/etc/cni/net.d"
								},
								{
									"name": "istio-cni-token-s6fzj",
									"readOnly": true,
									"mountPath": "/var/run/secrets/kubernetes.io/serviceaccount"
								}
							],
							"terminationMessagePath": "/dev/termination-log",
							"terminationMessagePolicy": "File",
							"imagePullPolicy": "IfNotPresent",
							"securityContext": {
								"capabilities": {
									"drop": [
										"ALL"
									]
								},
								"allowPrivilegeEscalation": false
							}
						}
					],
					"restartPolicy": "Always",
					"terminationGracePeriodSeconds": 5,
					"dnsPolicy": "ClusterFirst",
					"nodeSelector": {
						"beta.kubernetes.io/os": "linux"
					},
					"serviceAccountName": "istio-cni",
					"serviceAccount": "istio-cni",
					"hostNetwork": true,
					"securityContext": {},
					"affinity": {
						"nodeAffinity": {
							"requiredDuringSchedulingIgnoredDuringExecution": {
								"nodeSelectorTerms": [
									{
										"matchFields": [
											{
												"key": "metadata.name",
												"operator": "In",
												"values": [
													"wn1.kube-cluster.local"
												]
											}
										]
									}
								]
							}
						}
					},
					"schedulerName": "default-scheduler",
					"tolerations": [
						{
							"operator": "Exists",
							"effect": "NoSchedule"
						},
						{
							"operator": "Exists",
							"effect": "NoExecute"
						},
						{
							"key": "CriticalAddonsOnly",
							"operator": "Exists"
						},
						{
							"key": "node.kubernetes.io/not-ready",
							"operator": "Exists",
							"effect": "NoExecute"
						},
						{
							"key": "node.kubernetes.io/unreachable",
							"operator": "Exists",
							"effect": "NoExecute"
						},
						{
							"key": "node.kubernetes.io/disk-pressure",
							"operator": "Exists",
							"effect": "NoSchedule"
						},
						{
							"key": "node.kubernetes.io/memory-pressure",
							"operator": "Exists",
							"effect": "NoSchedule"
						},
						{
							"key": "node.kubernetes.io/pid-pressure",
							"operator": "Exists",
							"effect": "NoSchedule"
						},
						{
							"key": "node.kubernetes.io/unschedulable",
							"operator": "Exists",
							"effect": "NoSchedule"
						},
						{
							"key": "node.kubernetes.io/network-unavailable",
							"operator": "Exists",
							"effect": "NoSchedule"
						}
					],
					"priority": 0,
					"enableServiceLinks": true
				}
			}
			`,
			wantName: "istio-cni-node",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pod := corev1.Pod{}
			err := decodeObject([]byte(tt.json), &pod)
			if err != nil {
				t.Error(err)
				t.FailNow()
			}
			if gotName := GetResourceName(pod.ObjectMeta); gotName != tt.wantName {
				t.Errorf("GetResourceName() = %v, want %v", gotName, tt.wantName)
			}
		})
	}
}
