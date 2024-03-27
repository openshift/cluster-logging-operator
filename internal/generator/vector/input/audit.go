package input

import (
	logging "github.com/openshift/cluster-logging-operator/api/logging/v1"
	generator "github.com/openshift/cluster-logging-operator/internal/generator/framework"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
	sources "github.com/openshift/cluster-logging-operator/internal/generator/vector/source"
)

func NewAuditSources(input logging.InputSpec, op generator.Options) ([]generator.Element, []string) {

	adapter := &Input{}
	adapter.Add(NewAuditAuditdSource(input, op)).
		Add(NewK8sAuditSource(input, op)).
		Add(NewOpenshiftAuditSource(input, op)).
		Add(NewOVNAuditSource(input, op))

	return adapter.Elements(), adapter.InputIDs()
}

func NewAuditAuditdSource(input logging.InputSpec, op generator.Options) ([]generator.Element, []string) {
	hostID := helpers.MakeInputID(input.Name, "host")
	metaID := helpers.MakeID(hostID, "meta")
	el := []generator.Element{
		sources.NewHostAuditLog(hostID),
		NewLogSourceAndType(metaID, logging.AuditSourceAuditd, logging.InputNameAudit, hostID),
	}
	return el, []string{metaID}
}

func NewK8sAuditSource(input logging.InputSpec, op generator.Options) ([]generator.Element, []string) {
	id := helpers.MakeInputID(input.Name, "kube")
	metaID := helpers.MakeID(id, "meta")
	el := []generator.Element{
		sources.NewK8sAuditLog(id),
		NewLogSourceAndType(metaID, logging.AuditSourceKube, logging.InputNameAudit, id),
	}
	return el, []string{metaID}
}

func NewOpenshiftAuditSource(input logging.InputSpec, op generator.Options) ([]generator.Element, []string) {
	id := helpers.MakeInputID(input.Name, "openshift")
	metaID := helpers.MakeID(id, "meta")
	el := []generator.Element{
		sources.NewOpenshiftAuditLog(id),
		NewLogSourceAndType(metaID, logging.AuditSourceOpenShift, logging.InputNameAudit, id),
	}
	return el, []string{metaID}
}

func NewOVNAuditSource(input logging.InputSpec, op generator.Options) ([]generator.Element, []string) {
	id := helpers.MakeInputID(input.Name, "ovn")
	metaID := helpers.MakeID(id, "meta")
	el := []generator.Element{
		sources.NewOVNAuditLog(id),
		NewLogSourceAndType(metaID, logging.AuditSourceOVN, logging.InputNameAudit, id),
	}
	return el, []string{metaID}
}
