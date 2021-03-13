package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"

	"github.com/hashicorp/go-plugin"
	admissionv1 "k8s.io/api/admission/v1"

	"github.com/cruise-automation/k-rail/plugins"
	"github.com/cruise-automation/k-rail/policies"
	"github.com/cruise-automation/k-rail/resource"
)

const pluginName = "example_plugin"

// This implements the plugins.KRailPlugin interface in-order to implement
// the GRPC KRailPlugin service defined in the plugins.proto specification
type ExamplePlugin struct {
	Config   Config
	Policies map[string]Policy
}

type Config struct {
	Threshold float64
}

func (p ExamplePlugin) PluginName() (string, error) {
	return pluginName, nil
}

func (p ExamplePlugin) RegisterPolicy(policy Policy) {
	p.Policies[policy.Name()] = policy
}

func (p ExamplePlugin) PolicyNames() ([]string, error) {
	names := []string{}
	for name, _ := range p.Policies {
		names = append(names, name)
	}
	return names, nil
}

func (p ExamplePlugin) ConfigurePlugin(config map[string]interface{}) error {
	if threshold, ok := config["threshold"]; ok {
		if threshold64, ok := threshold.(float64); ok {
			p.Config.Threshold = threshold64
			log.Printf("Configured luck threshold to %.2f\n", threshold64)
		}
	}
	return nil
}

func (p ExamplePlugin) Validate(policyName string, ar *admissionv1.AdmissionRequest) ([]policies.ResourceViolation, []policies.PatchOperation, error) {
	if policy, ok := p.Policies[policyName]; ok {
		violations, patchOps := policy.Validate(context.Background(), p.Config, ar)
		return violations, patchOps, nil
	}
	return []policies.ResourceViolation{}, nil, fmt.Errorf("unknown policy %s for plugin %s\n", policyName, pluginName)
}

// This is mostly a copy of the server.Policies with the custom Config object being the only difference
type Policy interface {
	Name() string
	Validate(ctx context.Context,
		config Config,
		ar *admissionv1.AdmissionRequest,
	) ([]policies.ResourceViolation, []policies.PatchOperation)
}

type ThresholdPolicy struct{}

func (t ThresholdPolicy) Name() string {
	return "luck_threshold"
}

func (t ThresholdPolicy) Validate(ctx context.Context,
	config Config,
	ar *admissionv1.AdmissionRequest,
) ([]policies.ResourceViolation, []policies.PatchOperation) {

	resourceViolations := []policies.ResourceViolation{}
	podResource := resource.GetPodResource(ctx, ar)
	if podResource == nil {
		return resourceViolations, nil
	}

	if rand.Float64() > config.Threshold {
		violationText := fmt.Sprintf("This Pod was unlucky and didn't clear the random %.2f%% threshold, rejecting", config.Threshold)
		resourceViolations = append(resourceViolations, policies.ResourceViolation{
			Namespace:    ar.Namespace,
			ResourceName: podResource.ResourceName,
			ResourceKind: podResource.ResourceKind,
			Violation:    violationText,
			Policy:       t.Name(),
		})

	}
	return resourceViolations, nil
}

func main() {
	// Default the luck threshold to 99% (it's very lucky)
	examplePlugin := &ExamplePlugin{Config: Config{Threshold: 0.99}, Policies: map[string]Policy{}}
	examplePlugin.RegisterPolicy(ThresholdPolicy{})

	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: plugins.Handshake,
		Plugins: map[string]plugin.Plugin{
			plugins.GRPCPluginName: &plugins.KRailGRPCPlugin{Impl: examplePlugin},
		},

		// A non-nil value here enables gRPC serving for this plugin...
		GRPCServer: plugin.DefaultGRPCServer,
	})
}
