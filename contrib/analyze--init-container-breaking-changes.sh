#!/usr/bin/env bash

# This script was used to analyze some existing workloads for point-in-time analysis and verification.
# We're providing it in case it is useful. There may be later usage and documentation at some point in the future.

# This was used to analyzing the backwards-incompatible policy changes introduced in https://github.com/cruise-automation/k-rail/pull/21
# and present in k-rail v1.0

mkdir -p /tmp/pods
rm /tmp/pods/filtered.json

for cluster in $(kubectx)
do
    kubectx $cluster

    echo "dumping pod specs for $cluster"
    kubectl get po --all-namespaces -o json > /tmp/pods/pods-$cluster.json

    policy="privileged"
    echo "filtering $policy for $cluster"
    jq '.items[] | select(.spec.initContainers[].securityContext.privileged==true)' /tmp/pods/pods-$cluster.json | \
        jq "{namespace: .metadata.namespace, name: .metadata.name, cluster: \"$cluster\", policy: \"$policy\"}" >> /tmp/pods/filtered.json

    policy="capabilities"
    echo "filtering $policy for $cluster"
    jq '.items[] | select(.spec.initContainers[].securityContext.capabilities.add)' /tmp/pods/pods-$cluster.json | \
        jq "{namespace: .metadata.namespace, name: .metadata.name, cluster: \"$cluster\", policy: \"$policy\"}" >> /tmp/pods/filtered.json

done