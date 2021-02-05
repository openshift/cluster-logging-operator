package client

import (
	"fmt"

	loggingv1 "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/test"
	"k8s.io/apimachinery/pkg/watch"
)

func ClusterLogForwarderReady(e watch.Event) (bool, error) {
	clf := e.Object.(*loggingv1.ClusterLogForwarder)
	cond := clf.Status.Conditions
	switch {
	case cond.IsTrueFor(loggingv1.ConditionReady):
		return true, nil
	case cond.IsFalseFor(loggingv1.ConditionReady), cond.IsTrueFor(loggingv1.ConditionDegraded):
		return false, fmt.Errorf("ClusterLogForwarder unexpected condition: %v", test.YAMLString(clf.Status))
	default:
		return false, nil
	}
}
