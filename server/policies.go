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

	"github.com/cruise-automation/k-rail/policies"
	"github.com/cruise-automation/k-rail/policies/ingress"
	"github.com/cruise-automation/k-rail/policies/pod"
	log "github.com/sirupsen/logrus"
	admissionv1beta1 "k8s.io/api/admission/v1beta1"
)

// Policy specifies how a Policy is implemented
type Policy interface {

	// Name returns the name of the policy for rendering and for referencing configuration.
	Name() string

	// Validate is called if the Policy is enabled to detect violations. Violations will cause a resource to be
	// blocked unless there is an exemption for it.
	Apply(ctx context.Context,
		config policies.Config,
		ar *admissionv1beta1.AdmissionRequest,
	) ([]policies.ResourceViolation, []policies.PatchOperation)

	// Action will be called if the Policy is in violation and not in report-only mode.
	Action(ctx context.Context,
		exempt bool,
		config policies.Config,
		ar *admissionv1beta1.AdmissionRequest,
	) error
}

func (s *Server) registerPolicies() {
	// Policies will be run in the order that they are registered.
	// Policies that mutate will have their resulting patch merged with any previous patches in that order as well.

	s.registerPolicy(pod.PolicyNoExec{})
	s.registerPolicy(pod.PolicyBindMounts{})
	s.registerPolicy(pod.PolicyDockerSock{})
	s.registerPolicy(pod.PolicyImageImmutableReference{})
	s.registerPolicy(pod.PolicyNoTiller{})
	s.registerPolicy(pod.PolicyTrustedRepository{})
	s.registerPolicy(pod.PolicyNoHostNetwork{})
	s.registerPolicy(pod.PolicyNoPrivilegedContainer{})
	s.registerPolicy(pod.PolicyNoNewCapabilities{})
	s.registerPolicy(pod.PolicyNoHostPID{})
	s.registerPolicy(pod.PolicySafeToEvict{})
	s.registerPolicy(pod.PolicyMutateSafeToEvict{})
	s.registerPolicy(ingress.PolicyRequireIngressExemption{})
}

func (s *Server) registerPolicy(v Policy) {
	found := false
	for _, val := range s.Config.Policies {
		if val.Name == v.Name() {
			found = true
			if val.Enabled {
				if s.Config.GlobalReportOnly {
					s.ReportOnlyPolicies = append(s.ReportOnlyPolicies, v)
					log.Infof("enabling %s validator in REPORT ONLY mode because GLOBAL REPORT ONLY MODE is on", v.Name())
				} else if val.ReportOnly {
					s.ReportOnlyPolicies = append(s.ReportOnlyPolicies, v)
					log.Infof("enabling %s validator in REPORT ONLY mode", v.Name())
				} else {
					s.EnforcedPolicies = append(s.EnforcedPolicies, v)
					log.Infof("enabling %s validator in ENFORCE mode", v.Name())
				}
			} else {
				log.Infof("validator %s is NOT ENABLED", v.Name())

			}
		}
	}
	if !found {
		s.ReportOnlyPolicies = append(s.ReportOnlyPolicies, v)
		log.Warnf("configuration not present for %s validator, enabling REPORT ONLY mode", v.Name())
	}
}
