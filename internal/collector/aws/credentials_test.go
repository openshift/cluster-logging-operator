package aws_test

import (
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/collector/aws"
	v1 "k8s.io/api/core/v1"
)

var _ = Describe("AWS credentials generation", func() {
	const (
		roleArn      = "arn:aws:iam::123456789012:role/foo"
		altRoleArn   = "arn:aws-us-gov:iam::123456789012:role/anli-sts-25690-openshift-logging-cloudwatch-credentials"
		saTokenPath  = "/var/run/ocp-collector/serviceaccount/token"
		secretName   = "cw-secret"
		altSecretKey = "alt-role-arn"
	)

	var (
		cwSecret = map[string]*v1.Secret{
			secretName: {
				Data: map[string][]byte{
					"role_arn":  []byte(roleArn),
					altSecretKey: []byte(altRoleArn),
				},
			},
		}
	)

	Describe("#GatherAWSWebIdentities", func() {
		It("should be nil if no cloudwatch outputs", func() {
			outputs := []obs.OutputSpec{
				{
					Name: "es-out",
					Type: obs.OutputTypeElasticsearch,
				},
			}
			Expect(aws.GatherAWSWebIdentities(outputs, cwSecret)).To(BeNil())
		})

		It("should be nil if secrets are nil and no cloudwatch outputs", func() {
			outputs := []obs.OutputSpec{
				{
					Name: "es-out",
					Type: obs.OutputTypeElasticsearch,
				},
			}

			Expect(aws.GatherAWSWebIdentities(outputs, nil)).To(BeNil())
		})

		It("should be nil if secrets are nil and outputs are nil", func() {
			Expect(aws.GatherAWSWebIdentities(nil, nil)).To(BeNil())
		})

		DescribeTable("should produce expected CloudWatch profiles for ",
			func(webIds []aws.AWSWebIdentity, expFile string) {
				exp, err := credFiles.ReadFile(expFile)
				Expect(err).To(BeNil())

				w := &strings.Builder{}
				err = aws.AWSCredentialsTemplate.Execute(w, webIds)
				Expect(err).To(BeNil())
				Expect(w.String()).To(Equal(string(exp)))
			},
			Entry("single credentials", []aws.AWSWebIdentity{
				{
					Name:                 "default",
					RoleARN:              "arn:aws:iam::123456789012:role/test-default",
					WebIdentityTokenFile: saTokenPath,
				},
			}, "cw_single_credentials"),
			Entry("multiple credentials", []aws.AWSWebIdentity{
				{
					Name:                 "foo",
					RoleARN:              "arn:aws:iam::123456789012:role/test-foo",
					WebIdentityTokenFile: saTokenPath,
				},
				{
					Name:                 "bar",
					RoleARN:              "arn:aws:iam::123456789012:role/test-bar",
					WebIdentityTokenFile: saTokenPath,
				},
			}, "cw_multiple_credentials"),
			Entry("assume role profile", []aws.AWSWebIdentity{
				{
					Name:                 "default",
					RoleARN:              "arn:aws:iam::123456789012:role/test-default",
					WebIdentityTokenFile: saTokenPath,
					AssumeRoleARN:        "arn:aws:iam::987654321098:role/cross-account-role",
					ExternalID:           "unique-external-id",
					SessionName:          "custom-session-name",
				},
			}, "cw_assume_role_single"),
		)
	})
})