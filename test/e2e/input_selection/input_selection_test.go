package input_selection

import (
	"fmt"
	"sync/atomic"
	"time"

	"github.com/openshift/cluster-logging-operator/internal/runtime"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	obsruntime "github.com/openshift/cluster-logging-operator/internal/runtime/observability"
	framework "github.com/openshift/cluster-logging-operator/test/framework/e2e"
	testruntime "github.com/openshift/cluster-logging-operator/test/runtime/observability"
)

// specCounter yields a monotonically increasing value that, combined with the
// Ginkgo parallel process number, produces a token that is unique across every
// spec in every parallel process. This keeps the namespaces, namespace globs and
// pod-label selectors of concurrently running specs from colliding so the suite
// can be run with `ginkgo -p`.
var specCounter int64

// namespaces holds the per-spec unique namespace names and the token they share
// as a prefix. Threading these through the input builders and verifiers (rather
// than referencing shared constants) is what makes the specs parallel-safe.
type namespaces struct {
	token    string
	frontend string
	backend  string
	middle   string
}

// These tests exist as e2e because vector interacts directly with the API server
// and various bits of functionality are not testable using the functional
// framework
var _ = Describe("[e2e][InputSelection]", func() {

	const (
		valueBackend  = "backend"
		valueFrontend = "frontend"
		valueMiddle   = "middle"
		component     = "component"
		// tokenLabel isolates a spec's log generators from those of any other
		// spec running in parallel so label-based inputs only match this spec.
		tokenLabel = "clo-test-token"
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

	var _ = DescribeTable("filtering", func(buildInput func(ns namespaces) obs.InputSpec, generatorName func(string) string, verify func(ns namespaces)) {
		e2e = framework.NewE2ETestFramework()

		// A token unique to this spec across all parallel processes. The
		// application namespaces share it as a prefix; the infrastructure
		// ("middle") namespace embeds it after the required openshift- prefix.
		token := fmt.Sprintf("clo-test-p%d-%d", GinkgoParallelProcess(), atomic.AddInt64(&specCounter, 1))

		forwarder := obsruntime.NewClusterLogForwarder(e2e.CreateTestNamespace(), "my-log-collector", runtime.Initialize)
		forwarder.Name = "my-log-collector"
		if generatorName == nil {
			generatorName = func(component string) string {
				return component
			}
		}

		ns := namespaces{
			token:    token,
			frontend: e2e.CreateNamespace(token + "-frontend"),
			backend:  e2e.CreateTestNamespaceWithPrefix(token),
			middle:   e2e.CreateTestNamespaceWithPrefix("openshift-" + token),
		}
		for componentName, namespace := range map[string]string{
			valueFrontend: ns.frontend,
			valueBackend:  ns.backend,
			valueMiddle:   ns.middle} {
			options := framework.NewDefaultLogGeneratorOptions()
			options.Labels = map[string]string{
				"testtype": "myinfra",
				component:  componentName,
				tokenLabel: token,
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
		input := buildInput(ns)
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
		verify(ns)
	},
		Entry("infrastructure inputs should allow specifying only node logs",
			func(ns namespaces) obs.InputSpec {
				return obs.InputSpec{
					Type: obs.InputTypeInfrastructure,
					Infrastructure: &obs.Infrastructure{
						Sources: []obs.InfrastructureSource{obs.InfrastructureSourceNode},
					},
				}
			},
			nil,
			func(ns namespaces) {
				Expect(receiver.ListJournalLogs()).ToNot(HaveLen(0), "exp only journal logs to be collected")
				Expect(receiver.ListNamespaces(15*time.Second)).To(HaveLen(0), "exp no containers logs to be collected")
			}),
		Entry("infrastructure inputs should allow specifying only container logs",
			func(ns namespaces) obs.InputSpec {
				return obs.InputSpec{
					Type: obs.InputTypeInfrastructure,
					Infrastructure: &obs.Infrastructure{
						Sources: []obs.InfrastructureSource{obs.InfrastructureSourceContainer},
					},
				}
			},
			nil,
			func(ns namespaces) {
				Expect(receiver.ListNamespaces()).To(HaveEach(MatchRegexp("^(openshift.*|kube.*|default)$")))
				Expect(receiver.ListJournalLogs(15*time.Second)).To(HaveLen(0), "exp no journal logs to be collected")
			}),
		Entry("application inputs should only collect from matching pod label 'notin' expressions",
			func(ns namespaces) obs.InputSpec {
				return obs.InputSpec{
					Type: obs.InputTypeApplication,
					Application: &obs.Application{
						Selector: &metav1.LabelSelector{
							MatchExpressions: []metav1.LabelSelectorRequirement{
								{Key: tokenLabel, Operator: metav1.LabelSelectorOpIn, Values: []string{ns.token}},
								{Key: component, Operator: metav1.LabelSelectorOpNotIn, Values: []string{valueFrontend}},
							},
						},
					}}
			},
			nil,
			func(ns namespaces) {
				containers := receiver.ListContainers()
				Expect(containers).ToNot(BeEmpty(), "Exp. to collect some logs")
				Expect(containers).To(Not(HaveEach(MatchRegexp(fmt.Sprintf("^(%s)$", valueFrontend)))))
			}),
		Entry("application inputs should only collect from matching pod label 'in' expressions",
			func(ns namespaces) obs.InputSpec {
				return obs.InputSpec{
					Type: obs.InputTypeApplication,
					Application: &obs.Application{
						Selector: &metav1.LabelSelector{
							MatchExpressions: []metav1.LabelSelectorRequirement{
								{Key: tokenLabel, Operator: metav1.LabelSelectorOpIn, Values: []string{ns.token}},
								{Key: component, Operator: metav1.LabelSelectorOpIn, Values: []string{valueFrontend}},
							},
						},
					}}
			},
			nil,
			func(ns namespaces) {
				containers := receiver.ListContainers()
				Expect(containers).ToNot(BeEmpty(), "Exp. to collect some logs")
				Expect(containers).To(HaveEach(MatchRegexp(fmt.Sprintf("^(%s)$", valueFrontend))))
			}),
		Entry("application inputs should only collect from matching pod labels",
			func(ns namespaces) obs.InputSpec {
				return obs.InputSpec{
					Type: obs.InputTypeApplication,
					Application: &obs.Application{
						Selector: &metav1.LabelSelector{
							MatchLabels: map[string]string{
								tokenLabel: ns.token,
								component:  valueFrontend,
							},
						},
					}}
			},
			func(component string) string {
				if component == valueFrontend {
					return valueFrontend
				}
				return logGeneratorNameFn(component)
			},
			func(ns namespaces) {
				containers := receiver.ListContainers()
				Expect(containers).ToNot(BeEmpty(), "Exp. to collect some logs")
				Expect(containers).To(HaveEach(valueFrontend), "Expected to collect logs from only the the 'frontend' services")
			}),
		Entry("application inputs should only collect from included namespaces with wildcards",
			func(ns namespaces) obs.InputSpec {
				return obs.InputSpec{
					Type: obs.InputTypeApplication,
					Application: &obs.Application{
						Includes: []obs.NamespaceContainerSpec{
							{Namespace: ns.token + "*"},
						},
					}}
			},
			logGeneratorNameFn,
			func(ns namespaces) {
				collected := receiver.ListNamespaces()
				Expect(collected).ToNot(BeEmpty(), "Exp. to collect some logs")
				Expect(collected).To(HaveEach(MatchRegexp("^" + ns.token + ".*$")))
			}),
		Entry("application inputs should only collect from explicit namespaces",
			func(ns namespaces) obs.InputSpec {
				return obs.InputSpec{
					Type: obs.InputTypeApplication,
					Application: &obs.Application{
						Includes: []obs.NamespaceContainerSpec{
							{Namespace: ns.frontend},
						},
					}}
			},
			logGeneratorNameFn,
			func(ns namespaces) {
				collected := receiver.ListNamespaces()
				Expect(collected).ToNot(BeEmpty(), "Exp. to collect some logs")
				Expect(collected).To(HaveEach(Equal(ns.frontend)))
			}),
		Entry("application inputs should not collect from excluded namespaces",
			func(ns namespaces) obs.InputSpec {
				return obs.InputSpec{
					Type: obs.InputTypeApplication,
					Application: &obs.Application{
						Includes: []obs.NamespaceContainerSpec{
							{Namespace: ns.token + "*"},
						},
						Excludes: []obs.NamespaceContainerSpec{
							{Namespace: ns.token + "*"},
						},
					}}
			},
			logGeneratorNameFn,
			func(ns namespaces) {
				Expect(receiver.ListNamespaces()).To(HaveLen(0), "exp no logs to be collected")
			}),
		Entry("application inputs should collect from included containers",
			func(ns namespaces) obs.InputSpec {
				return obs.InputSpec{
					Type: obs.InputTypeApplication,
					Application: &obs.Application{
						Includes: []obs.NamespaceContainerSpec{
							{Namespace: ns.token + "*", Container: "log-*"},
						},
					}}
			},
			func(name string) string {
				if name == valueFrontend {
					return name
				}
				return logGeneratorNameFn(name)
			},
			func(ns namespaces) {
				containers := receiver.ListContainers()
				Expect(containers).ToNot(BeEmpty(), "Exp. to collect some logs")
				Expect(containers).To(HaveEach(MatchRegexp("^log-.*$")))
			}),
		Entry("should not collect from excluded containers",
			func(ns namespaces) obs.InputSpec {
				return obs.InputSpec{
					Type: obs.InputTypeApplication,
					Application: &obs.Application{
						Includes: []obs.NamespaceContainerSpec{
							{Namespace: ns.token + "*"},
						},
						Excludes: []obs.NamespaceContainerSpec{
							{Namespace: ns.token + "*", Container: "log-*"},
						},
					}}
			},
			func(name string) string {
				if name == valueFrontend {
					return name
				}
				return logGeneratorNameFn(name)
			},
			func(ns namespaces) {
				containers := receiver.ListContainers()
				Expect(containers).ToNot(BeEmpty(), "Exp. to collect some logs")
				Expect(containers).To(Not(HaveEach(MatchRegexp("^log-.*$"))))
			}),
	)
})
