package cloudwatch

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
)

var _ = Describe("Parsing strings for sts functionality", func() {
	var (
		altRoleArn        = "arn:aws-us-gov:iam::225746144451:role/anli-sts-25690-openshift-logging-cloudwatch-credentials"
		roleArn           = "arn:aws:iam::123456789012:role/my-role-from-secret"
		credentialsString = "[default]\nrole_arn = " + roleArn + "\nweb_identity_token_file = /var/run/secrets/token"
		secrets           = map[string]*corev1.Secret{
			"my-secret": {
				Data: map[string][]byte{
					"role_arn": []byte(roleArn),
				},
			},
		}
	)

	Context("pass a string containing a valid role_arn only", func() {
		Context("to ParseRoleArn() helper", func() {
			It("should return our specified valid role_arn", func() {
				results := ParseRoleArn(secrets["my-secret"])
				Expect(results).To(Equal(roleArn))
			})
			It("should return our specified valid role_arn when the partition is more than 'aws'", func() {
				secrets["other"] = &corev1.Secret{
					Data: map[string][]byte{
						"role_arn": []byte(altRoleArn),
					},
				}
				results := ParseRoleArn(secrets["other"])
				Expect(results).To(Equal(altRoleArn))
			})
		})
	})

	Context("pass a fully formatted sts secret with 'credentials' as key", func() {
		BeforeEach(func() {
			delete(secrets["my-secret"].Data, "role_arn")
			secrets["my-secret"] = &corev1.Secret{
				Data: map[string][]byte{
					"credentials": []byte(credentialsString),
				},
			}
		})
		Context("to ParseRoleArn() helper", func() {
			It("should be able to parse and return our specified valid role_arn value", func() {
				results := ParseRoleArn(secrets["my-secret"])
				Expect(results).To(Equal(roleArn))
			})
		})
	})

	Context("pass a fully formatted sts secret with 'role_arn' as the key", func() {
		BeforeEach(func() {
			// A properly formatted role arn is matched and should be found by regex
			secrets["my-secret"].Data["role_arn"] = []byte(credentialsString)
		})
		Context("to ParseRoleArn() helper", func() {
			It("should match and return only our valid role_arn", func() {
				results := ParseRoleArn(secrets["my-secret"])
				Expect(results).To(Equal(roleArn))
			})
		})
	})

	Context("pass an incorrectly formatted role_arn", func() {
		BeforeEach(func() {
			delete(secrets["my-secret"].Data, "credentials")
			roleArn = "arn:aws:iam::12345:role/my-role-from-secret" // incorrect format since not "arn:aws:iam::<12-digit-account-id>:"
			secrets["my-secret"].Data["role_arn"] = []byte(roleArn)
		})
		Context("to ParseRoleArn() helper", func() {
			It("should return an empty string since arn is not in valid format", func() {
				results := ParseRoleArn(secrets["my-secret"])
				Expect(results).To(BeEmpty())
			})
		})
	})
})
