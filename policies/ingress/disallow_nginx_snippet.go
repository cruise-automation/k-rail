// Copyright 2021 Cruise LLC
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

package ingress

import (
	"context"
	"fmt"
	"regexp"

	admissionv1 "k8s.io/api/admission/v1"

	"github.com/cruise-automation/k-rail/v3/policies"
	"github.com/cruise-automation/k-rail/v3/resource"
)

var nginxSnippetAnnotationRegex = regexp.MustCompile("^nginx.ingress.kubernetes.io/.*-snippet$")

type PolicyDisallowNGINXSnippet struct{}

func (p PolicyDisallowNGINXSnippet) Name() string {
	return "ingress_disallow_nginx_snippet"
}

func (p PolicyDisallowNGINXSnippet) Validate(ctx context.Context, config policies.Config, ar *admissionv1.AdmissionRequest) ([]policies.ResourceViolation, []policies.PatchOperation) {

	resourceViolations := []policies.ResourceViolation{}

	ingressResource := resource.GetIngressResource(ctx, ar)
	if ingressResource == nil {
		return resourceViolations, nil
	}

	if ingressResource.IngressExt.Annotations == nil {
		return resourceViolations, nil
	}
	for key := range ingressResource.IngressExt.Annotations {
		if nginxSnippetAnnotationRegex.MatchString(key) {
			resourceViolations = append(resourceViolations, policies.ResourceViolation{
				Namespace:    ar.Namespace,
				ResourceName: ingressResource.ResourceName,
				ResourceKind: ingressResource.ResourceKind,
				Violation:    fmt.Sprintf("NGINX Snippets are not allowed, found %q", key),
				Policy:       p.Name(),
			})
		}
	}
	return resourceViolations, nil
}
