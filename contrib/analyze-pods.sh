#!/usr/bin/env bash

# This script was used to analyze some existing workloads for point-in-time analysis and verification.
# We're providing it in case it is useful. There may be later usage and documentation at some point in the future.

mkdir -p /tmp/pods
rm /tmp/pods/filtered.json

for cluster in $(kubectx)
do
    kubectx $cluster

    echo "dumping pod specs for $cluster"
    kubectl get po --all-namespaces -o json > /tmp/pods/pods-$cluster.json

    policy="hostnetwork"
    echo "filtering $policy for $cluster"
    jq '.items[] | select(.spec.hostNetwork==true)' /tmp/pods/pods-$cluster.json | \
        jq "{namespace: .metadata.namespace, name: .metadata.name, cluster: \"$cluster\", policy: \"$policy\"}" >> /tmp/pods/filtered.json

    policy="hostpid"
    echo "filtering $policy for $cluster"
    jq '.items[] | select(.spec.hostPID==true)' /tmp/pods/pods-$cluster.json | \
        jq "{namespace: .metadata.namespace, name: .metadata.name, cluster: \"$cluster\", policy: \"$policy\"}" >> /tmp/pods/filtered.json

    policy="privileged"
    echo "filtering $policy for $cluster"
    jq '.items[] | select(.spec.containers[].securityContext.privileged==true)' /tmp/pods/pods-$cluster.json | \
        jq "{namespace: .metadata.namespace, name: .metadata.name, cluster: \"$cluster\", policy: \"$policy\"}" >> /tmp/pods/filtered.json

    policy="hostpath"
    echo "filtering $policy for $cluster"
    jq '.items[] | select(.spec.volumes[].hostPath)' /tmp/pods/pods-$cluster.json | \
        jq "{namespace: .metadata.namespace, name: .metadata.name, cluster: \"$cluster\", policy: \"$policy\"}" >> /tmp/pods/filtered.json

    policy="dockersock"
    echo "filtering $policy for $cluster"
    jq '.items[] | select(.spec.containers[].volumeMounts[].mountPath=="/var/run/docker.sock")' /tmp/pods/pods-$cluster.json | \
        jq "{namespace: .metadata.namespace, name: .metadata.name, cluster: \"$cluster\", policy: \"$policy\"}" >> /tmp/pods/filtered.json

    policy="capabilities"
    echo "filtering $policy for $cluster"
    jq '.items[] | select(.spec.containers[].securityContext.capabilities.add)' /tmp/pods/pods-$cluster.json | \
        jq "{namespace: .metadata.namespace, name: .metadata.name, cluster: \"$cluster\", policy: \"$policy\"}" >> /tmp/pods/filtered.json

done