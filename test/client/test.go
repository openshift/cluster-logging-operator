package client

import (
	"fmt"
	"testing"

	"github.com/openshift/cluster-logging-operator/internal/runtime"

	"github.com/onsi/ginkgo"
	"github.com/openshift/cluster-logging-operator/test"
	corev1 "k8s.io/api/core/v1"

	"github.com/ViaQ/logerr/v2/log"
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
		NS:     runtime.NewNamespace(test.UniqueNameForTest()),
	}
	test.Must(t.Create(t.NS))
	if _, ok := test.GinkgoCurrentTest(); ok {
		fmt.Fprintf(ginkgo.GinkgoWriter, "test namespace: %v\n", t.NS.Name)
	}
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
	log.NewLogger("test").Info("testhack", "namespace", t.NS.Name)
	return t
}
func (t *NamespaceClient) Close() {
	_ = t.Remove(t.NS)
}

// ForTest returns a Test for a testing.T.
// The client.Test is closed when the test ends.
func ForTest(t *testing.T) *Test {
	// UniqueName is a lowercase DNS label, add "-" separator to make camelCase readable.
	c := NewTest()
	t.Cleanup(func() {
		if !t.Failed() {
			_ = c.Delete(c.NS)
		}
	})
	t.Logf("test namespace: %v", c.NS.Name)
	return c
}
