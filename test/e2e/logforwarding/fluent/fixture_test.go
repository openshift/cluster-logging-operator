package fluent_test

import (
	loggingv1 "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/test"
	"github.com/openshift/cluster-logging-operator/test/client"
	"github.com/openshift/cluster-logging-operator/test/helpers/fluentd"
	. "github.com/openshift/cluster-logging-operator/test/matchers"
	"github.com/openshift/cluster-logging-operator/test/runtime"
	corev1 "k8s.io/api/core/v1"
)

type Fixture struct {
	ClusterLogging      *loggingv1.ClusterLogging
	ClusterLogForwarder *loggingv1.ClusterLogForwarder
	Receiver            *fluentd.Receiver
	LogGenerator        *corev1.Pod
}

// NewTest returns a new test, the clf and receiver are not yet configured.
func NewFixture(namespace, message string) *Fixture {
	return &Fixture{
		ClusterLogging:      runtime.NewClusterLogging(),
		ClusterLogForwarder: runtime.NewClusterLogForwarder(),
		Receiver:            fluentd.NewReceiver(namespace, "receiver"),
		LogGenerator:        runtime.NewLogGenerator(namespace, "log-generator", 1000, 0, message),
	}
}

// Create resources, wait for them to be ready.
func (f *Fixture) Create(c *client.Client) {
	g := test.FailGroup{}
	// Recreate resources in shared openshift-logging namespace.
	g.Go(func() { ExpectOK(c.Recreate(f.ClusterLogging)) })
	g.Go(func() {
		ExpectOK(c.Recreate(f.ClusterLogForwarder))
		ExpectOK(c.WaitFor(f.ClusterLogForwarder, client.ClusterLogForwarderReady))
	})
	g.Go(func() {
		ExpectOK(c.WaitForType(&corev1.Pod{}, client.PodRunning,
			client.MatchingLabels{"component": "fluentd"},
			client.InNamespace(f.ClusterLogging.Namespace)))
	})
	// Create resources in temporary test namespace.
	g.Go(func() {
		ExpectOK(c.Create(f.LogGenerator))
		ExpectOK(c.WaitFor(f.LogGenerator, client.PodRunning))
	})
	g.Go(func() { ExpectOK(f.Receiver.Create(c)) })
	g.Wait()
}
