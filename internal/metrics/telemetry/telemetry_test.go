package telemetry

import (
	"testing"

	loggingv1 "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/test"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func init() {
	test.Must(loggingv1.AddToScheme(scheme.Scheme))
}

func TestTelemetry(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "clo telemetry test")
}

// newFakeClient is a workaround for the deprecation of fake.NewFakeClient, which was removed again in future versions
// of controller-runtime, so it does not make sense to change the newer code.
func newFakeClient(initObjs ...runtime.Object) client.Client {
	return fake.NewClientBuilder().WithRuntimeObjects(initObjs...).Build()
}
