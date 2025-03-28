package observability

import (
	"k8s.io/utils/set"

	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/utils/sets"
)

var (
	ReservedInputTypes = sets.NewString(
		string(obs.InputTypeApplication),
		string(obs.InputTypeAudit),
		string(obs.InputTypeInfrastructure),
	)

	ReservedApplicationSources    = sets.NewString(string(obs.ApplicationSourceContainer))
	ReservedInfrastructureSources = sets.NewString()
	ReservedAuditSources          = sets.NewString()
)

func init() {
	for _, i := range obs.InfrastructureSources {
		ReservedInfrastructureSources.Insert(string(i))
	}
	for _, i := range obs.AuditSources {
		ReservedAuditSources.Insert(string(i))
	}
}

func MaxRecordsPerSecond(input obs.InputSpec) (int64, bool) {
	if input.Application != nil &&
		input.Application.Tuning != nil &&
		input.Application.Tuning.RateLimitPerContainer != nil {
		return Threshold(input.Application.Tuning.RateLimitPerContainer)
	}
	return 0, false
}

func Threshold(ls *obs.LimitSpec) (int64, bool) {
	if ls == nil {
		return 0, false
	}
	return ls.MaxRecordsPerSecond, true
}

type Inputs []obs.InputSpec

// Names returns a slice of input names
func (inputs Inputs) Names() (names []string) {
	for _, i := range inputs {
		names = append(names, i.Name)
	}
	return names
}

// InputTypes returns a unique set of input types
func (inputs Inputs) InputTypes() []obs.InputType {
	types := set.New[obs.InputType]()
	for _, i := range inputs {
		types.Insert(i.Type)
	}
	return types.UnsortedList()
}

// Map returns a map of input name to input spec
func (inputs Inputs) Map() map[string]obs.InputSpec {
	m := map[string]obs.InputSpec{}
	for _, i := range inputs {
		m[i.Name] = i
	}
	return m
}

// ConfigmapNames returns a unique set of unordered configmap names
func (inputs Inputs) ConfigmapNames() []string {
	names := set.New[string]()
	for _, i := range inputs {
		if i.Receiver != nil && i.Receiver.TLS != nil {
			names.Insert(ConfigmapsForTLS(obs.TLSSpec(*i.Receiver.TLS))...)
		}
	}
	return names.UnsortedList()
}

// SecretNames returns a unique set of unordered secret names
func (inputs Inputs) SecretNames() []string {
	secrets := set.New[string]()
	for _, i := range inputs {
		if i.Receiver != nil && i.Receiver.TLS != nil {
			secrets.Insert(SecretsForTLS(obs.TLSSpec(*i.Receiver.TLS))...)
		}
	}
	return secrets.UnsortedList()
}

func (inputs Inputs) HasJournalSource() bool {
	for _, i := range inputs {
		if i.Type == obs.InputTypeInfrastructure && i.Infrastructure != nil && (len(i.Infrastructure.Sources) == 0 || set.New(i.Infrastructure.Sources...).Has(obs.InfrastructureSourceNode)) {
			return true
		}
	}
	return false
}

func (inputs Inputs) HasContainerSource() bool {
	for _, i := range inputs {
		if i.Type == obs.InputTypeApplication {
			return true
		}
		if i.Type == obs.InputTypeInfrastructure && i.Infrastructure != nil && (len(i.Infrastructure.Sources) == 0 || set.New(i.Infrastructure.Sources...).Has(obs.InfrastructureSourceContainer)) {
			return true
		}
	}
	return false
}
func (inputs Inputs) HasAnyAuditSource() bool {
	for _, i := range inputs {
		if i.Type == obs.InputTypeAudit && i.Audit != nil {
			return true
		}
	}
	return false
}

func (inputs Inputs) HasAuditSource(logSource obs.AuditSource) bool {
	for _, i := range inputs {
		if i.Type == obs.InputTypeAudit && i.Audit != nil && (set.New(i.Audit.Sources...).Has(logSource) || len(i.Audit.Sources) == 0) ||
			// Also true if HTTP input receiver with `kubeApiAudit` format is defined and logSource == `kubeAPI`
			(logSource == obs.AuditSourceKube &&
				(i.Type == obs.InputTypeReceiver && i.Receiver != nil && i.Receiver.HTTP != nil && i.Receiver.HTTP.Format == obs.HTTPReceiverFormatKubeAPIAudit)) {
			return true
		}
	}
	return false
}

func (inputs Inputs) HasReceiverSource() bool {
	for _, i := range inputs {
		if i.Type == obs.InputTypeReceiver && i.Receiver != nil {
			return true
		}
	}
	return false
}
