package logstash

import (
	"path"

	testruntime "github.com/openshift/cluster-logging-operator/test/runtime/observability"

	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	"github.com/openshift/cluster-logging-operator/test/framework/functional"

	log "github.com/ViaQ/logerr/v2/log/static"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	"github.com/openshift/cluster-logging-operator/test/helpers/types"
	"github.com/openshift/cluster-logging-operator/test/matchers"
)

var _ = Describe("[Functional][Outputs][Logstash] Output to Logstash", func() {

	//LogstashApplicationLog is the log format as received by
	//Logstash over http
	type LogstashApplicationLog struct {
		types.ApplicationLog
		Tags    []string          `json:"tags"`
		Version string            `json:"@version"`
		Host    string            `json:"host"`
		Port    int               `json:"port"`
		Headers map[string]string `json:"headers"`
	}

	const (
		logStashImage        = "logstash:7.10.1"
		logstashConfFileName = "logstash.yml"
		logstashConf         = `xpack.monitoring.enabled: false`
		pipelineConfFileName = "pipeline.conf"
		//pipelineConf to validate: ./bin/logstash --config.test_and_exit -f <file>
		//replace tabs with spaces
		pipelineConf = `
input {
  http {
	additional_codecs => { "application/json" => "json_lines" }
    port => 24224
  }
}

output {
  file {
    path => '/tmp/app-logs'
  }
}
`
	)
	var (
		framework *functional.CollectorFunctionalFramework

		newVisitor = func(f *functional.CollectorFunctionalFramework) runtime.PodBuilderVisitor {
			return func(b *runtime.PodBuilder) error {
				log.V(2).Info("Adding forward output to logstash", "name", obs.OutputTypeHTTP)
				configName := "logstash-config"
				log.V(2).Info("Creating configmap", "name", configName)
				config := runtime.NewConfigMap(b.Pod.Namespace, configName, map[string]string{
					pipelineConfFileName: pipelineConf,
					logstashConfFileName: logstashConf,
				})
				if err := f.Test.Client.Create(config); err != nil {
					return err
				}

				log.V(2).Info("Adding container", "name", obs.OutputTypeHTTP)
				b.AddContainer(string(obs.OutputTypeHTTP), logStashImage).
					AddVolumeMount("logstash-config", path.Join("/usr/share/logstash/pipeline", pipelineConfFileName), pipelineConfFileName, true).
					AddVolumeMount("logstash-config", path.Join("/usr/share/logstash/config", logstashConfFileName), logstashConfFileName, true).
					End().
					AddConfigMapVolume("logstash-config", config.Name)
				return nil
			}
		}

		// Template expected as output Log
		outputLogTemplate = LogstashApplicationLog{
			ApplicationLog: functional.NewApplicationLogTemplate(),
			Tags:           []string{},
			Version:        "1",
			Host:           "*",
			Port:           0,
		}
	)

	BeforeEach(func() {
		framework = functional.NewCollectorFunctionalFramework()
		addLogStashContainer := newVisitor(framework)
		testruntime.NewClusterLogForwarderBuilder(framework.Forwarder).
			FromInput(obs.InputTypeApplication).
			ToHttpOutput(func(output *obs.OutputSpec) {
				output.HTTP = &obs.HTTP{
					URLSpec: obs.URLSpec{
						URL: "http://0.0.0.0:24224",
					},
				}
			})
		Expect(framework.DeployWithVisitor(addLogStashContainer)).To(BeNil())
		Expect(framework.WritesApplicationLogs(1)).To(BeNil())
	})
	AfterEach(func() {
		framework.Cleanup()
	})

	Context("when sending to Logstash using HTTP", func() {
		It("should  be compatible", func() {
			raw, err := framework.ReadRawApplicationLogsFrom(string(obs.OutputTypeHTTP))
			Expect(err).To(BeNil(), "Expected no errors reading the logs")
			Expect(raw).To(Not(BeEmpty()))

			// Parse log line
			var logs []LogstashApplicationLog
			err = types.StrictlyParseLogs(utils.ToJsonLogs(raw), &logs)
			Expect(err).To(BeNil(), "Expected no errors parsing the logs")
			// Compare to expected template
			outputTestLog := logs[0]
			Expect(outputTestLog).To(matchers.FitLogFormatTemplate(outputLogTemplate))
		})
	})
})
