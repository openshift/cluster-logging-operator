package common

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	openshiftv1 "github.com/openshift/api/config/v1"
	logging "github.com/openshift/cluster-logging-operator/api/logging/v1"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("#MapTLS", func() {
	const secretName = "my-secret"
	var (
		secret = &corev1.Secret{
			ObjectMeta: v1.ObjectMeta{
				Name:      secretName,
				Namespace: "foo-space",
			},
			Data: map[string][]byte{
				constants.ClientCertKey:      []byte("cert"),
				constants.ClientPrivateKey:   []byte("privatekey"),
				constants.TrustedCABundleKey: []byte("cabundle"),
				constants.Passphrase:         []byte("pass"),
			},
		}
	)
	It("should map logging output TLS to observability TLS", func() {
		loggingTLS := &logging.OutputTLSSpec{
			InsecureSkipVerify: true,
			TLSSecurityProfile: &openshiftv1.TLSSecurityProfile{
				Type:   openshiftv1.TLSProfileType("foo"),
				Modern: &openshiftv1.ModernTLSProfile{},
			},
		}

		expTLS := &obs.OutputTLSSpec{
			InsecureSkipVerify: true,
			TLSSecurityProfile: &openshiftv1.TLSSecurityProfile{
				Type:   openshiftv1.TLSProfileType("foo"),
				Modern: &openshiftv1.ModernTLSProfile{},
			},
			TLSSpec: obs.TLSSpec{
				Certificate: &obs.ValueReference{
					Key:        constants.ClientCertKey,
					SecretName: secretName,
				},
				Key: &obs.SecretReference{
					Key:        constants.ClientPrivateKey,
					SecretName: secretName,
				},
				CA: &obs.ValueReference{
					Key:        constants.TrustedCABundleKey,
					SecretName: secretName,
				},
				KeyPassphrase: &obs.SecretReference{
					Key:        constants.Passphrase,
					SecretName: secretName,
				},
			},
		}
		actOutTLS := MapOutputTls(loggingTLS, secret)
		Expect(actOutTLS).To(Equal(expTLS))
	})
})
