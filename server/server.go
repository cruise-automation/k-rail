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
	"context"
	"flag"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cruise-automation/k-rail/actions"
	"github.com/cruise-automation/k-rail/policies"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	yaml "gopkg.in/yaml.v2"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

const (
	serviceName = "k-rail"
)

// Server contains configuration state needed for the API server
type Server struct {
	Config             Config
	RequestedShutdown  bool
	EnforcedPolicies   []Policy
	ReportOnlyPolicies []Policy
	Exemptions         []policies.CompiledExemption
	KubernetesClient   *kubernetes.Clientset
}

// Run starts the API server
func Run(ctx context.Context) {
	log.SetFormatter(&log.JSONFormatter{})
	log.SetLevel(log.InfoLevel)

	cfg := Config{}

	configPath := flag.String("config", "config.yml", "path to configuration file")
	exemptionsPathGlob := flag.String("exemptions-path-glob", "", "path glob that includes exemption configs")
	flag.Parse()

	yamlFile, err := ioutil.ReadFile(*configPath)
	if err != nil {
		log.WithError(err).Fatal("error loading yaml config")
	}
	err = yaml.Unmarshal(yamlFile, &cfg)
	if err != nil {
		log.WithError(err).Fatal("error unmarshalling yaml config")
	}

	if level, err := log.ParseLevel(cfg.LogLevel); err != nil {
		log.SetLevel(level)
	} else {
		log.Fatal("invalid log level set: ", err)
	}

	var exemptions []policies.CompiledExemption
	if *exemptionsPathGlob != "" {
		exemptions, err = policies.ExemptionsFromDirectory(*exemptionsPathGlob)
		if err != nil {
			log.WithError(err).Fatal("error loading exemptions")

		}
		log.Infof("loaded %d exemptions", len(exemptions))
	}

	config, err := rest.InClusterConfig()
	if err != nil {
		log.WithError(err).Fatal("could not configure kubernetes client")
	}
	kubernetesClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.WithError(err).Fatal("could not make kubernetes client")
	}

	// tainted Pod janitor for execs
	if cfg.PolicyConfig.PolicyNoExec.DeleteTaintedPodsAfter != time.Duration(0) {
		ticker := time.NewTicker(time.Minute * 5)
		done := make(chan bool)
		go func() {
			for {
				select {
				case <-done:
					return
				case <-ticker.C:
					log.Info("running exec taint janitor task")
					actions.DeletePodsByAnnotationAfterDuration(
						kubernetesClient,
						"k-rail.cruise-automation.github.com/taint/exec",
						cfg.PolicyConfig.PolicyNoExec.DeleteTaintedPodsAfter)
				}
			}
		}()
	}

	srv := Server{
		Config:           cfg,
		Exemptions:       exemptions,
		KubernetesClient: kubernetesClient,
	}

	log.Info("k-rail is running")

	srv.registerPolicies()

	router := mux.NewRouter()
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
		log.Fatal("got empty certificate")
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

	readinessServer := http.NewServeMux()
	readinessServer.HandleFunc("/", srv.readinessFunc)
	go func() {
		log.Fatal(http.ListenAndServe(":8000", readinessServer))
	}()

	log.Fatal(s.ListenAndServeTLS(cfg.TLS.Cert, cfg.TLS.Key))
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
