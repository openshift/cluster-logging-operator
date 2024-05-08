package telemetry

import (
	"testing"

	loggingv1 "github.com/openshift/cluster-logging-operator/api/logging/v1"
	loggingv1alpha1 "github.com/openshift/cluster-logging-operator/api/logging/v1alpha1"
	"github.com/openshift/cluster-logging-operator/test"

	"k8s.io/client-go/kubernetes/scheme"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func init() {
	test.Must(loggingv1.AddToScheme(scheme.Scheme))
	test.Must(loggingv1alpha1.AddToScheme(scheme.Scheme))
}

func TestTelemetry(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "clo telemetry test")
}
