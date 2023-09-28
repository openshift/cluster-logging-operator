package matchers

import (
	//"fmt"
	"github.com/onsi/gomega/types"
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	corev1 "k8s.io/api/core/v1"
	//"k8s.io/utils/diff"
	//"reflect"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
)

// Match condition by type, status and reason if reason != "".
// Also match messageRegex if it is not empty.
func matchCondition(t logging.ConditionType, s bool, r logging.ConditionReason, messageRegex string) types.GomegaMatcher {
	var status corev1.ConditionStatus
	if s {
		status = corev1.ConditionTrue
	} else {
		status = corev1.ConditionFalse
	}
	fields := Fields{"Type": Equal(t), "Status": Equal(status)}
	if r != "" {
		fields["Reason"] = Equal(r)
	}
	if messageRegex != "" {
		fields["Message"] = MatchRegexp(messageRegex)
	}
	return MatchFields(IgnoreExtras, fields)
}

func HaveCondition(t logging.ConditionType, s bool, r logging.ConditionReason, messageRegex string) types.GomegaMatcher {
	return ContainElement(matchCondition(t, s, r, messageRegex))
}
