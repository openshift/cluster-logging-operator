package clusterlogging

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/test/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Describe("[internal][validations] ClusterLogging", func() {

	Context("#validateClusterLoggingSpec", func() {
		var (
			cl        *logging.ClusterLogging
			k8sClient client.Client
		)
		BeforeEach(func() {
			cl = runtime.NewClusterLogging()
			cl.Spec.LogStore = &logging.LogStoreSpec{
				Type: logging.LogStoreTypeElasticsearch,
			}
			cl.Spec.Visualization = &logging.VisualizationSpec{}
			cl.Spec.Curation = &logging.CurationSpec{}
			cl.Spec.Forwarder = &logging.ForwarderSpec{}
			k8sClient = fake.NewClientBuilder().Build()
		})

		Context("for resource in openshift-logging named 'instance'", func() {
			It("should pass validation with no regressions since this is the legacy mode", func() {
				Expect(validateClusterLoggingSpec(*cl, k8sClient)).To(Succeed())
			})
		})

		Context("for resource not in openshift-logging or not named 'instance' in openshift-logging", func() {

			BeforeEach(func() {
				k8sClient = fake.NewClientBuilder().Build()
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
				err := validateClusterLoggingSpec(*cl, k8sClient)
				Expect(err).To(Not(Succeed()))
			})
			It("should fail when collection.logs is spec'd", func() {
				cl.Spec.Collection.Logs = &logging.LogCollectionSpec{}
				err := validateClusterLoggingSpec(*cl, k8sClient)
				Expect(err).To(Not(Succeed()))
			})
			It("should fail when collection.type of fluentd is spec'd", func() {
				cl.Spec.Collection.Type = logging.LogCollectionTypeFluentd
				err := validateClusterLoggingSpec(*cl, k8sClient)
				Expect(err).To(Not(Succeed()))
			})
			It("should fail when collection.type is empty", func() {
				cl.Spec.Collection.Type = ""
				err := validateClusterLoggingSpec(*cl, k8sClient)
				Expect(err).To(Not(Succeed()))
			})
			It("should pass when collection.type is spec'd", func() {
				err := validateClusterLoggingSpec(*cl, k8sClient)
				Expect(err).To(Succeed())
			})

		})

	})

	Context("#validateClusterLoggingSetUp", func() {
		var (
			cl             *logging.ClusterLogging
			clf            *logging.ClusterLogForwarder
			k8sClient      client.Client
			otherNamespace = "other-namespace"
			otherName      = "other-name"
		)
		BeforeEach(func() {
			cl = runtime.NewClusterLogging()
			clf = runtime.NewClusterLogForwarder()
			k8sClient = fake.NewClientBuilder().Build()
		})

		It("should pass when CL(openshift-logging/instance) only - This is a valid LEGACY use case", func() {
			err := validateSetup(*cl, k8sClient)
			Expect(err).To(Succeed())
		})

		It("should pass when CL(openshift-logging/instance) & CLF(openshift-logging/instance) - This is a valid LEGACY use case", func() {
			k8sClient = fake.NewClientBuilder().WithRuntimeObjects(clf).Build()
			err := validateSetup(*cl, k8sClient)
			Expect(err).To(Succeed())
		})

		It("should fail when CL(ANY_NS/instance) - This is an invalid use case", func() {
			cl.Namespace = otherNamespace
			err := validateSetup(*cl, k8sClient)
			Expect(err).To(Not(Succeed()))
		})

		It("should pass when CL(ANY_NS/instance) & CLF(ANY_NS/instance) - This is a valid mCLF use case", func() {
			cl.Namespace = otherNamespace
			clf.Namespace = otherNamespace
			k8sClient = fake.NewClientBuilder().WithRuntimeObjects(clf).Build()
			err := validateSetup(*cl, k8sClient)
			Expect(err).To(Succeed())
		})

		It("should fail when CL(openshift-logging/OTHER_NAME) - This is an invalid use case", func() {
			cl.Name = otherName
			err := validateSetup(*cl, k8sClient)
			Expect(err).To(Not(Succeed()))
		})

		It("should fail when CL(ANY_NS/OTHER_NAME) - This is an invalid use case", func() {
			cl.Name = otherName
			cl.Namespace = otherNamespace
			err := validateSetup(*cl, k8sClient)
			Expect(err).To(Not(Succeed()))
		})

		It("should pass when CL(ANY_NS/ANY_NAME) && CLF(ANY_NS/ANY_NAME) - This is a valid use case", func() {
			cl.Name = otherName
			cl.Namespace = otherNamespace
			clf.Name = otherName
			clf.Namespace = otherNamespace
			k8sClient = fake.NewClientBuilder().WithRuntimeObjects(clf).Build()
			err := validateSetup(*cl, k8sClient)
			Expect(err).To(Succeed())
		})
	})
})
