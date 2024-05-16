package helpers

import (
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/generator/framework"
	corev1 "k8s.io/api/core/v1"
)

type ConfGenerateTest struct {
	Desc    string
	CLFSpec obs.ClusterLogForwarderSpec
	CLSpec  obs.CollectorSpec
	// key:Output Name, value: secret for the Output
	Secrets      map[string]*corev1.Secret
	Options      framework.Options
	ExpectedConf string
}
