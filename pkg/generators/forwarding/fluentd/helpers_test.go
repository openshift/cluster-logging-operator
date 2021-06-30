package fluentd

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/pkg/constants"
	corev1 "k8s.io/api/core/v1"
)

var _ = Describe("Helpers for outputLabelConf", func() {
	var (
		conf   *outputLabelConf
		secret *corev1.Secret
	)
	BeforeEach(func() {
		secret = &corev1.Secret{
			Data: map[string][]byte{},
		}
		conf = &outputLabelConf{
			Name:   "my-output",
			Secret: secret,
		}
	})
	Context("#HasCABundle", func() {
		It("should recognize when the output secret is nil", func() {
			conf.Secret = nil
			Expect(conf.HasCABundle()).To(BeFalse())
		})
		It("should recognize when the output secret is missing the expected key", func() {
			Expect(conf.HasCABundle()).To(BeFalse())
		})
		It("should recognize when the output secret has the expected key", func() {
			conf.Secret.Data[constants.TrustedCABundleKey] = []byte{}
			Expect(conf.HasCABundle()).To(BeTrue())
		})
		It("should recognize when the output is the default", func() {
			conf.Name = logging.OutputNameDefault
			Expect(conf.HasCABundle()).To(BeTrue())
		})
	})
	Context("#HasTLSKeyAndCrt", func() {
		It("should recognize when the output secret is nil", func() {
			conf.Secret = nil
			Expect(conf.HasTLSKeyAndCrt()).To(BeFalse())
		})
		It("should recognize when the output secret is missing the private key", func() {
			conf.Secret.Data[constants.ClientCertKey] = []byte{}
			Expect(conf.HasTLSKeyAndCrt()).To(BeFalse())
		})
		It("should recognize when the output secret is missing the public key", func() {
			conf.Secret.Data[constants.ClientPrivateKey] = []byte{}
			Expect(conf.HasTLSKeyAndCrt()).To(BeFalse())
		})
		It("should recognize when the output secret has the private and public key", func() {
			conf.Secret.Data[constants.ClientPrivateKey] = []byte{}
			conf.Secret.Data[constants.ClientCertKey] = []byte{}
			Expect(conf.HasTLSKeyAndCrt()).To(BeTrue())
		})
		It("should recognize when the output is the default", func() {
			conf.Name = logging.OutputNameDefault
			Expect(conf.HasTLSKeyAndCrt()).To(BeTrue())
		})
	})
})
