# Evict tainted pods after period
Operator that runs within k8s to find and evict `tainted` pods.

* `k-rail/tainted-timestamp` store the unix timestamp when the root event happend
* `k-rail/tainted-prevent-eviction` is a break-glass annotation to prevent automated eviction
* `k-rail/reason` intended for humans
