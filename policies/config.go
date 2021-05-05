// Copyright 2019 Cruise LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//    https://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package policies

import (
	"encoding/json"
	"errors"
	"fmt"

	apiresource "k8s.io/apimachinery/pkg/api/resource"
)

// Config contains configuration for Policies
type Config struct {
	// PolicyRequireIngressExemptionClasses contains the Ingress classes that an exemption is required for
	// to use. Typically this would include your public ingress classes.
	PolicyRequireIngressExemptionClasses []string `json:"policy_require_ingress_exemption_classes"`
	// PolicyRequireServiceLoadBalancerAnnotations contains the Service LB types annotation that are allowed with this policy.
	PolicyRequireServiceLoadBalancerAnnotations []*AnnotationConfig `json:"policy_require_service_loadbalancer_annotations"`
	// PolicyTrustedRepositoryRegexes contains regexes that match image repositories that you want to allow.
	PolicyTrustedRepositoryRegexes []string `json:"policy_trusted_repository_regexes"`
	// PolicyDefaultSeccompPolicy contains the seccomp policy that you want to be applied on Pods by default.
	// Defaults to 'runtime/default'
	PolicyDefaultSeccompPolicy string `json:"policy_default_seccomp_policy"`
	// PolicyImagePullPolicy contains the images that needs to enforce to a specific ImagePullPolicy
	PolicyImagePullPolicy   map[string][]string     `json:"mutate_image_pull_policy"`
	MutateEmptyDirSizeLimit MutateEmptyDirSizeLimit `json:"mutate_empty_dir_size_limit"`
}

// AnnotationConfig defines a single annotation config
type AnnotationConfig struct {
	Annotation    string   `json:"annotation"`
	Annotations   []string `json:"annotations"`
	AllowedValues []string `json:"allowed_values"`
	AllowMissing  bool     `json:"allow_missing"`
}

type MutateEmptyDirSizeLimit struct {
	MaximumSizeLimit apiresource.Quantity `json:"maximum_size_limit"`
	DefaultSizeLimit apiresource.Quantity `json:"default_size_limit"`
}

func (m *MutateEmptyDirSizeLimit) UnmarshalJSON(value []byte) error {
	var v map[string]json.RawMessage
	if err := json.Unmarshal(value, &v); err != nil {
		return err
	}

	if max, ok := v["maximum_size_limit"]; ok {
		if err := m.MaximumSizeLimit.UnmarshalJSON(max); err != nil {
			return fmt.Errorf("maximum_size_limit failed: %s", err)
		}
	}
	if def, ok := v["default_size_limit"]; ok {
		if err := m.DefaultSizeLimit.UnmarshalJSON(def); err != nil {
			return fmt.Errorf("default_size_limit failed: %s", err)
		}
	}
	if m.DefaultSizeLimit.IsZero() {
		return errors.New("default size must not be empty")
	}
	if m.MaximumSizeLimit.IsZero() {
		return errors.New("max size must not be empty")
	}
	if m.DefaultSizeLimit.Cmp(m.MaximumSizeLimit) > 0 {
		return errors.New("default size must not be greater than max size")
	}
	return nil
}
