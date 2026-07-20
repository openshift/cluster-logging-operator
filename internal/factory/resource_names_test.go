package factory_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/openshift/cluster-logging-operator/internal/factory"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	obsruntime "github.com/openshift/cluster-logging-operator/internal/runtime/observability"
)

var _ = Describe("ResourceNames", func() {
	It("should prefix the collector ConfigMap with clf-", func() {
		clf := obsruntime.NewClusterLogForwarder("openshift-logging", "foo", runtime.Initialize)
		names := factory.ResourceNames(*clf)
		Expect(names.ConfigMap).To(Equal("clf-foo-config"))
		Expect(factory.LegacyCollectorConfigMapName(clf.Name)).To(Equal("foo-config"))
	})
})
