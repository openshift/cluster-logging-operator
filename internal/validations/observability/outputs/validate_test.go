package outputs

import (
	"fmt"
	"github.com/golang-collections/collections/set"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	internalcontext "github.com/openshift/cluster-logging-operator/internal/api/context"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

var _ = Describe("validating multiple CloudWatch outputs auth", func() {

	var (
		foo       = "foo"
		bar       = "bar"
		fooSecret = &corev1.Secret{
			ObjectMeta: v1.ObjectMeta{
				Name:      foo,
				Namespace: constants.OpenshiftNS,
			},
			Data: map[string][]byte{
				constants.AWSSecretAccessKey: []byte("some-key"),
				constants.AWSAccessKeyID:     []byte("some-key-id"),
				constants.AWSCredentialsKey:  []byte("some-creds-key"),
			},
		}
		barSecret = &corev1.Secret{
			ObjectMeta: v1.ObjectMeta{
				Name:      bar,
				Namespace: constants.OpenshiftNS,
			},
			Data: map[string][]byte{
				constants.AWSSecretAccessKey: []byte("other-key"),
				constants.AWSAccessKeyID:     []byte("other-key-id"),
				constants.AWSCredentialsKey:  []byte("other-creds-key"),
			},
		}
	)

	// Helper function to create context
	createContext := func(outputs []obs.OutputSpec) internalcontext.ForwarderContext {
		return internalcontext.ForwarderContext{
			Forwarder: &obs.ClusterLogForwarder{
				Spec: obs.ClusterLogForwarderSpec{
					Outputs: outputs,
				},
			},
			Secrets: map[string]*corev1.Secret{
				fooSecret.Name: fooSecret,
				barSecret.Name: barSecret,
			},
		}
	}

	// Helper function to create CloudWatch output spec with AccessKey auth
	//createAccessKeySpec := func(name, secretName string) obs.OutputSpec {
	//	return obs.OutputSpec{
	//		Name: name,
	//		Type: obs.OutputTypeCloudwatch,
	//		Cloudwatch: &obs.Cloudwatch{
	//			Authentication: &obs.CloudwatchAuthentication{
	//				Type: obs.CloudwatchAuthTypeAccessKey,
	//				AWSAccessKey: &obs.CloudwatchAWSAccessKey{
	//					KeySecret: &obs.SecretKey{
	//						Secret: &corev1.LocalObjectReference{
	//							Name: secretName,
	//						},
	//						Key: constants.AWSSecretAccessKey,
	//					},
	//					KeyID: &obs.SecretKey{
	//						Secret: &corev1.LocalObjectReference{
	//							Name: secretName,
	//						},
	//						Key: constants.AWSAccessKeyID,
	//					},
	//				},
	//			},
	//		},
	//	}
	//}

	// Helper function to create CloudWatch output spec with IAMRole auth
	createIAMRoleSpec := func(name, secretName string) obs.OutputSpec {
		return obs.OutputSpec{
			Name: name,
			Type: obs.OutputTypeCloudwatch,
			Cloudwatch: &obs.Cloudwatch{
				Authentication: &obs.CloudwatchAuthentication{
					Type: obs.CloudwatchAuthTypeIAMRole,
					IAMRole: &obs.CloudwatchIAMRole{
						RoleARN: &obs.SecretKey{
							Secret: &corev1.LocalObjectReference{
								Name: secretName,
							},
							Key: constants.AWSCredentialsKey,
						},
					},
				},
			},
		}
	}

	Context("#ValidateCloudWatchAuth", func() {

		DescribeTable("should validate CloudWatch output specs auth",
			func(outputs []obs.OutputSpec, expectedMessages []string) {
				context := createContext(outputs)
				Validate(context)
				Expect(context.Forwarder.Status.Outputs).To(HaveLen(len(outputs)))
				for i, message := range expectedMessages {
					Expect(context.Forwarder.Status.Outputs[i].Message).To(ContainSubstring(message))
				}
			},
			//Entry("multiple CloudWatch outputs with different static keys",
			//	[]obs.OutputSpec{createAccessKeySpec("output1", foo), createAccessKeySpec("output2", bar)},
			//	[]string{"is valid", "is valid"},
			//),
			//Entry("CloudWatch outputs with one static key and one IAM role",
			//	[]obs.OutputSpec{createAccessKeySpec("output1", foo), createIAMRoleSpec("output2", foo)},
			//	[]string{"is valid", "is valid"},
			//),
			Entry("multiple CloudWatch outputs with same IAM role",
				[]obs.OutputSpec{createIAMRoleSpec("output1", foo), createIAMRoleSpec("output2", foo)},
				[]string{"is valid", "is valid"},
			),
			//Entry("multiple CloudWatch outputs with different IAM roles",
			//	[]obs.OutputSpec{createIAMRoleSpec("output1", foo), createIAMRoleSpec("output2", bar)},
			//	[]string{"is valid", "found various CloudWatch RoleARN auth in outputs spec"},
			//),
			//Entry("multiple CloudWatch outputs with different IAM roles and static key",
			//	[]obs.OutputSpec{createIAMRoleSpec("output1", foo), createAccessKeySpec("output2", bar), createIAMRoleSpec("output3", bar)},
			//	[]string{"is valid", "is valid", "found various CloudWatch RoleARN auth in outputs spec"},
			//),
		)
	})
})

func TestName(t *testing.T) {
	roles := set.New()

	role1 := &obs.CloudwatchIAMRole{
		RoleARN: &obs.SecretKey{
			Secret: &corev1.LocalObjectReference{
				Name: "foo",
			},
			Key: constants.AWSCredentialsKey,
		},
	}

	role2 := &obs.CloudwatchIAMRole{
		RoleARN: &obs.SecretKey{
			Secret: &corev1.LocalObjectReference{
				Name: "foo",
			},
			Key: constants.AWSCredentialsKey,
		},
	}

	roles.Insert(*role1)
	roles.Insert(*role2)

	//must be 1 because role1 == role2 but got 2
	fmt.Println(roles.Len())

}
