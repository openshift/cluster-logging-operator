package input

import (
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	generator "github.com/openshift/cluster-logging-operator/internal/generator/framework"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
	vector "github.com/openshift/cluster-logging-operator/internal/generator/vector/normalize"
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
	el := []generator.Element{
		sources.NewHostAuditLog(hostID),
	}
	hostViaqID := helpers.MakeID(hostID, "viaq")
	el = append(el, vector.NormalizeHostAuditLogs(hostID, hostViaqID)...)
	return el, []string{hostViaqID}
}

func NewK8sAuditSource(input logging.InputSpec, op generator.Options) ([]generator.Element, []string) {
	kubeID := helpers.MakeInputID(input.Name, "kube")
	el := []generator.Element{sources.NewK8sAuditLog(kubeID)}
	kubeViaQID := helpers.MakeID(kubeID, "viaq")
	el = append(el, vector.NormalizeK8sAuditLogs(kubeID, kubeViaQID)...)
	return el, []string{kubeViaQID}
}

func NewOpenshiftAuditSource(input logging.InputSpec, op generator.Options) ([]generator.Element, []string) {
	openshiftID := helpers.MakeInputID(input.Name, "openshift")
	el := []generator.Element{sources.NewOpenshiftAuditLog(openshiftID)}
	openshiftViaQID := helpers.MakeID(openshiftID, "viaq")
	el = append(el, vector.NormalizeOpenshiftAuditLogs(openshiftID, openshiftViaQID)...)
	return el, []string{openshiftViaQID}
}

func NewOVNAuditSource(input logging.InputSpec, op generator.Options) ([]generator.Element, []string) {
	ovnID := helpers.MakeInputID(input.Name, "ovn")
	el := []generator.Element{sources.NewOVNAuditLog(ovnID)}
	ovnViaQID := helpers.MakeID(ovnID, "viaq")
	el = append(el, vector.NormalizeOVNAuditLogs(ovnID, ovnViaQID)...)
	return el, []string{ovnViaQID}
}
