package clusterlogforwarder

import (
	"context"
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	loggingv1 "github.com/openshift/cluster-logging-operator/api/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	authorizationapi "k8s.io/api/authorization/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Describe("[internal][validations] validate clusterlogforwarder permissions", func() {
	var (
		k8sClient      client.Client
		customClf      loggingv1.ClusterLogForwarder
		extras         map[string]bool
		providedSAName = "test-serviceAccount"
	)

	BeforeEach(func() {
		customClf = loggingv1.ClusterLogForwarder{
			ObjectMeta: v1.ObjectMeta{
				Name:      "test-clf",
				Namespace: constants.OpenshiftNS,
			},
			Spec: loggingv1.ClusterLogForwarderSpec{
				ServiceAccountName: providedSAName,
			},
		}
		extras = map[string]bool{}
	})

	Context("service account existence", func() {

		It("should fail when no service account found", func() {
			customClf.Spec.ServiceAccountName = providedSAName
			k8sClient = fake.NewClientBuilder().Build()
			err, _ := ValidateServiceAccount(customClf, k8sClient, extras)
			Expect(err).To(MatchError(MatchRegexp("service account not found: .+")))
		})

		It("should succeed when service account is found", func() {

			clfServiceAccount := &corev1.ServiceAccount{
				ObjectMeta: v1.ObjectMeta{
					Name:      providedSAName,
					Namespace: customClf.Namespace,
				},
			}
			k8sClient = fake.NewClientBuilder().WithObjects(clfServiceAccount).Build()
			//customClf.Spec.ServiceAccountName = providedSAName

			serviceAccount, err := getServiceAccount(customClf.Spec.ServiceAccountName, customClf.Namespace, k8sClient)
			Expect(serviceAccount).ToNot(BeNil())
			Expect(err).To(BeNil())
		})

		It("should return an error if custom clusterlogforwarder does not include a service account name", func() {
			customClf.Spec.ServiceAccountName = ""
			Expect(ValidateServiceAccount(customClf, k8sClient, extras)).ToNot(Succeed())
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
				fake.NewClientBuilder().WithObjects(clfServiceAccount).Build(),
			}
		})
		It("should pass validation for application logs", func() {
			inputName := "some-custom-namespace"
			customClf.Spec = loggingv1.ClusterLogForwarderSpec{
				ServiceAccountName: clfServiceAccount.Name,
				Pipelines: []loggingv1.PipelineSpec{
					{
						Name: "pipeline1",
						InputRefs: []string{
							loggingv1.InputNameApplication,
							inputName,
						},
					},
				},
			}
			Expect(ValidateServiceAccount(customClf, k8sClient, extras)).To(Succeed())
		})
		It("should pass validation when service account can collect specified inputs", func() {
			inputName := "some-custom-namespace"
			customClf.Spec = loggingv1.ClusterLogForwarderSpec{
				ServiceAccountName: clfServiceAccount.Name,
				Inputs: []loggingv1.InputSpec{
					{
						Name: "my-custom-input",
						Application: &loggingv1.Application{
							Namespaces: []string{inputName},
						},
					},
				},
				Pipelines: []loggingv1.PipelineSpec{
					{
						Name: "pipeline1",
						InputRefs: []string{
							loggingv1.InputNameApplication,
							inputName,
						},
					},
					{
						Name: "pipeline2",
						InputRefs: []string{
							loggingv1.InputNameApplication,
						},
					},
				},
			}
			Expect(ValidateServiceAccount(customClf, k8sClient, extras)).To(Succeed())
		})

		It("should return validation error if service account cannot collect specified inputs", func() {
			customClf.Spec = loggingv1.ClusterLogForwarderSpec{
				ServiceAccountName: clfServiceAccount.Name,

				Pipelines: []loggingv1.PipelineSpec{
					{
						Name: "pipeline1",
						InputRefs: []string{
							loggingv1.InputNameApplication,
							loggingv1.InputNameAudit,
						},
					},
					{
						Name: "pipeline2",
						InputRefs: []string{
							loggingv1.InputNameApplication,
						},
					},
				},
			}
			Expect(ValidateServiceAccount(customClf, k8sClient, extras)).To(Not(Succeed()))
		})

		It("should pass validation if service account can collect audit logs and there is an HTTP receiver", func() {
			k8sAuditClient := &mockAuditSARClient{
				fake.NewClientBuilder().WithObjects(clfServiceAccount).Build(),
			}

			const httpInputName = `http-receiver`
			customClf.Spec = loggingv1.ClusterLogForwarderSpec{
				ServiceAccountName: clfServiceAccount.Name,

				Inputs: []loggingv1.InputSpec{
					{
						Name: httpInputName,
						Receiver: &loggingv1.ReceiverSpec{
							Type: loggingv1.ReceiverTypeHttp,
							ReceiverTypeSpec: &loggingv1.ReceiverTypeSpec{
								HTTP: &loggingv1.HTTPReceiver{
									Port:   8080,
									Format: loggingv1.FormatKubeAPIAudit,
								},
							},
						},
					},
				},

				Pipelines: []loggingv1.PipelineSpec{
					{
						Name: "pipeline1",
						InputRefs: []string{
							httpInputName,
						},
					},
				},
			}
			Expect(ValidateServiceAccount(customClf, k8sAuditClient, extras)).To(Succeed())
		})

		It("should pass validation if service account can collect external logs and there is a Syslog receiver", func() {
			const syslogInputName = `syslog-receiver`
			customClf.Spec = loggingv1.ClusterLogForwarderSpec{
				ServiceAccountName: clfServiceAccount.Name,

				Inputs: []loggingv1.InputSpec{
					{
						Name: syslogInputName,
						Receiver: &loggingv1.ReceiverSpec{
							Type: loggingv1.ReceiverTypeSyslog,
							ReceiverTypeSpec: &loggingv1.ReceiverTypeSpec{
								Syslog: &loggingv1.SyslogReceiver{
									Port: 10514,
								},
							},
						},
					},
				},

				Pipelines: []loggingv1.PipelineSpec{
					{
						Name: "pipeline1",
						InputRefs: []string{
							syslogInputName,
						},
					},
				},
			}
			Expect(ValidateServiceAccount(customClf, k8sClient, extras)).To(Succeed())
		})

		It("should pass validation if service account can collect external logs and there is a Syslog receiver without receiverTypeSpec", func() {
			const syslogInputName = `syslog-receiver`
			customClf.Spec = loggingv1.ClusterLogForwarderSpec{
				ServiceAccountName: clfServiceAccount.Name,

				Inputs: []loggingv1.InputSpec{
					{
						Name: syslogInputName,
						Receiver: &loggingv1.ReceiverSpec{
							Type: loggingv1.ReceiverTypeSyslog,
						},
					},
				},

				Pipelines: []loggingv1.PipelineSpec{
					{
						Name: "pipeline1",
						InputRefs: []string{
							syslogInputName,
						},
					},
				},
			}
			Expect(ValidateServiceAccount(customClf, k8sClient, extras)).To(Succeed())
		})

		Context("when evaluating custom application inputs that spec infrastructure namespaces", func() {
			const appWithInfraNSInputName = "appWithInfra"
			var k8sAppClient client.Client
			BeforeEach(func() {
				k8sAppClient = &mockAppSARClient{
					fake.NewClientBuilder().WithObjects(clfServiceAccount).Build(),
				}
			})

			Context("when service account only has collect-application-logs permission", func() {
				It("should pass validation for non-infra namespaces", func() {
					appInput := "my-app"
					customClf.Spec = loggingv1.ClusterLogForwarderSpec{
						ServiceAccountName: clfServiceAccount.Name,
						Inputs: []loggingv1.InputSpec{
							{
								Name: appInput,
								Application: &loggingv1.Application{
									Namespaces: []string{"sample-kube-namespace", "my-default-ns", "custom-openshift-namespace", "default-custom"},
								},
							},
						},
						Pipelines: []loggingv1.PipelineSpec{
							{
								Name: "pipeline1",
								InputRefs: []string{
									appInput,
								},
							},
						},
					}
					Expect(ValidateServiceAccount(customClf, k8sAppClient, extras)).To(Succeed(), "should pass validation for non-infra namespaces")
				})
				DescribeTable("application input with infrastructure namespaces included", func(infraNS []string) {
					customClf.Spec = loggingv1.ClusterLogForwarderSpec{
						ServiceAccountName: clfServiceAccount.Name,
						Inputs: []loggingv1.InputSpec{
							{
								Name: appWithInfraNSInputName,
								Application: &loggingv1.Application{
									Namespaces: infraNS,
								},
							},
						},
						Pipelines: []loggingv1.PipelineSpec{
							{
								Name: "pipeline1",
								InputRefs: []string{
									appWithInfraNSInputName,
								},
							},
						},
					}
					Expect(ValidateServiceAccount(customClf, k8sAppClient, extras)).ToNot(Succeed(), "should fail validation for infra namespaces")
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
					customClf.Spec = loggingv1.ClusterLogForwarderSpec{
						ServiceAccountName: clfServiceAccount.Name,

						Inputs: []loggingv1.InputSpec{
							{
								Name: appWithInfraNSInputName,
								Application: &loggingv1.Application{
									Namespaces: []string{
										"openshift-image-registry",
										"kube*",
										"foo",
									},
									Excludes: []loggingv1.NamespaceContainerSpec{
										{
											Namespace: "foobar",
										},
									},
								},
							},
						},
						Pipelines: []loggingv1.PipelineSpec{
							{
								Name: "pipeline1",
								InputRefs: []string{
									appWithInfraNSInputName,
								},
							},
						},
					}
					Expect(ValidateServiceAccount(customClf, k8sAppClient, extras)).ToNot(Succeed())
				})
			})

			Context("when service account has collect-application-logs & collect-infra-logs permissions", func() {
				DescribeTable("application input with infrastructure namespaces included", func(infraNS []string) {
					customClf.Spec = loggingv1.ClusterLogForwarderSpec{
						ServiceAccountName: clfServiceAccount.Name,
						Inputs: []loggingv1.InputSpec{
							{
								Name: appWithInfraNSInputName,
								Application: &loggingv1.Application{
									Namespaces: infraNS,
								},
							},
						},
						Pipelines: []loggingv1.PipelineSpec{
							{
								Name: "pipeline1",
								InputRefs: []string{
									appWithInfraNSInputName,
								},
							},
						},
					}
					Expect(ValidateServiceAccount(customClf, k8sClient, extras)).To(Succeed(), "should pass validation for infra namespaces")
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
					customClf.Spec = loggingv1.ClusterLogForwarderSpec{
						ServiceAccountName: clfServiceAccount.Name,

						Inputs: []loggingv1.InputSpec{
							{
								Name: appWithInfraNSInputName,
								Application: &loggingv1.Application{
									Includes: []loggingv1.NamespaceContainerSpec{
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
									Excludes: []loggingv1.NamespaceContainerSpec{
										{
											Namespace: "foobar",
										},
									},
								},
							},
						},
						Pipelines: []loggingv1.PipelineSpec{
							{
								Name: "pipeline1",
								InputRefs: []string{
									appWithInfraNSInputName,
								},
							},
						},
					}
					Expect(ValidateServiceAccount(customClf, k8sClient, extras)).To(Succeed())
				})
			})
		})
	})

	It("should not validate clusterlogforwarder named 'instance' in the namespace 'openshift-logging'", func() {
		singletonClf := loggingv1.ClusterLogForwarder{
			ObjectMeta: v1.ObjectMeta{
				Name:      constants.SingletonName,
				Namespace: constants.OpenshiftNS,
			},
			Spec: loggingv1.ClusterLogForwarderSpec{
				ServiceAccountName: "test-sa",
			},
		}
		Expect(ValidateServiceAccount(singletonClf, k8sClient, extras)).To(Succeed())
	})

	It("should fail validation when namespace is openshift-logging and SA name is 'logcollector' because this is reserved for the legacy usecase", func() {
		singletonClf := loggingv1.ClusterLogForwarder{
			ObjectMeta: v1.ObjectMeta{
				Name:      "anything",
				Namespace: constants.OpenshiftNS,
			},
			Spec: loggingv1.ClusterLogForwarderSpec{
				ServiceAccountName: constants.CollectorServiceAccountName,
			},
		}
		Expect(ValidateServiceAccount(singletonClf, k8sClient, extras)).To(MatchError(MatchRegexp("reserved serviceaccount")))
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
	if inputName == loggingv1.InputNameApplication || inputName == loggingv1.InputNameInfrastructure {
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
	if inputName == loggingv1.InputNameAudit {
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
	if inputName == loggingv1.InputNameApplication {
		sar.Status.Allowed = true
	}
	return nil
}
