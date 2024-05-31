package common

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	. "github.com/openshift/cluster-logging-operator/test/matchers"
	corev1 "k8s.io/api/core/v1"
)

var _ = Describe("#ValidateConfigMapOrSecretKey", func() {

	const (
		name       = "anAttribute"
		secretName = "mySecret"
		keyName    = "foo"
		missingKey = "missingKey"
	)
	var (
		secretKeys []*obs.ConfigMapOrSecretKey
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
			secretKey = &obs.ConfigMapOrSecretKey{
				Key: keyName,
				Secret: &corev1.LocalObjectReference{
					Name: secretName,
				},
			}
			secretKeyWithKeyMissing = &obs.ConfigMapOrSecretKey{
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
			secretKeys = []*obs.ConfigMapOrSecretKey{
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
			Expect(ValidateConfigMapOrSecretKey(name, secretKeys, secrets, configMaps)).To(HaveCondition(obs.ValidationCondition, true, obs.ReasonSecretNotFound, `(secret|configmap) ".*" not found for "`+name))
		})
		It("should fail when the secretKey is missing from the secret", func() {
			secretKeys = append(secretKeys, secretKeyWithKeyMissing)
			secrets[secretKey.Secret.Name] = aSecret
			Expect(ValidateConfigMapOrSecretKey(name, secretKeys, secrets, configMaps)).To(HaveCondition(obs.ValidationCondition, true, obs.ReasonSecretKeyNotFound, `key ".*" not found in (secret|configmap) ".*" for "`+name))

		})
		It("should pass when the secret and secretKey exist", func() {
			secretKeys = append(secretKeys, secretKey)
			secrets[secretKey.Secret.Name] = aSecret
			Expect(ValidateConfigMapOrSecretKey(name, secretKeys, secrets, configMaps)).To(BeEmpty())
		})
		It("should pass when there are no secrets spec'd", func() {
			Expect(ValidateConfigMapOrSecretKey(name, []*obs.ConfigMapOrSecretKey{}, secrets, configMaps)).To(BeEmpty())
		})
	})
	Context("when validating configmaps", func() {
		const configmapName = "myconfigmap"
		var (
			secretKey = &obs.ConfigMapOrSecretKey{
				Key: keyName,
				ConfigMap: &corev1.LocalObjectReference{
					Name: configmapName,
				},
			}
			secretKeyWithKeyMissing = &obs.ConfigMapOrSecretKey{
				Key: missingKey,
				ConfigMap: &corev1.LocalObjectReference{
					Name: configmapName,
				},
			}
			aSecret = &corev1.ConfigMap{
				Data: map[string]string{
					keyName: "somevalue",
				},
			}
		)
		BeforeEach(func() {
			secretKeys = []*obs.ConfigMapOrSecretKey{
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
			Expect(ValidateConfigMapOrSecretKey(name, secretKeys, secrets, configMaps)).To(HaveCondition(obs.ValidationCondition, true, obs.ReasonConfigMapNotFound, `(secret|configmap) ".*" not found for "`+name))
		})
		It("should fail when the key is missing from the configmap", func() {
			secretKeys = append(secretKeys, secretKeyWithKeyMissing)
			configMaps[secretKey.ConfigMap.Name] = aSecret
			Expect(ValidateConfigMapOrSecretKey(name, secretKeys, secrets, configMaps)).To(HaveCondition(obs.ValidationCondition, true, obs.ReasonConfigMapKeyNotFound, `key ".*" not found in (secret|configmap) ".*" for "`+name))

		})
		It("should pass when the configmap and key exist", func() {
			secretKeys = append(secretKeys, secretKey)
			configMaps[secretKey.ConfigMap.Name] = aSecret
			Expect(ValidateConfigMapOrSecretKey(name, secretKeys, secrets, configMaps)).To(BeEmpty())
		})
		It("should pass when there are no secrets or configmaps are spec'd", func() {
			Expect(ValidateConfigMapOrSecretKey(name, []*obs.ConfigMapOrSecretKey{}, secrets, configMaps)).To(BeEmpty())
		})
	})
})
