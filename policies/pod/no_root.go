package pod

import (
	"context"
	"fmt"
	"github.com/cruise-automation/k-rail/v3/policies"
	"github.com/cruise-automation/k-rail/v3/resource"
	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
)

type PolicyNoRootUser struct{}

func (p PolicyNoRootUser) Name() string {
	return "pod_no_root_user"
}

func (p PolicyNoRootUser) Validate(ctx context.Context, _ policies.Config, ar *admissionv1.AdmissionRequest) ([]policies.ResourceViolation, []policies.PatchOperation) {
	resourceViolations := []policies.ResourceViolation{}
	podResource := resource.GetPodResource(ctx, ar)
	if podResource == nil {
		return resourceViolations, nil
	}
	violationText := "No Root user: Running as the root user is forbidden"

	// check the containers first
	validateSecurityContext := func(container corev1.Container) {
		if container.SecurityContext != nil && container.SecurityContext.RunAsNonRoot != nil && *container.SecurityContext.RunAsNonRoot {
			return
		} else if container.SecurityContext != nil && container.SecurityContext.RunAsUser != nil && *container.SecurityContext.RunAsUser > 0 {
			return
		}
		resourceViolations = append(resourceViolations, policies.ResourceViolation{
			Namespace:    ar.Namespace,
			ResourceName: podResource.ResourceName,
			ResourceKind: podResource.ResourceKind,
			Violation:    fmt.Sprintf("No Root user: Container %s can run as the root user which is forbidden", container.Name),
			Policy:       p.Name(),
		})
	}
	for _, container := range podResource.PodSpec.Containers {
		validateSecurityContext(container)
	}
	for _, container := range podResource.PodSpec.InitContainers {
		validateSecurityContext(container)
	}

	// if all the containers have the appropriate securityContext
	// and the podSecurityContext is unset, we can skip checking it
	if len(resourceViolations) == 0 && podResource.PodSpec.SecurityContext == nil {
		return resourceViolations, nil
	}

	if podResource.PodSpec.SecurityContext != nil && podResource.PodSpec.SecurityContext.RunAsNonRoot != nil && *podResource.PodSpec.SecurityContext.RunAsNonRoot {
		return resourceViolations, nil
	} else if podResource.PodSpec.SecurityContext != nil && podResource.PodSpec.SecurityContext.RunAsUser != nil && *podResource.PodSpec.SecurityContext.RunAsUser > 0 {
		return resourceViolations, nil
	}

	resourceViolations = append(resourceViolations, policies.ResourceViolation{
		Namespace:    ar.Namespace,
		ResourceName: podResource.ResourceName,
		ResourceKind: podResource.ResourceKind,
		Violation:    violationText,
		Policy:       p.Name(),
	})
	return resourceViolations, nil
}
