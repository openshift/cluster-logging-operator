package tls

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/adapters"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	corev1 "k8s.io/api/core/v1"
)

var _ = Describe("NewTlsEnabled", func() {
	const secretName = "test-tls"

	var secrets map[string]*corev1.Secret

	BeforeEach(func() {
		secrets = map[string]*corev1.Secret{
			secretName: {
				Data: map[string][]byte{
					constants.TrustedCABundleKey: []byte("ca-cert"),
					constants.ClientCertKey:      []byte("client-cert"),
					constants.ClientPrivateKey:   []byte("client-key"),
				},
			},
		}
	})

	It("should return TlsEnabled when CAFile is set", func() {
		adapter := adapters.NewOutput(obs.OutputSpec{
			Name: "test",
			Type: obs.OutputTypeKafka,
			TLS: &obs.OutputTLSSpec{
				TLSSpec: obs.TLSSpec{
					CA: &obs.ValueReference{
						Key:        constants.TrustedCABundleKey,
						SecretName: secretName,
					},
				},
			},
		})
		result := NewTlsEnabled(adapter, secrets, utils.NoOptions)
		Expect(result).NotTo(BeNil())
		Expect(result.Enabled).To(BeTrue())
		Expect(result.TLS.CAFile).NotTo(BeEmpty())
	})

	It("should return TlsEnabled when both CRTFile and KeyFile are set", func() {
		adapter := adapters.NewOutput(obs.OutputSpec{
			Name: "test",
			Type: obs.OutputTypeKafka,
			TLS: &obs.OutputTLSSpec{
				TLSSpec: obs.TLSSpec{
					Certificate: &obs.ValueReference{
						Key:        constants.ClientCertKey,
						SecretName: secretName,
					},
					Key: &obs.SecretReference{
						Key:        constants.ClientPrivateKey,
						SecretName: secretName,
					},
				},
			},
		})
		result := NewTlsEnabled(adapter, secrets, utils.NoOptions)
		Expect(result).NotTo(BeNil())
		Expect(result.Enabled).To(BeTrue())
		Expect(result.TLS.CRTFile).NotTo(BeEmpty())
		Expect(result.TLS.KeyFile).NotTo(BeEmpty())
	})

	It("should return TlsEnabled when CAFile, CRTFile and KeyFile are all set", func() {
		adapter := adapters.NewOutput(obs.OutputSpec{
			Name: "test",
			Type: obs.OutputTypeKafka,
			TLS: &obs.OutputTLSSpec{
				TLSSpec: obs.TLSSpec{
					CA: &obs.ValueReference{
						Key:        constants.TrustedCABundleKey,
						SecretName: secretName,
					},
					Certificate: &obs.ValueReference{
						Key:        constants.ClientCertKey,
						SecretName: secretName,
					},
					Key: &obs.SecretReference{
						Key:        constants.ClientPrivateKey,
						SecretName: secretName,
					},
				},
			},
		})
		result := NewTlsEnabled(adapter, secrets, utils.NoOptions)
		Expect(result).NotTo(BeNil())
		Expect(result.Enabled).To(BeTrue())
		Expect(result.TLS.CAFile).NotTo(BeEmpty())
		Expect(result.TLS.CRTFile).NotTo(BeEmpty())
		Expect(result.TLS.KeyFile).NotTo(BeEmpty())
	})

	It("should return nil when only KeyFile is set (without CRTFile)", func() {
		adapter := adapters.NewOutput(obs.OutputSpec{
			Name: "test",
			Type: obs.OutputTypeKafka,
			TLS: &obs.OutputTLSSpec{
				TLSSpec: obs.TLSSpec{
					Key: &obs.SecretReference{
						Key:        constants.ClientPrivateKey,
						SecretName: secretName,
					},
				},
			},
		})
		result := NewTlsEnabled(adapter, secrets, utils.NoOptions)
		Expect(result).To(BeNil())
	})

	It("should return nil when only CRTFile is set (without KeyFile)", func() {
		adapter := adapters.NewOutput(obs.OutputSpec{
			Name: "test",
			Type: obs.OutputTypeKafka,
			TLS: &obs.OutputTLSSpec{
				TLSSpec: obs.TLSSpec{
					Certificate: &obs.ValueReference{
						Key:        constants.ClientCertKey,
						SecretName: secretName,
					},
				},
			},
		})
		result := NewTlsEnabled(adapter, secrets, utils.NoOptions)
		Expect(result).To(BeNil())
	})

	It("should return nil when no TLS config is provided", func() {
		adapter := adapters.NewOutput(obs.OutputSpec{
			Name: "test",
			Type: obs.OutputTypeKafka,
		})
		result := NewTlsEnabled(adapter, secrets, utils.NoOptions)
		Expect(result).To(BeNil())
	})

	It("should return nil when TLS spec is empty", func() {
		adapter := adapters.NewOutput(obs.OutputSpec{
			Name: "test",
			Type: obs.OutputTypeKafka,
			TLS:  &obs.OutputTLSSpec{},
		})
		result := NewTlsEnabled(adapter, secrets, utils.NoOptions)
		Expect(result).To(BeNil())
	})
})
