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
		case spec.Application.Excludes != nil:
			for _, ex := range spec.Application.Excludes {
				if !globRE.MatchString(ex.Namespace) {
					status.Inputs.Set(spec.Name, loggingv1.NewCondition(loggingv1.ValidationCondition,
						corev1.ConditionTrue,
						loggingv1.ValidationFailureReason,
						globErrorFmt, "namespace excludes"))
				}
				if !globRE.MatchString(ex.Container) {
					status.Inputs.Set(spec.Name, loggingv1.NewCondition(loggingv1.ValidationCondition,
						corev1.ConditionTrue,
						loggingv1.ValidationFailureReason,
						globErrorFmt, "container excludes"))
				}
			}
		case spec.Application.Includes != nil:
			for _, in := range spec.Application.Includes {
				if !globRE.MatchString(in.Namespace) {
					status.Inputs.Set(spec.Name, loggingv1.NewCondition(loggingv1.ValidationCondition,
						corev1.ConditionTrue,
						loggingv1.ValidationFailureReason,
						globErrorFmt, "namespace includes"))
				}
				if !globRE.MatchString(in.Container) {
					status.Inputs.Set(spec.Name, loggingv1.NewCondition(loggingv1.ValidationCondition,
						corev1.ConditionTrue,
						loggingv1.ValidationFailureReason,
						globErrorFmt, "container includes"))
				}
			}
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
