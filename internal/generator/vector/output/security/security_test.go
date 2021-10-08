package security

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openshift/cluster-logging-operator/internal/constants"
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
	Context("#Has Keys", func() {
		It("should be true if and only if all keys present", func() {
			secret.Data = map[string][]byte{"a": nil, "b": nil, "c": nil, "d": nil}
			Expect(HasKeys(secret, "a", "b", "c")).To(BeTrue())
			secret.Data = map[string][]byte{"a": nil, "c": nil, "d": nil}
			Expect(HasKeys(secret, "a", "b", "c")).To(BeFalse())
			// nil/empty cases.
			Expect(HasKeys(nil, "a", "b", "c")).To(BeFalse())
			Expect(HasKeys(&corev1.Secret{}, "a", "b", "c")).To(BeFalse())
		})
	})
	Context("#HasKeys", func() {
		It("should be true if and only if all keys present", func() {
			secret.Data = map[string][]byte{"a": nil, "b": nil, "c": nil, "d": nil}
			Expect(HasKeys(secret, "a", "b", "c")).To(BeTrue())
			secret.Data = map[string][]byte{"a": nil, "c": nil, "d": nil}
			Expect(HasKeys(secret, "a", "b", "c")).To(BeFalse())
			// nil/empty cases.
			Expect(HasKeys(nil, "a", "b", "c")).To(BeFalse())
			Expect(HasKeys(&corev1.Secret{}, "a", "b", "c")).To(BeFalse())
		})
	})
	Context("#TryKeys", func() {
		It("should return first key present", func() {

			secret.Data = map[string][]byte{"x": {1}, "y": {2}}
			_, ok := TryKeys(secret, "a", "b", "c")
			Expect(ok).To(BeFalse())

			v, ok := TryKeys(secret, "a", "b", "x")
			Expect(ok).To(BeTrue())
			Expect(v).To(Equal([]byte{1}))

			v, ok = TryKeys(secret, "x", "b", "c")
			Expect(ok).To(BeTrue())
			Expect(v).To(Equal([]byte{1}))

			v, ok = TryKeys(secret, "y", "x")
			Expect(ok).To(BeTrue())
			Expect(v).To(Equal([]byte{2}))

			// nil/empty cases.
			_, ok = TryKeys(nil, "a", "b", "c")
			Expect(ok).To(BeFalse())
			_, ok = TryKeys(nil, "a", "b", "c")
			Expect(ok).To(BeFalse())
		})
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
