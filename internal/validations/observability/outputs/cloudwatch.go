package outputs

import (
	"fmt"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	internalobs "github.com/openshift/cluster-logging-operator/internal/api/observability"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func ValidateCloudWatchAuth(spec obs.OutputSpec, secrets map[string]*corev1.Secret) (results []metav1.Condition) {
	conds := func(reason, msg string) []metav1.Condition {
		return []metav1.Condition{
			internalobs.NewCondition(obs.ValidationCondition, metav1.ConditionTrue, reason, msg),
		}
	}

	if spec.Type != obs.OutputTypeCloudwatch {
		return nil
	}
	authSpec := spec.Cloudwatch.Authentication
	if authSpec == nil {
		return conds(obs.ReasonMissingSpec, fmt.Sprintf("%q requires auth configuration", spec.Name))
	}
	switch authSpec.Type {
	case obs.CloudwatchAuthTypeAccessKey:
		if authSpec.AWSAccessKey == nil {
			return conds(obs.ReasonMissingSpec, fmt.Sprintf("%q requires auth configuration", spec.Name))
		}
		if authSpec.AWSAccessKey.KeySecret == nil {
			return conds(obs.ReasonMissingSpec, fmt.Sprintf("%q auth requires a KeySecret", spec.Name))
		}
		if authSpec.AWSAccessKey.KeyID == nil {
			return conds(obs.ReasonMissingSpec, fmt.Sprintf("%q auth requires a KeyID", spec.Name))
		}
	case obs.CloudwatchAuthTypeIAMRole:
		if authSpec.IAMRole == nil {
			return conds(obs.ReasonMissingSpec, fmt.Sprintf("%q requires auth configuration", spec.Name))
		}
		if authSpec.IAMRole.RoleARN == nil {
			return conds(obs.ReasonMissingSpec, fmt.Sprintf("%q auth requires a RoleARN", spec.Name))
		}
		if authSpec.IAMRole.Token == nil {
			return conds(obs.ReasonMissingSpec, fmt.Sprintf("%q auth requires a token", spec.Name))
		}
		if authSpec.IAMRole.Token.Secret == nil && authSpec.IAMRole.Token.ServiceAccount == nil {
			return conds(obs.ReasonMissingSpec, fmt.Sprintf("%q auth requires a secret or serviceaccount token", spec.Name))
		}
	}
	return nil
}
