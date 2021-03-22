package plugins

import (
	"context"
	"os/exec"
	"path/filepath"

	"github.com/hashicorp/go-plugin"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	admissionv1 "k8s.io/api/admission/v1"

	"github.com/cruise-automation/k-rail/plugins/proto"
	"github.com/cruise-automation/k-rail/policies"
)

type Plugin struct {
	name        string
	policyNames []string
	client      plugin.Client
	kRailPlugin KRailPlugin
}

func (p *Plugin) Name() string {
	return p.name
}

func (p *Plugin) PolicyNames() []string {
	return p.policyNames
}

func (p *Plugin) Configure(config map[string]interface{}) error {
	return p.kRailPlugin.ConfigurePlugin(config)
}

func (p *Plugin) Kill() {
	p.client.Kill()
}

func (p *Plugin) Validate(policyName string, ar *admissionv1.AdmissionRequest) ([]policies.ResourceViolation, []policies.PatchOperation, error) {
	return p.kRailPlugin.Validate(policyName, ar)
}

// PluginPolicy implements the server.Policy interface
type PluginPolicy struct {
	name   string
	plugin Plugin
}

func NewPluginPolicy(name string, plugin Plugin) PluginPolicy {
	return PluginPolicy{name: name, plugin: plugin}
}

func (p PluginPolicy) Name() string {
	return p.name
}

func (p PluginPolicy) Validate(ctx context.Context,
	config policies.Config,
	ar *admissionv1.AdmissionRequest,
) ([]policies.ResourceViolation, []policies.PatchOperation) {

	violations, patchOps, err := p.plugin.Validate(p.name, ar)

	if err != nil {
		log.WithError(err).Errorf("error running Validate on Plugin %s Policy %s\n", p.plugin.name, p.name)
		return []policies.ResourceViolation{}, nil
	}
	return violations, patchOps
}

// KRailPlugin is the interface that we're exposing as a plugin.
type KRailPlugin interface {
	PluginName() (string, error)
	PolicyNames() ([]string, error)
	ConfigurePlugin(config map[string]interface{}) error
	Validate(policyName string, ar *admissionv1.AdmissionRequest) ([]policies.ResourceViolation, []policies.PatchOperation, error)
}

// This is the implementation of plugin.GRPCPlugin so we can serve/consume this.
type KRailGRPCPlugin struct {
	// GRPCPlugin must still implement the Plugin interface
	plugin.Plugin
	// Concrete implementation, written in Go. This is only used for plugins
	// that are written in Go.
	Impl KRailPlugin
}

func (p *KRailGRPCPlugin) GRPCServer(broker *plugin.GRPCBroker, s *grpc.Server) error {
	proto.RegisterKRailPluginServer(s, &GRPCServer{Impl: p.Impl})
	return nil
}

func (p *KRailGRPCPlugin) GRPCClient(ctx context.Context, broker *plugin.GRPCBroker, c *grpc.ClientConn) (interface{}, error) {
	return &GRPCClient{client: proto.NewKRailPluginClient(c)}, nil
}

func PluginsFromDirectory(directory string) ([]Plugin, error) {
	binaries, err := filepath.Glob(directory)
	if err != nil {
		return []Plugin{}, err
	}

	pluginClients := []Plugin{}
	for _, binary := range binaries {
		pluginClient, err := LaunchPluginProcess(binary)
		if err != nil {
			return pluginClients, err

		}
		pluginClients = append(pluginClients, *pluginClient)
	}
	return pluginClients, nil
}

var Handshake = plugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "K_RAIL_PLUGIN",
	MagicCookieValue: "TRUE",
}

const GRPCPluginName = "K_RAIL_GRPC"

var pluginMap = map[string]plugin.Plugin{
	GRPCPluginName: &KRailGRPCPlugin{},
}

func LaunchPluginProcess(binaryPath string) (*Plugin, error) {
	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig:  Handshake,
		Plugins:          pluginMap,
		Cmd:              exec.Command(binaryPath),
		AllowedProtocols: []plugin.Protocol{plugin.ProtocolGRPC},
	})

	// Connect via RPC
	rpcClient, err := client.Client()
	if err != nil {
		return nil, err
	}

	// Request the plugin
	raw, err := rpcClient.Dispense(GRPCPluginName)
	if err != nil {
		return nil, err
	}

	krailPlugin := raw.(KRailPlugin)

	pluginName, err := krailPlugin.PluginName()
	if err != nil {
		return nil, err
	}

	policyNames, err := krailPlugin.PolicyNames()
	if err != nil {
		return nil, err
	}

	return &Plugin{
		name:        pluginName,
		policyNames: policyNames,
		client:      *client,
		kRailPlugin: krailPlugin,
	}, nil
}
