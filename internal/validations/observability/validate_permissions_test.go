package observability

import (
	"context"
	"fmt"

	internalcontext "github.com/openshift/cluster-logging-operator/internal/api/context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	. "github.com/openshift/cluster-logging-operator/test/matchers"
	authorizationapi "k8s.io/api/authorization/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Describe("[internal][validations] validate clusterlogforwarder permissions", func() {
	var (
		k8sClient      client.Client
		customClf      obs.ClusterLogForwarder
		providedSAName = "test-serviceAccount"

		expectValidateToSucceed = func(success bool, reMsg string) {
			context := internalcontext.ForwarderContext{
				Client:    k8sClient,
				Reader:    k8sClient,
				Forwarder: &customClf,
			}
			ValidatePermissions(context)
			reason := obs.ReasonClusterRolesExist
			if !success {
				reason = obs.ReasonClusterRoleMissing
			}
			Expect(customClf.Status.Conditions).To(HaveCondition(obs.ConditionTypeAuthorized, success, reason, reMsg))
		}
	)

	BeforeEach(func() {
		customClf = obs.ClusterLogForwarder{
			ObjectMeta: v1.ObjectMeta{
				Name:      "test-clf",
				Namespace: constants.OpenshiftNS,
			},
			Spec: obs.ClusterLogForwarderSpec{
				ServiceAccount: obs.ServiceAccount{
					Name: providedSAName,
				},
			},
		}
	})

	Context("service account existence", func() {

		It("should fail when no service account found", func() {
			customClf.Spec.Inputs = []obs.InputSpec{
				{
					Name:        "my-custom-input",
					Type:        obs.InputTypeApplication,
					Application: &obs.Application{},
				},
			}
			customClf.Spec.Pipelines = []obs.PipelineSpec{
				{
					Name: "pipeline1",
					InputRefs: []string{
						string(obs.InputTypeApplication),
					},
				},
			}
			customClf.Spec.ServiceAccount.Name = providedSAName
			k8sClient = fake.NewFakeClient()
			ValidatePermissions(internalcontext.ForwarderContext{
				Client:    k8sClient,
				Reader:    k8sClient,
				Forwarder: &customClf,
			})
			Expect(customClf.Status.Conditions).To(HaveCondition(obs.ConditionTypeAuthorized, false, obs.ReasonServiceAccountDoesNotExist, ""))
		})
	})

	Context("when evaluating inputs", func() {
		var (
			clfServiceAccount = &corev1.ServiceAccount{
				ObjectMeta: v1.ObjectMeta{
					Name:      "test-serviceAccount",
					Namespace: constants.OpenshiftNS,
				},
			}
		)
		BeforeEach(func() {
			k8sClient = &mockSARClient{
				fake.NewFakeClient(clfServiceAccount),
			}
		})
		It("should pass validation for application logs", func() {
			inputName := "some-custom-namespace"
			customClf.Spec = obs.ClusterLogForwarderSpec{
				ServiceAccount: obs.ServiceAccount{
					Name: clfServiceAccount.Name,
				},
				Pipelines: []obs.PipelineSpec{
					{
						Name: "pipeline1",
						InputRefs: []string{
							string(obs.InputTypeApplication),
							inputName,
						},
					},
				},
			}
			expectValidateToSucceed(true, "")
		})

		It("should pass validation for application logs when missing spec", func() {
			customClf.Spec = obs.ClusterLogForwarderSpec{
				Inputs: []obs.InputSpec{
					{
						Name: string(obs.InputTypeApplication),
						Type: obs.InputTypeApplication,
					},
				},
				Pipelines: []obs.PipelineSpec{
					{
						Name: "pipeline1",
						InputRefs: []string{
							string(obs.InputTypeApplication),
						},
					},
				},
				ServiceAccount: obs.ServiceAccount{
					Name: clfServiceAccount.Name,
				},
			}
			expectValidateToSucceed(true, "")
		})

		It("should pass validation when service account can collect specified inputs", func() {
			inputName := "some-custom-namespace"
			customClf.Spec = obs.ClusterLogForwarderSpec{
				ServiceAccount: obs.ServiceAccount{
					Name: clfServiceAccount.Name,
				},
				Inputs: []obs.InputSpec{
					{
						Name: "my-custom-input",
						Application: &obs.Application{
							Includes: []obs.NamespaceContainerSpec{
								{Namespace: inputName},
							},
						},
					},
				},
				Pipelines: []obs.PipelineSpec{
					{
						Name: "pipeline1",
						InputRefs: []string{
							string(obs.InputTypeApplication),
							inputName,
						},
					},
					{
						Name: "pipeline2",
						InputRefs: []string{
							string(obs.InputTypeApplication),
						},
					},
				},
			}
			expectValidateToSucceed(true, "")
		})

		It("should return validation error if service account cannot collect specified inputs", func() {
			customClf.Spec = obs.ClusterLogForwarderSpec{
				ServiceAccount: obs.ServiceAccount{
					Name: clfServiceAccount.Name,
				},
				Pipelines: []obs.PipelineSpec{
					{
						Name: "pipeline1",
						InputRefs: []string{
							string(obs.InputTypeApplication),
							string(obs.InputTypeAudit),
						},
					},
					{
						Name: "pipeline2",
						InputRefs: []string{
							string(obs.InputTypeApplication),
						},
					},
				},
			}
			expectValidateToSucceed(false, "")
		})

		It("should pass validation if service account can collect audit logs and there is an HTTP receiver", func() {
			k8sAuditClient := &mockAuditSARClient{
				fake.NewFakeClient(clfServiceAccount),
			}

			const httpInputName = `http-receiver`
			customClf.Spec = obs.ClusterLogForwarderSpec{
				ServiceAccount: obs.ServiceAccount{
					Name: clfServiceAccount.Name,
				},
				Inputs: []obs.InputSpec{
					{
						Name: httpInputName,
						Type: obs.InputTypeReceiver,
						Receiver: &obs.ReceiverSpec{
							Type: obs.ReceiverTypeHTTP,
							Port: 8080,
							HTTP: &obs.HTTPReceiver{
								Format: obs.HTTPReceiverFormatKubeAPIAudit,
							},
						},
					},
				},

				Pipelines: []obs.PipelineSpec{
					{
						Name: "pipeline1",
						InputRefs: []string{
							httpInputName,
						},
					},
				},
			}
			ValidatePermissions(internalcontext.ForwarderContext{
				Client:    k8sAuditClient,
				Reader:    k8sAuditClient,
				Forwarder: &customClf,
			})
			Expect(customClf.Status.Conditions).To(HaveCondition(obs.ConditionTypeAuthorized, true, obs.ReasonClusterRolesExist, ""))
		})

		It("should pass validation if service account can collect external logs and there is a Syslog receiver", func() {
			const syslogInputName = `syslog-receiver`
			customClf.Spec = obs.ClusterLogForwarderSpec{
				ServiceAccount: obs.ServiceAccount{
					Name: clfServiceAccount.Name,
				},
				Inputs: []obs.InputSpec{
					{
						Name: syslogInputName,
						Type: obs.InputTypeReceiver,
						Receiver: &obs.ReceiverSpec{
							Type: obs.ReceiverTypeSyslog,
							Port: 10514,
						},
					},
				},

				Pipelines: []obs.PipelineSpec{
					{
						Name: "pipeline1",
						InputRefs: []string{
							syslogInputName,
						},
					},
				},
			}
			expectValidateToSucceed(true, "")
		})

		Context("when evaluating custom application inputs that spec infrastructure namespaces", func() {
			const appWithInfraNSInputName = "appWithInfra"
			var k8sAppClient client.Client
			BeforeEach(func() {
				k8sAppClient = &mockAppSARClient{
					fake.NewFakeClient(clfServiceAccount),
				}
			})

			Context("when service account only has collect-application-logs permission", func() {
				It("should pass validation for non-infra namespaces that include infra keywords kube, default, openshift", func() {
					namespaces := []string{"sample-kube-namespace", "my-default-ns", "custom-openshift-namespace", "default-custom"}
					var includes []obs.NamespaceContainerSpec
					for _, ns := range namespaces {
						includes = append(includes, obs.NamespaceContainerSpec{Namespace: ns})
					}
					appInput := "my-app"
					customClf.Spec = obs.ClusterLogForwarderSpec{
						ServiceAccount: obs.ServiceAccount{
							Name: clfServiceAccount.Name,
						},
						Inputs: []obs.InputSpec{
							{
								Name: appInput,
								Type: obs.InputTypeApplication,
								Application: &obs.Application{
									Includes: includes,
								},
							},
						},
						Pipelines: []obs.PipelineSpec{
							{
								Name: "pipeline1",
								InputRefs: []string{
									appInput,
								},
							},
						},
					}
					expectValidateToSucceed(true, "")
				})
				DescribeTable("application input with infrastructure namespaces included", func(infraNS []string) {
					var includes []obs.NamespaceContainerSpec
					for _, ns := range infraNS {
						includes = append(includes, obs.NamespaceContainerSpec{Namespace: ns})
					}
					customClf.Spec = obs.ClusterLogForwarderSpec{
						ServiceAccount: obs.ServiceAccount{
							Name: clfServiceAccount.Name,
						}, Inputs: []obs.InputSpec{
							{
								Name: appWithInfraNSInputName,
								Type: obs.InputTypeApplication,
								Application: &obs.Application{
									Includes: includes,
								},
							},
						},
						Pipelines: []obs.PipelineSpec{
							{
								Name: "pipeline1",
								InputRefs: []string{
									appWithInfraNSInputName,
								},
							},
						},
					}
					ValidatePermissions(internalcontext.ForwarderContext{
						Client:    k8sAppClient,
						Reader:    k8sAppClient,
						Forwarder: &customClf,
					})
					Expect(customClf.Status.Conditions).To(HaveCondition(obs.ConditionTypeAuthorized, false, obs.ReasonClusterRoleMissing, ""))
				},
					Entry("with default namespace", []string{"default"}),
					Entry("with openshift namespace", []string{"openshift"}),
					Entry("with openshift and wildcard namespace", []string{"openshift*"}),
					Entry("with openshift-operators-redhat namespace", []string{"openshift-operators-redhat"}),
					Entry("with kube namespace", []string{"kube"}),
					Entry("with kube and wildcard namespace", []string{"kube*"}),
					Entry("with kube-system namespace", []string{"kube-system"}),
					Entry("with multiple namespaces including an infra namespace", []string{"kube*", "custom-ns"}),
				)

				It("when including infra namespaces and excluding other namespaces", func() {
					customClf.Spec = obs.ClusterLogForwarderSpec{
						ServiceAccount: obs.ServiceAccount{
							Name: clfServiceAccount.Name,
						},
						Inputs: []obs.InputSpec{
							{
								Name: appWithInfraNSInputName,
								Type: obs.InputTypeApplication,
								Application: &obs.Application{
									Includes: []obs.NamespaceContainerSpec{
										{Namespace: "openshift-image-registry"},
										{Namespace: "kube*"},
										{Namespace: "foo"},
									},
									Excludes: []obs.NamespaceContainerSpec{
										{
											Namespace: "foobar",
										},
									},
								},
							},
						},
						Pipelines: []obs.PipelineSpec{
							{
								Name: "pipeline1",
								InputRefs: []string{
									appWithInfraNSInputName,
								},
							},
						},
					}
					expectValidateToSucceed(true, "")
				})
			})

			Context("when service account has collect-application-logs & collect-infra-logs permissions", func() {
				DescribeTable("application input with infrastructure namespaces included", func(infraNS []string) {
					var includes []obs.NamespaceContainerSpec
					for _, ns := range infraNS {
						includes = append(includes, obs.NamespaceContainerSpec{Namespace: ns})
					}
					customClf.Spec = obs.ClusterLogForwarderSpec{
						ServiceAccount: obs.ServiceAccount{
							Name: clfServiceAccount.Name,
						},
						Inputs: []obs.InputSpec{
							{
								Name: appWithInfraNSInputName,
								Type: obs.InputTypeApplication,
								Application: &obs.Application{
									Includes: includes,
								},
							},
						},
						Pipelines: []obs.PipelineSpec{
							{
								Name: "pipeline1",
								InputRefs: []string{
									appWithInfraNSInputName,
								},
							},
						},
					}
					expectValidateToSucceed(true, "")
				},
					Entry("with default namespace", []string{"default"}),
					Entry("with openshift namespace", []string{"openshift"}),
					Entry("with openshift and wildcard namespace", []string{"openshift*"}),
					Entry("with openshift-operators-redhat namespace", []string{"openshift-operators-redhat"}),
					Entry("with kube namespace", []string{"kube"}),
					Entry("with kube and wildcard namespace", []string{"kube*"}),
					Entry("with kube-system namespace", []string{"kube-system"}),
					Entry("with multiple namespaces including an infra namespace", []string{"kube*", "custom-ns"}),
				)

				It("when including infra namespaces and excluding other namespaces", func() {
					customClf.Spec = obs.ClusterLogForwarderSpec{
						ServiceAccount: obs.ServiceAccount{
							Name: clfServiceAccount.Name,
						},
						Inputs: []obs.InputSpec{
							{
								Name: appWithInfraNSInputName,
								Type: obs.InputTypeApplication,
								Application: &obs.Application{
									Includes: []obs.NamespaceContainerSpec{
										{
											Namespace: "openshift-image-registry",
										},
										{
											Namespace: "kube*",
										},
										{
											Namespace: "foo",
										},
									},
									Excludes: []obs.NamespaceContainerSpec{
										{
											Namespace: "foobar",
										},
									},
								},
							},
						},
						Pipelines: []obs.PipelineSpec{
							{
								Name: "pipeline1",
								InputRefs: []string{
									appWithInfraNSInputName,
								},
							},
						},
					}
					expectValidateToSucceed(true, "")
				})
			})
		})
	})
})

// Mocking a subject access review
type mockSARClient struct {
	client.Client
}

func (c *mockSARClient) Create(ctx context.Context, obj client.Object, opts ...client.CreateOption) error {
	sar, ok := obj.(*authorizationapi.SubjectAccessReview)
	if !ok {
		return fmt.Errorf("unexpected object type: %T", obj)
	}
	inputName := sar.Spec.ResourceAttributes.Name
	if inputName == string(obs.InputTypeApplication) || inputName == string(obs.InputTypeInfrastructure) {
		sar.Status.Allowed = true
	}
	return nil
}

type mockAuditSARClient struct {
	client.Client
}

func (c *mockAuditSARClient) Create(ctx context.Context, obj client.Object, opts ...client.CreateOption) error {
	sar, ok := obj.(*authorizationapi.SubjectAccessReview)
	if !ok {
		return fmt.Errorf("unexpected object type: %T", obj)
	}
	inputName := sar.Spec.ResourceAttributes.Name
	if inputName == string(obs.InputTypeAudit) {
		sar.Status.Allowed = true
	}
	return nil
}

type mockAppSARClient struct {
	client.Client
}

func (c *mockAppSARClient) Create(ctx context.Context, obj client.Object, opts ...client.CreateOption) error {
	sar, ok := obj.(*authorizationapi.SubjectAccessReview)
	if !ok {
		return fmt.Errorf("unexpected object type: %T", obj)
	}
	inputName := sar.Spec.ResourceAttributes.Name
	if inputName == string(obs.InputTypeApplication) {
		sar.Status.Allowed = true
	}
	return nil
}
