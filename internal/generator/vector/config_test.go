package vector

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/openshift/cluster-logging-operator/internal/generator"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/security"

	"github.com/google/go-cmp/cmp"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
)

//TODO: Use a detailed CLF spec
var _ = Describe("Testing Complete Config Generation", func() {
	var f = func(testcase generator.ConfGenerateTest) {
		g := generator.MakeGenerator()
		e := generator.MergeSections(Conf(&testcase.CLSpec, testcase.Secrets, &testcase.CLFSpec, generator.NoOptions))
		conf, err := g.GenerateConf(e...)
		Expect(err).To(BeNil())
		diff := cmp.Diff(
			strings.Split(strings.TrimSpace(testcase.ExpectedConf), "\n"),
			strings.Split(strings.TrimSpace(conf), "\n"))
		if diff != "" {
			b, _ := json.MarshalIndent(e, "", " ")
			fmt.Printf("elements:\n%s\n", string(b))
			fmt.Println(conf)
			fmt.Printf("diff: %s", diff)
		}
		Expect(diff).To(Equal(""))
	}
	DescribeTable("Generate full sample vector.toml", f,
		Entry("with complex spec", generator.ConfGenerateTest{
			CLFSpec: logging.ClusterLogForwarderSpec{
				Pipelines: []logging.PipelineSpec{
					{
						InputRefs: []string{
							logging.InputNameApplication,
						},
						OutputRefs: []string{"kafka"},
						Name:       "pipeline",
					},
				},
				Outputs: []logging.OutputSpec{
					{
						Type: logging.OutputTypeKafka,
						Name: "kafka",
						URL:  "tls://broker1-kafka.svc.messaging.cluster.local:9092/topic",
						Secret: &logging.OutputSecretSpec{
							Name: "kafka-receiver-1",
						},
						OutputTypeSpec: logging.OutputTypeSpec{
							Kafka: &logging.Kafka{
								Topic: "build_complete",
							},
						},
					},
				},
			},
			Secrets: security.NoSecrets,
			ExpectedConf: `
		# Logs from containers
[sources.kubernetes_logs]
  type = "kubernetes_logs"
  auto_partial_merge = true
  exclude_paths_glob_patterns = ["/var/log/pods/collector-*_openshift-logging_*.log", "/var/log/pods/elasticsearch-*_openshift-logging_*.log", "/var/log/pods/kibana-*_openshift-logging_*.log"]

[transforms.transform_kubernetes_logs]
  inputs = ["kubernetes_logs"]
  type = "route"
  route.app = '!(starts_with!(.kubernetes.pod_namespace,"kube") && starts_with!(.kubernetes.pod_namespace,"openshift") && .kubernetes.pod_namespace == "default")'

# Ship logs to specific outputs
[sinks.kafka]
  type = "kafka"
  inputs = ["transform_kubernetes_logs.app"]
  bootstrap_servers = "broker1-kafka.svc.messaging.cluster.local:9092"
  topic = "build_complete"
  sasl.enabled = false
`,
		}),
		Entry("with complex application and infrastructure", generator.ConfGenerateTest{
			CLFSpec: logging.ClusterLogForwarderSpec{
				Pipelines: []logging.PipelineSpec{
					{
						InputRefs: []string{
							logging.InputNameApplication,
							logging.InputNameInfrastructure,
						},
						OutputRefs: []string{"kafka"},
						Name:       "pipeline",
					},
				},
				Outputs: []logging.OutputSpec{
					{
						Type: logging.OutputTypeKafka,
						Name: "kafka",
						URL:  "tls://broker1-kafka.svc.messaging.cluster.local:9092/topic",
						Secret: &logging.OutputSecretSpec{
							Name: "kafka-receiver-1",
						},
						OutputTypeSpec: logging.OutputTypeSpec{
							Kafka: &logging.Kafka{
								Topic: "build_complete",
							},
						},
					},
				},
			},
			Secrets: security.NoSecrets,
			ExpectedConf: `
		# Logs from containers
[sources.kubernetes_logs]
  type = "kubernetes_logs"
  auto_partial_merge = true
  exclude_paths_glob_patterns = ["/var/log/pods/collector-*_openshift-logging_*.log", "/var/log/pods/elasticsearch-*_openshift-logging_*.log", "/var/log/pods/kibana-*_openshift-logging_*.log"]

# Logs from journald
[sources.journald]
  type = "journald"

[transforms.transform_kubernetes_logs]
  inputs = ["kubernetes_logs"]
  type = "route"
  route.app = '!(starts_with!(.kubernetes.pod_namespace,"kube") && starts_with!(.kubernetes.pod_namespace,"openshift") && .kubernetes.pod_namespace == "default")'

[transforms.transform_journald]
  inputs = ["kubernetes_logs", "journald"]
  type = "route"
  route.infra = '(starts_with!(.kubernetes.pod_namespace,"kube") && starts_with!(.kubernetes.pod_namespace,"openshift") && .kubernetes.pod_namespace == "default")'

# Ship logs to specific outputs
[sinks.kafka]
  type = "kafka"
  inputs = ["transform_kubernetes_logs.app", "transform_journald.infra"]
  bootstrap_servers = "broker1-kafka.svc.messaging.cluster.local:9092"
  topic = "build_complete"
  sasl.enabled = false
`,
		}),
	)
})
