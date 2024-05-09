package outputs

import (
	testruntime "github.com/openshift/cluster-logging-operator/test/runtime"
	"path"

	"github.com/openshift/cluster-logging-operator/internal/runtime"
	"github.com/openshift/cluster-logging-operator/test/framework/functional"

	log "github.com/ViaQ/logerr/v2/log/static"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	logging "github.com/openshift/cluster-logging-operator/api/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	"github.com/openshift/cluster-logging-operator/test/helpers/types"
	"github.com/openshift/cluster-logging-operator/test/matchers"
)

var _ = Describe("[Functional][Outputs][Logstash] FluentdForward Output to Logstash", func() {

	//LogstashApplicationLog is the log format as received by
	//Logstash over fluentd forward protocol
	type LogstashApplicationLog struct {
		types.ApplicationLog
		Tags    []string `json:"tags"`
		Version string   `json:"@version"`
		Host    string   `json:"host"`
		Port    int      `json:"port"`
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
  tcp {
    codec => fluent{
      nanosecond_precision => true
    }
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
				log.V(2).Info("Adding forward output to logstash", "name", logging.OutputTypeFluentdForward)
				configName := "logstash-config"
				log.V(2).Info("Creating configmap", "name", configName)
				config := runtime.NewConfigMap(b.Pod.Namespace, configName, map[string]string{
					pipelineConfFileName: pipelineConf,
					logstashConfFileName: logstashConf,
				})
				if err := f.Test.Client.Create(config); err != nil {
					return err
				}

				log.V(2).Info("Adding container", "name", logging.OutputTypeFluentdForward)
				b.AddContainer(logging.OutputTypeFluentdForward, logStashImage).
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
		Skip("Enable me for vector?  Over http?")
		framework = functional.NewCollectorFunctionalFramework()
		addLogStashContainer := newVisitor(framework)
		testruntime.NewClusterLogForwarderBuilder(framework.Forwarder).
			FromInput(logging.InputNameApplication).
			ToFluentForwardOutput()
		Expect(framework.DeployWithVisitor(addLogStashContainer)).To(BeNil())
		Expect(framework.WritesApplicationLogs(1)).To(BeNil())
	})
	AfterEach(func() {
		framework.Cleanup()
	})

	Context("when sending to Logstash using fluent's forward protocol", func() {
		It("should  be compatible", func() {
			raw, err := framework.ReadRawApplicationLogsFrom(logging.OutputTypeFluentdForward)
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
