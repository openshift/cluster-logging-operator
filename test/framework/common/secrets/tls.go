package secrets

import (
	"fmt"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	"github.com/openshift/cluster-logging-operator/test/helpers/certificate"
	corev1 "k8s.io/api/core/v1"
	"net"
)

// NewTLSSecret generates a CA and returns a set of signed certificates and TLS spec
func NewTLSSecret(namespace, name, serviceNamespace, serviceName string) (*corev1.Secret, obs.TLSSpec) {
	ca := certificate.NewCA(nil, "Test Self-signed Root CA")
	serverCert := certificate.NewCert(ca, "", serviceName,
		fmt.Sprintf("%s.%s.svc", serviceName, serviceNamespace),
		"functional",
		"localhost",
		net.IPv4(127, 0, 0, 1),
		net.IPv6loopback,
	)

	data := map[string][]byte{
		constants.ClientPrivateKey:   serverCert.PrivateKeyPEM(),
		constants.ClientCertKey:      serverCert.CertificatePEM(),
		constants.TrustedCABundleKey: ca.CertificatePEM(),
	}
	return runtime.NewSecret(namespace, name, data),
		obs.TLSSpec{
			CA: &obs.ConfigReference{
				Key: constants.TrustedCABundleKey,
				Secret: &corev1.LocalObjectReference{
					Name: name,
				},
			},
			Certificate: &obs.ConfigReference{
				Key: constants.ClientCertKey,
				Secret: &corev1.LocalObjectReference{
					Name: name,
				},
			},
			Key: &obs.SecretConfigReference{
				Key: constants.ClientPrivateKey,
				Secret: &corev1.LocalObjectReference{
					Name: name,
				},
			},
		}
}
