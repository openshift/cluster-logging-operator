package otlp

import (
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/api/observability"
	"github.com/openshift/cluster-logging-operator/internal/generator/framework"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
	. "github.com/openshift/cluster-logging-operator/test/matchers"
)

var _ = Describe("Generate vector config", func() {
	DescribeTable("for OTLP output", func(output obs.OutputSpec, secret observability.Secrets, op framework.Options, expFile string) {
		exp, err := tomlContent.ReadFile(expFile)
		if err != nil {
			Fail(fmt.Sprintf("Error reading the file %q with exp config: %v", expFile, err))
		}
		conf := New(helpers.MakeOutputID(output.Name), output, []string{"pipeline_my_pipeline_viaq_0"}, secret, nil, op) //, includeNS, excludes)
		Expect(string(exp)).To(EqualConfigFrom(conf))
	},
		Entry("with only URL spec'd",
			obs.OutputSpec{
				Type: obs.OutputTypeOTLP,
				Name: "otel-collector",
				OTLP: &obs.OTLP{
					URL: "http://localhost:4318/v1/logs",
				},
			},
			nil,
			framework.NoOptions,
			"otlp_all.toml",
		),
	)
})
