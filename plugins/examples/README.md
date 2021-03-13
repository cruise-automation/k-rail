Example Plugin
--------------

This directory contains an example K-Rail plugin in order to allow custom written K-Rail policies for your organization that are not general purpose enough for open-source usage.

# K-Rail plugins
K-Rail plugins are based on the [go-plugin Hashicorp](https://github.com/hashicorp/go-plugin/) interface which runs the provided plugin executable over localhost via GRPC and protobufs. This means that theoretically any language that supports GRPC can be written as a K-Rail plugin as long as it conforms to the [KRailPlugin protobuf service](../proto/plugin.proto).

# About the example
The example plugin is written as a template that can be copied for your own usage with a simple example "luck_threshold" policy included. The "luck_threshold" policy rejects pods from validation based on a configurable threshold. It can be run locally with the included config.yml by using the `make run-plugin` command. Definitely do not use this example policy on a production cluster as it will cause frustration by rejecting 1% of pod deployments by default.
