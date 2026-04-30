package outputs

import (
	"encoding/json"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	internalcontext "github.com/openshift/cluster-logging-operator/internal/api/context"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	corev1 "k8s.io/api/core/v1"
)

var _ = Describe("validating GCL auth", func() {
	const (
		secretName = "gcp-secret"
		credKey    = "credentials.json"
	)

	var (
		validServiceAccount = gcpCredentialFile{
			Type: "service_account",
		}

		validExternalAccount = gcpCredentialFile{
			Type: "external_account",
			CredentialSource: &gcpCredentialSource{
				File: filepath.Join(constants.ServiceAccountSecretPath, constants.TokenKey),
			},
		}

		makeSecret = func(creds interface{}) *corev1.Secret {
			data, _ := json.Marshal(creds)
			return &corev1.Secret{
				Data: map[string][]byte{
					credKey: data,
				},
			}
		}

		makeContext = func(spec obs.OutputSpec, secret *corev1.Secret) internalcontext.ForwarderContext {
			secrets := map[string]*corev1.Secret{}
			if secret != nil {
				secrets[secretName] = secret
			}
			return internalcontext.ForwarderContext{
				Forwarder: &obs.ClusterLogForwarder{
					Spec: obs.ClusterLogForwarderSpec{
						Outputs: []obs.OutputSpec{spec},
					},
				},
				Secrets: secrets,
			}
		}

		credRef = &obs.SecretReference{
			SecretName: secretName,
			Key:        credKey,
		}

		serviceAccountSpec = obs.OutputSpec{
			Name: "gcl-output",
			Type: obs.OutputTypeGoogleCloudLogging,
			GoogleCloudLogging: &obs.GoogleCloudLogging{
				Authentication: &obs.GoogleCloudLoggingAuthentication{
					Credentials: credRef,
				},
			},
		}

		workloadIdentitySpec = obs.OutputSpec{
			Name: "gcl-output",
			Type: obs.OutputTypeGoogleCloudLogging,
			GoogleCloudLogging: &obs.GoogleCloudLogging{
				Authentication: &obs.GoogleCloudLoggingAuthentication{
					Credentials: credRef,
					Token: &obs.BearerToken{
						From: obs.BearerTokenFromServiceAccount,
					},
				},
			},
		}
	)

	Context("service_account credentials", func() {
		It("should pass with valid service_account credentials", func() {
			ctx := makeContext(serviceAccountSpec, makeSecret(validServiceAccount))
			Expect(ValidateGCLAuth(serviceAccountSpec, ctx)).To(BeEmpty())
		})

		It("should fail with malformed JSON", func() {
			secret := &corev1.Secret{
				Data: map[string][]byte{
					credKey: []byte("not json"),
				},
			}
			ctx := makeContext(serviceAccountSpec, secret)
			res := ValidateGCLAuth(serviceAccountSpec, ctx)
			Expect(res).To(ContainElement(ContainSubstring("not valid JSON")))
		})

		It("should fail when secret does not exist", func() {
			ctx := makeContext(serviceAccountSpec, nil)
			res := ValidateGCLAuth(serviceAccountSpec, ctx)
			Expect(res).To(ContainElement(ContainSubstring("not found")))
		})
	})

	Context("external_account credentials (WIF)", func() {
		It("should pass with valid external_account credentials and token", func() {
			ctx := makeContext(workloadIdentitySpec, makeSecret(validExternalAccount))
			Expect(ValidateGCLAuth(workloadIdentitySpec, ctx)).To(BeEmpty())
		})

		It("should fail when credential_source is missing", func() {
			creds := gcpCredentialFile{
				Type: "external_account",
			}
			ctx := makeContext(workloadIdentitySpec, makeSecret(creds))
			res := ValidateGCLAuth(workloadIdentitySpec, ctx)
			Expect(res).To(ContainElement(ContainSubstring("credential_source")))
		})

		It("should fail when credential_source.file does not match expected token path", func() {
			creds := validExternalAccount
			creds.CredentialSource = &gcpCredentialSource{
				File: "/wrong/path/token",
			}
			ctx := makeContext(workloadIdentitySpec, makeSecret(creds))
			res := ValidateGCLAuth(workloadIdentitySpec, ctx)
			Expect(res).To(ContainElement(ContainSubstring("does not match expected token path")))
		})

		It("should fail when secret does not exist", func() {
			ctx := makeContext(workloadIdentitySpec, nil)
			res := ValidateGCLAuth(workloadIdentitySpec, ctx)
			Expect(res).To(ContainElement(ContainSubstring("not found")))
		})
	})

	Context("unsupported credentials type", func() {
		It("should fail with unsupported type", func() {
			creds := gcpCredentialFile{
				Type: "unknown_type",
			}
			ctx := makeContext(serviceAccountSpec, makeSecret(creds))
			res := ValidateGCLAuth(serviceAccountSpec, ctx)
			Expect(res).To(ContainElement(ContainSubstring("unsupported type")))
		})
	})
})
