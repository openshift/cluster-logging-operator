package outputs

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	internalcontext "github.com/openshift/cluster-logging-operator/internal/api/context"
	internalobs "github.com/openshift/cluster-logging-operator/internal/api/observability"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	"github.com/openshift/cluster-logging-operator/test/matchers"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Validating outputs", func() {

	var (
		foo = "foo"
		bar = "bar"

		myRoleArn    = "arn:aws:iam::123456789012:role/my-role-to-assume"
		otherRoleArn = "arn:aws:iam::123456789012:role/other-role-to-assume"

		fooSecret = &corev1.Secret{
			ObjectMeta: v1.ObjectMeta{
				Name:      foo,
				Namespace: constants.OpenshiftNS,
			},
			Data: map[string][]byte{
				constants.AwsSecretAccessKey: []byte("some-key"),
				constants.AwsAccessKeyID:     []byte("some-key-id"),
				constants.AwsCredentialsKey:  []byte(myRoleArn),
			},
		}

		barSecret = &corev1.Secret{
			ObjectMeta: v1.ObjectMeta{
				Name:      bar,
				Namespace: constants.OpenshiftNS,
			},
			Data: map[string][]byte{
				constants.AwsSecretAccessKey: []byte("other-key"),
				constants.AwsAccessKeyID:     []byte("other-key-id"),
				constants.AwsCredentialsKey:  []byte(otherRoleArn),
			},
		}
	)

	// Helper function to create context
	createContext := func(outputs []obs.OutputSpec) internalcontext.ForwarderContext {
		return internalcontext.ForwarderContext{
			Forwarder: &obs.ClusterLogForwarder{
				Spec: obs.ClusterLogForwarderSpec{
					Outputs: outputs,
					Pipelines: []obs.PipelineSpec{
						{
							Name:       "dummy",
							OutputRefs: internalobs.Outputs(outputs).Names(),
						},
					},
				},
			},
			Secrets: map[string]*corev1.Secret{
				fooSecret.Name: fooSecret,
				barSecret.Name: barSecret,
			},
			AdditionalContext: utils.Options{},
		}
	}

	// Helper function to create CloudWatch output spec with AccessKey auth
	createAccessKeySpec := func(name, secretName string) obs.OutputSpec {
		return obs.OutputSpec{
			Name: name,
			Type: obs.OutputTypeCloudwatch,
			Cloudwatch: &obs.Cloudwatch{
				Authentication: &obs.AwsAuthentication{
					Type: obs.AuthTypeAccessKey,
					AwsAccessKey: &obs.AwsAccessKey{
						KeySecret: obs.SecretReference{
							SecretName: secretName,
							Key:        constants.AwsSecretAccessKey,
						},
						KeyId: obs.SecretReference{
							SecretName: secretName,
							Key:        constants.AwsAccessKeyID,
						},
					},
				},
			},
		}
	}

	// Helper function to create CloudWatch output spec with IamRole auth
	createIAMRoleSpec := func(name, secretName string) obs.OutputSpec {
		return obs.OutputSpec{
			Name: name,
			Type: obs.OutputTypeCloudwatch,
			Cloudwatch: &obs.Cloudwatch{
				Authentication: &obs.AwsAuthentication{
					Type: obs.AuthTypeIAMRole,
					IamRole: &obs.AwsRole{
						RoleARN: obs.SecretReference{
							SecretName: secretName,
							Key:        constants.AwsCredentialsKey,
						},
						Token: obs.BearerToken{
							From: obs.BearerTokenFromServiceAccount,
						},
					},
				},
			},
		}
	}

	Context("with multiple CloudWatch", func() {
		DescribeTable("when validating auth",
			func(outputs []obs.OutputSpec, expectedMessages []string) {
				context := createContext(outputs)
				Validate(context)
				Expect(context.Forwarder.Status.OutputConditions).To(HaveLen(len(outputs)))
				for i, message := range expectedMessages {
					Expect(context.Forwarder.Status.OutputConditions[i].Message).To(ContainSubstring(message))
				}
			},
			Entry("should accept multiple CloudWatch outputs with different static keys",
				[]obs.OutputSpec{createAccessKeySpec("output1", foo), createAccessKeySpec("output2", bar)},
				[]string{"is valid", "is valid"},
			),
			Entry("should accept CloudWatch outputs with one static key and one IAM role",
				[]obs.OutputSpec{createAccessKeySpec("output1", foo), createIAMRoleSpec("output2", foo)},
				[]string{"is valid", "is valid"},
			),
			Entry("should accept multiple CloudWatch outputs with same IAM role",
				[]obs.OutputSpec{createIAMRoleSpec("output1", foo), createIAMRoleSpec("output2", foo)},
				[]string{"is valid", "is valid"},
			),
			Entry("should accept multiple CloudWatch outputs with different IAM roles",
				[]obs.OutputSpec{createIAMRoleSpec("output1", foo), createIAMRoleSpec("output2", bar)},
				[]string{"is valid", "is valid"},
			),
			Entry("should accept multiple CloudWatch outputs with different IAM roles and static key",
				[]obs.OutputSpec{createIAMRoleSpec("output1", foo), createAccessKeySpec("output2", bar), createIAMRoleSpec("output3", bar)},
				[]string{"is valid", "is valid", "is valid"},
			),
		)
	})
	Context("when not referenced by any pipeline", func() {
		It("should generate a failure validation message", func() {
			outputName := "unreferenced"
			context := createContext([]obs.OutputSpec{
				{
					Name: outputName,
					Type: obs.OutputTypeHTTP,
					HTTP: &obs.HTTP{
						URLSpec: obs.URLSpec{
							URL: "http://nowhere",
						},
					},
				},
			})
			context.Forwarder.Spec.Pipelines[0].OutputRefs = []string{}
			Validate(context)
			Expect(context.Forwarder.Status.OutputConditions[0]).To(matchers.MatchCondition(outputName, false, obs.ReasonValidationFailure, ".*"))
		})
	})
})
