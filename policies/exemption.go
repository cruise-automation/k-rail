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
// limitations under the License.package ingress

package policies

import (
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/gobwas/glob"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	authenticationv1 "k8s.io/api/authentication/v1"
)

// RawExemption is the configuration for a policy exemption
type RawExemption struct {
	ResourceName   string   `yaml:"resource_name"`
	Namespace      string   `yaml:"namespace"`
	Username       string   `yaml:"username"`
	Group          string   `yaml:"group"`
	ExemptPolicies []string `yaml:"exempt_policies"`
}

// CompiledExemption is the compiled configuration for a policy exemption
type CompiledExemption struct {
	ResourceName   glob.Glob
	Namespace      glob.Glob
	Username       glob.Glob
	Group          glob.Glob
	ExemptPolicies []glob.Glob
}

// Compile returns a CompiledExemption
func (r *RawExemption) Compile() CompiledExemption {
	// if not specified, assume it's the field matches all

	// ensure that ResourceName has a trailing glob so it can match the IDs added by certain resource types
	// ie, Deployment pod name test-pod, ReplicaSet name test-pod-sdf932, PodName test-pod-sdf932-ew92
	if !strings.HasSuffix(r.ResourceName, "*") {
		r.ResourceName = r.ResourceName + "*"
	}

	if r.Namespace == "" {
		r.Namespace = "*"
	}
	if r.Username == "" {
		r.Username = "*"
	}
	if r.Group == "" {
		r.Group = "*"
	}
	if len(r.ExemptPolicies) == 0 {
		r.ExemptPolicies = []string{"*"}
	}

	// compile the RawExemption
	var policies []glob.Glob
	for _, p := range r.ExemptPolicies {
		policies = append(policies, glob.MustCompile(p))
	}
	return CompiledExemption{
		ResourceName:   glob.MustCompile(r.ResourceName),
		Namespace:      glob.MustCompile(r.Namespace),
		Username:       glob.MustCompile(r.Username),
		Group:          glob.MustCompile(r.Group),
		ExemptPolicies: policies,
	}
}

// ExemptionsFromYAML returns compiled exemptions from YAML input
func ExemptionsFromYAML(exemptions []byte) ([]CompiledExemption, error) {
	var rawExemptions []RawExemption
	err := yaml.Unmarshal(exemptions, &rawExemptions)
	if err != nil {
		return []CompiledExemption{}, err
	}
	var c []CompiledExemption
	for _, e := range rawExemptions {
		log.WithFields(log.Fields{"exemption": e}).Info("loaded exemption")
		c = append(c, e.Compile())
	}
	return c, nil
}

// ExemptionsFromDirectory returns compiled exemptions a given directory
func ExemptionsFromDirectory(directory string) ([]CompiledExemption, error) {
	files, err := filepath.Glob(directory)
	if err != nil {
		return []CompiledExemption{}, err
	}

	var c []CompiledExemption
	for _, f := range files {
		contents, err := ioutil.ReadFile(f)
		if err != nil {
			return []CompiledExemption{}, err
		}
		e, err := ExemptionsFromYAML(contents)
		if err != nil {
			return []CompiledExemption{}, err
		}
		c = append(c, e...)
	}
	return c, nil
}

// IsExempt returns whether a resource is exempt from a given policy
func IsExempt(resourceName string, namespace string, userInfo authenticationv1.UserInfo, policyName string, exemptions []CompiledExemption) bool {
	for _, e := range exemptions {
		if e.Namespace.Match(namespace) &&
			e.ResourceName.Match(resourceName) &&
			e.Username.Match(userInfo.Username) {
			for _, p := range e.ExemptPolicies {
				if p.Match(policyName) {
					for _, g := range userInfo.Groups {
						if e.Group.Match(g) {
							return true
						}
					}
				}
			}
		}
	}

	return false
}
