package inputs

import (
	. "github.com/onsi/ginkgo"

	. "github.com/onsi/gomega"
	logging "github.com/openshift/cluster-logging-operator/api/logging/v1"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
)

var _ = Describe("#MigrateAuditInput", func() {
	It("should map logging.Audit to observability.Audit", func() {
		loggingAudit := logging.Audit{
			Sources: []string{"foo", "bar", "baz"},
		}

		expObsAudit := &obs.Audit{
			Sources: []obs.AuditSource{"foo", "bar", "baz"},
		}

		Expect(MapAuditInput(&loggingAudit)).To(Equal(expObsAudit))
	})
})
