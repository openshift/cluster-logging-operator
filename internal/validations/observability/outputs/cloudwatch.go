package outputs

import (
	"github.com/golang-collections/collections/set"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	internalcontext "github.com/openshift/cluster-logging-operator/internal/api/context"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/cloudwatch"
	"github.com/openshift/cluster-logging-operator/internal/utils"
)

const (
	RoleARNsOpt           = "roleARNs"
	ErrVariousRoleARNAuth = "Found multiple different CloudWatch RoleARN authorizations in the outputs spec"
)

func ValidateCloudWatchAuth(spec obs.OutputSpec, context internalcontext.ForwarderContext) (results []string) {
	secrets := helpers.Secrets(context.Secrets)
	additionalContext := context.AdditionalContext
	authSpec := spec.Cloudwatch.Authentication

	if authSpec.Type == obs.CloudwatchAuthTypeIAMRole {
		roleArn := cloudwatch.ParseRoleArn(authSpec, secrets)
		roleARNs := set.New(roleArn)
		utils.Update(additionalContext, RoleARNsOpt, roleARNs, func(existing *set.Set) *set.Set {
			existing = existing.Union(roleARNs)
			if existing.Len() > 1 {
				results = append(results, ErrVariousRoleARNAuth)
			}
			return existing
		})
	}
	return results
}
