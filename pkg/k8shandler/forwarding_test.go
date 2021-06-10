package k8shandler

import (
	"fmt"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	"github.com/onsi/gomega/types"
	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/test"
	. "github.com/openshift/cluster-logging-operator/test"
	. "github.com/openshift/cluster-logging-operator/test/matchers"
	core "k8s.io/api/core/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/yaml"
)

const (
	aNamespace      = "aNamespace"
	otherTargetName = "someothername"
)

// Match condition by type, status and reason if reason != "".
// Also match messageRegex if it is not empty.
func matchCondition(t logging.ConditionType, s bool, r logging.ConditionReason, messageRegex string) types.GomegaMatcher {
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

func HaveCondition(t logging.ConditionType, s bool, r logging.ConditionReason, messageRegex string) types.GomegaMatcher {
	return ContainElement(matchCondition(t, s, r, messageRegex))
}

var _ = Describe("Normalizing forwarder", func() {
	var (
		cluster     *logging.ClusterLogging
		output      logging.OutputSpec
		otherOutput logging.OutputSpec
		request     *ClusterLoggingRequest
	)
	BeforeEach(func() {
		output = logging.OutputSpec{
			Name: "myOutput",
			Type: "elasticsearch",
			URL:  "http://here",
		}
		otherOutput = logging.OutputSpec{
			Name: otherTargetName,
			Type: "elasticsearch",
			URL:  "http://there",
		}
		request = &ClusterLoggingRequest{
			Client: fake.NewFakeClient(),
			Cluster: &logging.ClusterLogging{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: aNamespace,
				},
			},
			ForwarderRequest: &logging.ClusterLogForwarder{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "instance",
					Namespace: aNamespace,
				},
			},
		}
		cluster = request.Cluster
	})

	Context("while validating ", func() {
		BeforeEach(func() {
			request.ForwarderSpec = logging.ClusterLogForwarderSpec{
				Outputs: []logging.OutputSpec{
					output,
					otherOutput,
				},
				Pipelines: []logging.PipelineSpec{
					{
						Name:       "aPipeline",
						OutputRefs: []string{output.Name, otherOutput.Name},
						InputRefs:  []string{logging.InputNameApplication},
					},
				},
			}
		})

		Context("pipelines", func() {
			It("should only include inputs if there is at least one valid pipeline", func() {
				request.ForwarderSpec.Pipelines = []logging.PipelineSpec{
					{
						Name:       "aPipeline",
						OutputRefs: []string{"someotherendpoint"},
						InputRefs:  []string{logging.InputNameApplication},
					},
				}
				spec, status := request.NormalizeForwarder()
				Expect(spec.Pipelines).To(BeEmpty(), "Exp. all pipelines to be dropped")
				Expect(status.Inputs).To(BeEmpty())
			})

			It("should drop pipelines that do not have unique names", func() {
				request.ForwarderSpec.Pipelines = append(request.ForwarderSpec.Pipelines,
					logging.PipelineSpec{
						Name:       "aPipeline",
						OutputRefs: []string{output.Name, otherOutput.Name},
						InputRefs:  []string{logging.InputNameApplication},
					})
				spec, status := request.NormalizeForwarder()
				Expect(spec.Pipelines).To(HaveLen(1), JSONString(spec))
				Expect(status.Pipelines).To(HaveKey("pipeline_1_"))
				Expect(status.Pipelines["pipeline_1_"]).To(HaveCondition(logging.ConditionReady, false, "Invalid", "duplicate"))
				Expect(status.Pipelines).To(HaveLen(2))
			})

			It("should allow pipelines with empty/missing names", func() {
				request.ForwarderSpec.Pipelines = append(request.ForwarderSpec.Pipelines,
					logging.PipelineSpec{
						OutputRefs: []string{output.Name},
						InputRefs:  []string{logging.InputNameInfrastructure},
					})
				spec, _ := request.NormalizeForwarder()
				Expect(spec.Pipelines).To(HaveLen(2), "Exp. all pipelines to be ok")
				Expect(spec.Pipelines[0].Name).To(Equal("aPipeline"))
				Expect(spec.Pipelines[1].Name).To(Equal("pipeline_1_"))
			})

			It("should drop pipelines that have unrecognized inputRefs", func() {
				request.ForwarderSpec.Pipelines = []logging.PipelineSpec{
					{
						Name:       "someDefinedPipeline",
						OutputRefs: []string{output.Name, otherOutput.Name},
						InputRefs:  []string{"foo"},
					},
				}
				spec, status := request.NormalizeForwarder()
				conds := status.Pipelines["someDefinedPipeline"]
				Expect(spec.Pipelines).To(BeEmpty(), "Exp. all pipelines to be dropped")
				Expect(conds).To(HaveCondition(logging.ConditionReady, false, logging.ReasonInvalid, `inputs:.*\[foo]`))
			})

			It("should drop pipelines that have no outputRefs", func() {
				request.ForwarderSpec.Pipelines = append(request.ForwarderSpec.Pipelines,
					logging.PipelineSpec{
						Name:       "someDefinedPipeline",
						OutputRefs: []string{},
						InputRefs:  []string{logging.InputNameApplication},
					})
				spec, status := request.NormalizeForwarder()
				conds := status.Pipelines["someDefinedPipeline"]
				Expect(conds).To(HaveCondition(logging.ConditionReady, false, logging.ReasonInvalid, "no valid outputs"))
				Expect(spec.Pipelines).NotTo(ContainElement(
					MatchFields(IgnoreExtras, Fields{"Name": Equal("someDefinedPipeline")})))
				Expect(spec.Pipelines).To(HaveLen(1))
			})

			It("should degrade pipelines with some bad outputRefs", func() {
				request.ForwarderSpec.Pipelines = append(request.ForwarderSpec.Pipelines,
					logging.PipelineSpec{
						Name:       "someDefinedPipeline",
						OutputRefs: []string{output.Name, otherOutput.Name, "aMissingOutput"},
						InputRefs:  []string{logging.InputNameApplication},
					})
				spec, status := request.NormalizeForwarder()
				Expect(spec.Pipelines).To(HaveLen(2), "Exp. all defined pipelines")
				Expect(status.Pipelines).To(HaveLen(2), "Exp. all defined pipelines")
				Expect(status.Pipelines).To(HaveKey("someDefinedPipeline"))
				conds := status.Pipelines["someDefinedPipeline"]
				Expect(conds).To(HaveCondition(logging.ConditionDegraded, true, "Invalid", "aMissingOutput"), YAMLString(status))
				Expect(conds).To(HaveCondition(logging.ConditionReady, true, "", ""))
			})
		})

		Context("outputs", func() {
			It("should drop outputs that do not have unique names", func() {
				request.ForwarderSpec.Outputs = append(request.ForwarderSpec.Outputs, logging.OutputSpec{
					Name: "myOutput",
					Type: "elasticsearch",
					URL:  "http://here",
				})
				// sanity check
				Expect(request.ForwarderSpec.Outputs).To(HaveLen(3))
				spec, status := request.NormalizeForwarder()
				Expect(spec.Outputs).To(HaveLen(2), "Exp. non-unique outputs to be dropped")
				Expect(status.Outputs["myOutput"]).To(HaveCondition(logging.ConditionReady, true, "", ""))
				Expect(status.Outputs["output_2_"]).To(HaveCondition(logging.ConditionReady, false, logging.ReasonInvalid, "duplicate"))
			})

			It("should drop outputs that have empty names", func() {
				request.ForwarderSpec.Outputs = append(request.ForwarderSpec.Outputs, logging.OutputSpec{
					Name: "",
					Type: "elasticsearch",
					URL:  "http://here",
				})
				spec, status := request.NormalizeForwarder()
				Expect(spec.Outputs).To(HaveLen(2), "Exp. outputs with an empty name to be dropped")
				Expect(status.Outputs["output_2_"]).To(HaveCondition("Ready", false, "Invalid", "must have a name"))
			})

			It("should drop outputs that conflict with the internally reserved name", func() {
				request.ForwarderSpec.Outputs = append(request.ForwarderSpec.Outputs, logging.OutputSpec{
					Name: "default",
					Type: "elasticsearch",
					URL:  "http://here",
				})
				spec, status := request.NormalizeForwarder()
				Expect(spec.Outputs).To(HaveLen(2), "Exp. outputs with an internal name conflict to be dropped")
				Expect(status.Outputs).To(HaveKey("output_2_"))
				Expect(status.Outputs["output_2_"]).To(HaveCondition("Ready", false, "Invalid", "reserved"))
			})

			It("should drop outputs that have empty types", func() {
				request.ForwarderSpec.Outputs = append(request.ForwarderSpec.Outputs, logging.OutputSpec{
					Name: "aName",
					URL:  "http://here",
				})
				spec, status := request.NormalizeForwarder()
				Expect(spec.Outputs).To(HaveLen(2), "Exp. outputs with an empty type to be dropped")
				Expect(status.Outputs["aName"]).To(HaveCondition("Ready", false, "Invalid", "unknown.*\"\""))
			})

			It("should drop outputs that have unrecognized types", func() {
				request.ForwarderSpec.Outputs = append(request.ForwarderSpec.Outputs, logging.OutputSpec{
					Name: "aName",
					Type: "foo",
					URL:  "http://here",
				})
				spec, status := request.NormalizeForwarder()
				Expect(spec.Outputs).To(HaveLen(2), "Exp. outputs with an unrecognized type to be dropped")
				Expect(status.Outputs["aName"]).To(HaveCondition("Ready", false, "Invalid", "unknown.*\"foo\""))
			})

			It("should drop outputs that have an invalid or non-absolute URL", func() {
				request.ForwarderSpec.Outputs = []logging.OutputSpec{
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
				spec, status := request.NormalizeForwarder()
				Expect(spec.Outputs).To(HaveLen(0), "Exp. bad endpoint to be dropped")
				Expect(status.Outputs["aName"]).To(HaveCondition("Ready", false, "Invalid", "relativeURLPath"))
				Expect(status.Outputs["bName"]).To(HaveCondition("Ready", false, "Invalid", ":invalid"))
			})

			It("should allow specific outputs that do not require URL", func() {
				request.ForwarderSpec.Outputs = []logging.OutputSpec{
					{
						Name: "aKafka",
						Type: logging.OutputTypeKafka,
					},
					{
						Name: "aCloudwatch",
						Type: logging.OutputTypeCloudwatch,
					},
				}
				spec, status := request.NormalizeForwarder()
				Expect(spec.Outputs).To(HaveLen(len(request.ForwarderSpec.Outputs)))
				Expect(status.Outputs["aKafka"]).To(HaveCondition("Ready", true, "", ""))
				Expect(status.Outputs["aCloudwatch"]).To(HaveCondition("Ready", true, "", ""))
			})

			It("should drop outputs that have secrets with no names", func() {
				request.ForwarderSpec.Outputs = append(request.ForwarderSpec.Outputs, logging.OutputSpec{
					Name:   "aName",
					Type:   "elasticsearch",
					URL:    "https://somewhere",
					Secret: &logging.OutputSecretSpec{},
				})
				spec, status := request.NormalizeForwarder()
				Expect(spec.Outputs).To(HaveLen(2), "Exp. outputs with empty secrets to be dropped")
				Expect(status.Outputs["aName"]).To(HaveCondition("Ready", false, "Invalid", "secret has empty name"))
			})

			It("should drop outputs that have secrets which don't exist", func() {
				request.ForwarderSpec.Outputs = append(request.ForwarderSpec.Outputs, logging.OutputSpec{
					Name:   "aName",
					Type:   "elasticsearch",
					URL:    "https://somewhere",
					Secret: &logging.OutputSecretSpec{Name: "mysecret"},
				})
				spec, status := request.NormalizeForwarder()
				Expect(spec.Outputs).To(HaveLen(2), "Exp. outputs with non-existent secrets to be dropped")
				Expect(status.Outputs["aName"]).To(HaveCondition("Ready", false, "MissingResource", "secret.*not found"))
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
							Namespace: aNamespace,
						},
						Data: map[string][]byte{},
					}
				})
				Context("for writing to Cloudwatch", func() {
					const missingMessage = "aws_access_key_id and aws_secret_access_key are required"
					BeforeEach(func() {
						output = logging.OutputSpec{
							Name:   "aName",
							Type:   logging.OutputTypeCloudwatch,
							Secret: &logging.OutputSecretSpec{Name: secret.Name},
						}
						request.ForwarderSpec.Outputs = []logging.OutputSpec{output}
					})
					It("should drop outputs with secrets that are missing aws_access_key_id and aws_secret_access_key", func() {
						request.Client = fake.NewFakeClient(secret)
						spec, status := request.NormalizeForwarder()
						Expect(spec.Outputs).To(BeEmpty(), fmt.Sprintf("secret %+v", secret))
						Expect(status.Outputs["aName"]).To(HaveCondition("Ready", false, "MissingResource", missingMessage))
					})
					It("should drop outputs with secrets that is missing aws_secret_access_id", func() {
						secret.Data["aws_secret_access_key"] = []byte{0, 1, 2}
						request.Client = fake.NewFakeClient(secret)
						spec, status := request.NormalizeForwarder()
						Expect(spec.Outputs).To(BeEmpty(), fmt.Sprintf("secret %+v", secret))
						Expect(status.Outputs["aName"]).To(HaveCondition("Ready", false, "MissingResource", missingMessage))
					})
					It("should drop outputs with secrets that has empty aws_secret_access_key", func() {
						secret.Data["aws_secret_access_key"] = []byte{}
						secret.Data["aws_access_key_id"] = []byte{1, 2, 3}
						request.Client = fake.NewFakeClient(secret)
						spec, status := request.NormalizeForwarder()
						Expect(spec.Outputs).To(BeEmpty(), fmt.Sprintf("secret %+v", secret))
						Expect(status.Outputs["aName"]).To(HaveCondition("Ready", false, "MissingResource", missingMessage))
					})
					It("should drop outputs with secrets that is missing aws_secret_access_key", func() {
						secret.Data["aws_access_key_id"] = []byte{0, 1, 2}
						request.Client = fake.NewFakeClient(secret)
						spec, status := request.NormalizeForwarder()
						Expect(spec.Outputs).To(BeEmpty(), fmt.Sprintf("secret %+v", secret))
						Expect(status.Outputs["aName"]).To(HaveCondition("Ready", false, "MissingResource", missingMessage))
					})
					It("should drop outputs with secrets that have empty aws_access_key_id", func() {
						secret.Data["aws_access_key_id"] = []byte{}
						secret.Data["aws_secret_access_key"] = []byte{1, 2, 3}
						request.Client = fake.NewFakeClient(secret)
						spec, status := request.NormalizeForwarder()
						Expect(spec.Outputs).To(BeEmpty(), fmt.Sprintf("secret %+v", secret))
						Expect(status.Outputs["aName"]).To(HaveCondition("Ready", false, "MissingResource", missingMessage))
					})
					It("should accept outputs with secrets that have aws_secret_access_key and aws_access_key_id", func() {
						secret.Data["aws_secret_access_key"] = []byte{0, 1, 2}
						secret.Data["aws_access_key_id"] = []byte{0, 1, 2}
						request.Client = fake.NewFakeClient(secret)
						spec, status := request.NormalizeForwarder()
						Expect(spec.Outputs).To(HaveLen(len(request.ForwarderSpec.Outputs)))
						Expect(status.Outputs["aName"]).To(HaveCondition("Ready", true, "", ""))
					})
				})
				Context("with certs", func() {
					BeforeEach(func() {
						output = logging.OutputSpec{
							Name:   "aName",
							Type:   "elasticsearch",
							URL:    "https://somewhere",
							Secret: &logging.OutputSecretSpec{Name: secret.Name},
						}
						request.ForwarderSpec.Outputs = []logging.OutputSpec{output}
					})
					It("should drop outputs with secrets that have missing tls.key", func() {
						secret.Data["tls.crt"] = []byte{0, 1, 2}
						request.Client = fake.NewFakeClient(secret)
						spec, status := request.NormalizeForwarder()
						Expect(spec.Outputs).To(BeEmpty(), fmt.Sprintf("secret %+v", secret))
						Expect(status.Outputs["aName"]).To(HaveCondition("Ready", false, "MissingResource", "cannot have.*without"))
					})
					It("should drop outputs with secrets that have empty tls.crt", func() {
						secret.Data["tls.crt"] = []byte{}
						secret.Data["tls.key"] = []byte{1, 2, 3}
						request.Client = fake.NewFakeClient(secret)
						spec, status := request.NormalizeForwarder()
						Expect(spec.Outputs).To(BeEmpty(), fmt.Sprintf("secret %+v", secret))
						Expect(status.Outputs["aName"]).To(HaveCondition("Ready", false, "MissingResource", "cannot have.*without"))
					})
					It("should drop outputs with secrets that have missing tls.crt", func() {
						secret.Data["tls.key"] = []byte{0, 1, 2}
						request.Client = fake.NewFakeClient(secret)
						spec, status := request.NormalizeForwarder()
						Expect(spec.Outputs).To(BeEmpty(), fmt.Sprintf("secret %+v", secret))
						Expect(status.Outputs["aName"]).To(HaveCondition("Ready", false, "MissingResource", "cannot have.*without"))
					})
					It("should drop outputs with secrets that have empty tls.key", func() {
						secret.Data["tls.key"] = []byte{}
						secret.Data["tls.crt"] = []byte{1, 2, 3}
						request.Client = fake.NewFakeClient(secret)
						spec, status := request.NormalizeForwarder()
						Expect(spec.Outputs).To(BeEmpty(), fmt.Sprintf("secret %+v", secret))
						Expect(status.Outputs["aName"]).To(HaveCondition("Ready", false, "MissingResource", "cannot have.*without"))
					})
					It("should accept outputs with secrets that have tls.key and tls.cert", func() {
						secret.Data["tls.key"] = []byte{0, 1, 2}
						secret.Data["tls.crt"] = []byte{0, 1, 2}
						request.Client = fake.NewFakeClient(secret)
						spec, status := request.NormalizeForwarder()
						Expect(spec.Outputs).To(HaveLen(len(request.ForwarderSpec.Outputs)))
						Expect(status.Outputs["aName"]).To(HaveCondition("Ready", true, "", ""))
					})
				})
			})

			It("should accept well formed outputs", func() {
				request.Client = fake.NewFakeClient(
					&corev1.Secret{
						TypeMeta: metav1.TypeMeta{
							Kind:       "Secret",
							APIVersion: corev1.SchemeGroupVersion.String(),
						},
						ObjectMeta: metav1.ObjectMeta{
							Name:      "mysecret",
							Namespace: aNamespace,
						},
					},
					&corev1.Secret{
						TypeMeta: metav1.TypeMeta{
							Kind:       "Secret",
							APIVersion: corev1.SchemeGroupVersion.String(),
						},
						ObjectMeta: metav1.ObjectMeta{
							Name:      "mycloudwatchsecret",
							Namespace: aNamespace,
						},
					},
				)
				request.ForwarderSpec.Outputs = append(request.ForwarderSpec.Outputs,
					logging.OutputSpec{
						Name:   "aName",
						Type:   "elasticsearch",
						URL:    "https://somewhere",
						Secret: &logging.OutputSecretSpec{Name: "mysecret"},
					},
				)
				spec, status := request.NormalizeForwarder()
				Expect(status.Outputs["aName"]).To(HaveCondition("Ready", true, "", ""), fmt.Sprintf("status: %+v", status))
				Expect(spec.Outputs).To(HaveLen(len(request.ForwarderSpec.Outputs)), fmt.Sprintf("status: %+v", status))
			})

			Context("with outputDefaults specified", func() {
				It("should accept default output", func() {
					cluster.Spec = logging.ClusterLoggingSpec{
						Collection: &logging.CollectionSpec{
							Logs: logging.LogCollectionSpec{
								Type: "fluentd",
							},
						},
						LogStore: &logging.LogStoreSpec{
							Type: logging.LogStoreTypeElasticsearch,
						},
					}
					request.ForwarderSpec = logging.ClusterLogForwarderSpec{
						OutputDefaults: &logging.OutputDefaults{
							Elasticsearch: &logging.Elasticsearch{
								StructuredIndexKey: "kubernetes.labels.mylabel",
							},
						},
						Pipelines: []logging.PipelineSpec{
							{
								InputRefs:  []string{"application"},
								OutputRefs: []string{"default"},
								Name:       "mypipe",
							},
						},
					}
					_, _ = request.generateCollectorConfig()

					Expect(len(request.ForwarderSpec.Outputs) == 1).To(BeTrue())
					Expect(request.ForwarderSpec.Outputs[0].Elasticsearch.StructuredIndexKey).To(Equal("kubernetes.labels.mylabel"))

				})
				It("should setup values for elasticsearch output", func() {
					cluster.Spec = logging.ClusterLoggingSpec{
						Collection: &logging.CollectionSpec{
							Logs: logging.LogCollectionSpec{
								Type: "fluentd",
							},
						},
						LogStore: &logging.LogStoreSpec{
							Type: logging.LogStoreTypeElasticsearch,
						},
					}
					request.ForwarderSpec = logging.ClusterLogForwarderSpec{
						OutputDefaults: &logging.OutputDefaults{
							Elasticsearch: &logging.Elasticsearch{
								StructuredIndexKey: "kubernetes.labels.mylabel",
							},
						},
						Outputs: []logging.OutputSpec{
							{
								Type: logging.OutputTypeElasticsearch,
								Name: "es-out",
								URL:  "http://some-url",
							},
						},
						Pipelines: []logging.PipelineSpec{
							{
								InputRefs:  []string{"application"},
								OutputRefs: []string{"es-out"},
								Name:       "mypipe",
							},
						},
					}
					_, _ = request.generateCollectorConfig()

					Expect(len(request.ForwarderSpec.Outputs) == 1).To(BeTrue())
					Expect(request.ForwarderSpec.Outputs[0].Elasticsearch.StructuredIndexKey).To(Equal("kubernetes.labels.mylabel"))

				})
			})

		})
	})

	Context("with empty forwarder spec", func() {
		BeforeEach(func() {
			request.ForwarderSpec = logging.ClusterLogForwarderSpec{}
			request.ForwarderRequest = &logging.ClusterLogForwarder{}
		})

		It("returns bad status on default output with no default logstore", func() {
			cluster.Spec.LogStore = nil
			spec, status := request.NormalizeForwarder()
			Expect(YAMLString(spec)).To(EqualLines("{}"))
			Expect(status.Conditions).To(HaveCondition("Ready", false, "", ""))
			Expect(spec).To(Equal(&logging.ClusterLogForwarderSpec{}))
		})

		It("generates default configuration for empty spec with default log store", func() {
			cluster.Spec.LogStore = &logging.LogStoreSpec{
				Type: logging.LogStoreTypeElasticsearch,
			}
			spec, status := request.NormalizeForwarder()
			Expect(YAMLString(spec)).To(EqualLines(`
outputs:
- name: default
	secret:
		name: fluentd
	type: elasticsearch
	url: https://elasticsearch.openshift-logging.svc:9200
pipelines:
- inputRefs:
	- application
	- infrastructure
	name: pipeline_0_
	outputRefs:
	- default
`))
			Expect(status.Conditions).To(HaveCondition("Ready", true, "", ""))
			Expect(status.Pipelines["pipeline_0_"]).To(HaveCondition("Ready", true, "", ""))
			Expect(status.Outputs["default"]).To(HaveCondition("Ready", true, "", ""))
			Expect(status.Inputs[logging.InputNameApplication]).To(HaveCondition("Ready", true, "", ""))
			Expect(status.Inputs[logging.InputNameInfrastructure]).To(HaveCondition("Ready", true, "", ""))
		})

		It("forwards logs to an explicit default logstore", func() {
			cluster.Spec.LogStore = &logging.LogStoreSpec{
				Type: logging.LogStoreTypeElasticsearch,
			}
			request.ForwarderSpec = logging.ClusterLogForwarderSpec{
				Pipelines: []logging.PipelineSpec{
					{
						InputRefs:  []string{"audit"},
						OutputRefs: []string{"default"},
					},
				},
			}
			spec, status := request.NormalizeForwarder()
			Expect(spec.Outputs).To(HaveLen(1))
			Expect(spec.Outputs[0].Name).To(Equal("default"))
			Expect(spec.Outputs[0].URL).To(Equal("https://elasticsearch.openshift-logging.svc:9200"))
			Expect(spec.Outputs[0].Secret.Name).To(Equal("fluentd"))
			Expect(spec.Outputs[0].Type).To(Equal("elasticsearch"))

			Expect(status.Conditions).To(HaveCondition("Ready", true, "", ""))
			Expect(status.Pipelines).To(HaveLen(1))
			Expect(status.Pipelines["pipeline_0_"]).To(HaveCondition("Ready", true, "", ""))
			Expect(status.Outputs["default"]).To(HaveCondition("Ready", true, "", ""))
			Expect(status.Inputs[logging.InputNameAudit]).To(HaveCondition("Ready", true, "", ""))
		})
	})

	It("parses spec with Inputs and Outputs", func() {
		request.ForwarderSpec = logging.ClusterLogForwarderSpec{
			Outputs: []logging.OutputSpec{
				{
					Name: "out",
					Type: "syslog",
					URL:  "udp://blahblah",
					OutputTypeSpec: logging.OutputTypeSpec{
						Syslog: &logging.Syslog{},
					},
				},
			},
			Pipelines: []logging.PipelineSpec{
				{
					Name:       "test",
					InputRefs:  []string{"audit"},
					OutputRefs: []string{"out"},
				},
			},
		}
		spec, status := request.NormalizeForwarder()
		Expect(status.Conditions).To(HaveCondition("Ready", true, "", ""), "unexpected "+YAMLString(status))
		Expect(status.Conditions).NotTo(HaveCondition("Degraded", true, "", ""), "unexpected "+YAMLString(status))
		Expect(*spec).To(EqualDiff(request.ForwarderSpec))
	})
})

func TestClusterLoggingRequest_generateCollectorConfig(t *testing.T) {
	_ = logging.SchemeBuilder.AddToScheme(scheme.Scheme)

	type fields struct {
		client           client.Client
		cluster          *logging.ClusterLogging
		ForwarderRequest *logging.ClusterLogForwarder
		ForwarderSpec    logging.ClusterLogForwarderSpec
	}
	tests := []struct {
		name       string
		fields     fields
		wantConfig string
		wantErr    bool
	}{
		{
			name: "Valid collector config",
			fields: fields{
				cluster: &logging.ClusterLogging{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "instance",
						Namespace: "openshift-logging",
					},
					Spec: logging.ClusterLoggingSpec{
						LogStore: nil,
						Collection: &logging.CollectionSpec{
							Logs: logging.LogCollectionSpec{
								Type: "fluentd",
								FluentdSpec: logging.FluentdSpec{
									Resources: &core.ResourceRequirements{
										Limits: core.ResourceList{
											"Memory": defaultFluentdMemory,
										},
										Requests: core.ResourceList{
											"Memory": defaultFluentdMemory,
										},
									},
									NodeSelector: map[string]string{"123": "123"},
								},
							},
						},
					},
				},
				ForwarderRequest: &logging.ClusterLogForwarder{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "instance",
						Namespace: "openshift-logging",
					},
				},
				ForwarderSpec: logging.ClusterLogForwarderSpec{},
			},
		},
		{
			name: "Collection not specified. Shouldn't crash",
			fields: fields{
				cluster: &logging.ClusterLogging{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "instance",
						Namespace: "openshift-logging",
					},
					Spec: logging.ClusterLoggingSpec{
						LogStore: nil,
					},
				},
				ForwarderRequest: &logging.ClusterLogForwarder{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "instance",
						Namespace: "openshift-logging",
					},
				},
				ForwarderSpec: logging.ClusterLogForwarderSpec{},
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			clusterRequest := &ClusterLoggingRequest{
				Client:           tt.fields.client,
				Cluster:          tt.fields.cluster,
				ForwarderRequest: tt.fields.ForwarderRequest,
				ForwarderSpec:    tt.fields.ForwarderSpec,
			}

			config := &core.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "secure-forward",
					Namespace: tt.fields.cluster.Namespace,
				},
				Data:       map[string]string{},
				BinaryData: nil,
			}

			clusterRequest.Client = fake.NewFakeClient(tt.fields.cluster, config)

			gotConfig, err := clusterRequest.generateCollectorConfig()
			if (err != nil) != tt.wantErr {
				t.Errorf("generateCollectorConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotConfig != tt.wantConfig {
				t.Errorf("generateCollectorConfig() gotConfig = %v, want %v", gotConfig, tt.wantConfig)
			}
		})
	}
}

var _ = DescribeTable("Normalizing round trip of valid YAML specs",

	func(yamlSpec string) {
		request := ClusterLoggingRequest{
			Client: fake.NewFakeClient(),
			Cluster: &logging.ClusterLogging{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: aNamespace,
				},
			},
		}
		Expect(yaml.Unmarshal([]byte(yamlSpec), &request.ForwarderSpec)).To(Succeed())
		spec, status := request.NormalizeForwarder()
		Expect(status.Conditions).To(HaveCondition("Ready", true, "", ""), JSONString(status))
		Expect(yamlSpec).To(EqualLines(test.YAMLString(spec)))
	},
	Entry("simple", `
outputs:
- name: myOutput
  type: elasticsearch
  url: http://here
- name: someothername
  type: elasticsearch
  url: http://there
pipelines:
- inputRefs:
  - application
  name: aPipeline
  outputRefs:
  - myOutput
  - someothername
`),
	Entry("syslog output", `
outputs:
- name: out
  syslog:
    severity: Alert
  type: syslog
  url: tcp://syslog-receiver.openshift-logging.svc:24224
pipelines:
  - inputRefs:
    - application
    name: foo
    outputRefs:
    - out
`),

	Entry("regression test 1", `
outputs:
- name: foo
  type: fluentdForward
  url: udp://blah:1234
pipelines:
- inputRefs:
  - application
  name: test-app
  outputRefs:
  - foo
- inputRefs:
  - infrastructure
  name: test-infra
  outputRefs:
  - foo
- inputRefs:
  - audit
  name: test-audit
  outputRefs:
  - foo
`),
	Entry("Bug 1866531", `
outputs:
- name: test
  type: fluentdForward
  url: tcp://test.openshift.logging.svc:24224
pipelines:
- inputRefs:
  - application
  name: test
  outputRefs:
  - test
`),
)

func TestClusterLoggingRequest_verifyOutputURL(t *testing.T) {
	type fields struct {
		Client           client.Client
		Cluster          *logging.ClusterLogging
		ForwarderRequest *logging.ClusterLogForwarder
		ForwarderSpec    logging.ClusterLogForwarderSpec
	}
	type args struct {
		output *logging.OutputSpec
		conds  logging.NamedConditions
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
				output: &logging.OutputSpec{
					Name: "test-output",
					Type: "fluentdForward",
					URL:  "",
				},
				conds: logging.NamedConditions{},
			},
			want: false,
		},
		{
			name: "With fluentdForward with URL",
			args: args{
				output: &logging.OutputSpec{
					Name: "test-output",
					Type: "fluentdForward",
					URL:  "http://123.local:9200",
				},
				conds: logging.NamedConditions{},
			},
			want: true,
		},
		{
			name: "With elastic without URL",
			args: args{
				output: &logging.OutputSpec{
					Name: "test-output",
					Type: "elasticsearch",
					URL:  "",
				},
				conds: logging.NamedConditions{},
			},
			want: false,
		},
		{
			name: "With elastic with URL",
			args: args{
				output: &logging.OutputSpec{
					Name: "test-output",
					Type: "elasticsearch",
					URL:  "http://123.local:9200",
				},
				conds: logging.NamedConditions{},
			},
			want: true,
		},
		{
			name: "With kafka without url",
			args: args{
				output: &logging.OutputSpec{
					Name: "test-output",
					Type: "kafka",
					URL:  "",
				},
				conds: logging.NamedConditions{},
			},
			want: true,
		},
		{
			name: "With kafka",
			args: args{
				output: &logging.OutputSpec{
					Name: "test-output",
					Type: "kafka",
					URL:  "https://local.svc",
				},
				conds: logging.NamedConditions{},
			},
			want: true,
		},
		{
			name: "With syslog",
			args: args{
				output: &logging.OutputSpec{
					Name: "test-output",
					Type: "syslog",
					URL:  "https://local.svc",
				},
				conds: logging.NamedConditions{},
			},
			want: true,
		},
		{
			name: "With syslog without url",
			args: args{
				output: &logging.OutputSpec{
					Name: "test-output",
					Type: "syslog",
					URL:  "",
				},
				conds: logging.NamedConditions{},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		tt := tt // Don't bind range variable.
		t.Run(tt.name, func(t *testing.T) {
			clusterRequest := &ClusterLoggingRequest{
				Client:           tt.fields.Client,
				Cluster:          tt.fields.Cluster,
				ForwarderRequest: tt.fields.ForwarderRequest,
				ForwarderSpec:    tt.fields.ForwarderSpec,
			}
			if got := clusterRequest.verifyOutputURL(tt.args.output, tt.args.conds); got != tt.want {
				t.Errorf("verifyOutputURL() = %v, want %v", got, tt.want)
			}
		})
	}
}
