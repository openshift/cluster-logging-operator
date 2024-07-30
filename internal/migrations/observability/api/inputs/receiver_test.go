package inputs

import (
	. "github.com/onsi/ginkgo"

	. "github.com/onsi/gomega"
	logging "github.com/openshift/cluster-logging-operator/api/logging/v1"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
)

var _ = Describe("#MigrateReceiverInputs", func() {
	It("should map logging http receiver to observability http receiver", func() {
		loggingReceiverSpec := logging.ReceiverSpec{
			Type: logging.ReceiverTypeHttp,
			ReceiverTypeSpec: &logging.ReceiverTypeSpec{
				HTTP: &logging.HTTPReceiver{
					Port:   9000,
					Format: logging.FormatKubeAPIAudit,
				},
			},
		}
		expObsReceiverSpec := &obs.ReceiverSpec{
			Type: obs.ReceiverTypeHTTP,
			Port: 9000,
			HTTP: &obs.HTTPReceiver{
				Format: obs.HTTPReceiverFormatKubeAPIAudit,
			},
		}

		Expect(MapReceiverInput(&loggingReceiverSpec)).To(Equal(expObsReceiverSpec))
	})

	It("should map logging syslog receiver to observability syslog receiver", func() {
		loggingReceiverSpec := logging.ReceiverSpec{
			Type: logging.ReceiverTypeSyslog,
			ReceiverTypeSpec: &logging.ReceiverTypeSpec{
				Syslog: &logging.SyslogReceiver{
					Port: 9000,
				},
			},
		}
		expObsReceiverSpec := &obs.ReceiverSpec{
			Type: obs.ReceiverTypeSyslog,
			Port: 9000,
		}

		Expect(MapReceiverInput(&loggingReceiverSpec)).To(Equal(expObsReceiverSpec))
	})
})
