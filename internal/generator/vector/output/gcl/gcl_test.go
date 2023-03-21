package gcl

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/generator"
	"github.com/openshift/cluster-logging-operator/test/helpers"
	corev1 "k8s.io/api/core/v1"
)

var _ = Describe("Generate Vector config", func() {
	inputPipeline := []string{"application"}
	var f = func(clspec logging.CollectionSpec, secrets map[string]*corev1.Secret, clfspec logging.ClusterLogForwarderSpec, op generator.Options) []generator.Element {
		e := []generator.Element{}
		for _, o := range clfspec.Outputs {
			e = generator.MergeElements(e, Conf(o, inputPipeline, secrets[o.Name], op))
		}
		return e
	}
	DescribeTable("For GoogleCloudLogging output", helpers.TestGenerateConfWith(f),
		Entry("with service account token", helpers.ConfGenerateTest{
			CLFSpec: logging.ClusterLogForwarderSpec{
				Outputs: []logging.OutputSpec{
					{
						Type: logging.OutputTypeGoogleCloudLogging,
						Name: "gcl-1",
						OutputTypeSpec: logging.OutputTypeSpec{
							GoogleCloudLogging: &logging.GoogleCloudLogging{
								BillingAccountID: "billing-1",
								LogID:            "vector-1",
							},
						},
						Secret: &logging.OutputSecretSpec{
							Name: "junk",
						},
					},
				},
			},
			Secrets: map[string]*corev1.Secret{
				"gcl-1": {
					Data: map[string][]byte{
						GoogleApplicationCredentialsKey: []byte("dummy-credentials"),
					},
				},
			},
			ExpectedConf: `
[transforms.gcl_1_dedot]
type = "lua"
inputs = ["application"]
version = "2"
hooks.init = "init"
hooks.process = "process"
source = '''
    function init()
        count = 0
    end
    function process(event, emit)
        count = count + 1
        event.log.openshift.sequence = count
        if event.log.kubernetes == nil then
            emit(event)
            return
        end
        if event.log.kubernetes.labels == nil then
            emit(event)
            return
        end
		dedot(event.log.kubernetes.namespace_labels)
        dedot(event.log.kubernetes.labels)
        emit(event)
    end
	
    function dedot(map)
        if map == nil then
            return
        end
        local new_map = {}
        local changed_keys = {}
        for k, v in pairs(map) do
            local dedotted = string.gsub(k, "[./]", "_")
            if dedotted ~= k then
                new_map[dedotted] = v
                changed_keys[k] = true
            end
        end
        for k in pairs(changed_keys) do
            map[k] = nil
        end
        for k, v in pairs(new_map) do
            map[k] = v
        end
    end
'''

[sinks.gcl_1]
type = "gcp_stackdriver_logs"
inputs = ["gcl_1_dedot"]
billing_account_id = "billing-1"
credentials_path = "/var/run/ocp-collector/secrets/junk/google-application-credentials.json"
log_id = "vector-1"
severity_key = "level"


[sinks.gcl_1.resource]
type = "k8s_node"
node_name = "{{hostname}}"
`,
		}),
	)
})

func TestVectorConfGenerator(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Vector Conf Generation")
}
