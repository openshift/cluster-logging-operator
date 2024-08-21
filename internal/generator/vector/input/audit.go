package input

import (
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	generator "github.com/openshift/cluster-logging-operator/internal/generator/framework"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/filter/openshift/viaq"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
	sources "github.com/openshift/cluster-logging-operator/internal/generator/vector/source"
)

func NewAuditAuditdSource(input obs.InputSpec, op generator.Options) ([]generator.Element, []string) {
	hostID := helpers.MakeInputID(input.Name, "host")
	metaID := helpers.MakeID(hostID, "meta")
	el := []generator.Element{
		sources.NewHostAuditLog(hostID),
		NewInternalNormalization(metaID, obs.AuditSourceAuditd, obs.InputTypeAudit, hostID, viaq.ParseHostAuditLogs),
	}
	return el, []string{metaID}
}

func NewK8sAuditSource(input obs.InputSpec, op generator.Options) ([]generator.Element, []string) {
	id := helpers.MakeInputID(input.Name, "kube")
	metaID := helpers.MakeID(id, "meta")
	el := []generator.Element{
		sources.NewK8sAuditLog(id),
		NewInternalNormalization(metaID, obs.AuditSourceKube, obs.InputTypeAudit, id, ParseStructured),
	}
	return el, []string{metaID}
}

func NewOpenshiftAuditSource(input obs.InputSpec, op generator.Options) ([]generator.Element, []string) {
	id := helpers.MakeInputID(input.Name, "openshift")
	metaID := helpers.MakeID(id, "meta")
	el := []generator.Element{
		sources.NewOpenshiftAuditLog(id),
		NewInternalNormalization(metaID, obs.AuditSourceOpenShift, obs.InputTypeAudit, id, ParseStructured),
	}
	return el, []string{metaID}
}

func NewOVNAuditSource(input obs.InputSpec, op generator.Options) ([]generator.Element, []string) {
	id := helpers.MakeInputID(input.Name, "ovn")
	metaID := helpers.MakeID(id, "meta")
	el := []generator.Element{
		sources.NewOVNAuditLog(id),
		NewInternalNormalization(metaID, obs.AuditSourceOVN, obs.InputTypeAudit, id, ParseStructured),
	}
	return el, []string{metaID}
}
