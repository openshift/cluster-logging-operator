package input

import (
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	generator "github.com/openshift/cluster-logging-operator/internal/generator/framework"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/filter/openshift/viaq/v1"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
	sources "github.com/openshift/cluster-logging-operator/internal/generator/vector/source"
)

func NewAuditAuditdSource(input obs.InputSpec, op generator.Options) ([]generator.Element, []string) {
	hostID := helpers.MakeInputID(input.Name, "host")
	metaID := helpers.MakeID(hostID, "meta")
	el := []generator.Element{
		sources.NewHostAuditLog(hostID),
		NewInternalNormalization(metaID, obs.AuditSourceAuditd, obs.InputTypeAudit, hostID, v1.ParseHostAuditLogs),
	}
	return el, []string{metaID}
}

func NewK8sAuditSource(input obs.InputSpec, op generator.Options) ([]generator.Element, []string) {
	id := helpers.MakeInputID(input.Name, "kube")
	metaID := helpers.MakeID(id, "meta")
	el := []generator.Element{
		sources.NewK8sAuditLog(id),
		NewAuditInternalNormalization(metaID, obs.AuditSourceKube, id, true),
	}
	return el, []string{metaID}
}

func NewOpenshiftAuditSource(input obs.InputSpec, op generator.Options) ([]generator.Element, []string) {
	id := helpers.MakeInputID(input.Name, "openshift")
	metaID := helpers.MakeID(id, "meta")
	el := []generator.Element{
		sources.NewOpenshiftAuditLog(id),
		NewAuditInternalNormalization(metaID, obs.AuditSourceOpenShift, id, true),
	}
	return el, []string{metaID}
}

func NewOVNAuditSource(input obs.InputSpec, op generator.Options) ([]generator.Element, []string) {
	id := helpers.MakeInputID(input.Name, "ovn")
	metaID := helpers.MakeID(id, "meta")
	el := []generator.Element{
		sources.NewOVNAuditLog(id),
		NewInternalNormalization(metaID, obs.AuditSourceOVN, obs.InputTypeAudit, id),
	}
	return el, []string{metaID}
}
