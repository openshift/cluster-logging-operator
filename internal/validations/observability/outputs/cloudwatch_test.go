package outputs

import (
	"github.com/golang-collections/collections/set"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	internalcontext "github.com/openshift/cluster-logging-operator/internal/api/context"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("validating CloudWatch auth", func() {
	Context("#ValidateCloudWatchAuth", func() {

		var (
			myRoleArn    = "arn:aws:iam::123456789012:role/my-role-to-assume"
			otherRoleArn = "arn:aws:iam::123456789012:role/other-role-to-assume"
			spec         = obs.OutputSpec{
				Name: "output",
				Type: obs.OutputTypeCloudwatch,
				Cloudwatch: &obs.Cloudwatch{
					Authentication: &obs.CloudwatchAuthentication{
						Type: obs.CloudwatchAuthTypeIAMRole,
						IAMRole: &obs.CloudwatchIAMRole{
							RoleARN: obs.SecretReference{
								SecretName: "foo",
								Key:        constants.AWSCredentialsKey,
							},
							Token: obs.BearerToken{
								From: obs.BearerTokenFromServiceAccount,
							},
						},
					},
				},
			}

			fooSecret = &corev1.Secret{
				ObjectMeta: v1.ObjectMeta{
					Name: "foo",
				},
				Data: map[string][]byte{
					constants.AWSCredentialsKey: []byte(myRoleArn),
				},
			}
			context = internalcontext.ForwarderContext{
				Forwarder: &obs.ClusterLogForwarder{
					Spec: obs.ClusterLogForwarderSpec{
						Outputs: []obs.OutputSpec{spec},
					},
				},
				Secrets: map[string]*corev1.Secret{
					fooSecret.Name: fooSecret,
				},
			}
		)

		It("should fail validation if meet different Role ARN", func() {
			roleARNs := set.New(otherRoleArn)
			context.AdditionalContext = utils.Options{
				RoleARNsOpt: roleARNs,
			}
			res := ValidateCloudWatchAuth(spec, context)
			Expect(res).ToNot(BeEmpty())
			Expect(len(res)).To(BeEquivalentTo(1))
			Expect(res[0]).To(BeEquivalentTo(ErrVariousRoleARNAuth))
		})

		It("should pass validation if Role ARNs are equals", func() {
			roleARNs := set.New(myRoleArn)
			context.AdditionalContext = utils.Options{
				RoleARNsOpt: roleARNs,
			}
			res := ValidateCloudWatchAuth(spec, context)
			Expect(res).To(BeEmpty())
		})

	})
})
