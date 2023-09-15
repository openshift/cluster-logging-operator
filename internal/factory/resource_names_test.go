package factory

import (
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
)

var _ = Describe("Collector Resource Name Generator", func() {
	var (
		clfName          string
		expectedResNames *ForwarderResourceNames
	)

	It("should name resources to match the legacy names", func() {
		clfName = "instance"
		expectedResNames = &ForwarderResourceNames{
			CommonName:                       constants.CollectorName,
			CaTrustBundle:                    constants.CollectorTrustedCAName,
			ServiceAccount:                   constants.CollectorServiceAccountName,
			InternalLogStoreSecret:           constants.CollectorName,
			SecretMetrics:                    constants.CollectorMetricSecretName,
			ServiceAccountTokenSecret:        constants.LogCollectorToken,
			MetadataReaderClusterRoleBinding: fmt.Sprintf("cluster-logging-%s-%s-metadata-reader", constants.OpenshiftNS, constants.CollectorName),
			ConfigMap:                        constants.CollectorConfigSecretName,
			ForwarderName:                    constants.SingletonName,
		}

		clf := *runtime.NewClusterLogForwarder(constants.OpenshiftNS, clfName)

		Expect(GenerateResourceNames(clf)).To(BeEquivalentTo(expectedResNames))

	})

	It("Collector resource names should include custom CLF name if not instance", func() {
		clfName = "custom-clf"

		expectedResNames = &ForwarderResourceNames{
			CommonName:                       clfName,
			CaTrustBundle:                    clfName + "-trustbundle",
			ServiceAccount:                   clfName,
			InternalLogStoreSecret:           clfName + "-default",
			SecretMetrics:                    clfName + "-metrics",
			ServiceAccountTokenSecret:        clfName + "-token",
			MetadataReaderClusterRoleBinding: fmt.Sprintf("cluster-logging-%s-%s-metadata-reader", constants.OpenshiftNS, clfName),
			ConfigMap:                        clfName + "-config",
			ForwarderName:                    clfName,
		}

		clf := *runtime.NewClusterLogForwarder(constants.OpenshiftNS, clfName)
		clf.Spec.ServiceAccountName = clfName
		Expect(GenerateResourceNames(clf)).To(BeEquivalentTo(expectedResNames))
	})
})
