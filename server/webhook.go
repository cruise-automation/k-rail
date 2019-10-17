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
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"encoding/json"

	"github.com/cruise-automation/k-rail/policies"

	"k8s.io/api/admission/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
)

var (
	runtimeScheme = runtime.NewScheme()
	codecs        = serializer.NewCodecFactory(runtimeScheme)
)

func writeAdmissionError(w http.ResponseWriter, ar v1beta1.AdmissionReview, e error) {
	w.WriteHeader(http.StatusBadRequest)
	ar.Response = &v1beta1.AdmissionResponse{
		Result: &metav1.Status{
			Message: e.Error(),
		},
	}
	payload, _ := json.Marshal(ar)
	w.Write(payload)
}

// ValidatingWebhook is a ValidatingWebhook endpoint that accepts K8s resources to process
func (s *Server) ValidatingWebhook(w http.ResponseWriter, r *http.Request) {

	ar := v1beta1.AdmissionReview{
		Response: &v1beta1.AdmissionResponse{
			Allowed: true,
			Result: &metav1.Status{
				Reason:  "k-rail admission review",
				Message: "errored while processing review",
			},
		},
	}

	// require application/json content-type
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		s.LogAndPrintError("wrong content type", fmt.Errorf("contentType=%s, expect application/json", contentType))
		writeAdmissionError(w, ar, errors.New("incorrect content type"))
	}

	// set the response content-type
	w.Header().Set("Content-Type", "application/json")

	// safely read the body into memory
	body, err := ioutil.ReadAll(http.MaxBytesReader(w, r.Body, 1024*1024))
	if err != nil {
		s.LogAndPrintError("error reading body", err)
		writeAdmissionError(w, ar, err)
	}
	defer r.Body.Close()

	if strings.ToLower(s.Config.LogLevel) == "debug" {
		fmt.Printf("%s\n", body)
	}

	// unmarshall review request
	deserializer := codecs.UniversalDeserializer()
	if _, _, err = deserializer.Decode(body, nil, &ar); err != nil {
		s.LogAndPrintError("error unmarshalling review request", err)
		writeAdmissionError(w, ar, err)
	}

	// validate the resources
	// this slips any results into the review structure
	ar = s.validateResources(ar)

	// write the review and results JSON response
	payload, err := json.Marshal(ar)
	if err != nil {
		s.LogAndPrintError("could not marshal response", err)
		writeAdmissionError(w, ar, err)
	}
	w.Write(payload)
}

// validateResources accepts K8s resources to process
func (s *Server) validateResources(ar v1beta1.AdmissionReview) v1beta1.AdmissionReview {
	ctx, cancelfn := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelfn()

	if ar.Request == nil {
		log.Warnf("got empty AdmissionRequest in AdmissionReview: %+v", ar)
		return ar
	}

	enforcedViolations := []policies.ResourceViolation{}
	reportedViolations := []policies.ResourceViolation{}
	exemptViolations := []policies.ResourceViolation{}

	// allow resource if namespace is blacklisted
	for _, namespace := range s.Config.BlacklistedNamespaces {
		if namespace == ar.Request.Namespace {
			ar.Response = &v1beta1.AdmissionResponse{
				Allowed: true,
				Result: &metav1.Status{
					Reason:  "k-rail admission review",
					Message: "blacklisted namespace",
				},
			}
			return ar
		}
	}

	for _, val := range s.EnforcedPolicies {
		violations := val.Validate(ctx, s.Config.PolicyConfig, ar.Request)
		if len(violations) > 0 {
			if policies.IsExempt(
				violations[0].ResourceName,
				ar.Request.Namespace,
				ar.Request.UserInfo,
				val.Name(),
				s.Exemptions,
			) {
				exemptViolations = append(exemptViolations,
					violations...)
			} else {
				enforcedViolations = append(enforcedViolations,
					violations...)
			}
		}
	}

	for _, val := range s.ReportOnlyPolicies {
		violations := val.Validate(ctx, s.Config.PolicyConfig, ar.Request)
		if len(violations) > 0 {
			if policies.IsExempt(
				violations[0].ResourceName,
				ar.Request.Namespace,
				ar.Request.UserInfo,
				val.Name(),
				s.Exemptions,
			) {
				exemptViolations = append(exemptViolations,
					violations...)
			} else {
				reportedViolations = append(reportedViolations,
					violations...)
			}
		}
	}

	// log exempt violations
	for _, v := range exemptViolations {
		log.WithFields(log.Fields{
			"kind":      v.ResourceKind,
			"resource":  v.ResourceName,
			"namespace": v.Namespace,
			"policy":    v.Policy,
			"user":      ar.Request.UserInfo.Username,
			"enforced":  false,
		}).Info("EXEMPT")
	}

	// log report-only violations
	for _, v := range reportedViolations {
		log.WithFields(log.Fields{
			"kind":      v.ResourceKind,
			"resource":  v.ResourceName,
			"namespace": v.Namespace,
			"policy":    v.Policy,
			"user":      ar.Request.UserInfo.Username,
			"enforced":  false,
		}).Info("NOT ENFORCED")
	}

	// log enforced violations when in global report-only mode
	if s.Config.GlobalReportOnly {
		for _, v := range enforcedViolations {
			log.WithFields(log.Fields{
				"kind":      v.ResourceKind,
				"resource":  v.ResourceName,
				"namespace": v.Namespace,
				"policy":    v.Policy,
				"user":      ar.Request.UserInfo.Username,
				"enforced":  false,
			}).Info("NOT ENFORCED")
		}
	}

	// log and respond to enforced violations
	if len(enforcedViolations) > 0 && s.Config.GlobalReportOnly == false {
		violations := ""
		for _, v := range enforcedViolations {
			log.WithFields(log.Fields{
				"kind":      v.ResourceKind,
				"resource":  v.ResourceName,
				"namespace": v.Namespace,
				"policy":    v.Policy,
				"user":      ar.Request.UserInfo.Username,
				"enforced":  true,
			}).Warn("ENFORCED")
			violations = violations + "\n" + v.HumanString()
		}

		ar.Response = &v1beta1.AdmissionResponse{
			UID:     ar.Request.UID,
			Allowed: false,
			Result: &metav1.Status{
				Reason:  "k-rail admission review",
				Message: violations,
			},
		}
		return ar
	}

	// allow other resources, but include any reported violations
	violations := ""
	for _, v := range reportedViolations {
		violations = violations + "\n" + v.HumanString()
	}
	if len(violations) != 0 {
		violations = "NOT ENFORCED:\n" + violations
	} else {
		violations = "NO VIOLATIONS"
	}

	ar.Response = &v1beta1.AdmissionResponse{
		UID:     ar.Request.UID,
		Allowed: true,
		Result: &metav1.Status{
			Reason:  "k-rail admission review",
			Message: violations,
		},
	}

	return ar
}
