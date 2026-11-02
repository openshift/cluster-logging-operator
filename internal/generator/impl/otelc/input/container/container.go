package container

import (
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/generator/forwarder/input/container"
	"github.com/openshift/cluster-logging-operator/internal/generator/helpers"
	"github.com/openshift/cluster-logging-operator/internal/generator/impl/otelc/api/receivers"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// AllContainerLogsGlobs match every CRI-O container log file including
// rolled-over rotations:
// /var/log/pods/<namespace>_<pod>_<uid>/<container>/<restart>.log
// /var/log/pods/<namespace>_<pod>_<uid>/<container>/<restart>.log.1, .2, ...
var AllContainerLogsGlobs = []string{
	"/var/log/pods/*/*/*.log",
	"/var/log/pods/*/*/*.log.*",
}

// New generates a FileLog receiver for Kubernetes container logs based on an InputSpec.
// It returns the receiver ID and the configured FileLog receiver.
//
// The function handles:
// - Building include/exclude path globs based on the InputSpec
// - Configuring CRI-O log parsing operators
// - Setting max message size limits
// - Applying label selectors
func New(spec obs.InputSpec) (receiver *receivers.FileLog) {
	var includes, excludes []string
	var selector *metav1.LabelSelector
	var maxMsgSize int64

	// Build include/exclude paths based on input type
	switch spec.Type {
	case obs.InputTypeApplication:
		ib := helpers.NewContainerPathGlobBuilder()
		eb := helpers.NewContainerPathGlobBuilder()
		appIncludes := []string{}

		if spec.Application != nil {
			// Process includes
			if len(spec.Application.Includes) > 0 {
				for _, in := range spec.Application.Includes {
					ncs := helpers.NamespaceContainer{
						Namespace: in.Namespace,
						Container: in.Container,
					}
					ib.AddCombined(ncs)
					appIncludes = append(appIncludes, ncs.Namespace)
				}
			} else {
				ib.AddContainers("*")
			}

			// Remove default excluded infra namespaces if they are part of the includes
			excludesList := container.PruneInfraNS(appIncludes)
			for _, ns := range excludesList {
				ncs := helpers.NamespaceContainer{
					Namespace: ns,
				}
				eb.AddCombined(ncs)
			}

			// Process excludes
			if len(spec.Application.Excludes) > 0 {
				for _, ex := range spec.Application.Excludes {
					ncs := helpers.NamespaceContainer{
						Namespace: ex.Namespace,
						Container: ex.Container,
					}
					eb.AddCombined(ncs)
				}
			}

			selector = spec.Application.Selector

			// Get max message size
			if spec.Application.Tuning != nil && spec.Application.Tuning.MaxMessageSize != nil {
				if size, ok := spec.Application.Tuning.MaxMessageSize.AsInt64(); ok {
					maxMsgSize = size
				}
			}
		} else {
			// Default excludes for infra namespaces
			excludesList := container.PruneInfraNS(appIncludes)
			for _, ns := range excludesList {
				ncs := helpers.NamespaceContainer{
					Namespace: ns,
				}
				eb.AddCombined(ncs)
			}
		}

		eb.AddExtensions(container.ExcludeExtensions...)
		includes = ib.Build()
		excludes = eb.Build(container.InfraNamespaces...)
		// An empty include list means every container log: the vector
		// kubernetes_logs source treats empty includes as match-all, but the
		// filelog receiver matches nothing, so default the glob explicitly
		if len(includes) == 0 {
			includes = AllContainerLogsGlobs
		}

	case obs.InputTypeInfrastructure:
		// Infrastructure logs - include infra namespaces
		includes = helpers.NewContainerPathGlobBuilder().AddNamespaces(container.InfraNamespaces...).Build()
		excludes = container.LoggingExcludes

		// Get max message size if configured
		if spec.Infrastructure != nil && spec.Infrastructure.Tuning != nil &&
			spec.Infrastructure.Tuning.Container != nil &&
			spec.Infrastructure.Tuning.Container.MaxMessageSize != nil {
			if size, ok := spec.Infrastructure.Tuning.Container.MaxMessageSize.AsInt64(); ok {
				maxMsgSize = size
			}
		}
	}

	// Default max message size if not set (3MB, matching vector config)
	if maxMsgSize == 0 {
		maxMsgSize = container.MaxReadBytes
	}

	// Create base receiver ID (format: input_<name>_container)
	base := helpers.MakeInputID(spec.Name, obs.InfrastructureSourceContainer.String())

	// Create FileLog receiver
	receiver = receivers.NewFileLog(base, includes...)
	receiver.Exclude = excludes
	receiver.StartAt = "end"
	receiver.IncludeFilePath = true

	// Set max log size (converted from bytes)
	maxLogSizeQuantity := resource.NewQuantity(maxMsgSize, resource.BinarySI)
	receiver.MaxLogSize, _ = maxLogSizeQuantity.AsInt64()

	// Add CRI-O parsing operators. The cluster UID and node name are resolved
	// by the collector at runtime from its environment (the same env vars the
	// vector config expands): OPENSHIFT_CLUSTER_ID and NODE_NAME
	receiver.Operators = NewOperators(
		"${env:OPENSHIFT_CLUSTER_ID}",
		"${env:NODE_NAME}",
		obs.InfrastructureSourceContainer.String(),
		spec.Type.String(),
	)

	// TODO: Apply label selector filtering if needed
	// Note: OpenTelemetry Collector doesn't have built-in label selector support like vector
	// This would need to be handled at a higher level (e.g., via processor or different approach)
	_ = selector

	return receiver
}
