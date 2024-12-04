package common

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	internalcontext "github.com/openshift/cluster-logging-operator/internal/api/context"
	corev1 "k8s.io/api/core/v1"
)

var _ = Describe("Common Validation", func() {

	Context("#ValidateValueReference", func() {
		const (
			name       = "anAttribute"
			secretName = "mySecret"
			keyName    = "foo"
			missingKey = "missingKey"
		)
		var (
			secretKeys []*obs.ValueReference
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

		When("validating secrets", func() {
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
				Expect(ValidateValueReference(secretKeys, secrets, configMaps)).To(ContainElement(MatchRegexp(`(secret|configmap)\[.*\] not found`)))
			})
			It("should fail when the secretKey is missing from the secret", func() {
				secretKeys = append(secretKeys, secretKeyWithKeyMissing)
				secrets[secretKey.SecretName] = aSecret
				Expect(ValidateValueReference(secretKeys, secrets, configMaps)).To(ContainElement(MatchRegexp(`(secret|configmap)\[.*\] not found`)))
			})
			It("should fail when the value identified by the key in the secret is empty", func() {
				secretKeys = append(secretKeys, secretKey)
				secrets[secretKey.SecretName] = &corev1.Secret{
					Data: map[string][]byte{
						keyName: []byte(""),
					},
				}
				Expect(ValidateValueReference(secretKeys, secrets, configMaps)).To(ContainElement(MatchRegexp(`(secret|configmap)\[.*\].*empty`)))
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

		When("validating configmaps", func() {
			const configmapName = "myconfigmap"
			var (
				secretKey = &obs.ValueReference{
					Key:           keyName,
					ConfigMapName: configmapName,
				}
				secretKeyWithKeyMissing = &obs.ValueReference{
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
				secretKeys = []*obs.ValueReference{
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
			It("should fail when the configmap does not exist", func() {
				secretKeys = append(secretKeys, secretKey)
				Expect(ValidateValueReference(secretKeys, secrets, configMaps)).To(ContainElement(MatchRegexp(`(secret|configmap)\[.*\] not found`)))
			})
			It("should fail when the key is missing from the configmap", func() {
				secretKeys = append(secretKeys, secretKeyWithKeyMissing)
				configMaps[secretKey.ConfigMapName] = aConfigMap
				Expect(ValidateValueReference(secretKeys, secrets, configMaps)).To(ContainElement(MatchRegexp(`(secret|configmap)\[.*\] not found`)))
			})
			It("should pass when the configmap and key exist", func() {
				secretKeys = append(secretKeys, secretKey)
				configMaps[secretKey.ConfigMapName] = aConfigMap
				Expect(ValidateValueReference(secretKeys, secrets, configMaps)).To(BeEmpty())
			})
			It("should pass when there are no secrets or configmaps are spec'd", func() {
				Expect(ValidateValueReference([]*obs.ValueReference{}, secrets, configMaps)).To(BeEmpty())
			})
		})
	})

	Context("#IsEnabledAnnotation", func() {
		var (
			forwarder obs.ClusterLogForwarder
			context   internalcontext.ForwarderContext
			key       string
		)

		When("validating an annotation by key", func() {
			BeforeEach(func() {
				key = "some.annotation/for-testing"
				forwarder.Annotations = map[string]string{key: "true"}
				context = internalcontext.ForwarderContext{
					Forwarder: &forwarder,
				}
			})
			It("should return true when annotation value is 'true'", func() {
				Expect(IsEnabledAnnotation(context, key)).To(BeTrue())
			})
			It("should return true when annotation value is 'enabled'", func() {
				forwarder.Annotations[key] = "enabled"
				Expect(IsEnabledAnnotation(context, key)).To(BeTrue())
			})
			It("should return false when annotation is not found or not the correct value", func() {
				Expect(IsEnabledAnnotation(context, "another.annotation")).To(BeFalse())
				forwarder.Annotations[key] = "on"
				Expect(IsEnabledAnnotation(context, key)).To(BeFalse())
				// verify case sensitive
				forwarder.Annotations[key] = "Enabled"
				Expect(IsEnabledAnnotation(context, key)).To(BeFalse())
			})
		})
	})
})
