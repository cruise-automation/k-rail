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

package server

import (
	"github.com/cruise-automation/k-rail/policies"
)

type PolicySettings struct {
	Name       string
	Enabled    bool
	ReportOnly bool `yaml:"report_only"`
}

type Config struct {
	LogLevel              string   `yaml:"log_level"`
	BlacklistedNamespaces []string `yaml:"blacklisted_namespaces"`
	TLS                   struct {
		Cert string
		Key  string
	}
	GlobalReportOnly bool `yaml:"global_report_only"`
	Policies         []PolicySettings
	PolicyConfig     policies.Config `yaml:"policy_config"`
}
