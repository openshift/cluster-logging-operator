package outputs

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	corev1 "k8s.io/api/core/v1"
)

var _ = Describe("validating output secrets", func() {
	var (
		spec = obs.OutputSpec{
			Type: obs.OutputTypeLoki,
			Loki: &obs.Loki{
				Authentication: &obs.HTTPAuthentication{
					Password: &obs.SecretKey{
						Key: "mypassword",
						Secret: &corev1.LocalObjectReference{
							Name: "mysecret",
						},
					},
				},
			},
			TLS: &obs.OutputTLSSpec{
				TLSSpec: obs.TLSSpec{
					CA: &obs.ConfigMapOrSecretKey{
						Key: "foo",
						ConfigMap: &corev1.LocalObjectReference{
							Name: "immissing",
						},
					},
				},
			},
		}
		secrets    = map[string]*corev1.Secret{}
		configMaps = map[string]*corev1.ConfigMap{}
	)
	Context("#ValidateSecretsAndConfigMaps", func() {
		It("should validate secrets if spec'd", func() {
			conds := ValidateSecretsAndConfigMaps(spec, secrets, configMaps)
			Expect(conds).To(Not(BeEmpty()))
		})
	})
})
