package server

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"
	admissionv1 "k8s.io/api/admission/v1"
)

func Check() {
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

	loadedPlugins, err := loadPlugins(pluginsPathGlob, cfg)
	if err != nil {
		log.Fatal(err)
	}

	srv := Server{
		Config:     cfg,
		Exemptions: exemptions,
		Plugins:    loadedPlugins,
	}

	srv.registerPolicies()

	inputFile := os.Args[1]
	stat, err := os.Stat(inputFile)
	if err != nil {
		log.Fatalf("cannot stat %s: %s", inputFile, err)
	}

	allowed := true

	results := make(map[string][]admissionv1.AdmissionReview)

	if stat.IsDir() {
		err := filepath.Walk(inputFile, func(path string, info fs.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() {
				ext := filepath.Ext(path)
				if ext == ".yml" || ext == ".yaml" {
					reviews, err := srv.validateFile(path)
					if err != nil {
						log.Errorf("error validating %s: %s", path, err)
						allowed = false
					}
					results[path] = reviews
				}
			}
			return nil
		})
		if err != nil {
			log.Errorf("error walking %s: %s", inputFile, err)
			allowed = false
		}
	} else {
		reviews, err := srv.validateFile(inputFile)
		if err != nil {
			log.Errorf("error validating %s: %s", inputFile, err)
			allowed = false
		}
		results[inputFile] = reviews
	}

	for filename, reviews := range results {
		for _, review := range reviews {
			if !review.Response.Allowed {
				allowed = false
				// validateFile can be refactored to return some struct to avoid parsing response message
				for _, violation := range strings.Split(review.Response.Result.Message, "\n") {
					if violation != "" {
						fmt.Printf("FAIL - %s - %s - %s\n", filename, review.Request.Name, violation)
					}
				}
			}
		}
	}

	if !allowed {
		os.Exit(1)
	}
}
