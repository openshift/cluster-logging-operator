package common

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	corev1 "k8s.io/api/core/v1"
)

var _ = Describe("#ValidateValueReference", func() {

	const (
		name       = "anAttribute"
		secretName = "mySecret"
		keyName    = "foo"
		missingKey = "missingKey"
	)
	var (
		secretKeys []*obs.ValueReference
		cmKeys     []*obs.ValueReference
		secrets    map[string]*corev1.Secret
		configMaps map[string]*corev1.ConfigMap
	)
	Context("when validating secrets", func() {
		var (
			secretKey = &obs.ValueReference{
				Key:        keyName,
				SecretName: secretName,
			}
			secretKeyWithKeyMissing = &obs.ValueReference{
				Key:        missingKey,
				SecretName: secretName,
			}
			aSecret = &corev1.Secret{
				Data: map[string][]byte{
					keyName: []byte("somevalue"),
				},
			}
		)
		BeforeEach(func() {
			secretKeys = []*obs.ValueReference{
				{
					Key:        "always",
					SecretName: "always",
				},
			}
			secrets = map[string]*corev1.Secret{
				"always": {
					Data: map[string][]byte{
						"always": []byte("somevalue"),
					},
				},
			}
		})
		It("should fail when the secret does not exist", func() {
			secretKeys = append(secretKeys, secretKey)
			Expect(ValidateValueReference(secretKeys, secrets, configMaps)).To(ContainElement(MatchRegexp(`secret\[.*\] not found`)))
		})
		It("should fail when the secretKey is missing from the secret", func() {
			secretKeys = append(secretKeys, secretKeyWithKeyMissing)
			secrets[secretKey.SecretName] = aSecret
			Expect(ValidateValueReference(secretKeys, secrets, configMaps)).To(ContainElement(MatchRegexp(`secret\[.*\] not found`)))
		})
		It("should fail when the value identified by the key in the secret is empty", func() {
			secretKeys = append(secretKeys, secretKey)
			secrets[secretKey.SecretName] = &corev1.Secret{
				Data: map[string][]byte{
					keyName: []byte(""),
				},
			}
			Expect(ValidateValueReference(secretKeys, secrets, configMaps)).To(ContainElement(MatchRegexp(`secret\[.*\].*empty`)))
		})
		It("should pass when the secret and secretKey exist", func() {
			secretKeys = append(secretKeys, secretKey)
			secrets[secretKey.SecretName] = aSecret
			Expect(ValidateValueReference(secretKeys, secrets, configMaps)).To(BeEmpty())
		})
		It("should pass when there are no secrets spec'd", func() {
			Expect(ValidateValueReference([]*obs.ValueReference{}, secrets, configMaps)).To(BeEmpty())
		})
	})

	Context("when validating configmaps", func() {
		const configmapName = "myconfigmap"
		var (
			cmKey = &obs.ValueReference{
				Key:           keyName,
				ConfigMapName: configmapName,
			}
			cmKeyWithKeyNotFound = &obs.ValueReference{
				Key:           missingKey,
				ConfigMapName: configmapName,
			}
			aConfigMap = &corev1.ConfigMap{
				Data: map[string]string{
					keyName: "somevalue",
				},
			}
		)
		BeforeEach(func() {
			cmKeys = []*obs.ValueReference{
				{
					Key:           "always",
					ConfigMapName: "always",
				},
			}
			configMaps = map[string]*corev1.ConfigMap{
				"always": {
					Data: map[string]string{
						"always": "somevalue",
					},
				},
			}
		})
		It("should fail when one of the configmaps does not exist", func() {
			cmKeys = append(cmKeys, cmKey)
			Expect(ValidateValueReference(cmKeys, secrets, configMaps)).To(ContainElement(MatchRegexp(`configmap\[.*\] not found`)))
		})
		It("should fail when the key is not found in configmap", func() {
			cmKeys = append(cmKeys, cmKeyWithKeyNotFound)
			configMaps[cmKey.ConfigMapName] = aConfigMap
			Expect(ValidateValueReference(cmKeys, secrets, configMaps)).To(ContainElement(MatchRegexp(`configmap\[.*\] not found`)))
		})
		It("should pass when all configmaps and keys exist", func() {
			cmKeys = append(cmKeys, cmKey)
			configMaps[cmKey.ConfigMapName] = aConfigMap
			Expect(ValidateValueReference(cmKeys, secrets, configMaps)).To(BeEmpty())
		})
		It("should pass when there are no keys or names are spec'd", func() {
			Expect(ValidateValueReference([]*obs.ValueReference{}, secrets, configMaps)).To(BeEmpty())
		})
		It("should fail when the value in the configmap is empty", func() {
			configMaps["always"].Data["always"] = ""
			Expect(ValidateValueReference(cmKeys, secrets, configMaps)).To(ContainElement(MatchRegexp(`configmap\[.*\].*empty`)))
		})
	})
})
