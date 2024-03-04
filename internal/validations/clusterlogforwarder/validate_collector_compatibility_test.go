package clusterlogforwarder

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	loggingv1 "github.com/openshift/cluster-logging-operator/api/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/test/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("[internal][validations] ClusterLogForwarder", func() {

	var (
		k8sClient client.Client
		clf       *loggingv1.ClusterLogForwarder

		es = loggingv1.OutputSpec{
			Type: loggingv1.OutputTypeElasticsearch,
		}
		fluentForward = loggingv1.OutputSpec{
			Type: loggingv1.OutputTypeFluentdForward,
		}
		gcl = loggingv1.OutputSpec{
			Type: loggingv1.OutputTypeGoogleCloudLogging,
		}
		http = loggingv1.OutputSpec{
			Type: loggingv1.OutputTypeHttp,
		}
		kafka = loggingv1.OutputSpec{
			Type: loggingv1.OutputTypeKafka,
		}
		loki = loggingv1.OutputSpec{
			Type: loggingv1.OutputTypeLoki,
		}
		splunk = loggingv1.OutputSpec{
			Type: loggingv1.OutputTypeSplunk,
		}
		syslog = loggingv1.OutputSpec{
			Type: loggingv1.OutputTypeSyslog,
		}
		cloudwatch = loggingv1.OutputSpec{
			Type: loggingv1.OutputTypeCloudwatch,
		}
		azureMonitor = loggingv1.OutputSpec{
			Type: loggingv1.OutputTypeAzureMonitor,
		}
	)

	BeforeEach(func() {
		clf = runtime.NewClusterLogForwarder()
		clf.Name = "collector"
	})

	Context("#validateCollectorCompatibility for Vector", func() {
		var vector = map[string]bool{
			constants.VectorName: true,
		}

		It("should fail validation when the collector is Vector and one of OutputType is FluentForward", func() {
			clf.Spec.Outputs = []loggingv1.OutputSpec{
				es, fluentForward, gcl, http, kafka, loki, splunk, syslog, cloudwatch, azureMonitor,
			}
			Expect(validateCollectorCompatibility(*clf, k8sClient, vector)).ToNot(Succeed())
		})

		It("should pass validation when the collector is Vector and all outputs are supported", func() {
			clf.Spec.Outputs = []loggingv1.OutputSpec{
				es, gcl, http, kafka, loki, splunk, syslog, cloudwatch, azureMonitor,
			}
			Expect(validateCollectorCompatibility(*clf, k8sClient, vector)).To(Succeed())
		})
	})

	Context("#validateCollectorCompatibility for Fluentd", func() {

		It("should fail validation when the collector is Fluentd and one of OutputType is Google Cloud Logging", func() {
			clf.Spec.Outputs = []loggingv1.OutputSpec{
				es, fluentForward, gcl, http, kafka, loki, syslog, cloudwatch,
			}
			Expect(validateCollectorCompatibility(*clf, k8sClient, map[string]bool{})).ToNot(Succeed())
		})

		It("should fail validation when the collector is Fluentd and one of OutputType is Splunk", func() {
			clf.Spec.Outputs = []loggingv1.OutputSpec{
				es, fluentForward, http, kafka, loki, splunk, syslog,
			}
			Expect(validateCollectorCompatibility(*clf, k8sClient, map[string]bool{})).ToNot(Succeed())
		})

		It("should fail validation when the collector is Vector and one of OutputType is Azure Monitor", func() {
			clf.Spec.Outputs = []loggingv1.OutputSpec{
				es, fluentForward, http, kafka, loki, syslog, cloudwatch, azureMonitor,
			}
			Expect(validateCollectorCompatibility(*clf, k8sClient, map[string]bool{})).ToNot(Succeed())
		})

		It("should pass validation when the collector is Fluentd and all outputs are supported", func() {
			clf.Spec.Outputs = []loggingv1.OutputSpec{
				es, http, kafka, loki, syslog, cloudwatch,
			}
			Expect(validateCollectorCompatibility(*clf, k8sClient, map[string]bool{})).To(Succeed())
		})
	})
})
