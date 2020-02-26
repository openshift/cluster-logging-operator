package indexmanagement

import (
	"fmt"
	"regexp"
	"strings"

	esapi "github.com/openshift/elasticsearch-operator/pkg/apis/logging/v1"
)

var (
	reTimeUnit = regexp.MustCompile("^(?P<number>\\d+)(?P<unit>[yMwdhHms])$")
)

const (
	pollIntervalFailMessage  = "The pollInterval is missing or requires a valid time unit (e.g. 3d)"
	phaseTimeUnitFailMessage = "The %s phase '%s' is missing or requires a valid time unit (e.g. 3d)"
	policyRefFailMessage     = "A policy mapping must reference a defined IndexManagement policy"
)

//VerifyAndNormalize validates the spec'd indexManagement and returns a spec which removes policies
//and mappings that are invalid
func VerifyAndNormalize(cluster *esapi.Elasticsearch) *esapi.IndexManagementSpec {
	result := &esapi.IndexManagementSpec{}
	status := esapi.NewIndexManagementStatus()
	cluster.Status.IndexManagementStatus = status
	if cluster.Spec.IndexManagement == nil || (len(cluster.Spec.IndexManagement.Mappings) == 0 && len(cluster.Spec.IndexManagement.Policies) == 0) {
		status.State = esapi.IndexManagementStateDropped
		status.Reason = esapi.IndexManagementStatusReasonUndefined
		status.Message = "IndexManagement was not defined"
		return nil
	}
	validatePolicies(cluster, result)
	validateMappings(cluster, result)
	if len(result.Mappings) != len(cluster.Spec.IndexManagement.Mappings) || len(result.Policies) != len(cluster.Spec.IndexManagement.Policies) {
		status.State = esapi.IndexManagementStateDegraded
		status.Reason = esapi.IndexManagementStatusReasonValidationFailed
	}
	if len(result.Mappings) == 0 && len(result.Policies) == 0 {
		status.State = esapi.IndexManagementStateDropped
	}
	return result
}

func validatePolicies(cluster *esapi.Elasticsearch, result *esapi.IndexManagementSpec) {
	if cluster.Spec.IndexManagement == nil {
		return
	}
	policyNames := map[string]interface{}{}
	for n, policy := range cluster.Spec.IndexManagement.Policies {
		status := esapi.NewIndexManagementPolicyStatus(policy.Name)
		if strings.TrimSpace(policy.Name) == "" {
			status.Name = fmt.Sprintf("policy[%d]", n)
			status.AddPolicyCondition(esapi.IndexManagementPolicyConditionTypeName, esapi.IndexManagementPolicyReasonMissing, "")
		} else {
			if len(policyNames) > 0 {
				if _, found := policyNames[policy.Name]; found {
					status.Name = fmt.Sprintf("policy[%d]", n)
					status.AddPolicyCondition(esapi.IndexManagementPolicyConditionTypeName, esapi.IndexManagementPolicyReasonNonUnique, "")
				}
			}
			policyNames[policy.Name] = ""
		}
		if !isValidTimeUnit(policy.PollInterval) {
			status.AddPolicyCondition(esapi.IndexManagementPolicyConditionTypePollInterval, esapi.IndexManagementPolicyReasonMalformed, pollIntervalFailMessage)
		}
		if policy.Phases.Hot != nil {
			if policy.Phases.Hot.Actions.Rollover == nil || !isValidTimeUnit(policy.Phases.Hot.Actions.Rollover.MaxAge) {
				message := fmt.Sprintf(phaseTimeUnitFailMessage, "hot", "maxAge")
				status.AddPolicyCondition(esapi.IndexManagementPolicyConditionTypeTimeUnit, esapi.IndexManagementPolicyReasonMalformed, message)
			}
		}
		if policy.Phases.Delete != nil {
			if !isValidTimeUnit(policy.Phases.Delete.MinAge) {
				message := fmt.Sprintf(phaseTimeUnitFailMessage, "delete", "minAge")
				status.AddPolicyCondition(esapi.IndexManagementPolicyConditionTypeTimeUnit, esapi.IndexManagementPolicyReasonMalformed, message)
			}
		}
		if len(status.Conditions) > 0 {
			status.State = esapi.IndexManagementPolicyStateDropped
			status.Reason = esapi.IndexManagementPolicyReasonConditionsNotMet
		} else {
			result.Policies = append(result.Policies, policy)
		}
		cluster.Status.IndexManagementStatus.Policies = append(cluster.Status.IndexManagementStatus.Policies, *status)
	}
}

func isValidTimeUnit(time esapi.TimeUnit) bool {
	return reTimeUnit.MatchString(string(time))
}

func validateMappings(cluster *esapi.Elasticsearch, result *esapi.IndexManagementSpec) {
	if cluster.Spec.IndexManagement == nil {
		return
	}
	policies := cluster.Spec.IndexManagement.PolicyMap()
	mappingNames := map[string]interface{}{}
	for n, mapping := range cluster.Spec.IndexManagement.Mappings {
		status := esapi.NewIndexManagementMappingStatus(mapping.Name)
		if strings.TrimSpace(mapping.Name) == "" {
			status.Name = fmt.Sprintf("mapping[%d]", n)
			status.AddPolicyMappingCondition(esapi.IndexManagementMappingConditionTypeName, esapi.IndexManagementMappingReasonMissing, "")
		} else {
			if len(mappingNames) > 0 {
				if _, found := mappingNames[mapping.Name]; found {
					status.Name = fmt.Sprintf("mapping[%d]", n)
					status.AddPolicyMappingCondition(esapi.IndexManagementMappingConditionTypeName, esapi.IndexManagementMappingReasonNonUnique, "")
				}
			}
			mappingNames[mapping.Name] = ""
		}
		if !policies.HasPolicy(mapping.PolicyRef) {
			status.AddPolicyMappingCondition(esapi.IndexManagementMappingConditionTypePolicyRef, esapi.IndexManagementMappingReasonMissing, policyRefFailMessage)
		}
		if len(status.Conditions) > 0 {
			status.State = esapi.IndexManagementMappingStateDropped
			status.Reason = esapi.IndexManagementMappingReasonConditionsNotMet
		} else {
			result.Mappings = append(result.Mappings, mapping)
		}
		cluster.Status.IndexManagementStatus.Mappings = append(cluster.Status.IndexManagementStatus.Mappings, *status)
	}
}
