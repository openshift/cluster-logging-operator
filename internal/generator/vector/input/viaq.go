package input

import (
	"fmt"
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/generator/framework"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/elements"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/normalize"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/source"
	"strings"
)

const (
	nsPodPathFmt       = "%s_%s-*/*/*.log"
	nsContainerPathFmt = "%s_*/%s/*.log"
)

var (
	loggingExcludes = source.NewContainerPathGlobBuilder().
			AddOther(
			fmt.Sprintf(nsPodPathFmt, constants.OpenshiftNS, constants.LogfilesmetricexporterName),
			fmt.Sprintf(nsPodPathFmt, constants.OpenshiftNS, constants.ElasticsearchName),
			fmt.Sprintf(nsPodPathFmt, constants.OpenshiftNS, constants.KibanaName),
		).AddOther(
		fmt.Sprintf(nsContainerPathFmt, constants.OpenshiftNS, "loki*"),
		fmt.Sprintf(nsContainerPathFmt, constants.OpenshiftNS, "gateway"),
		fmt.Sprintf(nsContainerPathFmt, constants.OpenshiftNS, "opa"),
	).AddExtensions(excludeExtensions...).
		Build()
	excludeExtensions = []string{"gz", "tmp"}
	infraNamespaces   = []string{"default", "openshift*", "kube*"}
	infraExcludes     = source.NewContainerPathGlobBuilder().
				AddNamespaces(infraNamespaces...).AddExtensions(excludeExtensions...).
				Build()
)

// NewViaQ creates an input adapter to generate config for ViaQ sources to collect logs excluding the
// collector container logs from the namespace where the collector is deployed
func NewViaQ(input logging.InputSpec, collectorNS string, op framework.Options) ([]framework.Element, []string) {
	els := []framework.Element{}
	ids := []string{}
	switch {
	case input.Name == logging.InputNameApplication:
		els, ids = NewViaqContainerSource(input, collectorNS, "", infraExcludes)
	case input.Name == logging.InputNameInfrastructure:
		infraIncludes := source.NewContainerPathGlobBuilder().AddNamespaces(infraNamespaces...).Build()
		cels, cids := NewViaqContainerSource(input, collectorNS, infraIncludes, loggingExcludes)
		els = append(els, cels...)
		ids = append(ids, cids...)
		jels, jids := NewViaqJournalSource(input)
		els = append(els, jels...)
		ids = append(ids, jids...)
	case input.Name == logging.InputNameAudit:
		els, ids = NewAuditSources(input, op)
	default:
		if input.Application != nil {
			includes := source.NewContainerPathGlobBuilder().
				AddNamespaces(input.Application.Namespaces...).
				Build()
			excludes := source.NewContainerPathGlobBuilder().AddExtensions(excludeExtensions...).Build()
			els, ids = NewViaqContainerSource(input, collectorNS, includes, excludes)
		} else if input.Infrastructure != nil {
			//TODO: modify when input api is added
			if true { //source == "container" {
				cels, cids := NewViaqContainerSource(input, collectorNS, "", "")
				els = append(els, cels...)
				ids = append(ids, cids...)
			}
		} else if input.Receiver != nil {
			els, ids = NewViaqReceiverSource(input, op)
		}
	}
	els, ids = addLogType(input, els, ids)
	return els, ids
}

func addLogType(spec logging.InputSpec, els []framework.Element, ids []string) ([]framework.Element, []string) {
	logType := ""
	switch {
	case spec.Application != nil:
		logType = logging.InputNameApplication
	case spec.Infrastructure != nil || logging.IsSyslogReceiver(&spec):
		logType = logging.InputNameInfrastructure
	case spec.Audit != nil || logging.IsAuditHttpReceiver(&spec):
		logType = logging.InputNameAudit
	}

	if logType != "" {
		id := helpers.MakeInputID(spec.Name, "viaq", "logtype")
		remap := elements.Remap{
			Desc:        `Set log_type`,
			ComponentID: id,
			Inputs:      helpers.MakeInputs(ids...),
			VRL:         fmt.Sprintf(".log_type = %q", logType),
		}
		if logType == logging.InputNameAudit {
			remap.VRL = strings.Join(helpers.TrimSpaces([]string{
				remap.VRL,
				normalize.FixHostname,
				normalize.FixTimestampField,
			}), "\n")
		}
		els = append(els, remap)
		return els, []string{id}
	}
	return els, ids
}

// NewViaqContainerSource generates config elements and the id reference of this input and normalizes
// the tomlContent to VIAQ api
func NewViaqContainerSource(spec logging.InputSpec, namespace, includes, excludes string) ([]framework.Element, []string) {
	base := helpers.MakeInputID(spec.Name, "container")
	el := []framework.Element{
		source.KubernetesLogs{
			ComponentID:  base,
			Desc:         "Logs from containers (including openshift containers)",
			IncludePaths: includes,
			ExcludePaths: excludes,
		},
	}
	inputID := base
	if spec.HasPolicy() {
		throttleID := helpers.MakeID(base, "throttle")
		inputID = throttleID
		el = append(el, AddThrottleToInput(throttleID, base, spec)...)
	}
	id := helpers.MakeID(base, "viaq")
	el = append(el, normalize.NormalizeContainerLogs(inputID, id)...)

	return el, []string{id}
}
