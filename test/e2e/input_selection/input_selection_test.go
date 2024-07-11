package input_selection

import (
	"fmt"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	obsruntime "github.com/openshift/cluster-logging-operator/internal/runtime/observability"
	framework "github.com/openshift/cluster-logging-operator/test/framework/e2e"
	testruntime "github.com/openshift/cluster-logging-operator/test/runtime/observability"
)

// These tests exist as e2e because vector interacts directly with the API server
// and various bits of functionality are not testable using the functional
// framework
var _ = Describe("[e2e][InputSelection]", func() {

	const (
		valueBackend    = "backend"
		valueFrontend   = "frontend"
		valueMiddle     = "middle"
		component       = "component"
		valueFrontendNS = "clo-test-frontend"
	)

	var (
		e2e      *framework.E2ETestFramework
		receiver *framework.VectorHttpReceiverLogStore
		err      error

		logGeneratorNameFn = func(name string) string {
			return "log-generator"
		}
	)

	AfterEach(func() {
		if e2e != nil {
			e2e.Cleanup()
		}
	})

	var _ = DescribeTable("filtering", func(input obs.InputSpec, generatorName func(string) string, verify func()) {
		e2e = framework.NewE2ETestFramework()
		forwarder := obsruntime.NewClusterLogForwarder(e2e.CreateTestNamespace(), "my-log-collector", runtime.Initialize)
		forwarder.Name = "my-log-collector"
		if generatorName == nil {
			generatorName = func(component string) string {
				return component
			}
		}
		e2e.CreateNamespace(valueFrontendNS)
		for componentName, namespace := range map[string]string{
			valueFrontend: valueFrontendNS,
			valueBackend:  e2e.CreateTestNamespace(),
			valueMiddle:   e2e.CreateTestNamespaceWithPrefix("openshift-test")} {
			options := framework.NewDefaultLogGeneratorOptions()
			options.Labels = map[string]string{
				"testtype": "myinfra",
				component:  componentName,
			}
			if err := e2e.DeployLogGeneratorWithNamespace(namespace, generatorName(componentName), options); err != nil {
				Fail(fmt.Sprintf("Timed out waiting for the log generator to deploy: %v", err))
			}
		}

		receiver, err = e2e.DeployHttpReceiver(forwarder.Namespace)
		Expect(err).To(BeNil())
		sa, err := e2e.BuildAuthorizationFor(forwarder.Namespace, forwarder.Name).
			AllowClusterRole("collect-application-logs").
			AllowClusterRole("collect-infrastructure-logs").
			AllowClusterRole("collect-audit-logs").
			Create()
		Expect(err).To(BeNil())
		forwarder.Spec.ServiceAccount.Name = sa.Name
		testruntime.NewClusterLogForwarderBuilder(forwarder).
			FromInputName("myinput", func(spec *obs.InputSpec) {
				spec.Type = input.Type
				spec.Application = input.Application
				spec.Infrastructure = input.Infrastructure
				spec.Audit = input.Audit
			}).ToHttpOutput(func(spec *obs.OutputSpec) {
			spec.HTTP.URL = receiver.ClusterLocalEndpoint()
		})
		if err := e2e.CreateObservabilityClusterLogForwarder(forwarder); err != nil {
			Fail(fmt.Sprintf("Unable to create an instance of logforwarder: %v", err))
		}
		if err := e2e.WaitForDaemonSet(forwarder.Namespace, forwarder.Name); err != nil {
			Fail(fmt.Sprintf("Failed waiting for component %s to be ready: %v", component, err))
		}
		verify()
	},
		Entry("infrastructure inputs should allow specifying only node logs",
			obs.InputSpec{
				Type: obs.InputTypeInfrastructure,
				Infrastructure: &obs.Infrastructure{
					Sources: []obs.InfrastructureSource{obs.InfrastructureSourceNode},
				},
			},
			nil,
			func() {
				Expect(receiver.ListJournalLogs()).ToNot(HaveLen(0), "exp only journal logs to be collected")
				Expect(receiver.ListNamespaces()).To(HaveLen(0), "exp no containers logs to be collected")
			}),
		Entry("infrastructure inputs should allow specifying only container logs",
			obs.InputSpec{
				Type: obs.InputTypeInfrastructure,
				Infrastructure: &obs.Infrastructure{
					Sources: []obs.InfrastructureSource{obs.InfrastructureSourceContainer},
				},
			},
			nil,
			func() {
				Expect(receiver.ListNamespaces()).To(HaveEach(MatchRegexp("^(openshift.*|kube.*|default)$")))
				Expect(receiver.ListJournalLogs()).To(HaveLen(0), "exp no journal logs to be collected")
			}),
		Entry("application inputs should only collect from matching pod label 'notin' expressions",
			obs.InputSpec{
				Type: obs.InputTypeApplication,
				Application: &obs.Application{
					Selector: &metav1.LabelSelector{
						MatchExpressions: []metav1.LabelSelectorRequirement{
							{
								Key: component, Operator: metav1.LabelSelectorOpNotIn, Values: []string{valueFrontend},
							},
						},
					},
				}},
			nil,
			func() {
				containers := receiver.ListContainers()
				Expect(containers).ToNot(BeEmpty(), "Exp. to collect some logs")
				Expect(containers).To(Not(HaveEach(MatchRegexp(fmt.Sprintf("^(%s)$", valueFrontend)))))
			}),
		Entry("application inputs should only collect from matching pod label 'in' expressions",
			obs.InputSpec{
				Type: obs.InputTypeApplication,
				Application: &obs.Application{
					Selector: &metav1.LabelSelector{
						MatchExpressions: []metav1.LabelSelectorRequirement{
							{
								Key: component, Operator: metav1.LabelSelectorOpIn, Values: []string{valueFrontend},
							},
						},
					},
				}},
			nil,
			func() {
				containers := receiver.ListContainers()
				Expect(containers).ToNot(BeEmpty(), "Exp. to collect some logs")
				Expect(containers).To(HaveEach(MatchRegexp(fmt.Sprintf("^(%s)$", valueFrontend))))
			}),
		Entry("application inputs should only collect from matching pod labels",
			obs.InputSpec{
				Type: obs.InputTypeApplication,
				Application: &obs.Application{
					Selector: &metav1.LabelSelector{
						MatchLabels: map[string]string{
							component: valueFrontend,
						},
					},
				}},
			func(component string) string {
				if component == valueFrontend {
					return valueFrontend
				}
				return logGeneratorNameFn(component)
			},
			func() {
				containers := receiver.ListContainers()
				Expect(containers).ToNot(BeEmpty(), "Exp. to collect some logs")
				Expect(containers).To(HaveEach(valueFrontend), "Expected to collect logs from only the the 'frontend' services")
			}),
		Entry("application inputs should only collect from included namespaces with wildcards",
			obs.InputSpec{
				Type: obs.InputTypeApplication,
				Application: &obs.Application{
					Includes: []obs.NamespaceContainerSpec{
						{
							Namespace: "clo-test*",
						},
					},
				}},
			logGeneratorNameFn,
			func() {
				namespaces := receiver.ListNamespaces()
				Expect(namespaces).ToNot(BeEmpty(), "Exp. to collect some logs")
				Expect(namespaces).To(HaveEach(MatchRegexp("^clo-test.*$")))
			}),
		Entry("application inputs should only collect from explicit namespaces",
			obs.InputSpec{
				Type: obs.InputTypeApplication,
				Application: &obs.Application{
					Includes: []obs.NamespaceContainerSpec{
						{
							Namespace: valueFrontendNS,
						},
					},
				}},
			logGeneratorNameFn,
			func() {
				namespaces := receiver.ListNamespaces()
				Expect(namespaces).ToNot(BeEmpty(), "Exp. to collect some logs")
				Expect(namespaces).To(HaveEach(Equal(valueFrontendNS)))
			}),
		Entry("application inputs should only collect from included namespaces with wildcards",
			obs.InputSpec{
				Type: obs.InputTypeApplication,
				Application: &obs.Application{
					Includes: []obs.NamespaceContainerSpec{
						{
							Namespace: "clo-test*",
						},
					},
				}},
			logGeneratorNameFn,
			func() {
				Expect(receiver.ListNamespaces()).To(HaveEach(MatchRegexp("^clo-test.*$")))
			}),
		Entry("application inputs should not collect from excluded namespaces",
			obs.InputSpec{
				Type: obs.InputTypeApplication,
				Application: &obs.Application{
					Excludes: []obs.NamespaceContainerSpec{
						{
							Namespace: "clo-test*",
						},
					},
				}},
			logGeneratorNameFn,
			func() {
				Expect(receiver.ListNamespaces()).To(HaveLen(0), "exp no logs to be collected")
			}),
		Entry("application inputs should collect from included containers",
			obs.InputSpec{
				Type: obs.InputTypeApplication,
				Application: &obs.Application{
					Includes: []obs.NamespaceContainerSpec{
						{
							Container: "log-*",
						},
					},
				}},
			func(name string) string {
				if name == valueFrontend {
					return name
				}
				return logGeneratorNameFn(name)
			},
			func() {
				containers := receiver.ListContainers()
				Expect(containers).ToNot(BeEmpty(), "Exp. to collect some logs")
				Expect(containers).To(HaveEach(MatchRegexp("^log-.*$")))
			}),
		Entry("should not collect from excluded containers",
			obs.InputSpec{
				Type: obs.InputTypeApplication,
				Application: &obs.Application{
					Excludes: []obs.NamespaceContainerSpec{
						{
							Container: "log-*",
						},
					},
				}},
			func(name string) string {
				if name == valueFrontend {
					return name
				}
				return logGeneratorNameFn(name)
			},
			func() {
				containers := receiver.ListContainers()
				Expect(containers).ToNot(BeEmpty(), "Exp. to collect some logs")
				Expect(containers).To(Not(HaveEach(MatchRegexp("^log-.*$"))))
			}),
	)
})
