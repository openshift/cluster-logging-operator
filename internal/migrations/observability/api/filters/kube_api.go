package filters

import (
	logging "github.com/openshift/cluster-logging-operator/api/logging/v1"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
)

func MapKubeApiAuditFilter(loggingKubeApiAudit logging.KubeAPIAudit) *obs.KubeAPIAudit {
	return &obs.KubeAPIAudit{
		Rules:             loggingKubeApiAudit.Rules,
		OmitStages:        loggingKubeApiAudit.OmitStages,
		OmitResponseCodes: loggingKubeApiAudit.OmitResponseCodes,
	}
}
