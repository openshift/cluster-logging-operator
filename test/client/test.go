package client

import (
	"fmt"

	"github.com/onsi/ginkgo"
	"github.com/openshift/cluster-logging-operator/test"
	"github.com/openshift/cluster-logging-operator/test/runtime"
	corev1 "k8s.io/api/core/v1"

	"github.com/ViaQ/logerr/log"
)

// Test wraps the singleton test client with setup/teardown and convenience methods
// for testing.
type Test struct {
	*Client
	NS *corev1.Namespace
}

// NewTest creates a new Test, which includes creating a new test namespace.
func NewTest() *Test {
	t := &Test{
		Client: Get(),
		NS:     runtime.NewUniqueNamespace(),
	}
	test.Must(t.Create(t.NS))
	fmt.Fprintf(ginkgo.GinkgoWriter, "test namespace: %v\n", t.NS.Name)
	return t
}

// Close removes the test namespace unless called from a failed test.
func (t *Test) Close() {
	if g, ok := test.GinkgoCurrentTest(); ok && !g.Failed {
		_ = t.Remove(t.NS)
	} else {
		fmt.Printf("\n\n============\n")
		fmt.Printf("Not removing functional test namespace since test failed. Run \"oc delete ns %s\" to delete namespace manually\n", t.NS.Name)
		fmt.Printf("To delete all lingering functional test namespaces, run \"oc delete ns -ltest-client=true\"\n")
		fmt.Printf("============\n\n")
	}
}

//NamespaceClient wraps the singleton test client for use with hack testing
type NamespaceClient struct {
	Test
}

func NewNamesapceClient() *NamespaceClient {
	namespace := test.UniqueName("testhack")
	t := &NamespaceClient{
		Test{
			Client: Get(),
			NS:     runtime.NewNamespace(namespace),
		},
	}
	test.Must(t.Create(t.NS))
	log.Info("testhack", "namespace", t.NS.Name)
	return t
}
func (t *NamespaceClient) Close() {
	_ = t.Remove(t.NS)
}
