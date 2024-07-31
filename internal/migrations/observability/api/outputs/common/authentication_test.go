package common

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("#MapAuthentication", func() {
	const secretName = "my-secret"
	var (
		secret = &corev1.Secret{
			ObjectMeta: v1.ObjectMeta{
				Name:      secretName,
				Namespace: "foo-space",
			},
			Data: map[string][]byte{
				constants.ClientUsername:     []byte("username"),
				constants.ClientPassword:     []byte("password"),
				constants.BearerTokenFileKey: []byte("token"),
			},
		}
	)
	It("should map logging HTTPAuthentication to observability HTTPAuthentication", func() {
		obsHttpAuth := &obs.HTTPAuthentication{
			Username: &obs.SecretReference{
				Key:        constants.ClientUsername,
				SecretName: secretName,
			},
			Password: &obs.SecretReference{
				Key:        constants.ClientPassword,
				SecretName: secretName,
			},
			Token: &obs.BearerToken{
				From: obs.BearerTokenFromSecret,
				Secret: &obs.BearerTokenSecretKey{
					Name: secretName,
					Key:  constants.BearerTokenFileKey,
				},
			},
		}
		actualHTTPAuth := MapHTTPAuth(secret)
		Expect(actualHTTPAuth).To(Equal(obsHttpAuth))
	})
})
