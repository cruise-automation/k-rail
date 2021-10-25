![k-rail](images/k-rail.png)

[![CI Status](https://github.com/cruise-automation/k-rail/workflows/CI/badge.svg)](https://github.com/cruise-automation/k-rail/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/cruise-automation/k-rail)](https://goreportcard.com/report/github.com/cruise-automation/k-rail)
[![Docker Hub Build Status](https://img.shields.io/docker/cloud/build/cruise/k-rail.svg)](https://hub.docker.com/r/cruise/k-rail/)

k-rail is a workload policy enforcement tool for Kubernetes. It can help you secure a multi tenant cluster with minimal disruption and maximum velocity.

- [Why k-rail?](#why-k-rail)
  - [Suggested usage](#suggested-usage)
- [Installation](#installation)
- [Removal](#removal)
- [Viewing policy violations](#viewing-policy-violations)
  - [Violations from realtime feedback](#violations-from-realtime-feedback)
  - [Violations from the Events API](#violations-from-the-events-api)
  - [Violations from logs](#violations-from-logs)
- [Supported policies](#supported-policies)
  - [No ShareProcessNamespace](#no-shareprocessnamespace)
  - [No Exec](#no-exec)
  - [No Bind Mounts](#no-bind-mounts)
  - [No Docker Sock Mount](#no-docker-sock-mount)
  - [EmptyDir size limit](#emptydir-size-limit)
    - [Policy configuration](#policy-configuration)
  - [Mutate Default Seccomp Profile](#mutate-default-seccomp-profile)
    - [Policy configuration](#policy-configuration-1)
  - [Immutable Image Reference](#immutable-image-reference)
  - [No Host Network](#no-host-network)
  - [No Host PID](#no-host-pid)
  - [No New Capabilities](#no-new-capabilities)
  - [No Privileged Container](#no-privileged-container)
  - [No Helm Tiller](#no-helm-tiller)
  - [Trusted Image Repository](#trusted-image-repository)
    - [Policy configuration](#policy-configuration-2)
  - [Safe to Evict (DEPRECATED)](#safe-to-evict-deprecated)
  - [Mutate Safe to Evict](#mutate-safe-to-evict)
  - [Mutate Image Pull Policy](#mutate-image-pull-policy)
    - [Policy configuration](#policy-configuration-3)
  - [Require Ingress Exemption](#require-ingress-exemption)
    - [Policy configuration](#policy-configuration-4)
  - [Unique Ingress Host](#unique-ingress-host)
  - [Service type LoadBalancer annotation check](#service-type-loadbalancer-annotation-check)
    - [Policy configuration](#policy-configuration-5)
  - [Istio VirtualService Gateways check](#istio-virtualservice-gateways-check)
    - [Policy configuration](#policy-configuration-6)
  - [No Persistent Volume Host Path](#no-persistent-volume-host-path)
  - [No Anonymous Cluster Role Binding](#no-anonymous-cluster-role-binding)
  - [No Anonymous Role Binding](#no-anonymous-role-binding)
  - [Invalid Pod Disruption Budget](#invalid-pod-disruption-budget)
  - [No External IP on Service](#no-external-ip-on-service)
  - [Deny Unconfined AppArmor Policies](#deny-unconfined-apparmor-policies)
  - [Protect CRD from accidental Deletion](#crd-protect)
- [Configuration](#configuration)
  - [Webhook Configuration](#webhook-configuration)
  - [Logging](#logging)
  - [Modes of operation](#modes-of-operation)
    - [Global report-only mode](#global-report-only-mode)
    - [Policy modes](#policy-modes)
  - [Policy exemptions](#policy-exemptions)
  - [Policy configuration](#policy-configuration-7)
- [Adding new policies](#adding-new-policies)
- [Debugging](#debugging)
  - [Resources are having timeout events](#resources-are-having-timeout-events)
    - [Checking kube-apiserver logs](#checking-kube-apiserver-logs)
    - [Viewing webhook latency and status code metrics](#viewing-webhook-latency-and-status-code-metrics)
  - [Policies are enabled, but are not triggering when they should](#policies-are-enabled-but-are-not-triggering-when-they-should)
  - [Policies are enabled, but a deployment is blocked and an exemption is needed](#policies-are-enabled-but-a-deployment-is-blocked-and-an-exemption-is-needed)
  - [Checking the mTLS certificate expiration](#checking-the-mtls-certificate-expiration)
- [License](#license)

# Why k-rail?

By default, the Kubernetes APIs allow for a variety of easy privilege escalation routes. When operating a multi-tenant cluster, many features can be dangerous or introduce instability and must be used judiciously. k-rail attempts to make workload policy enforcement easy in Kubernetes, even if you already have a large number of diverse workloads. Several features enable you to roll out policy enforcement safely without breaking existing workloads:

- Passive report-only mode of running policies
- Structured violation data logged, ready for analysis and dashboards
  ```json
  {
    "enforced": true,
    "kind": "PodExec",
    "namespace": "ecommerce",
    "policy": "pod_no_exec",
    "resource": "payment-processor",
    "user": "bob@amyshardware.com",
    "time": "2019-11-03T06:28:07Z"
  }
  ```
- Flexible and powerful policy exemptions by cluster, resource name, namespace, groups, and users
  ```yaml
  ---
  - cluster_name: paas-ci-us-west1
    resource_name: "*"
    namespace: "cluster-conformance-testing"
    username: "cluster-ci@paas-ci.iam.gserviceaccount.com"
    group: "*"
    exempt_policies:
      - "pod_no_privileged_containers"
      - "pod_no_bind_mounts"
      - "pod_no_host_network"
      - "pod_default_seccomp_policy"
      - "pod_no_host_pid"
      - "pod_no_exec"
  ```
- Realtime interactive feedback for engineers and systems that apply resources
  ```bash
  $ kubectl apply --namespace default -f examples/non-compliant-ingress.yaml
  Error from server (InternalError): error when creating "examples/non-compliant-ingress.yaml":
  Internal error occurred: admission webhook "k-rail.cruise-automation.github.com" denied the request:
  Ingress bad-ingress had violation: Require Ingress Exemption: Using the 'public' Ingress class requires an exemption
  Ingress bad-ingress had violation: Requires Unique Ingress Host: Ingress Host should not point to multiple namespaces
  ```

By leveraging the first three features you can quickly and easily roll out enforcement to deployments without breaking them and monitor violations with confidence. The interactive feedback informs and educates engineers during future policy violations.

Cruise was able to utilize this software to apply enforcement to more than a dozen clusters with thousands of existing diverse workloads in all environments in about a week without breaking existing deployments. Now you can too.

## Suggested usage

If you have a new cluster without existing workloads, just run k-rail in enforcement mode for the policies you desire and add exemptions as needed.

If you have a cluster with existing workloads, run it in monitor mode for a few weeks or until you have collected enough data. The violation events are emmitted in the logs in JSON, so it is suggested that you analyze that data collected to make exemptions as needed. Once the exemptions are applied, you can safely turn on enforcement mode without breaking existing workloads.

# Installation

You can install or update k-rail using the helm chart in [`charts/k-rail`](charts/k-rail). You can install the latest chart directly from our repo, by running:

```bash
# add the helm repo
helm repo add k-rail https://cruise-automation.github.io/k-rail/
helm repo update

# prepare the namespace
kubectl create namespace k-rail
kubectl label namespace k-rail k-rail/ignore=true

# install
helm install k-rail k-rail/k-rail --namespace k-rail
```

For the Helm deployment, all configuration for policies and exemptions are contained in [`charts/k-rail/values.yaml`](charts/k-rail/values.yaml). Feel free to override configuration values as you see fit per the various Helm provided [methods](https://helm.sh/docs/chart_template_guide/values_files/).

For Helm 2 and below, it is recommended to use `helm template` render the YAML for applying rather than using Helm Tiller:

```bash
helm template --namespace k-rail charts/k-rail | kubectl apply -f -
```

By default all policies are enforced (`report_only: false`).

Test the default configuration by applying the provided non-compliant deployment:

```bash
kubectl apply --namespace default -f examples/non-compliant-deployment.yaml
```

# Removal

If ussing `helm install`,

```bash
helm uninstall k-rail --namespace k-rail
```

If using `helm template`,

```bash
helm template --namespace k-rail charts/k-rail | kubectl delete -f -
```

# Viewing policy violations

There are a few ways of viewing violations. Violations from realtime feedback and the Events API is most useful for users, but violations from logs is most useful for presentation and analysis.

## Violations from realtime feedback

You may see violations when applying your resources:

```bash
$ kubectl apply -f examples/non-compliant-deployment.yaml

Error from server (k-rail): error when creating "examples/non-compliant-deployment.yaml": admission webhook "k-rail.cruise-automation.github.com" denied the request:
Deployment bad-deployment had violation: Host Bind Mounts: host bind mounts are forbidden
Deployment bad-deployment had violation: Docker Sock Mount: mounting the Docker socket is forbidden
Deployment bad-deployment had violation: Immutable Image Reference: image tag must include its sha256 digest
Deployment bad-deployment had violation: No Host Network: Using the host network is forbidden
Deployment bad-deployment had violation: No Privileged Container: Using privileged containers is forbidden
Deployment bad-deployment had violation: No New Capabilities: Adding additional capabilities is forbidden
Deployment bad-deployment had violation: No Host PID: Using the host PID namespace is forbidden
Deployment bad-deployment had violation: Safe to evict: annotation is required for Pods that use emptyDir or hostPath mounts to enable cluster autoscaling
```

## Violations from the Events API

You can also see violations that have occurred recently with the events API:

```bash
$ kubectl get events --namespace default
LAST SEEN   TYPE      REASON         KIND         MESSAGE
3m41s       Warning   FailedCreate   ReplicaSet   Error creating: admission webhook "k-rail.cruise-automation.github.com" denied the request:
bad-pod-5f7cd9bf45-rbhsb had violation: Docker Sock Mount: mounting the Docker socket is forbidden
```

## Violations from logs

Violations are also emitted as structured data in the logs:

```json
$ kubectl logs --namespace k-rail --selector name=k-rail | jq '.'

{
  "enforced": true,
  "kind": "Deployment",
  "namespace": "default",
  "policy": "pod_no_host_network",
  "resource": "evil-deployment",
  "time": "2019-10-23T19:54:24Z",
  "user": "dustin.decker@getcruise.com"
}
{
  "enforced": true,
  "kind": "Deployment",
  "namespace": "default",
  "policy": "pod_no_privileged_container",
  "resource": "evil-deployment",
  "time": "2019-10-23T19:54:24Z",
  "user": "dustin.decker@getcruise.com"
}
{
  "enforced": true,
  "kind": "Deployment",
  "namespace": "default",
  "policy": "pod_no_new_capabilities",
  "resource": "evil-deployment",
  "time": "2019-10-23T19:54:24Z",
  "user": "dustin.decker@getcruise.com"
}

```

Since the violations are outputted as structured data, you are encouraged to aggregate and display that information. GCP BigQuery + Data Studio, Sumologic, Elasticsearch + Kibana, Splunk, etc are all capable of this.

# Supported policies

Below are the policies built-in to K-Rail. Additional custom written policies can be created and configured for your organization if they are not general purpose enough for committing to upstream.  See the [example plugin provided](plugins/examples/README.md) for details on writing your own policy plugin.

## No ShareProcessNamespace

`shareProcessNamespace: true` is a Pod Spec directive that puts all containers in a Pod within
the same PID Namespace. When this occurs, containers can, for example, [access each others' filesystem and memory](https://kubernetes.io/docs/tasks/configure-pod-container/share-process-namespace/#understanding-process-namespace-sharing),
as long as they share user and group IDs. These effects could be unexpected, especially if security (e.g. egress controls)
occurs in a sidecar container.

## No Exec

The No Exec policy prevents users from execing into running pods unless they have an exemption. This policy is typically enforced within a production environment, but run in report-only mode in dev and staging environments to facilitate debugging.

Execing into a pod can enable someone to do many nefarious things to that workload. Eventually this policy will also apply a taint label to the Pod to indicate that it should no longer be trusted and can be evicted.

## No Bind Mounts

Host bind mounts (also called `hostPath` mounts) can be used to exfiltrate data from or escalate privileges on the host system. Using host bind mounts can cause unreliability of the node if it causes a partition to fill up.

## No Docker Sock Mount

The Docker socket bind mount provides API access to the host Docker daemon, which can be used for privilege escalation or otherwise control the container host. Using Docker sock mounts can cause unreliability of the node because of the extra workloads that the Kubernetes schedulers are not aware of.

**Note:** It is recommended to use the `No Bind Mounts` policy to disable all `hostPath` mounts rather than only this policy, because it is easily bypassed. This policy does not provide meaningful protection and is here for informative purposes.

## No Root User

Running as the root user is extremely dangerous and should be forbidden for all possible workloads.
This policy blocks pods when the security context doesn't explicitly set `runAsUser: [some uid > 0]` or `runAsNonRoot: true`

The securityContext can be set at the pod level or on each individual container.

## EmptyDir size limit

By [default](https://kubernetes.io/docs/concepts/storage/volumes/#example-pod), an `emptyDir` lacks a `sizeLimit` parameter, and is disk-based;
a Pod with access to said `emptyDir` can consume the Node's entire disk (i.e. the limit is unbounded) until the offending Pod is deleted or evicted, which can constitute a denial-of-service condition at the affected Node (i.e. DiskPressure).
This policy
* sets the configured default size when none is set for an `emptyDir` volume
* reports a violation when the size is greater then the configured max size

### Policy configuration

```yaml
policy_config:
    mutate_empty_dir_size_limit:
      maximum_size_limit: "1Gi"
      default_size_limit: "512Mi"
```

## Mutate Default Seccomp Profile

Sets a default seccomp profile (`runtime/default` or a configured one) for Pods if they have no existing seccomp configuration. The default seccomp policy for Docker and Containerd both block over 40 syscalls, [many of which](https://docs.docker.com/engine/security/seccomp/#significant-syscalls-blocked-by-the-default-profile) are potentially dangerous. The default policies are [usually very compatible](https://blog.jessfraz.com/post/containers-security-and-echo-chambers/#breaking-changes) with applications, too.

### Policy configuration

The Mutate Default Seccomp Profile policy can be configured in the k-rail configuration file.

Example

```yaml
policy_config:
  policy_default_seccomp_policy: "runtime/default"
```

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

The host PID namespace can be used to inspect process environment variables (which often contain secrets). It can also potentially be used to dump process memory, modify kernel parameters, and insert kprobes+uprobes into the kernel to exfiltrate information.

## No New Capabilities

Kernel Capabilities can be used to escalate to level of kernel API access available to the process. Some can enable loading kernel modules, changing namespace, load eBPF byte code in the kernel and other potentially dangerous things.

## No Privileged Container

Privileged containers have all capabilities and also removes cgroup resource accounting.

## No Helm Tiller

Helm Tiller installations often have an unauthenticated API open to the cluster which provides a privilege escalation route to ClusterAdmin or NamespaceEditor.

**Note:** This policy only blocks images that `/tiller` in their name from being deployed. It is not a robust policy and serves more as a reminder for engineers to seek an alternate route of deployment, such as using `helm template` or [isopod](https://github.com/cruise-automation/isopod#isopod).

## Trusted Image Repository

There are many malicious, poorly configured, and outdated and vulnerable images available in public Docker image repositories. Images must be sourced from configured trusted internal repositories or from an official Docker Hub repository.

### Policy configuration

The Trusted Image Repository policy can be configured in the k-rail configuration file.

Example

```yaml
policy_config:
  policy_trusted_repository_regexes:
    - '^gcr.io/some-gcr-repo/.*'   # private GCR repo
    - '^k8s.gcr.io/.*'             # official k8s GCR repo
    - '^[A-Za-z0-9\-:@]+$'         # official docker hub images
```

## Safe to Evict (DEPRECATED)

**DEPRECATED** - See `Mutate Safe to Evict` below

The Kubernetes autoscaler will not evict pods using hostPath or emptyDir mounts unless they have this annotation:

```javascript
cluster-autoscaler.kubernetes.io/safe-to-evict=true
```

This policy validates that Pods have this annotation. You'll probably find the mutation policy below more useful.

## Mutate Safe to Evict

The Kubernetes autoscaler will not evict pods using hostPath or emptyDir mounts unless they have this annotation:

```javascript
cluster-autoscaler.kubernetes.io/safe-to-evict=true
```

This policy mutates Pods that do not have the annotation specfied to be `true`. It will not override existing annotations with `false`.

You can also set the annoation on existing Pods with this one-liner:

```bash
$ kubectl get po --all-namespaces -o json | jq -r '.items[] | select(.spec.volumes[].hostPath or .spec.volumes[].emptyDir) | [ .metadata.namespace, .metadata.name ] | @tsv' | while IFS=$'\t' read -r namespace pod; do echo "\n NAMESPACE: $namespace \n POD: $pod \n"; kubectl annotate pod -n $namespace $pod "cluster-autoscaler.kubernetes.io/safe-to-evict=true"; done
```

## Mutate Image Pull Policy

There are cerntain images which require the enforcement of the ImagePullPolicy according to different user scenarios

- IfNotPresent
It can reduce the unnecessary traffic (Auth and Download requests) to Image repository and reuse the image which is cached on the node 
- Always
It can be useful when it requires the absolute isolation in multi-tenant cluster, which prevents others to reuse the image cached on the node, for example: The image protected with ImagePullSecret from private repository is cached on the node after first successful pull, other user can directly pull from node without proper auth.
However if we force the imagePullPolicy to Always, it would fail without proper ImagePullSecret

### Policy configuration

The Mutate Image Pull Policy can be configured in the k-rail configuration file.

Example

```yaml
policy_config:
  mutate_image_pull_policy:
    IfNotPresent: 
      - '^gcr.io/repo/image1.*'
      - '^gcr.io/repo/image2.*'
    Always:
      - '^gcr.io/private-repo/secretimage.*'
```

## Require Ingress Exemption

The Require Ingress Exemption policy requires the configured ingress classes to have an a Policy exemption to be used. This is typically useful if you want to gate the usage of public ingress.

### Policy configuration

The Require Ingress Exemption policy can be configured in the k-rail configuration file.

Example

```yaml
policy_config:
  policy_require_ingress_exemption_classes:
    - nginx-public
```

## Unique Ingress Host

Unique Ingress Host policy requires the configured ingress hosts to be unique across cluster namespaces. This is helps to prevent ingress host collisions.

## Service type LoadBalancer annotation check

Annotations used on services are used to configure public IPs or other cloud provider specific parameters. 

This policy validates the annotations put on a service and will reject services defined with annotations outside the acceptable range.

### Policy configuration

Each annotation to police is configured with a list of possible values and a parameter `allow_missing` which defines if a service is allowed without this annotation present.

```yaml
policy_config:
  policy_require_service_loadbalancer_annotations:
    - annotations:
        - "cloud.google.com/load-balancer-type"
        - "networking.gke.io/load-balancer-type"
      allowed_values:
        - internal
        - external
      allow_missing: false
    - annotation: "networking.gke.io/internal-load-balancer-allow-global-access"
      allowed_values:
        - true
      allow_missing: false
```
## Istio VirtualService Gateways check

Gateways set on Istio virtual services are used to configure public and private Istio ingress access along with potentially usage of sensitive domains.

This policy validates the gateways listed on an Istio virtual service and will reject virtual services defined with gateways outside the acceptable range.

### Policy configuration

A list of allowed gateways is configured along with a parameter to set if an empty list of gateways is allowed for virtual services. According to the [Istio virtual service documentation](https://istio.io/latest/docs/reference/config/networking/virtual-service/), an unset list of gateways will default to the `mesh` gateway which will apply the virtual service to all sidecars in the service mesh.

```yaml
policy_config:
  policy_require_virtualservice_gateways:
    allowed_gateways:
      - "istio-system/internal-gateway"
      - "mesh"
    allow_empty_gateways: true
```

## No Persistent Volume Host Path

Enforcing the policy prevents direct access to potentially sensitive files or directories at the Node-level via Persistent Volumes. Production clusters should not use HostPath. Instead a cluster administrator would provision a network resource like a Google Compute Engine persistent disk, an NFS share, or an Amazon Elastic Block Store volume.

## No Anonymous Cluster Role Binding

Enforcing the policy prevents the creation of cluster level role bindings that authorize unathenticated or anonymous users to access resources.

## No Anonymous Role Binding

Enforcing the policy prevents the creation of namespace level role bindings that authorize unathenticated or anonymous users to access resources. 

## Invalid Pod Disruption Budget

Prevent misconfigured pod disruption budgets from disrupting normal system maintenance such as node drains. Ensures that
- minAvailable is less than the items replicas
- maxUnavailable is greater than or equal to 1

## No External IP on Service

Prevents providing External IPs on a Service to mitigate [CVE-2020-8554](https://cve.mitre.org/cgi-bin/cvename.cgi?name=CVE-2020-8554).

## Deny Unconfined AppArmor Policies

Prevents users from specifing an unconfined apparmor policy which can be used with other conditions to lead to [container escape](https://blog.trailofbits.com/2019/07/19/understanding-docker-container-escapes/).

## Protect CRD From Accidental Deletion

When a Custom Resource Definition is deleted the corresponding Custom Resources are deleted as well. This creates the risk of accidentally destroying important data during regular maintenance. This policy allows the user to set the annotation `k-rail.crd.protect: enabled` on any CRD which will prevent its deletion if any children CRs exist.

## Disallow NGINX Snippet

In response to [NGINX Ingress Controller vulnerability CVE-2021-25742](https://github.com/kubernetes/ingress-nginx/issues/7837), this rule will disallow usages of all NGINX snippet annotations.

# Configuration

For the Helm deployment, all configuration is contained in [`charts/k-rail/values.yaml`](charts/k-rail/values.yaml).

## Webhook Configuration

By default, k-rail will "fail close" if it cannot be reached by the API server. k-rail can be changed to
"fail open" by changing the `failurePolicy` directive from `Fail` to `Ignore`, in [`charts/k-rail/values.yaml`](charts/k-rail/values.yaml).
See the Kubernetes [docs](https://kubernetes.io/docs/reference/access-authn-authz/extensible-admission-controllers/#failure-policy) for
more details.

In Kubernetes 1.15 and beyond, mutating admission webhooks (e.g. k-rail) can elect to be polled again, if a subsequent admission plugin
(such as another webhook) modifies an object the webhook has interacted with. 
They do so with a 
[`reinvocationPolicy`](https://kubernetes.io/docs/reference/access-authn-authz/extensible-admission-controllers/#reinvocation-policy) 
value of `IfNeeded`; the Kubernetes default value is `Never`, which does not reinvoke
the mutating admission webhook(s). 
Since this is a [newer](https://kubernetes.io/blog/2019/06/19/kubernetes-1-15-release-announcement/) type field, k-rail omits by default,
but operators can set a chosen value by commenting out `reinvocationPolicy` in [`charts/k-rail/values.yaml`](charts/k-rail/values.yaml).
See the [associated KEP](https://github.com/kubernetes/enhancements/blob/master/keps/sig-api-machinery/00xx-admission-webhooks-to-ga.md#mutating-plugin-ordering) for more details on `reinvocationPolicy` and admission plugin ordering.

## Logging

Log levels can be set in the k-rail configuration file. Acceptable values are `debug`, `warn`, and `info`. The default log level is `info`.

All reporting and enforcement operations are logged in a structured json blob per event.
It is useful to run policies in report-only mode, analyze your state in with the structured logs, and flip on enforcement mode when appropriate.

## Modes of operation

### Global report-only mode

When `global_report_only_mode` is toggled in the config, ALL policies run in `report_only` mode, even if configured otherwise.
This mode must be false to have any policies in enforcement mode.

### Policy modes

Policies can be enabled/disabled, and run in report-only or enforcement mode as specified in the config.

## Policy exemptions

A folder to load policy exemptions from can be specified from config. Load exemptions by specifying the `-exemptions-path-glob` parameter, and specify a path glob that includes the exemptions, such as `/config/policies/*.yaml`.

For the Helm deployment, all policy and exemption configuration is contained in [`charts/k-rail/values.yaml`](charts/k-rail/values.yaml).

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

*Note:* The resource name automatically has a trailing glob appended in order to match resources created by controllers. This could lead to unintended matches.

## Policy configuration

Some policies are configurable. Policy configuration is contained in the k-rail configuration file, and documentation for a policy's configuration can be found in the Supported policies heading above.

For the Helm deployment, all policy and exemption configuration is contained in [`charts/k-rail/values.yaml`](charts/k-rail/values.yaml).

## Plugin configuration

Custom-written plugins are configurable under the `plugin_config:` yaml key such as below

```yaml
plugin_config:
  <plugin_name>:
    <custom plugin configuration>
```

For an example of this see the [provided plugin example](plugins/examples/README.md) and associated [config.yml](plugins/examples/config.yml).

For additional reference the Helm deployment, contains this example plugin configuration as well, but is disabled by default [`charts/k-rail/values.yaml`](charts/k-rail/values.yaml).

# Adding new policies

## For general open-source use in k-rail

Policies must satisfy this interface:

```go
// Policy specifies how a Policy is implemented
// Returns an optional slice of violations and an optional slice of patch operations if mutation is desired.
type Policy interface {
  Name() string
  Validate(ctx context.Context,
    config policies.Config,
    ar *admissionv1.AdmissionRequest,
  ) ([]policies.ResourceViolation, []policies.PatchOperation)
}
```

`Name()` must return a string that matches a policy name that is provided in configuration.

`Validate`accepts an AdmissionRequest, and the resource of interest must be extracted from it. See `resource/pod.go` for an example of extracting PodSpecs from an AdmissionRequest. If mutation on a resource is desired, you can return a slice of JSONPatch operations and `nil` for the violations.

Policies can be registered in `internal/policies.go`. Any policies that are registered but do not have configuration provided get enabled in report-only mode.

## For a custom purpose using plugins

For custom written policies for your organization that are not general purpose enough for open-source usage, write a policy plugin. See the [example plugin provided](plugins/examples/README.md) for details on writing your own policy plugin in Go. Policy plugins must satisfy the following [GRPC protobuf KRailPlugin service specification](plugins/proto/plugin.proto).

```protobuf
service KRailPlugin {
    rpc PluginName(PluginNameRequest) returns (PluginNameResponse);
    rpc PolicyNames(PolicyNamesRequest) returns (PolicyNamesResponse);
    rpc ConfigurePlugin(ConfigurePluginRequest) returns (ConfigurePluginResponse);
    rpc Validate(ValidateRequest) returns (ValidateResponse);
}
```

`PluginName` returns the name of the plugin as a string which is then used in the `plugin_config` stanza for providing customizable yaml configuration.

`PolicyNames` returns the names of all the policies implemented by the plugin as an array of strings which is then used to configure them under the `policies` stanza as `enabled` and `report_only`

`ConfigurePlugin` provides the customizable yaml from under corresponding `plugin_config` and plugin name stanza to initialize the plugin

`Validate` accepts the policy name with an AdmissionRequest.  The resource of interest must be extracted from it. See `resource/pod.go` for an example of extracting PodSpecs from an AdmissionRequest. If mutation on a resource is desired, you can return a slice of JSONPatch operations and `nil` for the violations.

Again, it is highly recommended to see the [example plugin provided](plugins/examples/README.md) for details on writing your own policy plugin in Go.

# Debugging

## Resources are having timeout events

If you see timeout events on resources, this may be because the `k-rail` service is unreachable from the Kubernetes apiserver.
Newer versions (1.14+) of Kubernetes are not likely to have this issue if the `MutatingWebhookConfiguration` `failurePolicy` is set to `Ignore` and `timeoutSeconds` is set to a lower number (such as `5` or less).

To determine if this is occuring because the service is unreachable, check the `kube-apiserver` logs. You will see logs similar to,

```
E0911 04:57:22.686526       1 dispatcher.go:68] failed calling webhook "k-rail.cruise-automation.github.com": Post https://k-rail.k-rail.svc:443/?timeout=30s: dial tcp 10.110.63.191:443: connect: connection refused
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

## Checking the mTLS certificate expiration

By default, a 10 year certificate is generated during each apply of k-rail. Re-applying will renew it.

You can also check the expiration with this command:

```bash
$ kubectl get secret --namespace k-rail k-rail-cert -o json | jq -r '.data["cert.pem"]' | base64 -d | openssl x509 -noout -text | grep -A 3 "Validity"

        Validity
            Not Before: Oct 24 05:40:16 2019 GMT
            Not After : Oct 21 05:40:16 2029 GMT
        Subject: CN = k-rail.k-rail.svc
```

# License

Copyright (c) 2019-present, Cruise LLC

This source code is licensed under the Apache License, Version 2.0, found in the LICENSE file in the root directory of this source tree. You may not use this file except in compliance with the License.
