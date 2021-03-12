package plugins

import (
	"encoding/json"
	"errors"

	"golang.org/x/net/context"
	structpb "google.golang.org/protobuf/types/known/structpb"
	admissionv1 "k8s.io/api/admission/v1"

	"github.com/cruise-automation/k-rail/plugins/proto"
	"github.com/cruise-automation/k-rail/policies"
)

// GRPCClient is an implementation of KRailPlugin that talks over RPC.
type GRPCClient struct{ client proto.KRailPluginClient }

func (m *GRPCClient) PluginName() (string, error) {
	resp, err := m.client.PluginName(context.Background(), &proto.PluginNameRequest{})
	if err != nil {
		return "", err
	}
	return resp.PluginName, nil
}

func (m *GRPCClient) PolicyNames() ([]string, error) {
	resp, err := m.client.PolicyNames(context.Background(), &proto.PolicyNamesRequest{})
	if err != nil {
		return []string{}, err
	}
	return resp.PolicyNames, nil
}

func (m *GRPCClient) ConfigurePlugin(config map[string]interface{}) error {
	configStruct, err := structpb.NewStruct(config)
	if err != nil {
		return err
	}
	_, err = m.client.ConfigurePlugin(context.Background(), &proto.ConfigurePluginRequest{
		PluginConfig: configStruct,
	})
	return err
}

func (m *GRPCClient) Validate(ar *admissionv1.AdmissionRequest) (map[string]PolicyValidateResponse, error) {
	arJson, err := json.Marshal(ar)
	if err != nil {
		return map[string]PolicyValidateResponse{}, err
	}
	resp, err := m.client.Validate(context.Background(), &proto.ValidateRequest{
		AdmissionRequest: arJson,
	})
	if err != nil {
		return map[string]PolicyValidateResponse{}, err
	}
	policyResponses := map[string]PolicyValidateResponse{}
	for policy, policyResponse := range resp.PolicyResponses {
		resourceViolations := []policies.ResourceViolation{}
		for _, violation := range policyResponse.ResourceViolations {
			resourceViolations = append(resourceViolations, policies.ResourceViolation{
				ResourceName: violation.ResourceName,
				ResourceKind: violation.ResourceKind,
				Namespace:    violation.Namespace,
				Violation:    violation.Violation,
				Policy:       violation.Policy,
				Error:        errors.New(violation.Error),
			})
		}
		patchOperations := []policies.PatchOperation{}
		for _, patchOp := range policyResponse.PatchOperations {
			patchOperations = append(patchOperations, policies.PatchOperation{
				Op: patchOp.Op, Path: patchOp.Path,
				Value: patchOp.Value.AsInterface(),
			})
		}
		policyResponses[policy] = PolicyValidateResponse{
			ResourceViolations: resourceViolations,
			PatchOperations:    patchOperations,
		}
	}
	return policyResponses, err
}

// Here is the gRPC server that GRPCClient talks to.
type GRPCServer struct {
	// This is the real implementation
	Impl KRailPlugin
}

func (m *GRPCServer) PluginName(ctx context.Context, in *proto.PluginNameRequest) (*proto.PluginNameResponse, error) {
	pluginName, err := m.Impl.PluginName()
	return &proto.PluginNameResponse{PluginName: pluginName}, err
}

func (m *GRPCServer) PolicyNames(ctx context.Context, in *proto.PolicyNamesRequest) (*proto.PolicyNamesResponse, error) {
	policyNames, err := m.Impl.PolicyNames()
	return &proto.PolicyNamesResponse{PolicyNames: policyNames}, err
}

func (m *GRPCServer) ConfigurePlugin(ctx context.Context, in *proto.ConfigurePluginRequest) (*proto.ConfigurePluginResponse, error) {
	return &proto.ConfigurePluginResponse{}, m.Impl.ConfigurePlugin(in.PluginConfig.AsMap())
}

func (m *GRPCServer) Validate(ctx context.Context, in *proto.ValidateRequest) (*proto.ValidateResponse, error) {
	var ar admissionv1.AdmissionRequest
	err := json.Unmarshal(in.AdmissionRequest, &ar)
	if err != nil {
		return nil, err
	}
	resps, err := m.Impl.Validate(&ar)
	if err != nil {
		return nil, err
	}

	policyResponses := map[string]*proto.PolicyValidateResponse{}
	for policy, policyResponse := range resps {
		violations := []*proto.ResourceViolation{}
		for _, violation := range policyResponse.ResourceViolations {
			violations = append(violations, &proto.ResourceViolation{
				ResourceName: violation.ResourceName,
				ResourceKind: violation.ResourceKind,
				Namespace:    violation.Namespace,
				Violation:    violation.Violation,
				Policy:       violation.Policy,
				Error:        violation.Error.Error(),
			})
		}
		patchOps := []*proto.PatchOperation{}
		for _, patchOp := range policyResponse.PatchOperations {
			valueStruct, err := structpb.NewValue(patchOp.Value)
			if err != nil {
				return nil, err
			}
			patchOps = append(patchOps, &proto.PatchOperation{
				Op:    patchOp.Op,
				Path:  patchOp.Path,
				Value: valueStruct,
			})
		}
		policyResponses[policy] = &proto.PolicyValidateResponse{
			ResourceViolations: violations,
			PatchOperations:    patchOps,
		}
	}
	return &proto.ValidateResponse{PolicyResponses: policyResponses}, err
}
