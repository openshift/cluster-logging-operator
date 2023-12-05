package input

import (
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	generator "github.com/openshift/cluster-logging-operator/internal/generator/framework"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
	vector "github.com/openshift/cluster-logging-operator/internal/generator/vector/normalize"
	sources "github.com/openshift/cluster-logging-operator/internal/generator/vector/source"
)

func NewAuditSources(input logging.InputSpec, op generator.Options) ([]generator.Element, []string) {

	hostID := helpers.MakeInputID("audit", "host")
	el := []generator.Element{
		sources.NewHostAuditLog(hostID),
	}
	hostViaqID := helpers.MakeID(hostID, "viaq")
	el = append(el, vector.NormalizeHostAuditLogs(hostID, hostViaqID)...)

	kubeID := helpers.MakeInputID("audit", "kube")
	el = append(el, sources.NewK8sAuditLog(kubeID))
	kubeViaQID := helpers.MakeID(kubeID, "viaq")
	el = append(el, vector.NormalizeK8sAuditLogs(kubeID, kubeViaQID)...)

	openshiftID := helpers.MakeInputID("audit", "openshift")
	el = append(el, sources.NewOpenshiftAuditLog(openshiftID))
	openshiftViaQID := helpers.MakeID(openshiftID, "viaq")
	el = append(el, vector.NormalizeOpenshiftAuditLogs(openshiftID, openshiftViaQID)...)

	ovnID := helpers.MakeInputID("audit", "ovn")
	el = append(el, sources.NewOVNAuditLog(ovnID))
	ovnViaQID := helpers.MakeID(ovnID, "viaq")
	el = append(el, vector.NormalizeOVNAuditLogs(ovnID, ovnViaQID)...)

	return el, []string{hostViaqID, kubeViaQID, openshiftViaQID, ovnViaQID}
}
