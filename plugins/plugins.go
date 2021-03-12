package plugins

import (
	"context"
	"os/exec"
	"path/filepath"

	"github.com/hashicorp/go-plugin"
	"google.golang.org/grpc"
	admissionv1 "k8s.io/api/admission/v1"

	"github.com/cruise-automation/k-rail/plugins/proto"
	"github.com/cruise-automation/k-rail/policies"
)

type Plugin struct {
	Name        string
	PolicyNames []string
	Client      plugin.Client
	KRailPlugin KRailPlugin
}

func (p *Plugin) Configure(config map[string]interface{}) error {
	return p.KRailPlugin.ConfigurePlugin(config)
}

func (p *Plugin) Validate(ar *admissionv1.AdmissionRequest) (map[string]PolicyValidateResponse, error) {
	return p.KRailPlugin.Validate(ar)
}

type PolicyValidateResponse struct {
	ResourceViolations []policies.ResourceViolation
	PatchOperations    []policies.PatchOperation
}

// KRailPlugin is the interface that we're exposing as a plugin.
type KRailPlugin interface {
	PluginName() (string, error)
	PolicyNames() ([]string, error)
	ConfigurePlugin(config map[string]interface{}) error
	Validate(ar *admissionv1.AdmissionRequest) (map[string]PolicyValidateResponse, error)
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

var handshake = plugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "K_RAIL_PLUGIN",
	MagicCookieValue: "TRUE",
}

const grpcPluginName = "K_RAIL_GRPC"

var pluginMap = map[string]plugin.Plugin{
	grpcPluginName: &KRailGRPCPlugin{},
}

func LaunchPluginProcess(binaryPath string) (*Plugin, error) {
	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig:  handshake,
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
	raw, err := rpcClient.Dispense(grpcPluginName)
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
		Name:        pluginName,
		PolicyNames: policyNames,
		Client:      *client,
		KRailPlugin: krailPlugin,
	}, nil
}
