package common

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	corev1 "k8s.io/api/core/v1"
)

var _ = Describe("#ValidateConfigReferences", func() {

	const (
		name       = "anAttribute"
		secretName = "mySecret"
		keyName    = "foo"
		missingKey = "missingKey"
	)
	var (
		secretKeys []*obs.ConfigReference
		secrets    map[string]*corev1.Secret
		configMaps map[string]*corev1.ConfigMap
	)
	BeforeEach(func() {
		secrets = map[string]*corev1.Secret{
			"always": {
				Data: map[string][]byte{
					"always": []byte("somevalue"),
				},
			},
		}
	})

	Context("when validating secrets", func() {
		var (
			secretKey = &obs.ConfigReference{
				Key: keyName,
				Secret: &corev1.LocalObjectReference{
					Name: secretName,
				},
			}
			secretKeyWithKeyMissing = &obs.ConfigReference{
				Key: missingKey,
				Secret: &corev1.LocalObjectReference{
					Name: secretName,
				},
			}
			aSecret = &corev1.Secret{
				Data: map[string][]byte{
					keyName: []byte("somevalue"),
				},
			}
		)
		BeforeEach(func() {
			secretKeys = []*obs.ConfigReference{
				{
					Key: "always",
					Secret: &corev1.LocalObjectReference{
						Name: "always",
					},
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
			Expect(ValidateConfigReferences(secretKeys, secrets, configMaps)).To(ContainElement(MatchRegexp(`(secret|configmap)\[.*\] not found`)))
		})
		It("should fail when the secretKey is missing from the secret", func() {
			secretKeys = append(secretKeys, secretKeyWithKeyMissing)
			secrets[secretKey.Secret.Name] = aSecret
			Expect(ValidateConfigReferences(secretKeys, secrets, configMaps)).To(ContainElement(MatchRegexp(`(secret|configmap)\[.*\] not found`)))
		})
		It("should fail when the value identified by the key in the secret is empty", func() {
			secretKeys = append(secretKeys, secretKey)
			secrets[secretKey.Secret.Name] = &corev1.Secret{
				Data: map[string][]byte{
					keyName: []byte(""),
				},
			}
			Expect(ValidateConfigReferences(secretKeys, secrets, configMaps)).To(ContainElement(MatchRegexp(`(secret|configmap)\[.*\].*empty`)))
		})
		It("should pass when the secret and secretKey exist", func() {
			secretKeys = append(secretKeys, secretKey)
			secrets[secretKey.Secret.Name] = aSecret
			Expect(ValidateConfigReferences(secretKeys, secrets, configMaps)).To(BeEmpty())
		})
		It("should pass when there are no secrets spec'd", func() {
			Expect(ValidateConfigReferences([]*obs.ConfigReference{}, secrets, configMaps)).To(BeEmpty())
		})
	})
	Context("when validating configmaps", func() {
		const configmapName = "myconfigmap"
		var (
			secretKey = &obs.ConfigReference{
				Key: keyName,
				ConfigMap: &corev1.LocalObjectReference{
					Name: configmapName,
				},
			}
			secretKeyWithKeyMissing = &obs.ConfigReference{
				Key: missingKey,
				ConfigMap: &corev1.LocalObjectReference{
					Name: configmapName,
				},
			}
			aConfigMap = &corev1.ConfigMap{
				Data: map[string]string{
					keyName: "somevalue",
				},
			}
		)
		BeforeEach(func() {
			secretKeys = []*obs.ConfigReference{
				{
					Key: "always",
					ConfigMap: &corev1.LocalObjectReference{
						Name: "always",
					},
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
		It("should fail when the configmap does not exist", func() {
			secretKeys = append(secretKeys, secretKey)
			Expect(ValidateConfigReferences(secretKeys, secrets, configMaps)).To(ContainElement(MatchRegexp(`(secret|configmap)\[.*\] not found`)))
		})
		It("should fail when the key is missing from the configmap", func() {
			secretKeys = append(secretKeys, secretKeyWithKeyMissing)
			configMaps[secretKey.ConfigMap.Name] = aConfigMap
			Expect(ValidateConfigReferences(secretKeys, secrets, configMaps)).To(ContainElement(MatchRegexp(`(secret|configmap)\[.*\] not found`)))
		})
		It("should pass when the configmap and key exist", func() {
			secretKeys = append(secretKeys, secretKey)
			configMaps[secretKey.ConfigMap.Name] = aConfigMap
			Expect(ValidateConfigReferences(secretKeys, secrets, configMaps)).To(BeEmpty())
		})
		It("should pass when there are no secrets or configmaps are spec'd", func() {
			Expect(ValidateConfigReferences([]*obs.ConfigReference{}, secrets, configMaps)).To(BeEmpty())
		})
	})
})
