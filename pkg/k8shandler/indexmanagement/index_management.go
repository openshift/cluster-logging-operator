package indexmanagement

import (
	"fmt"
	"regexp"
	"strconv"

	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/pkg/logger"
	esapi "github.com/openshift/elasticsearch-operator/pkg/apis/logging/v1"
)

const (
	PolicyNameApp   = "app-policy"
	PolicyNameInfra = "infra-policy"
	PolicyNameAudit = "audit-policy"

	MappingNameApp   = "app"
	MappingNameInfra = "infra"
	MappingNameAudit = "audit.infra"

	PollInterval = "15m"

	HotPhaseAgeAsPercentOfMaxAge = 5
)

var (
	AliasesApp   = []string{"app", "logs.app"}
	AliasesInfra = []string{"infra", "logs.infra"}
	AliasesAudit = []string{"infra.audit", "logs.audit"}

	agePattern = regexp.MustCompile("^(?P<number>\\d+)(?P<unit>[yMwdhHms])$")
)

func NewSpec(retentionPolicy *logging.RetentionPoliciesSpec) *esapi.IndexManagementSpec {

	if retentionPolicy == nil {
		return nil
	}
	if retentionPolicy.App == nil && retentionPolicy.Infra == nil && retentionPolicy.Audit == nil {
		logger.Info("Retention policy not defined for any log source. Cannot create Index management spec.")
		return nil
	}

	indexManagement := esapi.IndexManagementSpec{}
	if retentionPolicy.App != nil {
		hotPhaseAgeApp, err := getHotPhaseAge(retentionPolicy.App.MaxAge)
		if err != nil {
			logger.Errorf("Error occured while getting hot phase age for App log source. err: %v", err)
			return nil
		}
		appPolicySpec := newPolicySpec(PolicyNameApp, retentionPolicy.App.MaxAge, hotPhaseAgeApp)
		indexManagement.Policies = append(indexManagement.Policies, appPolicySpec)
		appMappingSpec := newMappingSpec(MappingNameApp, PolicyNameApp, AliasesApp)
		indexManagement.Mappings = append(indexManagement.Mappings, appMappingSpec)
	}
	if retentionPolicy.Infra != nil {
		hotPhaseAgeInfra, err := getHotPhaseAge(retentionPolicy.Infra.MaxAge)
		if err != nil {
			logger.Errorf("Error occured while getting hot phase age for Infra log source. err: %v", err)
			return nil
		}
		infraPolicySpec := newPolicySpec(PolicyNameInfra, retentionPolicy.Infra.MaxAge, hotPhaseAgeInfra)
		indexManagement.Policies = append(indexManagement.Policies, infraPolicySpec)
		infraMappingSpec := newMappingSpec(MappingNameInfra, PolicyNameInfra, AliasesInfra)
		indexManagement.Mappings = append(indexManagement.Mappings, infraMappingSpec)
	}
	if retentionPolicy.Audit != nil {
		hotPhaseAgeAudit, err := getHotPhaseAge(retentionPolicy.Audit.MaxAge)
		if err != nil {
			logger.Errorf("Error occured while getting hot phase age for Audit log source. err: %v", err)
			return nil
		}
		auditPolicySpec := newPolicySpec(PolicyNameAudit, retentionPolicy.Audit.MaxAge, hotPhaseAgeAudit)
		indexManagement.Policies = append(indexManagement.Policies, auditPolicySpec)
		auditMappingSpec := newMappingSpec(MappingNameAudit, PolicyNameAudit, AliasesAudit)
		indexManagement.Mappings = append(indexManagement.Mappings, auditMappingSpec)
	}
	return &indexManagement
}

func newPolicySpec(name string, maxIndexAge esapi.TimeUnit, hotPhaseAge esapi.TimeUnit) esapi.IndexManagementPolicySpec {

	policySpec := esapi.IndexManagementPolicySpec{
		Name:         name,
		PollInterval: PollInterval,
		Phases: esapi.IndexManagementPhasesSpec{
			Hot: &esapi.IndexManagementHotPhaseSpec{
				Actions: esapi.IndexManagementActionsSpec{
					Rollover: &esapi.IndexManagementActionSpec{
						MaxAge: hotPhaseAge,
					},
				},
			},
			Delete: &esapi.IndexManagementDeletePhaseSpec{
				MinAge: maxIndexAge,
			},
		},
	}
	return policySpec
}

func newMappingSpec(name string, policyRef string, aliases []string) esapi.IndexManagementPolicyMappingSpec {
	mappingSpec := esapi.IndexManagementPolicyMappingSpec{
		Name:      name,
		PolicyRef: policyRef,
		Aliases:   aliases,
	}
	return mappingSpec
}

func getHotPhaseAge(maxAge esapi.TimeUnit) (esapi.TimeUnit, error) {
	var (
		age         int
		unit        byte
		err         error
		hotphaseAge int
	)
	age, unit, err = toAgeAndUnit(maxAge)
	if err == nil {
		hotphaseAge, unit, err = toHotPhaseAge(age, unit)
		if err == nil {
			return esapi.TimeUnit(fmt.Sprintf("%d%c", hotphaseAge, unit)), nil
		}
	}
	return esapi.TimeUnit(""), err
}

func toAgeAndUnit(timeunit esapi.TimeUnit) (int, byte, error) {
	strvalues := agePattern.FindStringSubmatch(string(timeunit))
	if len(strvalues) != 3 {
		return 0, 0, fmt.Errorf("age pattern mismatch")
	}
	age, _ := strconv.Atoi(strvalues[1])
	unit := strvalues[2][0]
	return age, unit, nil
}

func toHotPhaseAge(value int, unit byte) (int, byte, error) {
	newval := value * HotPhaseAgeAsPercentOfMaxAge / 100

	for newval == 0 {
		value, newunit, err := convertToLowerUnits(value, unit)
		if err != nil {
			return 0, 0, err
		}
		newval = value * HotPhaseAgeAsPercentOfMaxAge / 100
		unit = newunit
	}

	return newval, unit, nil
}

func convertToLowerUnits(value int, unit byte) (int, byte, error) {

	switch unit {
	case 's':
		return 0, 0, fmt.Errorf("cannot convert \"%d%c\" to lower units", value, unit)
	case 'm':
		newval := value * 60
		return newval, 's', nil
	case 'h', 'H':
		newval := value * 60
		return newval, 'm', nil
	case 'd':
		newval := value * 24
		return newval, 'h', nil
	case 'w':
		newval := value * 7
		return newval, 'd', nil
	case 'M':
		newval := value * 30
		return newval, 'd', nil
	case 'y':
		newval := value * 365
		return newval, 'd', nil
	}

	return 0, 0, fmt.Errorf("unknown units")
}
