package client

import (
	"fmt"
	"testing"

	"github.com/openshift/cluster-logging-operator/internal/runtime"

	"github.com/onsi/ginkgo"
	"github.com/openshift/cluster-logging-operator/test"
	corev1 "k8s.io/api/core/v1"

	log "github.com/ViaQ/logerr/v2/log/static"
)

// Test wraps the singleton test client with setup/teardown and convenience methods
// for testing.
type Test struct {
	*Client
	NS *corev1.Namespace
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

// NewTest creates a new Test, which includes creating a new test namespace.
func NewTest(testOptions ...TestOption) *Test {
	ns := test.UniqueNameForTest()
	if TestOptions(testOptions).Include(UseInfraNamespaceTestOption) {
		ns = test.UniqueName("openshift-test")
	}
	t := &Test{
		Client: Get(),
		NS:     runtime.NewNamespace(ns),
	}
	if !TestOptions(testOptions).Include(DryRunTestOption) {
		test.Must(t.Create(t.NS))
	}
	if _, ok := test.GinkgoCurrentTest(); ok {
		fmt.Fprintf(ginkgo.GinkgoWriter, "test namespace: %v\n", t.NS.Name)
	}
	return t
}

// TestOption is an option to alter a test in some way
type TestOption string

type TestOptions []TestOption

func (options TestOptions) Include(option TestOption) bool {
	for _, o := range options {
		if o == option {
			return true
		}
	}
	return false
}

const (
	//UseInfraNamespaceTestOption is the option to hint the test should be run in an infrastructure namespace
	UseInfraNamespaceTestOption TestOption = "useInfraNamespace"

	//DryRunTestOption is a hint to use in testing to not actually create resources against cluster (e.g. unit tests)
	DryRunTestOption TestOption = "dryRun"
)

// NamespaceClient wraps the singleton test client for use with hack testing
type NamespaceClient struct {
	Test
}

func NewNamespaceClient() *NamespaceClient {
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
