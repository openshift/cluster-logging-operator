package clusterlogging

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/test/runtime"
)

var _ = Describe("[internal][validations] ClusterLogging", func() {

	Context("#validateClusterLoggingSpec", func() {
		var (
			cl *logging.ClusterLogging
		)
		BeforeEach(func() {
			cl = runtime.NewClusterLogging()
			cl.Spec.LogStore = &logging.LogStoreSpec{
				Type: logging.LogStoreTypeElasticsearch,
			}
			cl.Spec.Visualization = &logging.VisualizationSpec{}
			cl.Spec.Curation = &logging.CurationSpec{}
			cl.Spec.Forwarder = &logging.ForwarderSpec{}
		})

		Context("for resource in openshift-logging named 'instance'", func() {
			It("should pass validation with no regressions since this is the legacy mode", func() {
				Expect(validateClusterLoggingSpec(*cl)).To(Succeed())
			})
		})

		Context("for resource not in openshift-logging or not named 'instance' in openshift-logging", func() {

			BeforeEach(func() {
				cl = runtime.NewClusterLogging()
				cl.Namespace = "a test namespace"
				cl.Name = "mycollector"
				cl.Spec.Collection = &logging.CollectionSpec{
					Type: logging.LogCollectionTypeVector,
					CollectorSpec: logging.CollectorSpec{
						NodeSelector: map[string]string{
							"foo": "bar",
						},
					},
				}
			})

			It("should fail when anything but collection is spec'd", func() {
				cl.Spec.LogStore = &logging.LogStoreSpec{
					Type: logging.LogStoreTypeElasticsearch,
				}
				cl.Spec.Visualization = &logging.VisualizationSpec{}
				cl.Spec.Curation = &logging.CurationSpec{}
				cl.Spec.Forwarder = &logging.ForwarderSpec{}
				err := validateClusterLoggingSpec(*cl)
				Expect(err).To(Not(Succeed()))
			})
			It("should fail when collection.logs is spec'd", func() {
				cl.Spec.Collection.Logs = &logging.LogCollectionSpec{}
				err := validateClusterLoggingSpec(*cl)
				Expect(err).To(Not(Succeed()))
			})
			It("should fail when collection.type of fluentd is spec'd", func() {
				cl.Spec.Collection.Type = logging.LogCollectionTypeFluentd
				err := validateClusterLoggingSpec(*cl)
				Expect(err).To(Not(Succeed()))
			})
			It("should fail when collection.type is empty", func() {
				cl.Spec.Collection.Type = ""
				err := validateClusterLoggingSpec(*cl)
				Expect(err).To(Not(Succeed()))
			})
			It("should pass when collection.type is spec'd", func() {
				err := validateClusterLoggingSpec(*cl)
				Expect(err).To(Succeed())
			})

		})

	})

})
