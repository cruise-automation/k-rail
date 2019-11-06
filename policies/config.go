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

import "time"

// Config contains configuration for Policies
type Config struct {
	// PolicyRequireIngressExemptionClasses contains the Ingress classes that an exemption is required for
	// to use. Typically this would include your public ingress classes.
	PolicyRequireIngressExemptionClasses []string `yaml:"policy_require_ingress_exemption_classes"`
	// PolicyTrustedRepositoryRegexes contains regexes that match image repositories that you want to allow.
	PolicyTrustedRepositoryRegexes []string `yaml:"policy_trusted_repository_regexes"`
	// PolicyNoExec controls the configuration for the No Exec policy
	PolicyNoExec struct {
		// LabelTaintedPods controls whether pods that have been execed into get labeled as tainted
		// with the annotation k-rail.cruise-automation.github.com/taint/exec=<epoch time>
		LabelTaintedPods bool
		// DeleteTaintedPodsAfter controls the duration after the pod exec event that the Pod will be deleted.
		// Setting the duration to 0 will disable deletion
		DeleteTaintedPodsAfter time.Duration
	} `yaml:"policy_no_exec"`
}
