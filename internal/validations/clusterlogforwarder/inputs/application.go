package inputs

import (
	loggingv1 "github.com/openshift/cluster-logging-operator/api/logging/v1"
	corev1 "k8s.io/api/core/v1"
	"regexp"
)

var (
	globRE = regexp.MustCompile(`^[a-zA-Z0-9\*\-]*$`)

	globErrorFmt = `invalid glob for %s. Must match '` + globRE.String() + `"`
)

func validApplication(spec loggingv1.InputSpec, status *loggingv1.ClusterLogForwarderStatus, extras map[string]bool) bool {
	if spec.Application != nil {
		switch {
		case spec.HasPolicy() && spec.Application.ContainerLimit != nil && spec.Application.GroupLimit != nil:
			status.Inputs.Set(spec.Name, loggingv1.NewCondition(loggingv1.ValidationCondition,
				corev1.ConditionTrue,
				loggingv1.ValidationFailureReason,
				"application input must define only one of container or group limit"))
		case spec.HasPolicy() && spec.GetMaxRecordsPerSecond() < 0:
			status.Inputs.Set(spec.Name, loggingv1.NewCondition(loggingv1.ValidationCondition,
				corev1.ConditionTrue,
				loggingv1.ValidationFailureReason,
				"application input cannot have a negative limit threshold"))
		case !validGlob(spec.Application.Namespaces):
			status.Inputs.Set(spec.Name, loggingv1.NewCondition(loggingv1.ValidationCondition,
				corev1.ConditionTrue,
				loggingv1.ValidationFailureReason,
				globErrorFmt, "namespaces"))
		case !validGlob(spec.Application.ExcludeNamespaces):
			status.Inputs.Set(spec.Name, loggingv1.NewCondition(loggingv1.ValidationCondition,
				corev1.ConditionTrue,
				loggingv1.ValidationFailureReason,
				globErrorFmt, "excludeNamespaces"))
		case spec.Application.Containers != nil && !validGlob(spec.Application.Containers.Include):
			status.Inputs.Set(spec.Name, loggingv1.NewCondition(loggingv1.ValidationCondition,
				corev1.ConditionTrue,
				loggingv1.ValidationFailureReason,
				globErrorFmt, "containers include"))
		case spec.Application.Containers != nil && !validGlob(spec.Application.Containers.Exclude):
			status.Inputs.Set(spec.Name, loggingv1.NewCondition(loggingv1.ValidationCondition,
				corev1.ConditionTrue,
				loggingv1.ValidationFailureReason,
				globErrorFmt, "containers exclude"))
		}
	}
	return len(status.Inputs[spec.Name]) == 0
}

func validGlob(values []string) bool {
	for _, v := range values {
		if !globRE.MatchString(v) {
			return false
		}
	}
	return true
}
