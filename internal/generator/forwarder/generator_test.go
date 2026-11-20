package forwarder_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/factory"
	"github.com/openshift/cluster-logging-operator/internal/generator/forwarder"
	"github.com/openshift/cluster-logging-operator/internal/generator/framework"
	"github.com/openshift/cluster-logging-operator/internal/tls"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/yaml"
)

var _ = Describe("ConfigGenerator", func() {

	It("generates the vector config as YAML", func() {
		g := forwarder.New()
		spec := obs.ClusterLogForwarderSpec{
			Pipelines: []obs.PipelineSpec{
				{
					Name:       "pipe",
					InputRefs:  []string{string(obs.InputTypeApplication)},
					OutputRefs: []string{"myhttp"},
				},
			},
			Outputs: []obs.OutputSpec{
				{
					Name: "myhttp",
					Type: obs.OutputTypeHTTP,
					HTTP: &obs.HTTP{
						URLSpec: obs.URLSpec{URL: "http://localhost:8090"},
					},
				},
			},
		}
		conf, err := g.GenerateConf(map[string]*corev1.Secret{}, spec, constants.OpenshiftNS, "instance",
			factory.ForwarderResourceNames{CommonName: constants.CollectorName},
			framework.Options{framework.ClusterTLSProfileSpec: tls.GetClusterTLSProfileSpec(nil)})
		Expect(err).ToNot(HaveOccurred())

		var generic map[string]interface{}
		Expect(yaml.Unmarshal([]byte(conf), &generic)).To(Succeed())
		Expect(generic).To(HaveKey("sinks"))
		sinks := generic["sinks"].(map[string]interface{})
		Expect(sinks).To(HaveKey("output_myhttp"))
	})

})
