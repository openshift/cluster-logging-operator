package security

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openshift/cluster-logging-operator/pkg/constants"
	corev1 "k8s.io/api/core/v1"
)

var _ = Describe("Helpers for outputLabelConf", func() {
	var (
		secret *corev1.Secret
	)
	BeforeEach(func() {
		secret = &corev1.Secret{
			Data: map[string][]byte{},
		}
	})
	Context("#HasCABundle", func() {
		It("should recognize when the output secret is nil", func() {
			secret = nil
			Expect(HasCABundle(secret)).To(BeFalse())
		})
		It("should recognize when the output secret is missing the expected key", func() {
			Expect(HasCABundle(secret)).To(BeFalse())
		})
		It("should recognize when the output secret has the expected key", func() {
			secret.Data[constants.TrustedCABundleKey] = []byte{}
			Expect(HasCABundle(secret)).To(BeTrue())
		})
	})
	Context("#HasTLSKeyAndCrt", func() {
		It("should recognize when the output secret is nil", func() {
			secret = nil
			Expect(HasTLSCertAndKey(secret)).To(BeFalse())
		})
		It("should recognize when the output secret is missing the private key", func() {
			secret.Data[constants.ClientCertKey] = []byte{}
			Expect(HasTLSCertAndKey(secret)).To(BeFalse())
		})
		It("should recognize when the output secret is missing the public key", func() {
			secret.Data[constants.ClientPrivateKey] = []byte{}
			Expect(HasTLSCertAndKey(secret)).To(BeFalse())
		})
		It("should recognize when the output secret has the private and public key", func() {
			secret.Data[constants.ClientPrivateKey] = []byte{}
			secret.Data[constants.ClientCertKey] = []byte{}
			Expect(HasTLSCertAndKey(secret)).To(BeTrue())
		})
	})
})

func TestFluendConfGenerator(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Fluend Conf Generation")
}
