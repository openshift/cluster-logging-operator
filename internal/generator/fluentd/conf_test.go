package fluentd

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/factory"
	"github.com/openshift/cluster-logging-operator/internal/tls"

	"github.com/openshift/cluster-logging-operator/internal/generator/helpers"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	testhelpers "github.com/openshift/cluster-logging-operator/test/helpers"

	"github.com/google/go-cmp/cmp"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/generator"
	corev1 "k8s.io/api/core/v1"
)

//go:embed conf_test/fluent.conf
var ExpectedFluentConf string

// TODO: Use a detailed CLF spec
var _ = Describe("Testing Complete Config Generation", func() {
	var f = func(testcase testhelpers.ConfGenerateTest) {
		g := generator.MakeGenerator()
		e := generator.MergeSections(Conf(&testcase.CLSpec, testcase.Secrets, &testcase.CLFSpec, constants.OpenshiftNS, constants.SingletonName, factory.GenerateResourceNames(*runtime.NewClusterLogForwarder(constants.OpenshiftNS, constants.SingletonName)), generator.Options{generator.ClusterTLSProfileSpec: tls.GetClusterTLSProfileSpec(nil)}))
		conf, err := g.GenerateConf(e...)
		Expect(err).To(BeNil())
		diff := cmp.Diff(
			strings.Split(strings.Replace(strings.TrimSpace(helpers.FormatFluentConf(testcase.ExpectedConf)), "\n\n", "\n", -1), "\n"),
			strings.Split(strings.Replace(strings.TrimSpace(helpers.FormatFluentConf(conf)), "\n\n", "\n", -1), "\n"))
		if diff != "" {
			b, _ := json.MarshalIndent(e, "", " ")
			fmt.Printf("elements:\n%s\n", string(b))
			fmt.Println(conf)
			fmt.Printf("diff: %s", diff)
		}
		Expect(diff).To(Equal(""))
	}
	DescribeTable("Generate full fluent.conf", f,
		Entry("with complex spec", testhelpers.ConfGenerateTest{
			CLSpec: logging.CollectionSpec{
				Fluentd: &logging.FluentdForwarderSpec{
					Buffer: &logging.FluentdBufferSpec{
						ChunkLimitSize: "8m",
						TotalLimitSize: "800000000",
						OverflowAction: "throw_exception",
					},
				},
			},
			CLFSpec: logging.ClusterLogForwarderSpec{
				Pipelines: []logging.PipelineSpec{
					{
						InputRefs: []string{
							logging.InputNameApplication,
							logging.InputNameInfrastructure,
							logging.InputNameAudit},
						OutputRefs:            []string{"es-1"},
						Name:                  "pipeline",
						DetectMultilineErrors: true,
					},
				},
				Outputs: []logging.OutputSpec{
					{
						Name: "es-1",
						Type: logging.OutputTypeElasticsearch,
						URL:  "https://es.svc.infra.cluster:9999",
						Secret: &logging.OutputSecretSpec{
							Name: "es-1-secret",
						},
					},
				},
			},
			Secrets: map[string]*corev1.Secret{
				"es-1": {
					Data: map[string][]byte{
						"tls.key":       []byte("junk"),
						"tls.crt":       []byte("junk"),
						"ca-bundle.crt": []byte("junk"),
					},
				},
			},
			ExpectedConf: ExpectedFluentConf,
		}),
	)
})
