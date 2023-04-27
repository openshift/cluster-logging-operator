package generator_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	configv1 "github.com/openshift/api/config/v1"
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	. "github.com/openshift/cluster-logging-operator/internal/generator"
	"github.com/openshift/cluster-logging-operator/internal/tls"
	"strings"
)

var _ = Describe("Options#TLSProfileInfo", func() {

	var (
		options = Options{}
	)

	Context("when a cluster profile is absent", func() {

		It("should use the defaults when clf profile is nil and output.TLS is nil", func() {
			minTLS, ciphers := options.TLSProfileInfo(logging.OutputSpec{}, ",")
			Expect(minTLS).To(BeEquivalentTo(tls.DefaultMinTLSVersion))
			Expect(ciphers).To(Equal(strings.Join(tls.DefaultTLSCiphers, ",")))
		})
	})

	Context("when a cluster profile exists", func() {

		var (
			clusterCiphers       = []string{"a", "b", "c"}
			clusterMinTLSVersion = configv1.VersionTLS12
			outputProfile        *configv1.TLSSecurityProfile
			outputSpec           logging.OutputSpec
		)
		BeforeEach(func() {
			options = Options{}
			options[ClusterTLSProfileSpec] = configv1.TLSProfileSpec{
				Ciphers:       clusterCiphers,
				MinTLSVersion: clusterMinTLSVersion,
			}
			outputProfile = &configv1.TLSSecurityProfile{
				Type: configv1.TLSProfileOldType,
			}
			outputSpec = logging.OutputSpec{
				TLS: &logging.OutputTLSSpec{
					TLSSecurityProfile: outputProfile,
				},
			}
		})

		It("should prefer the output profile over the cluster profile", func() {
			minTLS, ciphers := options.TLSProfileInfo(outputSpec, ",")
			spec := configv1.TLSProfiles[outputProfile.Type]
			Expect(minTLS).To(BeEquivalentTo(spec.MinTLSVersion))
			Expect(ciphers).To(Equal(strings.Join(spec.Ciphers, ",")))
		})

		It("should prefer the cluster profile when the forwarder and output.TLS are nil", func() {
			minTLS, ciphers := options.TLSProfileInfo(logging.OutputSpec{}, ",")
			Expect(minTLS).To(BeEquivalentTo(clusterMinTLSVersion))
			Expect(ciphers).To(Equal(strings.Join(clusterCiphers, ",")))
		})
	})

})
