package conf

import (
	"github.com/openshift/cluster-logging-operator/internal/constants"
	. "github.com/openshift/cluster-logging-operator/test/matchers"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("#Global", func() {

	Context("for namespaced forwarders", func() {
		It("should include a global data directory for that forwarder", func() {
			Expect(`
expire_metrics_secs = 60
data_dir = "/var/lib/vector/openshift-logging/my-forwarder"

[api]
enabled = true
address = "0.0.0.0:24686"

# Load sensitive data from files
[secret.kubernetes_secret]
type = "directory"
path = "/var/run/ocp-collector/secrets"
`,
			).To(EqualConfigFrom(Global(constants.OpenshiftNS, "my-forwarder")))
		})
	})
	Context("for the legacy forwarder", func() {
		It("should not include a global data directory for that forwarder", func() {
			Expect(`
expire_metrics_secs = 60

[api]
enabled = true
address = "0.0.0.0:24686"

# Load sensitive data from files
[secret.kubernetes_secret]
type = "directory"
path = "/var/run/ocp-collector/secrets"
`,
			).To(EqualConfigFrom(Global(constants.OpenshiftNS, constants.SingletonName)))

		})
	})
})
