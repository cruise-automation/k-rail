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

package server

import (
	"fmt"
	"io/ioutil"

	"github.com/cruise-automation/k-rail/v3/policies"
	log "github.com/sirupsen/logrus"
	"sigs.k8s.io/yaml"
)

type PolicySettings struct {
	Name       string
	Enabled    bool
	ReportOnly bool `json:"report_only"`
}

type Config struct {
	LogLevel              string   `json:"log_level"`
	ClusterName           string   `json:"cluster_name"`
	BlacklistedNamespaces []string `json:"blacklisted_namespaces"`
	TLS                   struct {
		Cert string
		Key  string
	}
	GlobalReportOnly     bool `json:"global_report_only"`
	GlobalMetricsEnabled bool `json:"global_metrics_enabled"`
	Policies         []PolicySettings
	PolicyConfig     policies.Config        `json:"policy_config"`
	PluginConfig     map[string]interface{} `json:"plugin_config"`
}

func (cfg *Config) load(configPath string) error {
	yamlFile, err := ioutil.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("error loading yaml config: %s", err)
	}
	err = yaml.Unmarshal(yamlFile, &cfg)
	if err != nil {
		return fmt.Errorf("error unmarshalling yaml config: %s", err)
	}

	if len(cfg.LogLevel) == 0 {
		log.SetLevel(log.InfoLevel)
	} else {
		level, err := log.ParseLevel(cfg.LogLevel)
		if err != nil {
			return fmt.Errorf("invalid log level set: %s", err)
		}
		log.SetLevel(level)
	}
	return nil
}
