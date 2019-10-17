# k-rail

k-rail is a workload policy enforcement tool for Kubernetes. It can help you secure a multi tenant cluster with minimal disruption and maximum velocity.

- [k-rail](#k-rail)
- [Installation](#installation)
- [Viewing policy violations](#viewing-policy-violations)
- [Supported policies](#supported-policies)
  * [No Bind Mounts](#no-bind-mounts)
  * [No Docker Sock Mount](#no-docker-sock-mount)
  * [Immutable Image Reference](#immutable-image-reference)
  * [No Host Network](#no-host-network)
  * [No Host PID](#no-host-pid)
  * [No New Capabilities](#no-new-capabilities)
  * [No Privileged Container](#no-privileged-container)
  * [No Helm Tiller](#no-helm-tiller)
  * [Trusted Image Repository](#trusted-image-repository)
    + [Policy configuration](#policy-configuration)
  * [Require Ingress Exemption](#require-ingress-exemption)
    + [Policy configuration](#policy-configuration-1)
- [Configuration](#configuration)
  * [Logging](#logging)
  * [Modes of operation](#modes-of-operation)
    + [Global report-only mode](#global-report-only-mode)
    + [Policy modes](#policy-modes)
  * [Policy exemptions](#policy-exemptions)
  * [Policy configuration](#policy-configuration-2)
- [Adding new policies](#adding-new-policies)
- [Debugging](#debugging)
  * [Resources are having timeout events](#resources-are-having-timeout-events)
    + [Checking kube-apiserver logs](#checking-kube-apiserver-logs)
    + [Viewing webhook latency and status code metrics](#viewing-webhook-latency-and-status-code-metrics)
  * [Policies are enabled, but are not triggering when they should](#policies-are-enabled--but-are-not-triggering-when-they-should)
  * [Policies are enabled, but a deployment is blocked and an exemption is needed](#policies-are-enabled--but-a-deployment-is-blocked-and-an-exemption-is-needed)
- [License](#license)


# Installation

You can install k-rail using the helm chart in [`deploy/helm`](deploy/helm).
For Helm 2 and below, it is recommended to use `helm template` rather than Tiller:

```bash
helm template --namespace k-rail deploy/helm | kubectl apply -f - 
```

By default all policies are enforced (`report_only: false`).
You can adjust the configuration and manage the exemptions in [`values.yaml`](deploy/helm/values.yaml).

Test the default configuration by applying the non-compliant deployment:

```bash
kubectl apply -n default -f deploy/non-compliant-deployment.yaml
```

# Viewing policy violations

You may see violations when applying your resources:

```bash
$ kubectl apply -f deploy/non-compliant-deployment.yaml
Error from server (k-rail): error when creating "deploy/non-compliant-deployment.yaml": admission webhook "k-rail.cruise-automation.github.com" denied the request:
bad-deployment had violation: Host Bind Mounts: host bind mounts are forbidden
bad-deployment had violation: Docker Sock Mount: mounting the Docker socket is forbidden
bad-deployment had violation: No Host Network: Using the host network is forbidden
bad-deployment had violation: No Host PID: Using the host PID namespace is forbidden
```
    
You can also see violations that have occurred recently with the events API:


```bash
$ kubectl get events --namespace enforced
LAST SEEN   TYPE      REASON         KIND         MESSAGE
3m41s       Warning   FailedCreate   ReplicaSet   Error creating: admission webhook "k-rail.cruise-automation.github.com" denied the request:
bad-pod-5f7cd9bf45-rbhsb had violation: Docker Sock Mount: mounting the Docker socket is forbidden
```

Since the violations are outputted as structured data, you are encouraged to aggregate and display that information. GCP BigQuery + Data Studio, Sumologic, Elasticsearch + Kibana, Splunk, etc are all capable of this.


# Supported policies

## No Bind Mounts

Host bind mounts can be used to exfiltrate data from or escalate privileges on the host system. Using host bind mounts can cause unreliability of the node if it causes a partition to fill up.

## No Docker Sock Mount

The Docker socket bind mount provides API access to the host Docker daemon, which can be used for privilege escalation or otherwise control the container host. Using Docker sock mounts can cause unreliability of the node because of the extra workloads that the Kubernetes schedulers are not aware of.

## Immutable Image Reference

Docker image tags in a registry are mutable, so if you reference a tag without specifying the image digest someone or something could change the image you were using without you knowing.

You can obtain the immutable reference for an image with this command:

```bash
$ docker inspect --format='{{index .RepoDigests 0}}' alpine:3.8  
alpine@sha256:dad671370a148e9d9573e3e10a9f8cc26ce937bea78f3da80b570c2442364406
You can also add the tag back in for it to be more human readable:

alpine:3.8@sha256:dad671370a148e9d9573e3e10a9f8cc26ce937bea78f3da80b570c2442364406
```

## No Host Network

Host networking enables packet capture of host network interfaces and a bypass to some cloud meta data APIs, such as the GKE metadata API. The metadata API can be used to escalate access.

## No Host PID

The host PID namespace can be used to inspect process environment variables (which often contain secrets). It can also potentially be used to dump process memory.

## No New Capabilities

Kernel Capabilities can be used to escalate to level of kernel API access available to the process. Some can enable loading kernel modules among other potentially dangerous things.

## No Privileged Container

Privileged containers have all capabilities and also removes cgroup resource accounting.

## No Helm Tiller

Helm Tiller installations often have an unauthenticated API open to the cluster which provides a privilege escalation route to ClusterAdmin or NamespaceEditor.

## Trusted Image Repository

There are many malicious, poorly configured, and outdated and vulnerable images available in public Docker image repositories. Images must be sourced from configured trusted internal repositories or from an official Docker Hub repository.

### Policy configuration

The Trusted Image Repository policy can be configured in the K-rail configuration file.

Example
```yaml
policy_config:
  policy_trusted_repository_regexes:
    - '^gcr.io/some-gcr-repo/.*'   # private GCR repo
    - '^k8s.gcr.io/.*'             # official k8s GCR repo
    - '^[A-Za-z0-9\-:@]+$'         # official docker hub images
```

## Require Ingress Exemption

The Require Ingress Exemption policy requires the configured ingress classes to have an a Policy exemption to be used. This is typically useful if you want to gate the usage of public ingress.

### Policy configuration

The Require Ingress Exemption policy can be configured in the K-rail configuration file.

Example
```yaml
policy_config:
  policy_require_ingress_exemption_classes:
    - nginx-public
```


# Configuration

## Logging

Log levels can be set in the K-rail configuration file. Acceptable values are `debug`, `warn`, and `info`. The default log level is `info`.

All reporting and enforcement operations are logged in a structured json blob per event.
It is useful to run policies in report-only mode, analyze your state in with the structured logs, and flip on enforcement mode when appropriate.

## Modes of operation

### Global report-only mode

When `global_report_only_mode` is toggled in the config, ALL policies run in `report_only` mode, even if configured otherwise.
This mode must be false to have any policies in enforcement mode.

### Policy modes

Policies configure a validator. They can be enabled/disabled, and run in report-only or enforcement mode as specified in the config.

## Policy exemptions

A folder to load policy exemptions from can be specified from config. Load exemptions by specifying the `-exemptions-path-glob` parameter, and specify a path glob that includes the exemptions, such as `/config/policies/*.yaml`.

The format of an exemption config is YAML, and looks like this:

```yaml
---
# exempt all kube-system pods since they are largely provided by GKE
- resource_name: "*"
  namespace: "kube-system"
  exempt_policies: 
  - "*"

# malicious-pod needs host network to escalate access via GCE metadata API
- resource_name: malicious-pod
  namespace: malicious
  exempt_policies: ["pod_no_host_network"]

# allow everything
# - resource_name: "*"
#   namespace: "*"
#   username: "*"
#   group: "*"
#   exempt_policies: ["*"]
```

## Policy configuration

Some policies are configurable. Policy configuration is contained in the K-rail configuration file, and documentation for a policy's configuration can be found in the Supported policies heading above.


# Adding new policies

Policies must satisfy this interface:

```go
type Policy interface {
	Name() string
	Validate(ctx context.Context, config policies.Config, ar *admissionv1beta1.AdmissionRequest) []internal.ResourceViolation
}
```

`Name()` must return a string that matches a policy name that is provided in configuration.
Validate accepts an AdmissionRequest, and the resource of interest must be extracted from it. See `resource/pod.go` for an example of extracting PodSpecs from an AdmissionRequest.


Policies can be registered in `internal/policies.go`. Any policies that are registered but do not have configuration provided get enabled in report-only mode.



# Debugging

## Resources are having timeout events

If you see timeout events on resources, this may be because the `k-rail` service is unreachable from the Kubernetes apiserver.
Newer versions (1.14+) of Kubernetes are not likely to have this issue if the `ValidationWebhookConfiguration` `failurePolicy` is set to `Ignore` and `timeoutSeconds` is set to a lower number (such as `5` or less).

To determine if this is occuring because the service is unreachable, check the `kube-apiserver` logs. You will see logs similar to,
```
E0911 04:57:22.686526       1 dispatcher.go:68] failed calling webhook "K-rail.cruise-automation.github.com": Post https://K-rail.K-rail.svc:443/?timeout=30s: dial tcp 10.110.63.191:443: connect: connection refused
```

### Checking kube-apiserver logs

Checking `kube-apiserver` logs depends on what Kubernetes distribution you use.
- For minikube and other self hosted (meaning Kubernetes runs its infra on itself) clusters, you can typically just view the logs for the `apiserver` pod in the `kube-system` namespace.
- For some non self-hosted clusters, such as GKE, you can download the `apiserver` logs through the Kubernetes proxy:
```bash
kubectl proxy --port 8001 &
curl http://localhost:8001/logs/kube-apiserver.log > /tmp/out.log
```

### Viewing webhook latency and status code metrics

The apiserver tracks latency and status code metrics for webhooks. This may be useful for debugging timeouts or assurance of performance.

```bash
kubectl proxy --port 8001 &
curl -s http://localhost:8001/metrics | grep k-rail
```

## Policies are enabled, but are not triggering when they should

This may be caused by the `k-rail` service being unreachable. To determine this, see [Resources are having timeout events](#resources-are-having-timeout-events).

## Policies are enabled, but a deployment is blocked and an exemption is needed

If you need to make an exemption to a policy, see [Policy exemptions](#policy-exemptions).

# License

Copyright (c) 2019-present, Cruise LLC

This source code is licensed under the Apache License, Version 2.0, found in the LICENSE file in the root directory of this source tree. You may not use this file except in compliance with the License.