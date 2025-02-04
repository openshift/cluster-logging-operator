package input

import (
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	internalobs "github.com/openshift/cluster-logging-operator/internal/api/observability"
	"github.com/openshift/cluster-logging-operator/internal/factory"
	"github.com/openshift/cluster-logging-operator/internal/generator/framework"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/source"
	"k8s.io/utils/set"
)

// NewSource creates an input adapter to generate config for ViaQ sources to collect logs excluding the
// collector container logs from the namespace where the collector is deployed
func NewSource(input obs.InputSpec, collectorNS string, resNames factory.ForwarderResourceNames, secrets internalobs.Secrets, op framework.Options) ([]framework.Element, []string) {
	els := []framework.Element{}
	ids := []string{}
	switch input.Type {
	case obs.InputTypeApplication:
		ib := source.NewContainerPathGlobBuilder()
		eb := source.NewContainerPathGlobBuilder()
		appIncludes := []string{}
		if input.Application != nil {
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
		} else {
			// Need to remove any of the default excluded infra namespaces if they are part of the includes
			excludesList := pruneInfraNS(appIncludes)
			for _, ns := range excludesList {
				ncs := source.NamespaceContainer{
					Namespace: ns,
				}
				eb.AddCombined(ncs)
			}
		}
		eb.AddExtensions(excludeExtensions...)
		includes := ib.Build()
		excludes := eb.Build(infraNamespaces...)
		return NewContainerSource(input, collectorNS, includes, excludes, obs.InputTypeApplication, obs.InfrastructureSourceContainer)
	case obs.InputTypeInfrastructure:
		sources := set.Set[obs.InfrastructureSource]{}
		if input.Infrastructure == nil {
			sources.Insert(obs.InfrastructureSources...)
		} else {
			sources = set.New(input.Infrastructure.Sources...)
			if sources.Len() == 0 {
				sources.Insert(obs.InfrastructureSources...)
			}
		}
		if sources.Has(obs.InfrastructureSourceContainer) {
			infraIncludes := source.NewContainerPathGlobBuilder().AddNamespaces(infraNamespaces...).Build()
			cels, cids := NewContainerSource(input, collectorNS, infraIncludes, loggingExcludes, obs.InputTypeInfrastructure, obs.InfrastructureSourceContainer)
			els = append(els, cels...)
			ids = append(ids, cids...)
		}
		if sources.Has(obs.InfrastructureSourceNode) {
			jels, jids := NewJournalSource(input)
			els = append(els, jels...)
			ids = append(ids, jids...)
		}
		return els, ids
	case obs.InputTypeAudit:
		sources := set.Set[obs.AuditSource]{}
		if input.Audit == nil || len(input.Audit.Sources) == 0 {
			sources.Insert(obs.AuditSources...)
		} else {
			sources = set.New(input.Audit.Sources...)
			if sources.Len() == 0 {
				sources.Insert(obs.AuditSources...)
			}
		}
		if sources.Has(obs.AuditSourceAuditd) {
			cels, cids := NewAuditAuditdSource(input, op)
			els = append(els, cels...)
			ids = append(ids, cids...)
		}
		if sources.Has(obs.AuditSourceKube) {
			cels, cids := NewK8sAuditSource(input, op)
			els = append(els, cels...)
			ids = append(ids, cids...)
		}
		if sources.Has(obs.AuditSourceOpenShift) {
			cels, cids := NewOpenshiftAuditSource(input, op)
			els = append(els, cels...)
			ids = append(ids, cids...)
		}
		if sources.Has(obs.AuditSourceOVN) {
			cels, cids := NewOVNAuditSource(input, op)
			els = append(els, cels...)
			ids = append(ids, cids...)
		}
	case obs.InputTypeReceiver:
		return NewViaqReceiverSource(input, resNames, secrets, op)
	}
	return els, ids
}
