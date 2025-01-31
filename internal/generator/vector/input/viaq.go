package input

import (
	"fmt"
	"regexp"
	"strings"

	logging "github.com/openshift/cluster-logging-operator/api/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/factory"
	"github.com/openshift/cluster-logging-operator/internal/generator/framework"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/elements"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/normalize"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/source"
	"github.com/openshift/cluster-logging-operator/internal/utils/sets"
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
	excludeExtensions = []string{"gz", "tmp", "log.*"}
	infraNamespaces   = []string{"default", "openshift", "openshift-*", "kube", "kube-*"}
	infraNSRegex      = regexp.MustCompile(
		`(?P<default>^default$)|(?P<openshift_dash>^openshift-.+$)|(?P<kube_dash>^kube-.+$)|(?P<openshift>^openshift$)|(?P<kube>^kube$)|(?P<openshift_all>^openshift\*$)|(?P<kube_all>^kube\*$)`)

	infraExcludes = source.NewContainerPathGlobBuilder().
			AddNamespaces(infraNamespaces...).AddExtensions(excludeExtensions...).
			Build()
)

// NewViaQ creates an input adapter to generate config for ViaQ sources to collect logs excluding the
// collector container logs from the namespace where the collector is deployed
func NewViaQ(input logging.InputSpec, collectorNS string, resNames *factory.ForwarderResourceNames, op framework.Options) ([]framework.Element, []string) {
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
			ib := source.NewContainerPathGlobBuilder()
			eb := source.NewContainerPathGlobBuilder()
			appIncludes := []string{}
			// Migrate existing namespaces
			if len(input.Application.Namespaces) > 0 {
				for _, ns := range input.Application.Namespaces {
					ncs := source.NamespaceContainer{
						Namespace: ns,
					}
					ib.AddCombined(ncs)
					appIncludes = append(appIncludes, ncs.Namespace)
				}
			}
			if len(input.Application.Includes) > 0 {
				for _, in := range input.Application.Includes {
					ncs := source.NamespaceContainer{
						Namespace: in.Namespace,
						Container: in.Container,
					}
					ib.AddCombined(ncs)
					appIncludes = append(appIncludes, ncs.Namespace)
				}
			}
			// Need to remove any of the default excluded infra namespaces if they are part of the includes
			excludesList := pruneInfraNS(appIncludes)
			for _, ns := range excludesList {
				ncs := source.NamespaceContainer{
					Namespace: ns,
				}
				eb.AddCombined(ncs)
			}
			if len(input.Application.Excludes) > 0 {
				for _, ex := range input.Application.Excludes {
					ncs := source.NamespaceContainer{
						Namespace: ex.Namespace,
						Container: ex.Container,
					}
					eb.AddCombined(ncs)
				}
			}
			eb.AddExtensions(excludeExtensions...)
			includes := ib.Build()
			excludes := eb.Build(infraNamespaces...)
			els, ids = NewViaqContainerSource(input, collectorNS, includes, excludes)
		} else if input.Infrastructure != nil {
			sources := sets.NewString(input.Infrastructure.Sources...)
			if sources.Has(logging.InfrastructureSourceContainer) {
				infraIncludes := source.NewContainerPathGlobBuilder().AddNamespaces(infraNamespaces...).Build()
				cels, cids := NewViaqContainerSource(input, collectorNS, infraIncludes, loggingExcludes)
				els = append(els, cels...)
				ids = append(ids, cids...)
			}
			if sources.Has(logging.InfrastructureSourceNode) {
				jels, jids := NewViaqJournalSource(input)
				els = append(els, jels...)
				ids = append(ids, jids...)
			}
		} else if input.Audit != nil {
			sources := sets.NewString(input.Audit.Sources...)
			if sources.Has(logging.AuditSourceAuditd) {
				cels, cids := NewAuditAuditdSource(input, op)
				els = append(els, cels...)
				ids = append(ids, cids...)
			}
			if sources.Has(logging.AuditSourceKube) {
				cels, cids := NewK8sAuditSource(input, op)
				els = append(els, cels...)
				ids = append(ids, cids...)
			}
			if sources.Has(logging.AuditSourceOpenShift) {
				cels, cids := NewOpenshiftAuditSource(input, op)
				els = append(els, cels...)
				ids = append(ids, cids...)
			}
			if sources.Has(logging.AuditSourceOVN) {
				cels, cids := NewOVNAuditSource(input, op)
				els = append(els, cels...)
				ids = append(ids, cids...)
			}
		} else if input.Receiver != nil {
			els, ids = NewViaqReceiverSource(input, resNames, op)
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
	case spec.Infrastructure != nil || spec.Receiver.IsSyslogReceiver():
		logType = logging.InputNameInfrastructure
	case spec.Audit != nil || spec.Receiver.IsAuditHttpReceiver():
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

		switch logType {
		case logging.InputNameApplication:
			remap.VRL = fmt.Sprintf(`
# If namespace is infra, label log_type as infra
if match_any(string!(.kubernetes.namespace_name), [r'^default$', r'^openshift(-.+)?$', r'^kube(-.+)?$']) {
	.log_type = %q
    ._internal.log_source = "node"
} else {
	.log_type = %q
    ._internal.log_source = "container"
}`, logging.InputNameInfrastructure,
				logType)
		case logging.InputNameAudit:
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
	var selector *logging.LabelSelector
	if spec.Application != nil {
		selector = spec.Application.Selector
	}
	el := []framework.Element{
		source.KubernetesLogs{
			ComponentID:        base,
			Desc:               "Logs from containers (including openshift containers)",
			IncludePaths:       includes,
			ExcludePaths:       excludes,
			ExtraLabelSelector: source.LabelSelectorFrom(selector),
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

// pruneInfraNS returns a pruned infra namespace list depending on which infra namespaces were included
// since the exclusion list includes all infra namespaces by default
// Example:
// Include: ["openshift-logging"]
// Default Exclude: ["default", "openshift", "openshift-*", "kube", "kube-*"]
// Final infra namespaces in Exclude: ["default", "openshift", "kube", "kube-*"]
func pruneInfraNS(includes []string) []string {
	foundInfraNamespaces := make(map[string]string)
	for _, ns := range includes {
		matches := infraNSRegex.FindStringSubmatch(ns)
		if matches != nil {
			for i, name := range infraNSRegex.SubexpNames() {
				if i != 0 && matches[i] != "" {
					foundInfraNamespaces[name] = matches[i]
				}
			}
		}
	}

	infraNSSet := sets.NewString(infraNamespaces...)
	// Remove infra namespace depending on the named capture group
	for k := range foundInfraNamespaces {
		switch k {
		case "default":
			infraNSSet.Remove("default")
		case "openshift":
			infraNSSet.Remove("openshift")
		case "openshift_dash":
			infraNSSet.Remove("openshift-*")
		case "openshift_all":
			infraNSSet.Remove("openshift")
			infraNSSet.Remove("openshift-*")
		case "kube":
			infraNSSet.Remove("kube")
		case "kube_dash":
			infraNSSet.Remove("kube-*")
		case "kube_all":
			infraNSSet.Remove("kube")
			infraNSSet.Remove("kube-*")
		}
	}
	return infraNSSet.List()
}
