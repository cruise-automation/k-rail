Example Plugin
--------------

This directory contains an example K-Rail plugin in order to allow custom written K-Rail policies for your organization that are not general purpose enough for open-source usage.

# K-Rail plugins
K-Rail plugins are based on the [go-plugin Hashicorp](https://github.com/hashicorp/go-plugin/) interface which runs the provided plugin executable over localhost via GRPC and protobufs. This means that theoretically any language that supports GRPC can be written as a K-Rail plugin as long as it conforms to the [KRailPlugin protobuf service](../proto/plugin.proto) as seen below.

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

# About the example
The example plugin is written in Go as a template that can be copied for your own usage with a simple example "luck_threshold" policy included. The "luck_threshold" policy rejects pods from validation based on a configurable threshold. It can be run locally with the included config.yml by using the `make run-plugin` command. Definitely do not use this example policy on a production cluster as it will cause frustration by rejecting 1% of pod deployments by default.
