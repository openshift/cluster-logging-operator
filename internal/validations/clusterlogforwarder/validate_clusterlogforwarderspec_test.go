package clusterlogforwarder

import (
	"fmt"
	"github.com/openshift/cluster-logging-operator/test/helpers/rand"

	"github.com/openshift/cluster-logging-operator/internal/migrations/clusterlogforwarder"
	v1 "k8s.io/apiserver/pkg/apis/audit/v1"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	loggingv1 "github.com/openshift/cluster-logging-operator/api/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
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

var _ = Describe("Validate clusterlogforwarderspec", func() {
	var clfStatus *loggingv1.ClusterLogForwarderStatus

	BeforeEach(func() {
		clfStatus = &loggingv1.ClusterLogForwarderStatus{}
	})

	Context("output specs", func() {
		const secretName = "mytestsecret"
		var (
			client           client.Client
			namespace        = constants.OpenshiftNS
			extras           map[string]bool
			clfStatus        *loggingv1.ClusterLogForwarderStatus
			output           loggingv1.OutputSpec
			otherOutput      loggingv1.OutputSpec
			forwarderSpec    *loggingv1.ClusterLogForwarderSpec
			cloudWatchSecret *corev1.Secret
		)

		BeforeEach(func() {
			client = fake.NewClientBuilder().WithRuntimeObjects(runtime.NewSecret(
				constants.OpenshiftNS, constants.CollectorSecretName, nil,
			)).Build()
			clfStatus = &loggingv1.ClusterLogForwarderStatus{}
			extras = map[string]bool{}

			cloudWatchSecret = runtime.NewSecret(constants.OpenshiftNS, secretName, map[string][]byte{
				constants.AWSSecretAccessKey: {0, 1, 2},
				constants.AWSAccessKeyID:     {0, 1, 2},
			})

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
				Pipelines: []loggingv1.PipelineSpec{
					{OutputRefs: []string{output.Name, otherOutput.Name}},
				},
				Outputs: []loggingv1.OutputSpec{
					output,
					otherOutput,
				},
			}
		})

		DescribeTable("googlecloudlogging output validation",
			func(gcl *loggingv1.GoogleCloudLogging, expectedPass bool) {
				forwarderSpec = &loggingv1.ClusterLogForwarderSpec{
					Pipelines: []loggingv1.PipelineSpec{{OutputRefs: []string{"X"}}},
					Outputs: []loggingv1.OutputSpec{
						{
							Name: "X",
							Type: "googleCloudLogging",
							OutputTypeSpec: loggingv1.OutputTypeSpec{
								GoogleCloudLogging: gcl,
							},
						},
					},
				}
				verifyOutputs(namespace, client, forwarderSpec, clfStatus, extras)
				if expectedPass {
					Expect(clfStatus.Outputs["X"]).To(HaveCondition("Ready", true, "", ""))
				} else {
					Expect(clfStatus.Outputs["X"]).To(HaveCondition("Ready", false, loggingv1.ReasonInvalid,
						"output \"X\": Exactly one of billingAccountId, folderId, organizationId, or projectId must be set."))
				}
			},
			// number of subsets: 2^4 = 16
			// 4C0
			Entry("empty", &loggingv1.GoogleCloudLogging{}, false),
			// 4C1
			Entry("billingAccountID", &loggingv1.GoogleCloudLogging{BillingAccountID: "billingAccountID"}, true),
			Entry("organizationID", &loggingv1.GoogleCloudLogging{OrganizationID: "organizationID"}, true),
			Entry("folderID", &loggingv1.GoogleCloudLogging{FolderID: "folderID"}, true),
			Entry("projectID", &loggingv1.GoogleCloudLogging{ProjectID: "projectID"}, true),
			// 4C2
			Entry("billingAccountID, organizationID", &loggingv1.GoogleCloudLogging{BillingAccountID: "billingAccountID", OrganizationID: "organizationID"}, false),
			Entry("billingAccountID, folderID", &loggingv1.GoogleCloudLogging{BillingAccountID: "billingAccountID", FolderID: "folderID"}, false),
			Entry("billingAccountID, projectID", &loggingv1.GoogleCloudLogging{BillingAccountID: "billingAccountID", ProjectID: "projectID"}, false),
			Entry("organizationID, folderID", &loggingv1.GoogleCloudLogging{OrganizationID: "organizationID", FolderID: "folderID"}, false),
			Entry("organizationID, projectID", &loggingv1.GoogleCloudLogging{OrganizationID: "organizationID", ProjectID: "projectID"}, false),
			Entry("projectID, folderID", &loggingv1.GoogleCloudLogging{ProjectID: "projectID", FolderID: "folderID"}, false),
			// 4C3
			Entry("billingAccountID, organizationID, projectID", &loggingv1.GoogleCloudLogging{BillingAccountID: "billingAccountID", OrganizationID: "organizationID", ProjectID: "projectID"}, false),
			Entry("billingAccountID, organizationID, folderID", &loggingv1.GoogleCloudLogging{BillingAccountID: "billingAccountID", OrganizationID: "organizationID", FolderID: "folderID"}, false),
			Entry("organizationID, projectID, folderID", &loggingv1.GoogleCloudLogging{OrganizationID: "organizationID", ProjectID: "projectID", FolderID: "folderID"}, false),
			Entry("billingAccountID, ProjectID, folderID", &loggingv1.GoogleCloudLogging{BillingAccountID: "billingAccountID", ProjectID: "projectID", FolderID: "folderID"}, false),
			// 4C4
			Entry("all", &loggingv1.GoogleCloudLogging{OrganizationID: "organizationID", BillingAccountID: "billingAccountID", ProjectID: "projectID", FolderID: "folderID"}, false),
		)

		// Ref: https://issues.redhat.com/browse/LOG-3228
		It("should validate the default output as any other without adding a new one", func() {
			forwarderSpec := &loggingv1.ClusterLogForwarderSpec{
				Pipelines: []loggingv1.PipelineSpec{{OutputRefs: []string{"default"}}},
				Outputs: []loggingv1.OutputSpec{
					clusterlogforwarder.NewDefaultOutput(nil, constants.CollectorName),
				},
			}

			extras[constants.MigrateDefaultOutput] = true

			// sanity check
			Expect(forwarderSpec.Outputs).To(HaveLen(1))
			verifyOutputs(namespace, client, forwarderSpec, clfStatus, extras)
			Expect(clfStatus.Outputs[loggingv1.OutputNameDefault]).To(HaveCondition(loggingv1.ConditionReady, true, "", ""))
			Expect(forwarderSpec.Outputs).To(HaveLen(1), "Exp. the number of outputs to remain unchanged")
		})

		It("should be invalid if output is not referenced in a pipeline", func() {
			forwarderSpec.Pipelines = []loggingv1.PipelineSpec{{OutputRefs: []string{otherOutput.Name}}}
			verifyOutputs(namespace, client, forwarderSpec, clfStatus, extras)
			Expect(clfStatus.Outputs[output.Name]).To(HaveCondition(loggingv1.ConditionReady, false, loggingv1.ReasonInvalid, "not referenced"))
		})

		It("should be invalid if outputs do not have unique names", func() {
			forwarderSpec.Outputs = append(forwarderSpec.Outputs, loggingv1.OutputSpec{
				Name: "myOutput",
				Type: "elasticsearch",
				URL:  "http://here",
			})
			// sanity check
			Expect(forwarderSpec.Outputs).To(HaveLen(3))
			verifyOutputs(namespace, client, forwarderSpec, clfStatus, extras)
			Expect(clfStatus.Outputs["myOutput"]).To(HaveCondition(loggingv1.ConditionReady, true, "", ""))
			Expect(clfStatus.Outputs["output_2_"]).To(HaveCondition(loggingv1.ConditionReady, false, loggingv1.ReasonInvalid, "duplicate"))
		})

		It("should be invalid if any outputs have empty names", func() {
			forwarderSpec.Outputs = append(forwarderSpec.Outputs, loggingv1.OutputSpec{
				Name: "",
				Type: "elasticsearch",
				URL:  "http://here",
			})
			verifyOutputs(namespace, client, forwarderSpec, clfStatus, extras)
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
			verifyOutputs(namespace, client, forwarderSpec, clfStatus, extras)
			Expect(len(forwarderSpec.Outputs)).To(Equal(len(outputs)), "Exp. outputs with an internal name of 'default' do be kept")
		})

		It("should be invalid if outputs have empty types", func() {
			forwarderSpec.Outputs = append(forwarderSpec.Outputs, loggingv1.OutputSpec{
				Name: "aName",
				URL:  "http://here",
			})
			verifyOutputs(namespace, client, forwarderSpec, clfStatus, extras)
			Expect(clfStatus.Outputs["aName"]).To(HaveCondition("Ready", false, "Invalid", "unknown.*\"\""))
		})

		It("should fail outputs that have unrecognized types", func() {
			forwarderSpec.Outputs = append(forwarderSpec.Outputs, loggingv1.OutputSpec{
				Name: "aName",
				Type: "foo",
				URL:  "http://here",
			})
			verifyOutputs(namespace, client, forwarderSpec, clfStatus, extras)
			Expect(clfStatus.Outputs["aName"]).To(HaveCondition("Ready", false, "Invalid", "unknown.*\"foo\""))
		})

		It("should fail Cloudwatch output without OutputTypeSpec", func() {
			client = fake.NewFakeClient(cloudWatchSecret) //nolint
			forwarderSpec.Outputs = []loggingv1.OutputSpec{
				{
					Name: "cw",
					Type: loggingv1.OutputTypeCloudwatch,
					Secret: &loggingv1.OutputSecretSpec{
						Name: secretName,
					},
				},
			}
			verifyOutputs(namespace, client, forwarderSpec, clfStatus, extras)
			Expect(clfStatus.Outputs["cw"]).To(HaveCondition("Ready", false, "Invalid", "Cloudwatch output requires type spec"))
		})

		It("should allow specific outputs that do not require URL", func() {
			client = fake.NewFakeClient(cloudWatchSecret) //nolint
			forwarderSpec.Pipelines = []loggingv1.PipelineSpec{{OutputRefs: []string{"aCloudwatch"}}}
			forwarderSpec.Outputs = []loggingv1.OutputSpec{
				{
					Name: "aCloudwatch",
					Type: loggingv1.OutputTypeCloudwatch,
					OutputTypeSpec: loggingv1.OutputTypeSpec{
						Cloudwatch: &loggingv1.Cloudwatch{},
					},
					Secret: &loggingv1.OutputSecretSpec{
						Name: secretName,
					},
				},
			}
			verifyOutputs(namespace, client, forwarderSpec, clfStatus, extras)
			Expect(clfStatus.Outputs["aCloudwatch"]).To(HaveCondition("Ready", true, "", ""))
		})

		It("should fail outputs that have secrets with no names", func() {
			forwarderSpec.Outputs = append(forwarderSpec.Outputs, loggingv1.OutputSpec{
				Name:   "aName",
				Type:   "elasticsearch",
				URL:    "https://somewhere",
				Secret: &loggingv1.OutputSecretSpec{},
			})
			verifyOutputs(namespace, client, forwarderSpec, clfStatus, extras)
			Expect(clfStatus.Outputs["aName"]).To(HaveCondition("Ready", false, "Invalid", "secret has empty name"))
		})

		It("should fail outputs that have secrets which don't exist", func() {
			forwarderSpec.Outputs = append(forwarderSpec.Outputs, loggingv1.OutputSpec{
				Name:   "aName",
				Type:   "elasticsearch",
				URL:    "https://somewhere",
				Secret: &loggingv1.OutputSecretSpec{Name: "mysecret"},
			})
			verifyOutputs(namespace, client, forwarderSpec, clfStatus, extras)
			Expect(clfStatus.Outputs["aName"]).To(HaveCondition("Ready", false, "MissingResource", "secret.*not found"))
		})

		It("should be valid if output has a positive limit threshold", func() {
			forwarderSpec.Pipelines = []loggingv1.PipelineSpec{{OutputRefs: []string{"custom-output"}}}
			forwarderSpec.Outputs = append(forwarderSpec.Outputs, loggingv1.OutputSpec{
				Name: "custom-output",
				Type: "elasticsearch",
				URL:  "https://somewhere",
				Limit: &loggingv1.LimitSpec{
					MaxRecordsPerSecond: 100,
				},
			})
			verifyOutputs(namespace, client, forwarderSpec, clfStatus, extras)
			Expect(clfStatus.Outputs["custom-output"]).To(HaveCondition("Ready", true, "", ""))

		})

		//TODO: FIXME
		//It("should fail if output has a negative limit threshold", func() {
		//	forwarderSpec.Outputs = append(forwarderSpec.Outputs, loggingv1.OutputSpec{
		//		Name: "custom-output",
		//		Type: "elasticsearch",
		//		URL:  "https://somewhere",
		//		Limit: &loggingv1.LimitSpec{
		//			MaxRecordsPerSecond: -100,
		//		},
		//	})
		//	verifyOutputs(namespace, client, forwarderSpec, clfStatus, extras)
		//	Expect(clfStatus.Outputs["custom-output"]).To(HaveCondition("Ready", false, loggingv1.ReasonInvalid, "output \"custom-output\": Output cannot have negative limit threshold"))
		//})

		It("should pass Azure Monitor Logs with valid conf", func() {
			forwarderSpec.Outputs = []loggingv1.OutputSpec{
				{
					Name: "azml",
					Type: loggingv1.OutputTypeAzureMonitor,
					OutputTypeSpec: loggingv1.OutputTypeSpec{
						AzureMonitor: &loggingv1.AzureMonitor{
							CustomerId: "customer",
							LogType:    "application",
						},
					},
				},
			}
			verifyOutputs(namespace, client, forwarderSpec, clfStatus, extras)
			Expect(clfStatus.Outputs["azml"]).To(HaveCondition("Ready", true, "", ""))
		})

		It("should fail Azure Monitor Logs with not valid LogType", func() {
			forwarderSpec.Outputs = []loggingv1.OutputSpec{
				{
					Name: "azml",
					Type: loggingv1.OutputTypeAzureMonitor,
					OutputTypeSpec: loggingv1.OutputTypeSpec{
						AzureMonitor: &loggingv1.AzureMonitor{
							CustomerId: "customer",
							LogType:    "my.application",
						},
					},
				},
			}
			verifyOutputs(namespace, client, forwarderSpec, clfStatus, extras)
			Expect(clfStatus.Outputs["azml"]).To(HaveCondition("Ready", false, "Invalid", "output \"azml\": LogType names must start with a letter/number, contain only letters/numbers/underscores \\(_\\), and be between 1-100 characters."))
		})

		It("should fail Azure Monitor Logs output without OutputTypeSpec", func() {
			forwarderSpec.Outputs = []loggingv1.OutputSpec{
				{
					Name: "azml",
					Type: loggingv1.OutputTypeAzureMonitor,
				},
			}
			verifyOutputs(namespace, client, forwarderSpec, clfStatus, extras)
			Expect(clfStatus.Outputs["azml"]).To(HaveCondition("Ready", false, "Invalid", "Azure Monitor Logs output requires type spec"))
		})

		It("should fail Azure Monitor Logs output without LogType", func() {
			forwarderSpec.Outputs = []loggingv1.OutputSpec{
				{
					Name: "azml",
					Type: loggingv1.OutputTypeAzureMonitor,
					OutputTypeSpec: loggingv1.OutputTypeSpec{
						AzureMonitor: &loggingv1.AzureMonitor{
							CustomerId: string(rand.Word(10)),
						},
					},
				},
			}
			verifyOutputs(namespace, client, forwarderSpec, clfStatus, extras)
			Expect(clfStatus.Outputs["azml"]).To(HaveCondition("Ready", false, "Invalid", "LogType must be set"))
		})

		It("should fail Azure Monitor Logs output with LogType name more than 100 characters", func() {
			forwarderSpec.Outputs = []loggingv1.OutputSpec{
				{
					Name: "azml",
					Type: loggingv1.OutputTypeAzureMonitor,
					OutputTypeSpec: loggingv1.OutputTypeSpec{
						AzureMonitor: &loggingv1.AzureMonitor{
							CustomerId: string(rand.Word(10)),
							LogType:    string(rand.Word(101)),
						},
					},
				},
			}
			verifyOutputs(namespace, client, forwarderSpec, clfStatus, extras)
			Expect(clfStatus.Outputs["azml"]).To(HaveCondition("Ready", false, "Invalid", "output \"azml\": LogType names must start with a letter/number, contain only letters/numbers/underscores \\(_\\), and be between 1-100 characters."))
		})

		It("should fail Azure Monitor Logs output without CustomerId", func() {
			forwarderSpec.Outputs = []loggingv1.OutputSpec{
				{
					Name: "azml",
					Type: loggingv1.OutputTypeAzureMonitor,
					OutputTypeSpec: loggingv1.OutputTypeSpec{
						AzureMonitor: &loggingv1.AzureMonitor{
							LogType: string(rand.Word(10)),
						},
					},
				},
			}
			verifyOutputs(namespace, client, forwarderSpec, clfStatus, extras)
			Expect(clfStatus.Outputs["azml"]).To(HaveCondition("Ready", false, "Invalid", "CustomerId must be set"))
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
						Name:      secretName,
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
					forwarderSpec.Pipelines = []loggingv1.PipelineSpec{{OutputRefs: []string{output.Name}}}
					forwarderSpec.Outputs = []loggingv1.OutputSpec{output}
				})

				It("should fail outputs with secrets that are missing aws_access_key_id and aws_secret_access_key and role_arn", func() {
					client = fake.NewFakeClient(secret) //nolint
					verifyOutputs(namespace, client, forwarderSpec, clfStatus, extras)
					Expect(clfStatus.Outputs["aName"]).To(HaveCondition("Ready", false, "MissingResource", missingMessage))
				})

				It("should fail outputs without secrets", func() {
					client = fake.NewFakeClient(secret) //nolint
					output = loggingv1.OutputSpec{
						Name: "aName",
						Type: loggingv1.OutputTypeCloudwatch,
						OutputTypeSpec: loggingv1.OutputTypeSpec{
							Cloudwatch: &loggingv1.Cloudwatch{},
						},
					}
					forwarderSpec = &loggingv1.ClusterLogForwarderSpec{}
					forwarderSpec.Pipelines = []loggingv1.PipelineSpec{{OutputRefs: []string{output.Name}}}
					forwarderSpec.Outputs = []loggingv1.OutputSpec{output}
					verifyOutputs(namespace, client, forwarderSpec, clfStatus, extras)
					Expect(clfStatus.Outputs["aName"]).To(HaveCondition("Ready", false, "MissingResource", "secret must be provided for cloudwatch output"))
				})

				It("should fail outputs without secrets that is missing aws_secret_access_id", func() {
					secret.Data[constants.AWSSecretAccessKey] = []byte{0, 1, 2}
					client = fake.NewFakeClient(secret) //nolint
					verifyOutputs(namespace, client, forwarderSpec, clfStatus, extras)
					Expect(clfStatus.Outputs["aName"]).To(HaveCondition("Ready", false, "MissingResource", missingMessage))
				})

				It("should fail outputs with secrets that has empty aws_secret_access_key", func() {
					secret.Data[constants.AWSSecretAccessKey] = []byte{}
					secret.Data[constants.AWSAccessKeyID] = []byte{1, 2, 3}
					client = fake.NewFakeClient(secret) //nolint
					verifyOutputs(namespace, client, forwarderSpec, clfStatus, extras)
					Expect(clfStatus.Outputs["aName"]).To(HaveCondition("Ready", false, "MissingResource", missingMessage))
				})
				It("should fail outputs with secrets that is missing aws_secret_access_key", func() {
					secret.Data[constants.AWSAccessKeyID] = []byte{0, 1, 2}
					client = fake.NewFakeClient(secret) //nolint
					verifyOutputs(namespace, client, forwarderSpec, clfStatus, extras)
					Expect(clfStatus.Outputs["aName"]).To(HaveCondition("Ready", false, "MissingResource", missingMessage))
				})
				It("should fail outputs with secrets that have empty aws_access_key_id", func() {
					secret.Data[constants.AWSAccessKeyID] = []byte{}
					secret.Data[constants.AWSSecretAccessKey] = []byte{1, 2, 3}
					client = fake.NewFakeClient(secret) //nolint
					verifyOutputs(namespace, client, forwarderSpec, clfStatus, extras)
					Expect(clfStatus.Outputs["aName"]).To(HaveCondition("Ready", false, "MissingResource", missingMessage))
				})
				It("should pass outputs with secrets that have aws_secret_access_key and aws_access_key_id", func() {
					secret.Data[constants.AWSSecretAccessKey] = []byte{0, 1, 2}
					secret.Data[constants.AWSAccessKeyID] = []byte{0, 1, 2}
					client = fake.NewFakeClient(secret) //nolint
					verifyOutputs(namespace, client, forwarderSpec, clfStatus, extras)
					Expect(forwarderSpec.Outputs).To(HaveLen(len(forwarderSpec.Outputs)))
					Expect(clfStatus.Outputs["aName"]).To(HaveCondition("Ready", true, "", ""))
				})
				// TODO: FIXME
				//It("should pass outputs with secrets that have role_arn key with valid arn specified", func() {
				//	secret.Data[constants.AWSWebIdentityRoleKey] = []byte("arn:aws:iam::123456789012:role/my-role")
				//	client = fake.NewFakeClient(secret) //nolint
				//	verifyOutputs(namespace, client, forwarderSpec, clfStatus, extras)
				//	Expect(forwarderSpec.Outputs).To(HaveLen(len(forwarderSpec.Outputs)))
				//	Expect(clfStatus.Outputs["aName"]).To(HaveCondition("Ready", true, "", ""))
				//})
				It("should fail outputs with role_arn key but without formatted arn specified", func() {
					secret.Data[constants.AWSWebIdentityRoleKey] = []byte("role/my-role")
					client = fake.NewFakeClient(secret) //nolint
					verifyOutputs(namespace, client, forwarderSpec, clfStatus, extras)
					stsMessage := "auth keys: a 'role_arn' or 'credentials' key is required containing a valid arn value"
					Expect(clfStatus.Outputs["aName"]).To(HaveCondition("Ready", false, "MissingResource", stsMessage))
				})
				// TODO: FIXME
				//It("should pass outputs with secrets that have credentials key with valid arn specified", func() {
				//	secret.Data[constants.AWSCredentialsKey] = []byte("role_arn = arn:aws:iam::123456789012:role/my-role")
				//	client = fake.NewFakeClient(secret) //nolint
				//	verifyOutputs(namespace, client, forwarderSpec, clfStatus, extras)
				//	Expect(forwarderSpec.Outputs).To(HaveLen(len(forwarderSpec.Outputs)))
				//	Expect(clfStatus.Outputs["aName"]).To(HaveCondition("Ready", true, "", ""))
				//})
				It("should fail outputs with credential key but without formatted arn specified", func() {
					secret.Data[constants.AWSCredentialsKey] = []byte("role/my-role")
					client = fake.NewFakeClient(secret) //nolint
					verifyOutputs(namespace, client, forwarderSpec, clfStatus, extras)
					stsMessage := "auth keys: a 'role_arn' or 'credentials' key is required containing a valid arn value"
					Expect(clfStatus.Outputs["aName"]).To(HaveCondition("Ready", false, "MissingResource", stsMessage))
				})
			})

			Context("for writing to Splunk", func() {
				BeforeEach(func() {
					output = loggingv1.OutputSpec{
						Name: "aName",
						Type: loggingv1.OutputTypeSplunk,
						URL:  "https://splunk-web:8088/endpoint",
						OutputTypeSpec: loggingv1.OutputTypeSpec{
							Splunk: &loggingv1.Splunk{},
						},
						Secret: &loggingv1.OutputSecretSpec{Name: secret.Name},
					}
					forwarderSpec.Pipelines = []loggingv1.PipelineSpec{{OutputRefs: []string{output.Name}}}
					forwarderSpec.Outputs = []loggingv1.OutputSpec{output}
				})

				It("should fail outputs with secrets that is missing hecToken", func() {
					client = fake.NewFakeClient(secret) //nolint
					verifyOutputs(namespace, client, forwarderSpec, clfStatus, extras)
					Expect(clfStatus.Outputs["aName"]).To(HaveCondition("Ready", false, "MissingResource", "A non-empty hecToken entry is required"))
				})

				It("should fail outputs without secrets", func() {
					client = fake.NewFakeClient(secret) //nolint
					output = loggingv1.OutputSpec{
						Name: "aName",
						URL:  "https://splunk-web:8088/endpoint",
						Type: loggingv1.OutputTypeSplunk,
						OutputTypeSpec: loggingv1.OutputTypeSpec{
							Splunk: &loggingv1.Splunk{},
						},
					}
					forwarderSpec = &loggingv1.ClusterLogForwarderSpec{}
					forwarderSpec.Pipelines = []loggingv1.PipelineSpec{{OutputRefs: []string{output.Name}}}
					forwarderSpec.Outputs = []loggingv1.OutputSpec{output}
					verifyOutputs(namespace, client, forwarderSpec, clfStatus, extras)
					Expect(clfStatus.Outputs["aName"]).To(HaveCondition("Ready", false, "MissingResource", "secret must be provided for splunk output"))
				})

				It("should fail outputs with secrets that have empty hecToken", func() {
					secret.Data[constants.SplunkHECTokenKey] = []byte{}
					client = fake.NewFakeClient(secret) //nolint
					verifyOutputs(namespace, client, forwarderSpec, clfStatus, extras)
					Expect(clfStatus.Outputs["aName"]).To(HaveCondition("Ready", false, "MissingResource", "A non-empty hecToken entry is required"))
				})

				It("should pass outputs with secrets that have hecToken", func() {
					secret.Data[constants.SplunkHECTokenKey] = []byte{0, 1, 2}
					client = fake.NewFakeClient(secret) //nolint
					verifyOutputs(namespace, client, forwarderSpec, clfStatus, extras)
					Expect(forwarderSpec.Outputs).To(HaveLen(len(forwarderSpec.Outputs)))
					Expect(clfStatus.Outputs["aName"]).To(HaveCondition("Ready", true, "", ""))
				})

				It("should fail outputs without URL", func() {
					forwarderSpec.Outputs[0].URL = ""
					secret.Data[constants.SplunkHECTokenKey] = []byte{0, 1, 2}
					client = fake.NewFakeClient(secret) //nolint
					verifyOutputs(namespace, client, forwarderSpec, clfStatus, extras)
					Expect(clfStatus.Outputs["aName"]).To(HaveCondition("Ready", false, "Invalid", "URL is required for output type splunk"))
				})
			})

			Context("for writing to Azure Monitor", func() {
				const missingMessage = "A non-empty shared_key entry is required"
				BeforeEach(func() {
					output = loggingv1.OutputSpec{
						Name: "aName",
						Type: loggingv1.OutputTypeAzureMonitor,
						OutputTypeSpec: loggingv1.OutputTypeSpec{
							AzureMonitor: &loggingv1.AzureMonitor{
								LogType:    string(rand.Word(10)),
								CustomerId: string(rand.Word(20)),
							},
						},
						Secret: &loggingv1.OutputSecretSpec{Name: secret.Name},
					}
					forwarderSpec.Pipelines = []loggingv1.PipelineSpec{{OutputRefs: []string{output.Name}}}
					forwarderSpec.Outputs = []loggingv1.OutputSpec{output}
				})

				It("should fail outputs with secrets that are missing shared_key", func() {
					client = fake.NewFakeClient(secret) //nolint
					verifyOutputs(namespace, client, forwarderSpec, clfStatus, extras)
					Expect(clfStatus.Outputs["aName"]).To(HaveCondition("Ready", false, "MissingResource", missingMessage))
				})

				It("should pass outputs with secrets that have credentials key with valid arn specified", func() {
					secret.Data[constants.SharedKey] = rand.Word(20)
					client = fake.NewFakeClient(secret) //nolint
					verifyOutputs(namespace, client, forwarderSpec, clfStatus, extras)
					Expect(forwarderSpec.Outputs).To(HaveLen(len(forwarderSpec.Outputs)))
					Expect(clfStatus.Outputs["aName"]).To(HaveCondition("Ready", true, "", ""))
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
					forwarderSpec.Pipelines = []loggingv1.PipelineSpec{{OutputRefs: []string{output.Name}}}
					forwarderSpec.Outputs = []loggingv1.OutputSpec{output}
				})
				It("should fail outputs with secrets that have missing tls.key", func() {
					secret.Data["tls.crt"] = []byte{0, 1, 2}
					client = fake.NewFakeClient(secret) //nolint
					verifyOutputs(namespace, client, forwarderSpec, clfStatus, extras)
					Expect(clfStatus.Outputs["aName"]).To(HaveCondition("Ready", false, "MissingResource", "cannot have.*without"))
				})
				It("should fail outputs with secrets that have empty tls.crt", func() {
					secret.Data["tls.crt"] = []byte{}
					secret.Data["tls.key"] = []byte{1, 2, 3}
					client = fake.NewFakeClient(secret) //nolint
					verifyOutputs(namespace, client, forwarderSpec, clfStatus, extras)
					Expect(clfStatus.Outputs["aName"]).To(HaveCondition("Ready", false, "MissingResource", "cannot have.*without"))
				})
				It("should fail outputs with secrets that have missing tls.crt", func() {
					secret.Data["tls.key"] = []byte{0, 1, 2}
					client = fake.NewFakeClient(secret) //nolint
					verifyOutputs(namespace, client, forwarderSpec, clfStatus, extras)
					Expect(clfStatus.Outputs["aName"]).To(HaveCondition("Ready", false, "MissingResource", "cannot have.*without"))
				})
				It("should fail outputs with secrets that have empty tls.key", func() {
					secret.Data["tls.key"] = []byte{}
					secret.Data["tls.crt"] = []byte{1, 2, 3}
					client = fake.NewFakeClient(secret) //nolint
					verifyOutputs(namespace, client, forwarderSpec, clfStatus, extras)
					Expect(clfStatus.Outputs["aName"]).To(HaveCondition("Ready", false, "MissingResource", "cannot have.*without"))
				})
				It("should pass outputs with secrets that have tls.key and tls.cert", func() {
					secret.Data["tls.key"] = []byte{0, 1, 2}
					secret.Data["tls.crt"] = []byte{0, 1, 2}
					client = fake.NewFakeClient(secret) //nolint
					verifyOutputs(namespace, client, forwarderSpec, clfStatus, extras)
					Expect(forwarderSpec.Outputs).To(HaveLen(len(forwarderSpec.Outputs)))
					Expect(clfStatus.Outputs["aName"]).To(HaveCondition("Ready", true, "", ""))
				})
			})
		})

		It("should pass well formed outputs", func() {
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
			forwarderSpec.Pipelines = []loggingv1.PipelineSpec{{OutputRefs: []string{"aName"}}}
			forwarderSpec.Outputs = append(forwarderSpec.Outputs,
				loggingv1.OutputSpec{
					Name:   "aName",
					Type:   "elasticsearch",
					URL:    "https://somewhere",
					Secret: &loggingv1.OutputSecretSpec{Name: "mysecret"},
				},
			)
			verifyOutputs(namespace, client, forwarderSpec, clfStatus, extras)
			Expect(clfStatus.Outputs["aName"]).To(HaveCondition("Ready", true, "", ""), fmt.Sprintf("status: %+v", clfStatus))
			Expect(forwarderSpec.Outputs).To(HaveLen(len(forwarderSpec.Outputs)), fmt.Sprintf("status: %+v", clfStatus))
		})

		Context("splunk custom index", func() {
			var (
				forwarderSpec    = loggingv1.ClusterLogForwarderSpec{}
				splunkOutputName = "splunk-index"
				splunkSecret     *corev1.Secret
			)

			BeforeEach(func() {
				splunkSecret = runtime.NewSecret(constants.OpenshiftNS, "mysecret", map[string][]byte{
					constants.SplunkHECTokenKey: {'t', 'o', 'k', 'e', 'n'},
				})
				forwarderSpec.Pipelines = []loggingv1.PipelineSpec{{OutputRefs: []string{splunkOutputName}}}
				client = fake.NewFakeClient(splunkSecret) //nolint
			})

			It("should pass if only IndexKey is spec'd", func() {
				forwarderSpec.Outputs = []loggingv1.OutputSpec{
					{
						Name: splunkOutputName,
						Type: loggingv1.OutputTypeSplunk,
						URL:  "https://splunk-web:8088/endpoint",
						OutputTypeSpec: loggingv1.OutputTypeSpec{
							Splunk: &loggingv1.Splunk{
								IndexKey: "kubernetes.namespace_name",
							},
						},
						Secret: &loggingv1.OutputSecretSpec{Name: splunkSecret.Name},
					},
				}
				verifyOutputs(namespace, client, &forwarderSpec, clfStatus, extras)
				Expect(forwarderSpec.Outputs).To(HaveLen(len(forwarderSpec.Outputs)))
				Expect(clfStatus.Outputs[splunkOutputName]).To(HaveCondition("Ready", true, "", ""), fmt.Sprintf("status: %+v", clfStatus))
			})

			It("should pass if only IndexName is spec'd", func() {
				forwarderSpec.Outputs = []loggingv1.OutputSpec{
					{
						Name: splunkOutputName,
						Type: loggingv1.OutputTypeSplunk,
						URL:  "https://splunk-web:8088/endpoint",
						OutputTypeSpec: loggingv1.OutputTypeSpec{
							Splunk: &loggingv1.Splunk{
								IndexName: "custom-index",
							},
						},
						Secret: &loggingv1.OutputSecretSpec{Name: splunkSecret.Name},
					},
				}
				verifyOutputs(namespace, client, &forwarderSpec, clfStatus, extras)
				Expect(forwarderSpec.Outputs).To(HaveLen(len(forwarderSpec.Outputs)))
				Expect(clfStatus.Outputs[splunkOutputName]).To(HaveCondition("Ready", true, "", ""), fmt.Sprintf("status: %+v", clfStatus))
			})

			It("should pass if IndexKey && IndexName are not spec'd", func() {
				forwarderSpec.Outputs = []loggingv1.OutputSpec{
					{
						Name:           splunkOutputName,
						Type:           loggingv1.OutputTypeSplunk,
						URL:            "https://splunk-web:8088/endpoint",
						OutputTypeSpec: loggingv1.OutputTypeSpec{},
						Secret:         &loggingv1.OutputSecretSpec{Name: splunkSecret.Name},
					},
				}
				verifyOutputs(namespace, client, &forwarderSpec, clfStatus, extras)
				Expect(forwarderSpec.Outputs).To(HaveLen(len(forwarderSpec.Outputs)))
				Expect(clfStatus.Outputs[splunkOutputName]).To(HaveCondition("Ready", true, "", ""), fmt.Sprintf("status: %+v", clfStatus))
			})

			It("should fail if both IndexKey && IndexName is spec'd", func() {
				forwarderSpec.Outputs = []loggingv1.OutputSpec{
					{
						Name: splunkOutputName,
						Type: loggingv1.OutputTypeSplunk,
						URL:  "https://splunk-web:8088/endpoint",
						OutputTypeSpec: loggingv1.OutputTypeSpec{
							Splunk: &loggingv1.Splunk{
								IndexKey:  "kubernetes.namespace_name",
								IndexName: "custom-index",
							},
						},
						Secret: &loggingv1.OutputSecretSpec{Name: splunkSecret.Name},
					},
				}
				verifyOutputs(namespace, client, &forwarderSpec, clfStatus, extras)
				Expect(forwarderSpec.Outputs).To(HaveLen(len(forwarderSpec.Outputs)))
				Expect(clfStatus.Outputs[splunkOutputName]).To(HaveCondition("Ready", false, loggingv1.ReasonInvalid, "output \""+splunkOutputName+"\": Only one of indexKey or indexName can be set, not both."))
			})

			It("should pass with condition Degraded if Fields is included in spec", func() {
				forwarderSpec.Outputs = []loggingv1.OutputSpec{
					{
						Name: splunkOutputName,
						Type: loggingv1.OutputTypeSplunk,
						URL:  "https://splunk-web:8088/endpoint",
						OutputTypeSpec: loggingv1.OutputTypeSpec{
							Splunk: &loggingv1.Splunk{
								Fields: []string{"foo1", "foo2"},
							},
						},
						Secret: &loggingv1.OutputSecretSpec{Name: splunkSecret.Name},
					},
				}
				verifyOutputs(namespace, client, &forwarderSpec, clfStatus, extras)
				Expect(forwarderSpec.Outputs).To(HaveLen(len(forwarderSpec.Outputs)))
				Expect(clfStatus.Outputs[splunkOutputName]).To(HaveCondition(loggingv1.ConditionDegraded, true, loggingv1.ReasonUnused, "Warning: Support for 'fields' is not implemented and deprecated for output \""+splunkOutputName+"\""))
			})
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
					clusterlogforwarder.NewDefaultOutput(nil, constants.CollectorName),
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
					output.Name:                 []status.Condition{condReady},
					otherOutput.Name:            []status.Condition{condReady},
					loggingv1.OutputNameDefault: []status.Condition{condReady},
				},
			}
		})

		It("should fail all pipelines if output refs are invalid.", func() {
			forwarderSpec.Pipelines = []loggingv1.PipelineSpec{
				{
					Name:       "aPipeline",
					OutputRefs: []string{"someotherendpoint"},
					InputRefs:  []string{loggingv1.InputNameApplication},
				},
			}
			verifyPipelines(constants.SingletonName, &forwarderSpec, clfStatus)
			Expect(clfStatus.Pipelines["aPipeline"]).To(HaveCondition(loggingv1.ValidationCondition, true, loggingv1.ValidationFailureReason, `unrecognized outputs: \[someotherendpoint\], no valid outputs`))
		})

		It("should fail all pipelines if even one pipeline does not have a unique name", func() {
			forwarderSpec.Pipelines = append(forwarderSpec.Pipelines,
				loggingv1.PipelineSpec{
					Name:       "aPipeline",
					OutputRefs: []string{output.Name, otherOutput.Name},
					InputRefs:  []string{loggingv1.InputNameApplication},
				})
			verifyPipelines(constants.SingletonName, &forwarderSpec, clfStatus)
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
			verifyPipelines(constants.SingletonName, &forwarderSpec, clfStatus)
			Expect(clfStatus.Pipelines["pipeline_1_"]).To(HaveCondition(loggingv1.ConditionReady, false, "Invalid", "pipeline must have a name"))
		})

		It("should fail all pipelines if pipelines have unrecognized inputRefs", func() {
			forwarderSpec.Pipelines = []loggingv1.PipelineSpec{
				{
					Name:       "someDefinedPipeline",
					OutputRefs: []string{output.Name, otherOutput.Name},
					InputRefs:  []string{"foo"},
				},
			}
			verifyPipelines(constants.SingletonName, &forwarderSpec, clfStatus)
			conds := clfStatus.Pipelines["someDefinedPipeline"]
			Expect(conds).To(HaveCondition(loggingv1.ValidationCondition, true, loggingv1.ValidationFailureReason, `inputs:.*\[foo]`))
		})

		It("should fail all pipelines if pipelines have no outputRefs", func() {
			forwarderSpec.Pipelines = append(forwarderSpec.Pipelines,
				loggingv1.PipelineSpec{
					Name:       "someDefinedPipeline",
					OutputRefs: []string{},
					InputRefs:  []string{loggingv1.InputNameApplication},
				})
			verifyPipelines(constants.SingletonName, &forwarderSpec, clfStatus)
			conds := clfStatus.Pipelines["someDefinedPipeline"]
			Expect(conds).To(HaveCondition(loggingv1.ValidationCondition, true, loggingv1.ValidationFailureReason, "no valid outputs"))
		})

		// Degraded here means partially valid, which will not be supported
		It("should fail all pipelines if there are degraded pipelines with some bad outputRefs", func() {
			forwarderSpec.Pipelines = append(forwarderSpec.Pipelines,
				loggingv1.PipelineSpec{
					Name:       "someDefinedPipeline",
					OutputRefs: []string{output.Name, otherOutput.Name, "aMissingOutput"},
					InputRefs:  []string{loggingv1.InputNameApplication},
				})
			verifyPipelines(constants.SingletonName, &forwarderSpec, clfStatus)
			Expect(clfStatus.Pipelines).To(HaveLen(2), "Exp. all defined pipelines in clfStatus object")
			Expect(clfStatus.Pipelines).To(HaveKey("someDefinedPipeline"))
			conds := clfStatus.Pipelines["someDefinedPipeline"]
			Expect(conds).To(HaveCondition(loggingv1.ValidationCondition, true, loggingv1.ValidationFailureReason, `invalid: unrecognized outputs: \[aMissingOutput\]`))
		})

		It("should invalidate all pipelines if any input ref is not ready", func() {
			forwarderSpec.Pipelines = append(forwarderSpec.Pipelines,
				loggingv1.PipelineSpec{
					Name:       "someDefinedPipeline",
					OutputRefs: []string{output.Name, otherOutput.Name},
					InputRefs:  []string{input.Name, otherInput.Name},
				})
			verifyPipelines(constants.SingletonName, &forwarderSpec, clfStatus)
			Expect(clfStatus.Pipelines).To(HaveLen(2), "Exp. all defined pipelines in clfStatus object")
			Expect(clfStatus.Pipelines).To(HaveKey("someDefinedPipeline"))
			conds := clfStatus.Pipelines["someDefinedPipeline"]
			Expect(conds).To(HaveCondition(loggingv1.ValidationCondition, true, loggingv1.ValidationFailureReason, `invalid: unrecognized inputs: \[app-input\]`))
		})

		It("should fail if clusterlogforwarder not named instance forwarding to the default logstore", func() {
			forwarderSpec.Pipelines = append(forwarderSpec.Pipelines,
				loggingv1.PipelineSpec{
					Name:       "someDefinedPipeline",
					OutputRefs: []string{output.Name, otherOutput.Name, loggingv1.OutputNameDefault},
					InputRefs:  []string{otherInput.Name},
				})
			verifyPipelines("custom-clf-name", &forwarderSpec, clfStatus)
			Expect(clfStatus.Pipelines).To(HaveLen(2), "Exp. all defined pipelines in clfStatus object")
			Expect(clfStatus.Pipelines).To(HaveKey("someDefinedPipeline"))
			conds := clfStatus.Pipelines["someDefinedPipeline"]
			Expect(conds).To(HaveCondition(loggingv1.ValidationCondition, true, loggingv1.ValidationFailureReason, "invalid: custom ClusterLogForwarders cannot forward to the `default` log store, unrecognized outputs: .?default.?"))
		})

		DescribeTable("gcl and prune filter", func(pruneSpec loggingv1.PruneFilterSpec) {
			gclOutput := loggingv1.OutputSpec{

				Name: "gcl-out",
				Type: "googleCloudLogging",
				OutputTypeSpec: loggingv1.OutputTypeSpec{
					GoogleCloudLogging: &loggingv1.GoogleCloudLogging{BillingAccountID: "billingAccountID"},
				},
			}

			pruneHost := loggingv1.FilterSpec{
				Name: "prune",
				Type: loggingv1.FilterPrune,
				FilterTypeSpec: loggingv1.FilterTypeSpec{
					PruneFilterSpec: &pruneSpec,
				},
			}

			forwarderSpec.Outputs = []loggingv1.OutputSpec{gclOutput}
			forwarderSpec.Filters = []loggingv1.FilterSpec{pruneHost}

			forwarderSpec.Pipelines = []loggingv1.PipelineSpec{
				{
					Name:       "gclPruneHost",
					OutputRefs: []string{"gcl-out"},
					InputRefs:  []string{loggingv1.InputNameApplication},
					FilterRefs: []string{"prune"},
				},
			}
			clfStatus.Outputs.Set("gcl-out", condReady)

			verifyPipelines("custom-clf-name", &forwarderSpec, clfStatus)
			Expect(clfStatus.Pipelines).To(HaveLen(1))
			Expect(clfStatus.Pipelines).To(HaveKey("gclPruneHost"))
			conds := clfStatus.Pipelines["gclPruneHost"]
			Expect(conds).To(HaveCondition(loggingv1.ConditionReady, false, loggingv1.ReasonInvalid, "googleCloudLogging cannot prune `.hostname` field.+"))
		},
			Entry("when `in` prunes .hostname", loggingv1.PruneFilterSpec{In: []string{".hostname"}}),
			Entry("when `notIn` prunes .hostname", loggingv1.PruneFilterSpec{NotIn: []string{".foo"}}),
			Entry("when `in` prunes .hostname but `notIn` does not", loggingv1.PruneFilterSpec{In: []string{".hostname"}, NotIn: []string{".foo"}}),
			Entry("when `notIn` prunes .hostname but `in` does not", loggingv1.PruneFilterSpec{In: []string{".foo"}, NotIn: []string{".foo"}}),
			Entry("when `notIn` && `in` prunes .hostname", loggingv1.PruneFilterSpec{In: []string{".hostname"}, NotIn: []string{".foo"}}),
		)

		DescribeTable("valid prune and gcl pipeline", func(pruneSpec loggingv1.PruneFilterSpec) {
			gclOutput := loggingv1.OutputSpec{

				Name: "gcl-out",
				Type: "googleCloudLogging",
				OutputTypeSpec: loggingv1.OutputTypeSpec{
					GoogleCloudLogging: &loggingv1.GoogleCloudLogging{BillingAccountID: "billingAccountID"},
				},
			}

			pruneHost := loggingv1.FilterSpec{
				Name: "prune",
				Type: loggingv1.FilterPrune,
				FilterTypeSpec: loggingv1.FilterTypeSpec{
					PruneFilterSpec: &pruneSpec,
				},
			}

			forwarderSpec.Outputs = []loggingv1.OutputSpec{gclOutput}
			forwarderSpec.Filters = []loggingv1.FilterSpec{pruneHost}

			forwarderSpec.Pipelines = []loggingv1.PipelineSpec{
				{
					Name:       "gclPruneHost",
					OutputRefs: []string{"gcl-out"},
					InputRefs:  []string{loggingv1.InputNameApplication},
					FilterRefs: []string{"prune"},
				},
			}

			clfStatus.Outputs.Set("gcl-out", condReady)

			verifyPipelines("custom-clf-name", &forwarderSpec, clfStatus)
			Expect(clfStatus.Pipelines["gclPruneHost"]).To(HaveCondition("Ready", true, "", ""))
		},
			Entry("when `in` does not include .hostname", loggingv1.PruneFilterSpec{In: []string{".foo"}}),
			Entry("when `notIn` includes .hostname", loggingv1.PruneFilterSpec{NotIn: []string{".hostname"}}),
			Entry("when `in` does not include and `notIn` includes .hostname", loggingv1.PruneFilterSpec{In: []string{".foo"}, NotIn: []string{".hostname"}}))

		It("should pass validation when prune filters `.hostname` for pipeline without GCL output", func() {
			esOutput := loggingv1.OutputSpec{

				Name: "myOutput",
				Type: "elasticsearch",
				URL:  "http://here",
			}

			pruneHost := loggingv1.FilterSpec{
				Name: "prune",
				Type: loggingv1.FilterPrune,
				FilterTypeSpec: loggingv1.FilterTypeSpec{
					PruneFilterSpec: &loggingv1.PruneFilterSpec{
						In:    []string{".foo, .hostname"},
						NotIn: []string{".foo"},
					},
				},
			}

			forwarderSpec.Outputs = []loggingv1.OutputSpec{esOutput}
			forwarderSpec.Filters = []loggingv1.FilterSpec{pruneHost}

			forwarderSpec.Pipelines = []loggingv1.PipelineSpec{
				{
					Name:       "esPruneHost",
					OutputRefs: []string{"myOutput"},
					InputRefs:  []string{loggingv1.InputNameApplication},
					FilterRefs: []string{"prune"},
				},
			}

			verifyPipelines("custom-clf-name", &forwarderSpec, clfStatus)
			Expect(clfStatus.Pipelines["esPruneHost"]).To(HaveCondition("Ready", true, "", ""))
		})

	})

	Context("validating all", func() {
		var (
			client        client.Client
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

			clfInstance = testRunTime.NewClusterLogForwarder()
			forwarderSpec = &clfInstance.Spec

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

			forwarderSpec.Outputs = []loggingv1.OutputSpec{
				output,
				otherOutput,
			}

			forwarderSpec.Pipelines = []loggingv1.PipelineSpec{
				{
					Name:       "valid-pipeline",
					OutputRefs: []string{output.Name, otherOutput.Name},
					InputRefs:  []string{loggingv1.InputNameApplication},
				},
			}
		})

		It("should fail forwarder spec if outputref is invalid", func() {
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
				clusterlogforwarder.NewDefaultOutput(nil, constants.CollectorName),
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

			_, clfStatus := ValidateInputsOutputsPipelines(*clfInstance, client, extras)

			Expect(forwarderSpec.Pipelines).To(HaveLen(1), "Exp. not to mutate original spec pipelines")
			Expect(forwarderSpec.Inputs).To(BeEmpty(), "Exp. not to mutate original spec inputs")
			Expect(forwarderSpec.Outputs).To(HaveLen(2), "Exp. not to mutate original spec outputs")

			Expect(clfStatus.Outputs["my-cloudwatch"]).To(HaveCondition("Ready", false, "MissingResource", "secret \"inval-secret\" not found"))
			Expect(clfStatus.Pipelines).To(HaveLen(1), "Exp. all defined pipelines to have statuses")
			Expect(clfStatus.Pipelines).To(HaveKey("custom-pipeline"))
			Expect(clfStatus.Pipelines["custom-pipeline"]).To(HaveCondition(loggingv1.ValidationCondition, true, loggingv1.ValidationFailureReason, "invalid*"))
			Expect(clfStatus.Conditions).To(HaveCondition(loggingv1.ConditionReady, false, loggingv1.ReasonInvalid, "invalid clf spec; one or more errors present: *"))
		})

		It("invalid forwarder spec if inputref is invalid", func() {

			invalInput := loggingv1.InputSpec{
				Name: "inval-input",
			}

			forwarderSpec.Outputs = []loggingv1.OutputSpec{
				clusterlogforwarder.NewDefaultOutput(nil, constants.CollectorName),
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

			_, clfStatus := ValidateInputsOutputsPipelines(*clfInstance, client, extras)

			Expect(forwarderSpec.Pipelines).To(HaveLen(1), "Exp. not to mutate original spec pipelines")
			Expect(forwarderSpec.Inputs).To(HaveLen(1), "Exp. not to mutate original spec inputs")
			Expect(forwarderSpec.Outputs).To(HaveLen(1), "Exp. not to mutate original spec outputs")

			Expect(clfStatus.Inputs["inval-input"]).To(HaveCondition("Ready", false, loggingv1.ReasonInvalid, "inputspec must define one and only one of"))
			Expect(clfStatus.Pipelines).To(HaveLen(1), "Exp. all defined pipelines to have statuses")
			Expect(clfStatus.Pipelines).To(HaveKey("custom-pipeline"))
			Expect(clfStatus.Pipelines["custom-pipeline"]).To(HaveCondition(loggingv1.ValidationCondition, true, loggingv1.ValidationFailureReason, "invalid*"))
			Expect(clfStatus.Conditions).To(HaveCondition(loggingv1.ConditionReady, false, loggingv1.ReasonInvalid, "invalid clf spec; one or more errors present: *"))
		})

		It("invalid forwarder spec if pipeline has unknown inputRef", func() {

			appInput := loggingv1.InputSpec{
				Name:        "app-logs",
				Application: &loggingv1.Application{},
			}

			forwarderSpec.Outputs = []loggingv1.OutputSpec{
				clusterlogforwarder.NewDefaultOutput(nil, constants.CollectorName),
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

			_, clfStatus := ValidateInputsOutputsPipelines(*clfInstance, client, extras)

			Expect(forwarderSpec.Pipelines).To(HaveLen(1), "Exp. not to mutate original spec pipelines")
			Expect(forwarderSpec.Inputs).To(HaveLen(1), "Exp. not to mutate original spec inputs")
			Expect(forwarderSpec.Outputs).To(HaveLen(1), "Exp. not to mutate original spec outputs")

			Expect(clfStatus.Inputs["app-logs"]).To(HaveCondition("Ready", true, "", ""))
			Expect(clfStatus.Pipelines).To(HaveLen(1), "Exp. all defined pipelines to have statuses")
			Expect(clfStatus.Pipelines).To(HaveKey("custom-pipeline"))
			Expect(clfStatus.Pipelines["custom-pipeline"]).To(HaveCondition(loggingv1.ValidationCondition, true, loggingv1.ValidationFailureReason, "invalid*"))
			Expect(clfStatus.Conditions).To(HaveCondition(loggingv1.ConditionReady, false, loggingv1.ReasonInvalid, "invalid clf spec; one or more errors present: *"))
		})

		It("invalid forwarder spec if even one pipeline is invalid", func() {

			appInput := loggingv1.InputSpec{
				Name:        "app-logs",
				Application: &loggingv1.Application{},
			}

			forwarderSpec.Outputs = []loggingv1.OutputSpec{
				clusterlogforwarder.NewDefaultOutput(nil, constants.CollectorName),
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

			_, clfStatus := ValidateInputsOutputsPipelines(*clfInstance, client, extras)

			Expect(forwarderSpec.Pipelines).To(HaveLen(2), "Exp. not to mutate original spec pipelines")
			Expect(forwarderSpec.Inputs).To(HaveLen(1), "Exp. not to mutate original spec inputs")
			Expect(forwarderSpec.Outputs).To(HaveLen(1), "Exp. not to mutate original spec outputs")

			Expect(clfStatus.Inputs["app-logs"]).To(HaveCondition("Ready", true, "", ""))
			Expect(clfStatus.Pipelines).To(HaveLen(2), "Exp. all defined pipelines to have statuses")
			Expect(clfStatus.Pipelines).To(HaveKey("inval-pipeline"))
			Expect(clfStatus.Pipelines).To(HaveKey("valid-pipeline"))
			Expect(clfStatus.Pipelines["inval-pipeline"]).To(HaveCondition(loggingv1.ValidationCondition, true, loggingv1.ValidationFailureReason, `invalid: unrecognized inputs: \[missingInRef\]`))
			Expect(clfStatus.Pipelines["valid-pipeline"]).To(HaveCondition(loggingv1.ValidationCondition, true, loggingv1.ValidationFailureReason, "invalid: unrecognized outputs*"))
			Expect(clfStatus.Conditions).To(HaveCondition(loggingv1.ConditionReady, false, loggingv1.ReasonInvalid, "invalid clf spec; one or more errors present: *"))
		})

		It("should have no status if spec has empty pipelines and no forwarder instance", func() {
			forwarderSpec = &loggingv1.ClusterLogForwarderSpec{
				Inputs:    []loggingv1.InputSpec{},
				Outputs:   []loggingv1.OutputSpec{},
				Pipelines: []loggingv1.PipelineSpec{},
			}
			clfInstance.Spec = *forwarderSpec
			Expect(YAMLString(forwarderSpec)).To(EqualLines("{}"))
			_, clfStatus := ValidateInputsOutputsPipelines(*clfInstance, client, extras)
			Expect(YAMLString(clfStatus)).To(EqualLines("{}"))
		})
	})

	Context("filter specs", func() {

		client := fake.NewClientBuilder().WithRuntimeObjects(runtime.NewSecret(
			constants.OpenshiftNS, constants.CollectorSecretName, nil,
		)).Build()
		extras := map[string]bool{}

		const (
			pipelineName = "test"
			outputName   = "my-output"
			filterName   = "my-policy"
			esURL        = "https://es.svc.infra.cluster:9999"
		)

		It("should pass with filter", func() {
			forwarderSpec := &loggingv1.ClusterLogForwarderSpec{
				Filters: []loggingv1.FilterSpec{
					{
						Name: filterName,
						Type: loggingv1.FormatKubeAPIAudit,
						FilterTypeSpec: loggingv1.FilterTypeSpec{
							KubeAPIAudit: &loggingv1.KubeAPIAudit{
								Rules: []v1.PolicyRule{{
									Level: "RequestResponse",
									Resources: []v1.GroupResources{
										{
											Group:     "",
											Resources: []string{"pods"},
										},
									},
								}},
								OmitStages: []v1.Stage{v1.StageRequestReceived},
							},
						},
					},
				},
				Outputs: []loggingv1.OutputSpec{
					{
						Name: outputName,
						Type: loggingv1.OutputTypeElasticsearch,
						URL:  esURL,
						OutputTypeSpec: loggingv1.OutputTypeSpec{
							Elasticsearch: &loggingv1.Elasticsearch{},
						},
					},
				},
				Pipelines: []loggingv1.PipelineSpec{
					{
						FilterRefs: []string{filterName},
						InputRefs:  []string{loggingv1.InputNameApplication, loggingv1.InputNameInfrastructure, loggingv1.InputNameAudit},
						OutputRefs: []string{outputName},
						Name:       pipelineName,
					},
				},
			}
			clf := loggingv1.ClusterLogForwarder{}
			clf.Spec = *forwarderSpec
			clf.Name = constants.SingletonName
			clf.Namespace = constants.OpenshiftNS

			_, clfStatus = ValidateInputsOutputsPipelines(clf, client, extras)
			Expect(clfStatus.Pipelines[pipelineName]).To(HaveCondition(loggingv1.ConditionReady, true, "", ""))
		})

		It("should fail with undefined filter in pipeline", func() {
			forwarderSpec := &loggingv1.ClusterLogForwarderSpec{
				Outputs: []loggingv1.OutputSpec{
					{
						Name: outputName,
						Type: loggingv1.OutputTypeElasticsearch,
						URL:  esURL,
						OutputTypeSpec: loggingv1.OutputTypeSpec{
							Elasticsearch: &loggingv1.Elasticsearch{},
						},
					},
				},
				Pipelines: []loggingv1.PipelineSpec{
					{
						FilterRefs: []string{"does-not-exist"},
						InputRefs:  []string{loggingv1.InputNameApplication, loggingv1.InputNameInfrastructure, loggingv1.InputNameAudit},
						OutputRefs: []string{outputName},
						Name:       pipelineName,
					},
				},
			}
			clf := loggingv1.ClusterLogForwarder{}
			clf.Spec = *forwarderSpec
			clf.Name = constants.SingletonName
			clf.Namespace = constants.OpenshiftNS

			_, clfStatus = ValidateInputsOutputsPipelines(clf, client, extras)
			Expect(clfStatus.Pipelines[pipelineName]).To(HaveCondition(loggingv1.ValidationCondition, true, loggingv1.ValidationFailureReason, "invalid: unrecognized filters*"))
		})

	})

	Context("tuning specs", func() {

		client := fake.NewClientBuilder().WithRuntimeObjects(runtime.NewSecret(
			constants.OpenshiftNS, constants.CollectorSecretName, nil,
		)).Build()
		extras := map[string]bool{}

		const (
			pipelineName = "test"
			outputName   = "my-output"
			esURL        = "https://es.svc.infra.cluster:9999"
		)

		It("should pass with valid tuning parameters", func() {
			forwarderSpec := &loggingv1.ClusterLogForwarderSpec{
				Outputs: []loggingv1.OutputSpec{
					{
						Name: outputName,
						Type: loggingv1.OutputTypeElasticsearch,
						URL:  esURL,
						OutputTypeSpec: loggingv1.OutputTypeSpec{
							Elasticsearch: &loggingv1.Elasticsearch{},
						},
						Tuning: &loggingv1.OutputTuningSpec{
							Compression: "zlib",
							Delivery:    loggingv1.OutputDeliveryModeAtLeastOnce,
						},
					},
				},
				Pipelines: []loggingv1.PipelineSpec{
					{
						InputRefs:  []string{loggingv1.InputNameApplication, loggingv1.InputNameInfrastructure, loggingv1.InputNameAudit},
						OutputRefs: []string{outputName},
						Name:       pipelineName,
					},
				},
			}
			clf := loggingv1.ClusterLogForwarder{}
			clf.Spec = *forwarderSpec
			clf.Name = constants.SingletonName
			clf.Namespace = constants.OpenshiftNS

			_, clfStatus = ValidateInputsOutputsPipelines(clf, client, extras)
			Expect(clfStatus.Pipelines[pipelineName]).To(HaveCondition(loggingv1.ConditionReady, true, "", ""))
		})

		It("should fail with invalid tuning parameters", func() {
			forwarderSpec := &loggingv1.ClusterLogForwarderSpec{
				Outputs: []loggingv1.OutputSpec{
					{
						Name: outputName,
						Type: loggingv1.OutputTypeElasticsearch,
						URL:  esURL,
						OutputTypeSpec: loggingv1.OutputTypeSpec{
							Elasticsearch: &loggingv1.Elasticsearch{},
						},
						Tuning: &loggingv1.OutputTuningSpec{
							Compression: "lz4",
							Delivery:    loggingv1.OutputDeliveryModeAtLeastOnce,
						},
					},
				},
				Pipelines: []loggingv1.PipelineSpec{
					{
						InputRefs:  []string{loggingv1.InputNameApplication, loggingv1.InputNameInfrastructure, loggingv1.InputNameAudit},
						OutputRefs: []string{outputName},
						Name:       pipelineName,
					},
				},
			}
			clf := loggingv1.ClusterLogForwarder{}
			clf.Spec = *forwarderSpec
			clf.Name = constants.SingletonName
			clf.Namespace = constants.OpenshiftNS

			_, clfStatus = ValidateInputsOutputsPipelines(clf, client, extras)
			Expect(clfStatus.Pipelines[pipelineName]).To(HaveCondition(loggingv1.ValidationCondition, true, loggingv1.ValidationFailureReason, "invalid: unrecognized outputs*"))
			Expect(clfStatus.Outputs[outputName]).To(HaveCondition(loggingv1.ConditionReady, false, loggingv1.ReasonInvalid, `output "my-output": compression is not supported for the output type`))
		})
	})
})
