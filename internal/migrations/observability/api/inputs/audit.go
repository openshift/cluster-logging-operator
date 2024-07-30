package inputs

import (
	logging "github.com/openshift/cluster-logging-operator/api/logging/v1"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
)

func MapAuditInput(loggingAudit *logging.Audit) *obs.Audit {
	obsAudit := &obs.Audit{
		Sources: []obs.AuditSource{},
	}
	for _, source := range loggingAudit.Sources {
		obsAudit.Sources = append(obsAudit.Sources, obs.AuditSource(source))
	}
	return obsAudit
}
