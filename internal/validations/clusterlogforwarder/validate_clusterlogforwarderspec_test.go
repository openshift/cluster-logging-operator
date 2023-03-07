package clusterlogforwarder

import (
	"fmt"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	"github.com/onsi/gomega/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	loggingv1 "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/migrations"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	"github.com/openshift/cluster-logging-operator/internal/status"

	. "github.com/openshift/cluster-logging-operator/test"
	. "github.com/openshift/cluster-logging-operator/test/matchers"
	testRunTime "github.com/openshift/cluster-logging-operator/test/runtime"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	otherTargetName = "someothername"
)

// Match condition by type, status and reason if reason != "".
// Also match messageRegex if it is not empty.
func matchCondition(t loggingv1.ConditionType, s bool, r loggingv1.ConditionReason, messageRegex string) types.GomegaMatcher {
	var status corev1.ConditionStatus
	if s {
		status = corev1.ConditionTrue
	} else {
		status = corev1.ConditionFalse
	}
	fields := Fields{"Type": Equal(t), "Status": Equal(status)}
	if r != "" {
		fields["Reason"] = Equal(r)
	}
	if messageRegex != "" {
		fields["Message"] = MatchRegexp(messageRegex)
	}
	return MatchFields(IgnoreExtras, fields)
}

func HaveCondition(t loggingv1.ConditionType, s bool, r loggingv1.ConditionReason, messageRegex string) types.GomegaMatcher {
	return ContainElement(matchCondition(t, s, r, messageRegex))
}

var _ = Describe("Validate clusterlogforwarderspec", func() {
	var clfStatus *loggingv1.ClusterLogForwarderStatus

	BeforeEach(func() {
		clfStatus = &loggingv1.ClusterLogForwarderStatus{}
	})

	Context("input specs", func() {

		It("should fail if input does not have a name", func() {
			forwarderSpec := &loggingv1.ClusterLogForwarderSpec{
				Inputs: []loggingv1.InputSpec{
					{Name: ""},
				},
			}
			verifyInputs(forwarderSpec, clfStatus)
			Expect(clfStatus.Inputs["input_0_"]).To(HaveCondition("Ready", false, loggingv1.ReasonInvalid, "input must have a name"))
		})

		It("should fail if input name is one of the reserved names: application, infrastructure, audit", func() {
			forwarderSpec := &loggingv1.ClusterLogForwarderSpec{
				Inputs: []loggingv1.InputSpec{
					{Name: loggingv1.InputNameApplication},
				},
			}
			verifyInputs(forwarderSpec, clfStatus)
			Expect(clfStatus.Inputs[loggingv1.InputNameApplication]).To(HaveCondition("Ready", false, loggingv1.ReasonInvalid, "input name \"application\" is reserved"))
		})
		It("should fail if inputspec names are not unique", func() {
			forwarderSpec := &loggingv1.ClusterLogForwarderSpec{
				Inputs: []loggingv1.InputSpec{
					{Name: "my-app-logs",
						Application: &loggingv1.Application{}},
					{Name: "my-app-logs",
						Application: &loggingv1.Application{}},
				},
			}
			verifyInputs(forwarderSpec, clfStatus)
			Expect(clfStatus.Inputs["my-app-logs"]).To(HaveCondition("Ready", false, loggingv1.ReasonInvalid, "duplicate name: \"my-app-logs\""))
		})

		It("should fail when inputspec doesn't define one of application, infrastructure, or audit", func() {
			forwarderSpec := &loggingv1.ClusterLogForwarderSpec{
				Inputs: []loggingv1.InputSpec{
					{Name: "my-app-logs"},
				},
			}
			verifyInputs(forwarderSpec, clfStatus)
			Expect(clfStatus.Inputs["my-app-logs"]).To(HaveCondition("Ready", false, loggingv1.ReasonInvalid, "inputspec must define one or more of application, infrastructure, or audit"))
		})

		It("should remove all inputs if even one inputspec is invalid", func() {
			forwarderSpec := &loggingv1.ClusterLogForwarderSpec{
				Inputs: []loggingv1.InputSpec{
					{Name: "my-app-logs",
						Application: &loggingv1.Application{}},
					{Name: "invalid-input"},
				},
			}
			verifyInputs(forwarderSpec, clfStatus)
			Expect(clfStatus.Inputs["my-app-logs"]).To(HaveCondition("Ready", true, "", ""))
			Expect(clfStatus.Inputs["invalid-input"]).To(HaveCondition("Ready", false, loggingv1.ReasonInvalid, "inputspec must define one or more of application, infrastructure, or audit"))
		})

		It("should validate correctly with one valid input spec", func() {
			forwarderSpec := &loggingv1.ClusterLogForwarderSpec{
				Inputs: []loggingv1.InputSpec{
					{Name: "my-app-logs",
						Application: &loggingv1.Application{}},
				},
			}
			verifyInputs(forwarderSpec, clfStatus)
			Expect(clfStatus.Inputs["my-app-logs"]).To(HaveCondition("Ready", true, "", ""))
		})

		It("should validate correctly with more than one valid input spec", func() {
			forwarderSpec := &loggingv1.ClusterLogForwarderSpec{
				Inputs: []loggingv1.InputSpec{
					{Name: "my-app-logs",
						Application: &loggingv1.Application{}},
					{Name: "my-infra-logs",
						Infrastructure: &loggingv1.Infrastructure{}},
					{Name: "my-audit-logs",
						Audit: &loggingv1.Audit{}},
				},
			}
			verifyInputs(forwarderSpec, clfStatus)
			Expect(forwarderSpec.Inputs).To(HaveLen(3))
			Expect(clfStatus.Inputs["my-app-logs"]).To(HaveCondition("Ready", true, "", ""))
			Expect(clfStatus.Inputs["my-infra-logs"]).To(HaveCondition("Ready", true, "", ""))
			Expect(clfStatus.Inputs["my-audit-logs"]).To(HaveCondition("Ready", true, "", ""))
		})

		It("should validate correctly when input spec defines all three input source specs", func() {
			forwarderSpec := &loggingv1.ClusterLogForwarderSpec{
				Inputs: []loggingv1.InputSpec{
					{Name: "all-logs",
						Application:    &loggingv1.Application{},
						Infrastructure: &loggingv1.Infrastructure{},
						Audit:          &loggingv1.Audit{}},
				},
			}
			verifyInputs(forwarderSpec, clfStatus)
			Expect(forwarderSpec.Inputs).To(HaveLen(1))
			Expect(clfStatus.Inputs["all-logs"]).To(HaveCondition("Ready", true, "", ""))
		})

		It("should be valid with multiple input specs, multiple input source specs", func() {
			forwarderSpec := &loggingv1.ClusterLogForwarderSpec{
				Inputs: []loggingv1.InputSpec{
					{Name: "all-logs",
						Application:    &loggingv1.Application{},
						Infrastructure: &loggingv1.Infrastructure{},
						Audit:          &loggingv1.Audit{}},
					{Name: "app-infra-logs",
						Application:    &loggingv1.Application{},
						Infrastructure: &loggingv1.Infrastructure{},
					},
				},
			}
			verifyInputs(forwarderSpec, clfStatus)
			Expect(clfStatus.Inputs["all-logs"]).To(HaveCondition("Ready", true, "", ""))
			Expect(clfStatus.Inputs["app-infra-logs"]).To(HaveCondition("Ready", true, "", ""))
		})
	})

	Context("output specs", func() {
		var (
			client        client.Client
			cluster       *loggingv1.ClusterLogging
			extras        map[string]bool
			clfStatus     *loggingv1.ClusterLogForwarderStatus
			output        loggingv1.OutputSpec
			otherOutput   loggingv1.OutputSpec
			forwarderSpec *loggingv1.ClusterLogForwarderSpec
		)

		BeforeEach(func() {
			client = fake.NewClientBuilder().WithRuntimeObjects(runtime.NewSecret(
				constants.OpenshiftNS, constants.CollectorSecretName, nil,
			)).Build()
			cluster = &loggingv1.ClusterLogging{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: constants.OpenshiftNS,
				},
			}
			extras = map[string]bool{}
			clfStatus = &loggingv1.ClusterLogForwarderStatus{}

			output = loggingv1.OutputSpec{
				Name: "myOutput",
				Type: "elasticsearch",
				URL:  "http://here",
			}
			otherOutput = loggingv1.OutputSpec{
				Name: otherTargetName,
				Type: "elasticsearch",
				URL:  "http://there",
			}

			forwarderSpec = &loggingv1.ClusterLogForwarderSpec{
				Outputs: []loggingv1.OutputSpec{
					output,
					otherOutput,
				},
			}
		})

		It("should fail googlecloudlogging output with more than one spec", func() {
			forwarderSpec = &loggingv1.ClusterLogForwarderSpec{
				Outputs: []loggingv1.OutputSpec{
					{
						Name: "X",
						Type: "googleCloudLogging",
						OutputTypeSpec: loggingv1.OutputTypeSpec{
							GoogleCloudLogging: &loggingv1.GoogleCloudLogging{
								OrganizationID: "redhat",
								ProjectID:      "project1",
							},
						},
					},
				},
			}
			verifyOutputs(cluster, client, forwarderSpec, clfStatus, extras)
			Expect(clfStatus.Outputs["X"]).To(HaveCondition("Ready", false, loggingv1.ReasonInvalid,
				"output \"X\": Exactly one of billingAccountId, folderId, organizationId, or projectId must be set."))
		})

		It("should validate googlecloudlogging output", func() {
			forwarderSpec := &loggingv1.ClusterLogForwarderSpec{
				Outputs: []loggingv1.OutputSpec{
					{
						Name: "X",
						Type: "googleCloudLogging",
						OutputTypeSpec: loggingv1.OutputTypeSpec{
							GoogleCloudLogging: &loggingv1.GoogleCloudLogging{
								ProjectID: "project1",
							},
						},
					},
				},
			}
			verifyOutputs(cluster, client, forwarderSpec, clfStatus, extras)
			Expect(clfStatus.Outputs["X"]).To(HaveCondition("Ready", true, "", ""))
		})

		// Ref: https://issues.redhat.com/browse/LOG-3228
		It("should validate the default output as any other without adding a new one", func() {
			cluster.Spec.LogStore = &loggingv1.LogStoreSpec{
				Type: loggingv1.LogStoreTypeElasticsearch,
			}

			forwarderSpec := &loggingv1.ClusterLogForwarderSpec{
				Outputs: []loggingv1.OutputSpec{
					migrations.NewDefaultOutput(nil),
				},
			}

			extras[constants.MigrateDefaultOutput] = true

			// sanity check
			Expect(forwarderSpec.Outputs).To(HaveLen(1))
			verifyOutputs(cluster, client, forwarderSpec, clfStatus, extras)
			Expect(clfStatus.Outputs[loggingv1.OutputNameDefault]).To(HaveCondition(loggingv1.ConditionReady, true, "", ""))
			Expect(forwarderSpec.Outputs).To(HaveLen(1), "Exp. the number of outputs to remain unchanged")
		})
		It("should be invalid if outputs do not have unique names", func() {
			forwarderSpec.Outputs = append(forwarderSpec.Outputs, loggingv1.OutputSpec{
				Name: "myOutput",
				Type: "elasticsearch",
				URL:  "http://here",
			})
			// sanity check
			Expect(forwarderSpec.Outputs).To(HaveLen(3))
			verifyOutputs(cluster, client, forwarderSpec, clfStatus, extras)
			Expect(clfStatus.Outputs["myOutput"]).To(HaveCondition(loggingv1.ConditionReady, true, "", ""))
			Expect(clfStatus.Outputs["output_2_"]).To(HaveCondition(loggingv1.ConditionReady, false, loggingv1.ReasonInvalid, "duplicate"))
		})

		It("should be invalid if any outputs have empty names", func() {
			forwarderSpec.Outputs = append(forwarderSpec.Outputs, loggingv1.OutputSpec{
				Name: "",
				Type: "elasticsearch",
				URL:  "http://here",
			})
			verifyOutputs(cluster, client, forwarderSpec, clfStatus, extras)
			Expect(clfStatus.Outputs["output_2_"]).To(HaveCondition("Ready", false, "Invalid", "must have a name"))
		})

		It("should be valid even for outputs that conflict with the internally reserved name 'default' ", func() {
			outputs := append(forwarderSpec.Outputs, loggingv1.OutputSpec{
				Name: "default",
				Type: "elasticsearch",
				URL:  "http://here",
			})
			forwarderSpec.Outputs = outputs
			extras[constants.MigrateDefaultOutput] = true
			verifyOutputs(cluster, client, forwarderSpec, clfStatus, extras)
			Expect(len(forwarderSpec.Outputs)).To(Equal(len(outputs)), "Exp. outputs with an internal name of 'default' do be kept")
		})

		It("should be invalid if outputs have empty types", func() {
			forwarderSpec.Outputs = append(forwarderSpec.Outputs, loggingv1.OutputSpec{
				Name: "aName",
				URL:  "http://here",
			})
			verifyOutputs(cluster, client, forwarderSpec, clfStatus, extras)
			Expect(clfStatus.Outputs["aName"]).To(HaveCondition("Ready", false, "Invalid", "unknown.*\"\""))
		})

		It("should drop outputs that have unrecognized types", func() {
			forwarderSpec.Outputs = append(forwarderSpec.Outputs, loggingv1.OutputSpec{
				Name: "aName",
				Type: "foo",
				URL:  "http://here",
			})
			verifyOutputs(cluster, client, forwarderSpec, clfStatus, extras)
			Expect(clfStatus.Outputs["aName"]).To(HaveCondition("Ready", false, "Invalid", "unknown.*\"foo\""))
		})

		It("should drop outputs that have an invalid or non-absolute URL", func() {
			forwarderSpec.Outputs = []loggingv1.OutputSpec{
				{
					Name: "aName",
					Type: "fluentdForward",
					URL:  "relativeURLPath",
				},
				{
					Name: "bName",
					Type: "fluentdForward",
					URL:  ":invalid",
				},
			}
			verifyOutputs(cluster, client, forwarderSpec, clfStatus, extras)
			Expect(clfStatus.Outputs["aName"]).To(HaveCondition("Ready", false, "Invalid", "relativeURLPath"))
			Expect(clfStatus.Outputs["bName"]).To(HaveCondition("Ready", false, "Invalid", ":invalid"))
		})

		It("should drop Cloudwatch output without OutputTypeSpec", func() {
			forwarderSpec.Outputs = []loggingv1.OutputSpec{
				{
					Name: "cw",
					Type: loggingv1.OutputTypeCloudwatch,
				},
			}
			verifyOutputs(cluster, client, forwarderSpec, clfStatus, extras)
			Expect(clfStatus.Outputs["cw"]).To(HaveCondition("Ready", false, "Invalid", "Cloudwatch output requires type spec"))
		})

		It("should allow specific outputs that do not require URL", func() {
			forwarderSpec.Outputs = []loggingv1.OutputSpec{
				{
					Name: "aKafka",
					Type: loggingv1.OutputTypeKafka,
				},
				{
					Name: "aCloudwatch",
					Type: loggingv1.OutputTypeCloudwatch,
					OutputTypeSpec: loggingv1.OutputTypeSpec{
						Cloudwatch: &loggingv1.Cloudwatch{},
					},
				},
			}
			verifyOutputs(cluster, client, forwarderSpec, clfStatus, extras)
			Expect(clfStatus.Outputs["aKafka"]).To(HaveCondition("Ready", true, "", ""))
			Expect(clfStatus.Outputs["aCloudwatch"]).To(HaveCondition("Ready", true, "", ""))
		})

		It("should drop outputs that have secrets with no names", func() {
			forwarderSpec.Outputs = append(forwarderSpec.Outputs, loggingv1.OutputSpec{
				Name:   "aName",
				Type:   "elasticsearch",
				URL:    "https://somewhere",
				Secret: &loggingv1.OutputSecretSpec{},
			})
			verifyOutputs(cluster, client, forwarderSpec, clfStatus, extras)
			Expect(clfStatus.Outputs["aName"]).To(HaveCondition("Ready", false, "Invalid", "secret has empty name"))
		})

		It("should drop outputs that have secrets which don't exist", func() {
			forwarderSpec.Outputs = append(forwarderSpec.Outputs, loggingv1.OutputSpec{
				Name:   "aName",
				Type:   "elasticsearch",
				URL:    "https://somewhere",
				Secret: &loggingv1.OutputSecretSpec{Name: "mysecret"},
			})
			verifyOutputs(cluster, client, forwarderSpec, clfStatus, extras)
			Expect(clfStatus.Outputs["aName"]).To(HaveCondition("Ready", false, "MissingResource", "secret.*not found"))
		})

		Context("when validating secrets", func() {
			var secret *corev1.Secret
			BeforeEach(func() {
				secret = &corev1.Secret{
					TypeMeta: metav1.TypeMeta{
						Kind:       "Secret",
						APIVersion: corev1.SchemeGroupVersion.String(),
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "mytestsecret",
						Namespace: constants.OpenshiftNS,
					},
					Data: map[string][]byte{},
				}
			})

			Context("for writing to Cloudwatch", func() {
				const missingMessage = "auth keys: " + constants.AWSAccessKeyID + " and " + constants.AWSSecretAccessKey + " are required"
				BeforeEach(func() {
					output = loggingv1.OutputSpec{
						Name: "aName",
						Type: loggingv1.OutputTypeCloudwatch,
						OutputTypeSpec: loggingv1.OutputTypeSpec{
							Cloudwatch: &loggingv1.Cloudwatch{},
						},
						Secret: &loggingv1.OutputSecretSpec{Name: secret.Name},
					}
					forwarderSpec.Outputs = []loggingv1.OutputSpec{output}
				})

				It("should drop outputs with secrets that are missing aws_access_key_id and aws_secret_access_key and role_arn", func() {
					client = fake.NewFakeClient(secret) //nolint
					verifyOutputs(cluster, client, forwarderSpec, clfStatus, extras)
					Expect(clfStatus.Outputs["aName"]).To(HaveCondition("Ready", false, "MissingResource", missingMessage))
				})

				It("should drop outputs with secrets that is missing aws_secret_access_id", func() {
					secret.Data[constants.AWSSecretAccessKey] = []byte{0, 1, 2}
					client = fake.NewFakeClient(secret) //nolint
					verifyOutputs(cluster, client, forwarderSpec, clfStatus, extras)
					Expect(clfStatus.Outputs["aName"]).To(HaveCondition("Ready", false, "MissingResource", missingMessage))
				})

				It("should drop outputs with secrets that has empty aws_secret_access_key", func() {
					secret.Data[constants.AWSSecretAccessKey] = []byte{}
					secret.Data[constants.AWSAccessKeyID] = []byte{1, 2, 3}
					client = fake.NewFakeClient(secret) //nolint
					verifyOutputs(cluster, client, forwarderSpec, clfStatus, extras)
					Expect(clfStatus.Outputs["aName"]).To(HaveCondition("Ready", false, "MissingResource", missingMessage))
				})
				It("should drop outputs with secrets that is missing aws_secret_access_key", func() {
					secret.Data[constants.AWSAccessKeyID] = []byte{0, 1, 2}
					client = fake.NewFakeClient(secret) //nolint
					verifyOutputs(cluster, client, forwarderSpec, clfStatus, extras)
					Expect(clfStatus.Outputs["aName"]).To(HaveCondition("Ready", false, "MissingResource", missingMessage))
				})
				It("should drop outputs with secrets that have empty aws_access_key_id", func() {
					secret.Data[constants.AWSAccessKeyID] = []byte{}
					secret.Data[constants.AWSSecretAccessKey] = []byte{1, 2, 3}
					client = fake.NewFakeClient(secret) //nolint
					verifyOutputs(cluster, client, forwarderSpec, clfStatus, extras)
					Expect(clfStatus.Outputs["aName"]).To(HaveCondition("Ready", false, "MissingResource", missingMessage))
				})
				It("should accept outputs with secrets that have aws_secret_access_key and aws_access_key_id", func() {
					secret.Data[constants.AWSSecretAccessKey] = []byte{0, 1, 2}
					secret.Data[constants.AWSAccessKeyID] = []byte{0, 1, 2}
					client = fake.NewFakeClient(secret) //nolint
					verifyOutputs(cluster, client, forwarderSpec, clfStatus, extras)
					Expect(forwarderSpec.Outputs).To(HaveLen(len(forwarderSpec.Outputs)))
					Expect(clfStatus.Outputs["aName"]).To(HaveCondition("Ready", true, "", ""))
				})
				It("should accept outputs with secrets that have role_arn key with valid arn specified", func() {
					secret.Data[constants.AWSWebIdentityRoleKey] = []byte("arn:aws:iam::123456789012:role/my-role")
					client = fake.NewFakeClient(secret) //nolint
					verifyOutputs(cluster, client, forwarderSpec, clfStatus, extras)
					Expect(forwarderSpec.Outputs).To(HaveLen(len(forwarderSpec.Outputs)))
					Expect(clfStatus.Outputs["aName"]).To(HaveCondition("Ready", true, "", ""))
				})
				It("should drop outputs with role_arn key but without formatted arn specified", func() {
					secret.Data[constants.AWSWebIdentityRoleKey] = []byte("role/my-role")
					client = fake.NewFakeClient(secret) //nolint
					verifyOutputs(cluster, client, forwarderSpec, clfStatus, extras)
					stsMessage := "auth keys: a 'role_arn' or 'credentials' key is required containing a valid arn value"
					Expect(clfStatus.Outputs["aName"]).To(HaveCondition("Ready", false, "MissingResource", stsMessage))
				})
				It("should accept outputs with secrets that have credentials key with valid arn specified", func() {
					secret.Data[constants.AWSCredentialsKey] = []byte("role_arn = arn:aws:iam::123456789012:role/my-role")
					client = fake.NewFakeClient(secret) //nolint
					verifyOutputs(cluster, client, forwarderSpec, clfStatus, extras)
					Expect(forwarderSpec.Outputs).To(HaveLen(len(forwarderSpec.Outputs)))
					Expect(clfStatus.Outputs["aName"]).To(HaveCondition("Ready", true, "", ""))
				})
				It("should drop outputs with credential key but without formatted arn specified", func() {
					secret.Data[constants.AWSCredentialsKey] = []byte("role/my-role")
					client = fake.NewFakeClient(secret) //nolint
					verifyOutputs(cluster, client, forwarderSpec, clfStatus, extras)
					stsMessage := "auth keys: a 'role_arn' or 'credentials' key is required containing a valid arn value"
					Expect(clfStatus.Outputs["aName"]).To(HaveCondition("Ready", false, "MissingResource", stsMessage))
				})
			})

			Context("with certs", func() {
				BeforeEach(func() {
					output = loggingv1.OutputSpec{
						Name:   "aName",
						Type:   "elasticsearch",
						URL:    "https://somewhere",
						Secret: &loggingv1.OutputSecretSpec{Name: secret.Name},
					}
					forwarderSpec.Outputs = []loggingv1.OutputSpec{output}
				})
				It("should drop outputs with secrets that have missing tls.key", func() {
					secret.Data["tls.crt"] = []byte{0, 1, 2}
					client = fake.NewFakeClient(secret) //nolint
					verifyOutputs(cluster, client, forwarderSpec, clfStatus, extras)
					Expect(clfStatus.Outputs["aName"]).To(HaveCondition("Ready", false, "MissingResource", "cannot have.*without"))
				})
				It("should drop outputs with secrets that have empty tls.crt", func() {
					secret.Data["tls.crt"] = []byte{}
					secret.Data["tls.key"] = []byte{1, 2, 3}
					client = fake.NewFakeClient(secret) //nolint
					verifyOutputs(cluster, client, forwarderSpec, clfStatus, extras)
					Expect(clfStatus.Outputs["aName"]).To(HaveCondition("Ready", false, "MissingResource", "cannot have.*without"))
				})
				It("should drop outputs with secrets that have missing tls.crt", func() {
					secret.Data["tls.key"] = []byte{0, 1, 2}
					client = fake.NewFakeClient(secret) //nolint
					verifyOutputs(cluster, client, forwarderSpec, clfStatus, extras)
					Expect(clfStatus.Outputs["aName"]).To(HaveCondition("Ready", false, "MissingResource", "cannot have.*without"))
				})
				It("should drop outputs with secrets that have empty tls.key", func() {
					secret.Data["tls.key"] = []byte{}
					secret.Data["tls.crt"] = []byte{1, 2, 3}
					client = fake.NewFakeClient(secret) //nolint
					verifyOutputs(cluster, client, forwarderSpec, clfStatus, extras)
					Expect(clfStatus.Outputs["aName"]).To(HaveCondition("Ready", false, "MissingResource", "cannot have.*without"))
				})
				It("should accept outputs with secrets that have tls.key and tls.cert", func() {
					secret.Data["tls.key"] = []byte{0, 1, 2}
					secret.Data["tls.crt"] = []byte{0, 1, 2}
					client = fake.NewFakeClient(secret) //nolint
					verifyOutputs(cluster, client, forwarderSpec, clfStatus, extras)
					Expect(forwarderSpec.Outputs).To(HaveLen(len(forwarderSpec.Outputs)))
					Expect(clfStatus.Outputs["aName"]).To(HaveCondition("Ready", true, "", ""))
				})
			})
		})

		It("should accept well formed outputs", func() {
			client = fake.NewFakeClient( //nolint
				&corev1.Secret{
					TypeMeta: metav1.TypeMeta{
						Kind:       "Secret",
						APIVersion: corev1.SchemeGroupVersion.String(),
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "mysecret",
						Namespace: constants.OpenshiftNS,
					},
				},
				&corev1.Secret{
					TypeMeta: metav1.TypeMeta{
						Kind:       "Secret",
						APIVersion: corev1.SchemeGroupVersion.String(),
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "mycloudwatchsecret",
						Namespace: constants.OpenshiftNS,
					},
				},
			)
			forwarderSpec.Outputs = append(forwarderSpec.Outputs,
				loggingv1.OutputSpec{
					Name:   "aName",
					Type:   "elasticsearch",
					URL:    "https://somewhere",
					Secret: &loggingv1.OutputSecretSpec{Name: "mysecret"},
				},
			)
			verifyOutputs(cluster, client, forwarderSpec, clfStatus, extras)
			Expect(clfStatus.Outputs["aName"]).To(HaveCondition("Ready", true, "", ""), fmt.Sprintf("status: %+v", clfStatus))
			Expect(forwarderSpec.Outputs).To(HaveLen(len(forwarderSpec.Outputs)), fmt.Sprintf("status: %+v", clfStatus))
		})
	})

	Context("pipelines", func() {

		var (
			forwarderSpec loggingv1.ClusterLogForwarderSpec
			output        loggingv1.OutputSpec
			otherOutput   loggingv1.OutputSpec
			input         loggingv1.InputSpec
			otherInput    loggingv1.InputSpec
			condReady     status.Condition
			condNotReady  status.Condition
		)

		BeforeEach(func() {
			condReady = status.Condition{Type: loggingv1.ConditionReady, Status: corev1.ConditionTrue}
			condNotReady = status.Condition{Type: loggingv1.ConditionReady, Status: corev1.ConditionFalse}

			output = loggingv1.OutputSpec{
				Name: "myOutput",
				Type: "elasticsearch",
				URL:  "http://here",
			}
			otherOutput = loggingv1.OutputSpec{
				Name: otherTargetName,
				Type: "elasticsearch",
				URL:  "http://there",
			}

			input = loggingv1.InputSpec{
				Name: "app-input",
			}

			otherInput = loggingv1.InputSpec{
				Name:           "all-input",
				Application:    &loggingv1.Application{},
				Infrastructure: &loggingv1.Infrastructure{},
				Audit:          &loggingv1.Audit{},
			}

			forwarderSpec = loggingv1.ClusterLogForwarderSpec{
				Inputs: []loggingv1.InputSpec{
					input,
					otherInput,
				},
				Outputs: []loggingv1.OutputSpec{
					output,
					otherOutput,
				},
				Pipelines: []loggingv1.PipelineSpec{
					{
						Name:       "aPipeline",
						OutputRefs: []string{output.Name, otherOutput.Name},
						InputRefs:  []string{loggingv1.InputNameApplication},
					},
				},
			}

			clfStatus = &loggingv1.ClusterLogForwarderStatus{
				Inputs: loggingv1.NamedConditions{
					input.Name:      []status.Condition{condNotReady},
					otherInput.Name: []status.Condition{condReady},
				},
				Outputs: loggingv1.NamedConditions{
					output.Name:      []status.Condition{condReady},
					otherOutput.Name: []status.Condition{condReady},
				},
			}
		})

		It("should drop all pipelines if output refs are invalid.", func() {
			forwarderSpec.Pipelines = []loggingv1.PipelineSpec{
				{
					Name:       "aPipeline",
					OutputRefs: []string{"someotherendpoint"},
					InputRefs:  []string{loggingv1.InputNameApplication},
				},
			}
			verifyPipelines(&forwarderSpec, clfStatus)
			Expect(clfStatus.Pipelines["aPipeline"]).To(HaveCondition(loggingv1.ConditionReady, false, loggingv1.ReasonInvalid, "invalid:*"))
		})

		It("should drop all pipelines if even one pipeline does not have a unique name", func() {
			forwarderSpec.Pipelines = append(forwarderSpec.Pipelines,
				loggingv1.PipelineSpec{
					Name:       "aPipeline",
					OutputRefs: []string{output.Name, otherOutput.Name},
					InputRefs:  []string{loggingv1.InputNameApplication},
				})
			verifyPipelines(&forwarderSpec, clfStatus)
			Expect(clfStatus.Pipelines).To(HaveKey("pipeline_1_"))
			Expect(clfStatus.Pipelines["pipeline_1_"]).To(HaveCondition(loggingv1.ConditionReady, false, "Invalid", "duplicate"))
			Expect(clfStatus.Pipelines).To(HaveLen(2))
		})

		It("should not allow pipelines with empty/missing names", func() {
			forwarderSpec.Pipelines = append(forwarderSpec.Pipelines,
				loggingv1.PipelineSpec{
					OutputRefs: []string{otherOutput.Name},
					InputRefs:  []string{loggingv1.InputNameInfrastructure},
				})
			verifyPipelines(&forwarderSpec, clfStatus)
			Expect(clfStatus.Pipelines["pipeline_1_"]).To(HaveCondition(loggingv1.ConditionReady, false, "Invalid", "pipeline must have a name"))
		})

		It("should drop all pipelines if pipelines have unrecognized inputRefs", func() {
			forwarderSpec.Pipelines = []loggingv1.PipelineSpec{
				{
					Name:       "someDefinedPipeline",
					OutputRefs: []string{output.Name, otherOutput.Name},
					InputRefs:  []string{"foo"},
				},
			}
			verifyPipelines(&forwarderSpec, clfStatus)
			conds := clfStatus.Pipelines["someDefinedPipeline"]
			Expect(conds).To(HaveCondition(loggingv1.ConditionReady, false, loggingv1.ReasonInvalid, `inputs:.*\[foo]`))
		})

		It("should drop all pipelines if pipelines have no outputRefs", func() {
			forwarderSpec.Pipelines = append(forwarderSpec.Pipelines,
				loggingv1.PipelineSpec{
					Name:       "someDefinedPipeline",
					OutputRefs: []string{},
					InputRefs:  []string{loggingv1.InputNameApplication},
				})
			verifyPipelines(&forwarderSpec, clfStatus)
			conds := clfStatus.Pipelines["someDefinedPipeline"]
			Expect(conds).To(HaveCondition(loggingv1.ConditionReady, false, loggingv1.ReasonInvalid, "no valid outputs"))
		})

		// Degraded here means partially valid, which will not be supported
		It("should drop all pipelines if there are degraded pipelines with some bad outputRefs", func() {
			forwarderSpec.Pipelines = append(forwarderSpec.Pipelines,
				loggingv1.PipelineSpec{
					Name:       "someDefinedPipeline",
					OutputRefs: []string{output.Name, otherOutput.Name, "aMissingOutput"},
					InputRefs:  []string{loggingv1.InputNameApplication},
				})
			verifyPipelines(&forwarderSpec, clfStatus)
			Expect(clfStatus.Pipelines).To(HaveLen(2), "Exp. all defined pipelines in clfStatus object")
			Expect(clfStatus.Pipelines).To(HaveKey("someDefinedPipeline"))
			conds := clfStatus.Pipelines["someDefinedPipeline"]
			Expect(conds).To(HaveCondition(loggingv1.ConditionReady, false, loggingv1.ReasonInvalid, "invalid: unrecognized outputs*"))
		})

		It("should invalidate all pipelines if any input ref is not ready", func() {
			forwarderSpec.Pipelines = append(forwarderSpec.Pipelines,
				loggingv1.PipelineSpec{
					Name:       "someDefinedPipeline",
					OutputRefs: []string{output.Name, otherOutput.Name},
					InputRefs:  []string{input.Name, otherInput.Name},
				})
			verifyPipelines(&forwarderSpec, clfStatus)
			Expect(clfStatus.Pipelines).To(HaveLen(2), "Exp. all defined pipelines in clfStatus object")
			Expect(clfStatus.Pipelines).To(HaveKey("someDefinedPipeline"))
			conds := clfStatus.Pipelines["someDefinedPipeline"]
			Expect(conds).To(HaveCondition(loggingv1.ConditionReady, false, loggingv1.ReasonInvalid, "invalid:*"))
		})

	})

	Context("validating all", func() {
		var (
			client        client.Client
			cluster       *loggingv1.ClusterLogging
			extras        map[string]bool
			output        loggingv1.OutputSpec
			otherOutput   loggingv1.OutputSpec
			forwarderSpec *loggingv1.ClusterLogForwarderSpec
			clfInstance   *loggingv1.ClusterLogForwarder
		)

		BeforeEach(func() {
			client = fake.NewClientBuilder().WithRuntimeObjects(runtime.NewSecret(
				constants.OpenshiftNS, constants.CollectorSecretName, nil,
			)).Build()

			cluster = &loggingv1.ClusterLogging{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: constants.OpenshiftNS,
				},
			}

			clfInstance = testRunTime.NewClusterLogForwarder()

			extras = map[string]bool{}

			output = loggingv1.OutputSpec{
				Name: "myOutput",
				Type: "elasticsearch",
				URL:  "http://here",
			}
			otherOutput = loggingv1.OutputSpec{
				Name: otherTargetName,
				Type: "elasticsearch",
				URL:  "http://there",
			}

			forwarderSpec = &loggingv1.ClusterLogForwarderSpec{
				Outputs: []loggingv1.OutputSpec{
					output,
					otherOutput,
				},
			}

			forwarderSpec.Pipelines = []loggingv1.PipelineSpec{
				{
					Name:       "valid-pipeline",
					OutputRefs: []string{output.Name, otherOutput.Name},
					InputRefs:  []string{loggingv1.InputNameApplication},
				},
			}
		})

		It("invalid forwarder spec if outputref is invalid", func() {
			var clusterName = "cluster"

			invalidCW := loggingv1.OutputSpec{
				Name: "my-cloudwatch",
				Type: loggingv1.OutputTypeCloudwatch,
				OutputTypeSpec: loggingv1.OutputTypeSpec{
					Cloudwatch: &loggingv1.Cloudwatch{
						GroupPrefix: &clusterName,
					},
				},
				Secret: &loggingv1.OutputSecretSpec{Name: "inval-secret"},
			}

			forwarderSpec.Outputs = []loggingv1.OutputSpec{
				invalidCW,
				migrations.NewDefaultOutput(nil),
			}

			extras[constants.MigrateDefaultOutput] = true

			forwarderSpec.Pipelines = []loggingv1.PipelineSpec{
				{
					Name:       "custom-pipeline",
					OutputRefs: []string{invalidCW.Name, loggingv1.OutputNameDefault},
					InputRefs:  []string{loggingv1.InputNameApplication},
				},
			}
			Expect(forwarderSpec.Pipelines).To(HaveLen(1), "Exp 1 pipeline")
			Expect(forwarderSpec.Inputs).To(BeEmpty(), "Exp no inputs")
			Expect(forwarderSpec.Outputs).To(HaveLen(2), "Exp 1 output")

			clfStatus := ValidateInputsOutputsPipelines(cluster, client, clfInstance, *forwarderSpec, extras)

			Expect(forwarderSpec.Pipelines).To(HaveLen(1), "Exp. not to mutate original spec pipelines")
			Expect(forwarderSpec.Inputs).To(BeEmpty(), "Exp. not to mutate original spec inputs")
			Expect(forwarderSpec.Outputs).To(HaveLen(2), "Exp. not to mutate original spec outputs")

			Expect(clfStatus.Outputs["my-cloudwatch"]).To(HaveCondition("Ready", false, "MissingResource", "secret \"inval-secret\" not found"))
			Expect(clfStatus.Pipelines).To(HaveLen(1), "Exp. all defined pipelines to have statuses")
			Expect(clfStatus.Pipelines).To(HaveKey("custom-pipeline"))
			conds := clfStatus.Pipelines["custom-pipeline"]
			Expect(conds).To(HaveCondition(loggingv1.ConditionReady, false, loggingv1.ReasonInvalid, "invalid*"))
			Expect(clfStatus.Conditions).To(HaveCondition(loggingv1.ConditionReady, false, loggingv1.ReasonInvalid, "invalid clf spec; one or more errors present: *"))
		})

		It("invalid forwarder spec if inputref is invalid", func() {

			invalInput := loggingv1.InputSpec{
				Name: "inval-input",
			}

			forwarderSpec.Outputs = []loggingv1.OutputSpec{
				migrations.NewDefaultOutput(nil),
			}

			forwarderSpec.Inputs = []loggingv1.InputSpec{
				invalInput,
			}

			extras[constants.MigrateDefaultOutput] = true

			forwarderSpec.Pipelines = []loggingv1.PipelineSpec{
				{
					Name:       "custom-pipeline",
					OutputRefs: []string{loggingv1.OutputNameDefault},
					InputRefs:  []string{invalInput.Name, loggingv1.InputNameApplication},
				},
			}
			Expect(forwarderSpec.Pipelines).To(HaveLen(1), "Exp 1 pipeline")
			Expect(forwarderSpec.Inputs).To(HaveLen(1), "Exp 1 input")
			Expect(forwarderSpec.Outputs).To(HaveLen(1), "Exp 1 output")

			clfStatus := ValidateInputsOutputsPipelines(cluster, client, clfInstance, *forwarderSpec, extras)

			Expect(forwarderSpec.Pipelines).To(HaveLen(1), "Exp. not to mutate original spec pipelines")
			Expect(forwarderSpec.Inputs).To(HaveLen(1), "Exp. not to mutate original spec inputs")
			Expect(forwarderSpec.Outputs).To(HaveLen(1), "Exp. not to mutate original spec outputs")

			Expect(clfStatus.Inputs["inval-input"]).To(HaveCondition("Ready", false, loggingv1.ReasonInvalid, "inputspec must define one or more of application, infrastructure, or audit"))
			Expect(clfStatus.Pipelines).To(HaveLen(1), "Exp. all defined pipelines to have statuses")
			Expect(clfStatus.Pipelines).To(HaveKey("custom-pipeline"))
			conds := clfStatus.Pipelines["custom-pipeline"]
			Expect(conds).To(HaveCondition(loggingv1.ConditionReady, false, loggingv1.ReasonInvalid, "invalid*"))
			Expect(clfStatus.Conditions).To(HaveCondition(loggingv1.ConditionReady, false, loggingv1.ReasonInvalid, "invalid clf spec; one or more errors present: *"))
		})

		It("invalid forwarder spec if pipeline has unknown inputRef", func() {

			appInput := loggingv1.InputSpec{
				Name:        "app-logs",
				Application: &loggingv1.Application{},
			}

			forwarderSpec.Outputs = []loggingv1.OutputSpec{
				migrations.NewDefaultOutput(nil),
			}

			forwarderSpec.Inputs = []loggingv1.InputSpec{
				appInput,
			}

			extras[constants.MigrateDefaultOutput] = true

			forwarderSpec.Pipelines = []loggingv1.PipelineSpec{
				{
					Name:       "custom-pipeline",
					OutputRefs: []string{loggingv1.OutputNameDefault},
					InputRefs:  []string{appInput.Name, "missingInRef"},
				},
			}
			Expect(forwarderSpec.Pipelines).To(HaveLen(1), "Exp 1 pipeline")
			Expect(forwarderSpec.Inputs).To(HaveLen(1), "Exp 1 input")
			Expect(forwarderSpec.Outputs).To(HaveLen(1), "Exp 1 output")

			clfStatus := ValidateInputsOutputsPipelines(cluster, client, clfInstance, *forwarderSpec, extras)

			Expect(forwarderSpec.Pipelines).To(HaveLen(1), "Exp. not to mutate original spec pipelines")
			Expect(forwarderSpec.Inputs).To(HaveLen(1), "Exp. not to mutate original spec inputs")
			Expect(forwarderSpec.Outputs).To(HaveLen(1), "Exp. not to mutate original spec outputs")

			Expect(clfStatus.Inputs["app-logs"]).To(HaveCondition("Ready", true, "", ""))
			Expect(clfStatus.Pipelines).To(HaveLen(1), "Exp. all defined pipelines to have statuses")
			Expect(clfStatus.Pipelines).To(HaveKey("custom-pipeline"))
			conds := clfStatus.Pipelines["custom-pipeline"]
			Expect(conds).To(HaveCondition(loggingv1.ConditionReady, false, loggingv1.ReasonInvalid, "invalid*"))
			Expect(clfStatus.Conditions).To(HaveCondition(loggingv1.ConditionReady, false, loggingv1.ReasonInvalid, "invalid clf spec; one or more errors present: *"))
		})

		It("invalid forwarder spec if even one pipeline is invalid", func() {

			appInput := loggingv1.InputSpec{
				Name:        "app-logs",
				Application: &loggingv1.Application{},
			}

			forwarderSpec.Outputs = []loggingv1.OutputSpec{
				migrations.NewDefaultOutput(nil),
			}

			forwarderSpec.Inputs = []loggingv1.InputSpec{
				appInput,
			}

			extras[constants.MigrateDefaultOutput] = true

			forwarderSpec.Pipelines = append(forwarderSpec.Pipelines, loggingv1.PipelineSpec{
				Name:       "inval-pipeline",
				OutputRefs: []string{loggingv1.OutputNameDefault},
				InputRefs:  []string{appInput.Name, "missingInRef"},
			})
			Expect(forwarderSpec.Pipelines).To(HaveLen(2), "Exp 2 pipelines")
			Expect(forwarderSpec.Inputs).To(HaveLen(1), "Exp 1 input")
			Expect(forwarderSpec.Outputs).To(HaveLen(1), "Exp 1 output")

			clfStatus := ValidateInputsOutputsPipelines(cluster, client, clfInstance, *forwarderSpec, extras)

			Expect(forwarderSpec.Pipelines).To(HaveLen(2), "Exp. not to mutate original spec pipelines")
			Expect(forwarderSpec.Inputs).To(HaveLen(1), "Exp. not to mutate original spec inputs")
			Expect(forwarderSpec.Outputs).To(HaveLen(1), "Exp. not to mutate original spec outputs")

			Expect(clfStatus.Inputs["app-logs"]).To(HaveCondition("Ready", true, "", ""))
			Expect(clfStatus.Pipelines).To(HaveLen(2), "Exp. all defined pipelines to have statuses")
			Expect(clfStatus.Pipelines).To(HaveKey("inval-pipeline"))
			Expect(clfStatus.Pipelines).To(HaveKey("valid-pipeline"))
			conds := clfStatus.Pipelines["inval-pipeline"]
			Expect(conds).To(HaveCondition(loggingv1.ConditionReady, false, loggingv1.ReasonInvalid, "invalid*"))
			conds2 := clfStatus.Pipelines["valid-pipeline"]
			Expect(conds2).To(HaveCondition(loggingv1.ConditionReady, false, loggingv1.ReasonInvalid, "invalid*"))
			Expect(clfStatus.Conditions).To(HaveCondition(loggingv1.ConditionReady, false, loggingv1.ReasonInvalid, "invalid clf spec; one or more errors present: *"))
		})

		It("should have no status if spec has empty pipelines and no forwarder instance", func() {
			forwarderSpec = &loggingv1.ClusterLogForwarderSpec{
				Inputs:    []loggingv1.InputSpec{},
				Outputs:   []loggingv1.OutputSpec{},
				Pipelines: []loggingv1.PipelineSpec{},
			}
			Expect(YAMLString(forwarderSpec)).To(EqualLines("{}"))
			clfStatus := ValidateInputsOutputsPipelines(cluster, client, nil, *forwarderSpec, extras)
			Expect(YAMLString(clfStatus)).To(EqualLines("{}"))
		})
	})
})

func Test_verifyOutputURL(t *testing.T) {
	type fields struct {
		Client           client.Client
		Cluster          *loggingv1.ClusterLogging
		ForwarderRequest *loggingv1.ClusterLogForwarder
		ForwarderSpec    loggingv1.ClusterLogForwarderSpec
	}
	type args struct {
		output *loggingv1.OutputSpec
		conds  loggingv1.NamedConditions
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "With fluentdForward without URL",
			args: args{
				output: &loggingv1.OutputSpec{
					Name: "test-output",
					Type: "fluentdForward",
					URL:  "",
				},
				conds: loggingv1.NamedConditions{},
			},
			want: false,
		},
		{
			name: "With fluentdForward with URL",
			args: args{
				output: &loggingv1.OutputSpec{
					Name: "test-output",
					Type: "fluentdForward",
					URL:  "http://123.local:9200",
				},
				conds: loggingv1.NamedConditions{},
			},
			want: true,
		},
		{
			name: "With elastic without URL",
			args: args{
				output: &loggingv1.OutputSpec{
					Name: "test-output",
					Type: "elasticsearch",
					URL:  "",
				},
				conds: loggingv1.NamedConditions{},
			},
			want: false,
		},
		{
			name: "With elastic with URL",
			args: args{
				output: &loggingv1.OutputSpec{
					Name: "test-output",
					Type: "elasticsearch",
					URL:  "http://123.local:9200",
				},
				conds: loggingv1.NamedConditions{},
			},
			want: true,
		},
		{
			name: "With kafka without url",
			args: args{
				output: &loggingv1.OutputSpec{
					Name: "test-output",
					Type: "kafka",
					URL:  "",
				},
				conds: loggingv1.NamedConditions{},
			},
			want: true,
		},
		{
			name: "With kafka",
			args: args{
				output: &loggingv1.OutputSpec{
					Name: "test-output",
					Type: "kafka",
					URL:  "https://local.svc",
				},
				conds: loggingv1.NamedConditions{},
			},
			want: true,
		},
		{
			name: "With syslog",
			args: args{
				output: &loggingv1.OutputSpec{
					Name: "test-output",
					Type: "syslog",
					URL:  "https://local.svc",
				},
				conds: loggingv1.NamedConditions{},
			},
			want: true,
		},
		{
			name: "With syslog without url",
			args: args{
				output: &loggingv1.OutputSpec{
					Name: "test-output",
					Type: "syslog",
					URL:  "",
				},
				conds: loggingv1.NamedConditions{},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		tt := tt // Don't bind range variable.
		t.Run(tt.name, func(t *testing.T) {
			if got := verifyOutputURL(tt.args.output, tt.args.conds); got != tt.want {
				t.Errorf("verifyOutputURL() = %v, want %v", got, tt.want)
			}
		})
	}
}
