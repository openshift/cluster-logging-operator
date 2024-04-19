package clusterlogging

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	logging "github.com/openshift/cluster-logging-operator/api/logging/v1"
	"github.com/openshift/cluster-logging-operator/test/runtime"
)

var _ = Describe("[internal][validations] ClusterLogging", func() {
	defer GinkgoRecover()
	Skip("TODO: FIX ME if appropriate")
	Context("#validateClusterLoggingSpec", func() {
		var (
			cl *logging.ClusterLogging
		)

		Context("for resource in openshift-logging named 'instance'", func() {
			BeforeEach(func() {
				cl = runtime.NewClusterLogging()
				cl.Spec.Collection = nil
				cl.Spec.LogStore = &logging.LogStoreSpec{
					Type: logging.LogStoreTypeLokiStack,
				}
				cl.Spec.Visualization = &logging.VisualizationSpec{
					Type: logging.VisualizationTypeOCPConsole,
				}
			})

			It("should pass validation with no regressions since openshift-logging/instance is only supported", func() {
				Expect(validateClusterLoggingSpec(*cl, nil)).To(Succeed())
			})

			It("should fail when anything other then logstore and visualization is spec'd", func() {
				cl.Spec.Collection = &logging.CollectionSpec{}
				cl.Spec.Curation = &logging.CurationSpec{}
				cl.Spec.Forwarder = &logging.ForwarderSpec{}
				err := validateClusterLoggingSpec(*cl, nil)
				Expect(err).To(Not(Succeed()))
			})
			It("should fail when logStore.type of Elasticsearch is spec'd", func() {
				cl.Spec.LogStore.Type = logging.LogStoreTypeElasticsearch
				err := validateClusterLoggingSpec(*cl, nil)
				Expect(err).To(Not(Succeed()))
			})
			It("should fail when visualization.type of kibana is spec'd", func() {
				cl.Spec.Visualization.Type = logging.VisualizationTypeKibana
				err := validateClusterLoggingSpec(*cl, nil)
				Expect(err).To(Not(Succeed()))
			})

		})

		Context("for resource not in openshift-logging or not named 'instance' in openshift-logging", func() {
			BeforeEach(func() {
				cl = runtime.NewClusterLogging()
			})
			It("should fail validation with unsupported namespace since openshift-logging/instance is only supported", func() {
				cl = runtime.NewClusterLogging()
				cl.Namespace = "a test namespace"
				Expect(validateClusterLoggingSpec(*cl, nil)).To(Not(Succeed()))
			})
			It("should fail validation with unsupported name since openshift-logging/instance is only supported", func() {
				cl = runtime.NewClusterLogging()
				cl.Name = "myname"
				Expect(validateClusterLoggingSpec(*cl, nil)).To(Not(Succeed()))
			})

		})
	})
})
