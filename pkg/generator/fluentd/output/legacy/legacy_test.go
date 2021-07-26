package legacy_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	. "github.com/openshift/cluster-logging-operator/pkg/generator"
	"github.com/openshift/cluster-logging-operator/pkg/generator/fluentd"
	corev1 "k8s.io/api/core/v1"
)

var _ = Describe("fluentd conf generation", func() {
	Describe("generate legacy fluentdforward conf", func() {
		var f = func(clspec logging.ClusterLoggingSpec, secrets map[string]*corev1.Secret, clfspec logging.ClusterLogForwarderSpec, op Options) []Element {
			return MergeElements(
				fluentd.SourcesToInputs(&clfspec, op),
				fluentd.InputsToPipeline(&clfspec, op),
				fluentd.PipelineToOutputs(&clfspec, op),
				fluentd.Outputs(&clspec, secrets, &clfspec, op))
		}
		DescribeTable("for fluentdforward store", TestGenerateConfWith(f),
			Entry("", ConfGenerateTest{
				CLFSpec: logging.ClusterLogForwarderSpec{},
				Options: Options{
					IncludeLegacyForwardConfig: "",
					IncludeLegacySyslogConfig:  "",
				},
				ExpectedConf: `
# Include Infrastructure logs
<match **_default_** **_kube-*_** **_openshift-*_** **_openshift_** journal.** system.var.log**>
  @type relabel
  @label @_INFRASTRUCTURE
</match>

# Include Application logs
<match kubernetes.**>
  @type relabel
  @label @_APPLICATION
</match>

# Include Audit logs
<match linux-audit.log** k8s-audit.log** openshift-audit.log** ovn-audit.log**>
  @type relabel
  @label @_AUDIT
</match>

# Send any remaining unmatched tags to stdout
<match **>
 @type stdout
</match>

# Copying application source type to pipeline
<label @_APPLICATION>
  <filter **>
    @type record_modifier
    <record>
      log_type application
    </record>
  </filter>
  
  <match **>
    @type copy
    <store>
      @type relabel
      @label @_LEGACY_SECUREFORWARD
    </store>
    
    <store>
      @type relabel
      @label @_LEGACY_SYSLOG
    </store>
  </match>
</label>

# Copying infrastructure source type to pipeline
<label @_INFRASTRUCTURE>
  <filter **>
    @type record_modifier
    <record>
      log_type infrastructure
    </record>
  </filter>
  
  <match **>
    @type copy
    <store>
      @type relabel
      @label @_LEGACY_SECUREFORWARD
    </store>
    
    <store>
      @type relabel
      @label @_LEGACY_SYSLOG
    </store>
  </match>
</label>

# Sending audit source type to pipeline
<label @_AUDIT>
  <filter **>
    @type record_modifier
    <record>
      log_type audit
    </record>
  </filter>
  
  <match **>
    @type relabel
    @label @_LEGACY_SYSLOG
  </match>
</label>

# Ship logs to specific outputs
<label @_LEGACY_SECUREFORWARD>
  <match **>
    @type copy
    #include legacy secure-forward.conf
    @include /etc/fluent/configs.d/secure-forward/secure-forward.conf
  </match>
</label>

<label @_LEGACY_SYSLOG>
  <match **>
    @type copy
    #include legacy Syslog
    @include /etc/fluent/configs.d/syslog/syslog.conf
  </match>
</label>
`,
			}))
	})
})

func TestFluendConfGenerator(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Fluend Conf Generation")
}
