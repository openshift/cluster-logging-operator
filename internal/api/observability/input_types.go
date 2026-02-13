package observability

import (
	"regexp"
	"sort"

	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/utils/sets"
	"k8s.io/utils/set"
)

var (
	ReservedInputTypes = sets.NewString(
		obs.InputTypeApplication.String(),
		obs.InputTypeAudit.String(),
		obs.InputTypeInfrastructure.String(),
	)
	ReservedApplicationSources    = sets.NewString(obs.ApplicationSourceContainer.String())
	ReservedInfrastructureSources = sets.NewString(obs.InfrastructureSourceContainer.String(), obs.InfrastructureSourceNode.String())

	InfraNSRegex = regexp.MustCompile(`^(?P<default>default)|(?P<openshift>openshift.*)|(?P<kube>kube.*)$`)
)

func MaxRecordsPerSecond(input obs.InputSpec) (int64, bool) {
	if app := input.Application; app != nil {
		if tuning := app.Tuning; tuning != nil {
			if rateLimitPerContainer := tuning.RateLimitPerContainer; rateLimitPerContainer != nil {
				return Threshold(rateLimitPerContainer)
			}
		}
	}
	if infra := input.Infrastructure; infra != nil {
		if tuning := infra.Tuning; tuning != nil {
			if rateLimitPerContainer := tuning.Container.RateLimitPerContainer; rateLimitPerContainer != nil {
				return Threshold(rateLimitPerContainer)
			}
		}
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

func IncludesInfraNamespace(input *obs.Application) bool {
	if input != nil {
		for _, ns := range input.Includes {
			if matches := InfraNSRegex.FindStringSubmatch(ns.Namespace); matches != nil {
				return true
			}
		}
	}
	return false
}

// Names returns a slice of input names
func (inputs Inputs) Names() (names []string) {
	for _, i := range inputs {
		names = append(names, i.Name)
	}
	return names
}

// InputTypes returns a unique set of sorted input types.
func (inputs Inputs) InputTypes() []obs.InputType {
	types := set.New[obs.InputType]()
	for _, i := range inputs {
		types.Insert(i.Type)
	}
	typesList := types.UnsortedList()
	sort.Slice(typesList, func(i, j int) bool {
		return typesList[i].String() < typesList[j].String()
	})
	return typesList
}

// InputSources returns a unique set of input sources based upon the input type
func (inputs Inputs) InputSources(inputType obs.InputType) []string {
	types := set.New[string]()
	for _, i := range inputs {
		if i.Type == inputType {
			switch i.Type {
			case obs.InputTypeApplication:
				types.Insert(ReservedApplicationSources.List()...)
			case obs.InputTypeInfrastructure:
				types.Insert(InfrastructureSources(i.Infrastructure.Sources).AsStrings()...)
			case obs.InputTypeAudit:
				types.Insert(AuditSources(i.Audit.Sources).AsStrings()...)
			}
		}
	}
	return types.SortedList()
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

type InfrastructureSources []obs.InfrastructureSource

func (infraSources InfrastructureSources) AsStrings() (result []string) {
	for _, s := range infraSources {
		result = append(result, string(s))
	}
	return result
}

type AuditSources []obs.AuditSource

func (auditSources AuditSources) AsStrings() (result []string) {
	for _, s := range auditSources {
		result = append(result, string(s))
	}
	return result
}

// Input is an internal representation of the public API input
type Input struct {
	obs.InputSpec
	Ids []string
}

func (i *Input) InputIDs() []string {
	return i.Ids
}

func (i *Input) GetTlsSpec() *obs.TLSSpec {
	if i.Receiver == nil || i.Receiver.TLS == nil {
		return nil
	}
	tlsSpec := obs.TLSSpec(*i.Receiver.TLS)
	return &tlsSpec
}

func (i *Input) IsInsecureSkipVerify() bool {
	return false
}

func NewInput(spec obs.InputSpec) *Input {
	i := Input{
		InputSpec: spec,
	}
	return &i
}
