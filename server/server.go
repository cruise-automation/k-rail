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
	"context"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/cruise-automation/k-rail/v3/plugins"
	"github.com/cruise-automation/k-rail/v3/policies"
	"github.com/gertd/go-pluralize"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"k8s.io/api/admission/v1"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/json"
	"k8s.io/apimachinery/pkg/util/yaml"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/slok/go-http-metrics/middleware/std"
)

const (
	serviceName = "k-rail"
	yamlBufSize = 10 * 1024
)

// Server contains configuration state needed for the API server
type Server struct {
	Config             Config
	RequestedShutdown  bool
	EnforcedPolicies   []Policy
	ReportOnlyPolicies []Policy
	Exemptions         []policies.CompiledExemption
	Plugins            []plugins.Plugin
}

// Run starts the API server
func Run(ctx context.Context) {
	log.SetFormatter(&log.JSONFormatter{})
	log.SetLevel(log.InfoLevel)

	cfg := Config{}

	configPath, exemptionsPathGlob, pluginsPathGlob := parseFlags()

	err := cfg.load(configPath)
	if err != nil {
		log.Fatal(err)
	}

	exemptions, err := loadExemptions(exemptionsPathGlob)
	if err != nil {
		log.Fatal(err)
	}

	registerMetrics()

	loadedPlugins, err := loadPlugins(pluginsPathGlob, cfg)
	if err != nil {
		log.Fatal(err)
	}

	log.Info("k-rail is running")

	srv := Server{
		Config:     cfg,
		Exemptions: exemptions,
		Plugins:    loadedPlugins,
	}

	srv.registerPolicies()

	router := mux.NewRouter()
	router.Use(std.HandlerProvider("", prometheusMiddleware))
	router.HandleFunc("/", srv.ValidatingWebhook)

	s := &http.Server{
		Addr:              ":10250",
		Handler:           router,
		ReadTimeout:       10 * time.Second,
		ReadHeaderTimeout: 10 * time.Second,
		WriteTimeout:      10 * time.Second,
		MaxHeaderBytes:    1 << 20,
	}

	certBytes, err := ioutil.ReadFile(cfg.TLS.Cert)
	if len(certBytes) == 0 || err != nil {
		log.WithError(err).Fatal("got empty certificate")
	}

	// on ^C, or SIGTERM handle safe shutdown
	signals := make(chan os.Signal, 0)
	signal.Notify(signals, os.Interrupt)
	signal.Notify(signals, syscall.SIGTERM)
	go func() {
		for sig := range signals {
			log.Warnf("received %s signal, failing healthcheck to divert traffic", sig.String())
			srv.RequestedShutdown = true
			log.Warn("shutting down in 15 seconds")
			time.Sleep(15 * time.Second)
			os.Exit(0)
		}
	}()

	// serve metrics
	prometheusServer := http.NewServeMux()
	prometheusServer.Handle("/metrics", promhttp.Handler())
	go func() {
		log.Info("metrics listening at :2112")
		if err := http.ListenAndServe(":2112", prometheusServer); err != nil {
			log.Fatalf("error while serving metrics: %s", err)
		}
	}()

	readinessServer := http.NewServeMux()
	readinessServer.HandleFunc("/", srv.readinessFunc)
	go func() {
		log.Fatal(http.ListenAndServe(":8000", readinessServer))
	}()

	log.Fatal(s.ListenAndServeTLS(cfg.TLS.Cert, cfg.TLS.Key))
}

func registerMetrics() {
	prometheus.MustRegister(totalRegisteredPolicies)
	prometheus.MustRegister(policyViolations)
	prometheus.MustRegister(totalLoadedPlugins)
}

func parseFlags() (string, string, string) {
	configPath := flag.String("config", "config.yml", "path to configuration file")
	exemptionsPathGlob := flag.String("exemptions-path-glob", "", "path glob that includes exemption configs")
	pluginsPathGlob := flag.String("plugins-path-glob", "", "path glob that includes plugin binaries")
	flag.Parse()
	return *configPath, *exemptionsPathGlob, *pluginsPathGlob
}

func loadPlugins(pluginsPathGlob string, cfg Config) ([]plugins.Plugin, error) {
	var loadedPlugins []plugins.Plugin
	var err error
	if pluginsPathGlob != "" {
		loadedPlugins, err = plugins.PluginsFromDirectory(pluginsPathGlob)
		if err != nil {
			return loadedPlugins, fmt.Errorf("error launching plugins: %s", err)
		}

		for _, plugin := range loadedPlugins {
			defer plugin.Kill()
			if pluginConfig, ok := cfg.PluginConfig[plugin.Name()]; ok {
				if pluginConfigMap, ok := pluginConfig.(map[string]interface{}); ok {
					err = plugin.Configure(pluginConfigMap)
					if err != nil {
						return loadedPlugins, fmt.Errorf("error configuring plugin %s: %v\n", plugin.Name(), err)
					}
				} else {
					return loadedPlugins, fmt.Errorf("expected plugin config for plugin %s to be a map of values (eg. plugins_config: %s: <config key>: <config values>)", plugin.Name(), plugin.Name())
				}
			} else {
				log.Infof("no plugin config found for plugin %s, continuing on", plugin.Name())
			}

			totalLoadedPlugins.Add(1)
		}

		log.Infof("loaded %d plugins", len(loadedPlugins))
	}
	return loadedPlugins, err
}

func loadExemptions(exemptionsPathGlob string) ([]policies.CompiledExemption, error) {
	var exemptions []policies.CompiledExemption
	var err error
	if exemptionsPathGlob != "" {
		exemptions, err = policies.ExemptionsFromDirectory(exemptionsPathGlob)
		if err != nil {
			return exemptions, fmt.Errorf("error loading exemptions: %s", err)
		}
		log.Infof("loaded %d exemptions", len(exemptions))
	}
	return exemptions, err
}

func (s *Server) readinessFunc(w http.ResponseWriter, r *http.Request) {
	if s.RequestedShutdown {
		w.WriteHeader(http.StatusGone)
		w.Write([]byte("shutting down"))
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

func (s *Server) LogAndPrintError(user string, err error) {
	log.Error(err)
}

// validateFile validates all resources in file
func (s *Server) validateFile(inputFile string) ([]v1.AdmissionReview, error) {
	var results []v1.AdmissionReview
	// Read input manifest file
	inputFileReader, err := os.Open(inputFile)
	if err != nil {
		return results, fmt.Errorf("read input file: %s", err)
	}

	yamlReader := yaml.NewDocumentDecoder(inputFileReader)
	yamlBuf := make([]byte, yamlBufSize)

	for {
		yamlSize, err := yamlReader.Read(yamlBuf)
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return results, fmt.Errorf("error reading yaml document: %s", err)
			}
		}

		raw, err := yaml.ToJSON(yamlBuf[:yamlSize])
		if err != nil {
			return results, fmt.Errorf("error converting yaml to json: %s", err)
		}
		var partialObject v12.PartialObjectMetadata
		err = json.Unmarshal(raw, &partialObject)
		if err != nil {
			return results, fmt.Errorf("error unmarshaling object from json: %s", err)
		}
		gvk := partialObject.GroupVersionKind()
		pc := pluralize.NewClient()
		resource := strings.ToLower(gvk.Kind)
		if pc.IsSingular(gvk.Kind) {
			resource = strings.ToLower(pc.Plural(gvk.Kind))
		}
		ar := v1.AdmissionReview{
			Request: &v1.AdmissionRequest{
				Namespace: partialObject.GetNamespace(),
				Name:      partialObject.GetName(),
				Object:    runtime.RawExtension{Raw: raw},
				Resource:  v12.GroupVersionResource{Group: gvk.Group, Version: gvk.Version, Resource: resource},
				Kind:      v12.GroupVersionKind{Group: gvk.Group, Version: gvk.Version, Kind: gvk.Kind},
			},
		}
		results = append(results, s.validateResources(ar))
		continue
	}
	return results, nil
}
