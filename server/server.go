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

	"github.com/cruise-automation/k-rail/policies"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	yaml "gopkg.in/yaml.v2"
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
		log.Fatal("error loading yaml config: ", err)
	}
	err = yaml.Unmarshal(yamlFile, &cfg)
	if err != nil {
		log.Fatal("error unmarshalling yaml config: ", err)
	}

	if len(cfg.LogLevel) == 0 {
		log.SetLevel(log.InfoLevel)
	} else {
		level, err := log.ParseLevel(cfg.LogLevel)
		if err != nil {
			log.Fatal("invalid log level set: ", err)
		}
		log.SetLevel(level)
	}

	var exemptions []policies.CompiledExemption
	if *exemptionsPathGlob != "" {
		exemptions, err = policies.ExemptionsFromDirectory(*exemptionsPathGlob)
		if err != nil {
			log.Fatal("error loading exemptions: ", err)
		}
		log.Infof("loaded %d exemptions", len(exemptions))
	}

	log.Info("k-rail is running")

	srv := Server{
		Config:     cfg,
		Exemptions: exemptions,
	}

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

func (s *Server) LogAndPrintError(user string, err error) {
	log.Error(err)
}
