package observability

import (
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	defaultAudience = "openshift"
)

func InitServiceAccount(spec obs.ClusterLogForwarderSpec) (obs.ClusterLogForwarderSpec, []metav1.Condition) {
	if spec.ServiceAccount.Audience == "" {
		spec.ServiceAccount.Audience = defaultAudience
	}
	return spec, nil
}
