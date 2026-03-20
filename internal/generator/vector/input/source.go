package input

import (
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	internalobs "github.com/openshift/cluster-logging-operator/internal/api/observability"
	"github.com/openshift/cluster-logging-operator/internal/factory"
	"github.com/openshift/cluster-logging-operator/internal/generator/framework"
	"github.com/openshift/cluster-logging-operator/internal/generator/helpers"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/adapters"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	"k8s.io/utils/set"
)

// NewSource creates an input adapter to generate config for ViaQ sources to collect logs excluding the
// collector container logs from the namespace where the collector is deployed
func NewSource(input *adapters.Input, resNames factory.ForwarderResourceNames, secrets internalobs.Secrets, op utils.Options) (inputSources api.Sources, tfs api.Transforms) {
	framework.SetTLSProfileOptionsFrom(op, input)
	inputSources = api.Sources{}
	tfs = api.Transforms{}
	switch input.Type {
	case obs.InputTypeApplication:
		ib := helpers.NewContainerPathGlobBuilder()
		eb := helpers.NewContainerPathGlobBuilder()
		appIncludes := []string{}
		if input.Application != nil {
			if len(input.Application.Includes) > 0 {
				for _, in := range input.Application.Includes {
					ncs := helpers.NamespaceContainer{
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
				ncs := helpers.NamespaceContainer{
					Namespace: ns,
				}
				eb.AddCombined(ncs)
			}
			if len(input.Application.Excludes) > 0 {
				for _, ex := range input.Application.Excludes {
					ncs := helpers.NamespaceContainer{
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
				ncs := helpers.NamespaceContainer{
					Namespace: ns,
				}
				eb.AddCombined(ncs)
			}
		}
		eb.AddExtensions(excludeExtensions...)
		includes := ib.Build()
		excludes := eb.Build(infraNamespaces...)
		sourceId, source, ctfs := NewContainerSource(input, includes, excludes, obs.InputTypeApplication, obs.InfrastructureSourceContainer)
		inputSources.Add(sourceId, source)
		return inputSources, ctfs
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
			infraIncludes := helpers.NewContainerPathGlobBuilder().AddNamespaces(infraNamespaces...).Build()
			sourceId, source, ctfs := NewContainerSource(input, infraIncludes, loggingExcludes, obs.InputTypeInfrastructure, obs.InfrastructureSourceContainer)
			inputSources.Add(sourceId, source)
			tfs.Merge(ctfs)
		}
		if sources.Has(obs.InfrastructureSourceNode) {
			sourceId, source, ctfs := NewJournalInput(input)
			inputSources.Add(sourceId, source)
			tfs.Merge(ctfs)
		}
		return inputSources, tfs
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
			sourceId, source, ctfs := NewAuditAuditdSource(input)
			inputSources.Add(sourceId, source)
			tfs.Merge(ctfs)
		}
		if sources.Has(obs.AuditSourceKube) {
			sourceId, source, ctfs := NewK8sAuditSource(input)
			inputSources.Add(sourceId, source)
			tfs.Merge(ctfs)
		}
		if sources.Has(obs.AuditSourceOpenShift) {
			sourceId, source, ctfs := NewOpenshiftAuditSource(input)
			inputSources.Add(sourceId, source)
			tfs.Merge(ctfs)
		}
		if sources.Has(obs.AuditSourceOVN) {
			sourceId, source, ctfs := NewOVNAuditSource(input)
			inputSources.Add(sourceId, source)
			tfs.Merge(ctfs)
		}
	case obs.InputTypeReceiver:
		sourceId, source, ctfs := NewViaqReceiverSource(input, resNames, secrets, op)
		inputSources.Add(sourceId, source)
		tfs.Merge(ctfs)
	}
	return inputSources, tfs
}
