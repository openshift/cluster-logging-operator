package factory

import (
	"fmt"
	"reflect"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openshift/cluster-logging-operator/internal/constants"
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
			MetadataReaderClusterRoleBinding: fmt.Sprintf("cluster-logging-%s-%s-metadata-reader", constants.WatchNamespace, constants.CollectorName),
			ConfigMap:                        constants.CollectorConfigSecretName,
		}

		actualResNames := GenerateResourceNames(clfName, constants.WatchNamespace)

		equal := reflect.DeepEqual(actualResNames, expectedResNames)

		Expect(equal).To(BeTrue())

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
			MetadataReaderClusterRoleBinding: fmt.Sprintf("cluster-logging-%s-%s-metadata-reader", constants.WatchNamespace, clfName),
			ConfigMap:                        clfName + "-config",
		}

		actualResNames := GenerateResourceNames(clfName, constants.WatchNamespace)
		equal := reflect.DeepEqual(actualResNames, expectedResNames)
		Expect(equal).To(BeTrue())
	})
})
