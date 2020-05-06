package v1

import (
	"github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1/inputs"
	sets "k8s.io/apimachinery/pkg/util/sets"
)

// Reserved input names.
const (
	InputNameApplication    = "application"    // Non-infrastructure container logs.
	InputNameInfrastructure = "infrastructure" // Infrastructure containers and system logs.
	InputNameAudit          = "audit"          // System audit logs.
)

var ReservedInputNames = sets.NewString(InputNameApplication, InputNameInfrastructure, InputNameAudit)

func IsInputTypeName(s string) bool { return ReservedInputNames.Has(s) }

// InputSpec defines a selector of log messages.
type InputSpec struct {
	// Name used to refer to the input of a `pipeline`.
	//
	// +required
	Name string `json:"name"`

	// NOTE: the following fields in this struct are deliberately _not_ `omitempty`.
	// An empty field means enable that input type with no filter.

	// Enable `application` logs. Use `application: {}` to enable with no filter.
	// +optional
	Application *inputs.Application `json:"application"`

	// Enable `infrastructure` logs. Use `infrastructure: {}` to enable with no filter.
	// +optional
	Infrastructure *inputs.Infrastructure `json:"infrastructure"`

	// Enable `audit` logs. Use `infrastructure: {}` to enable with no filter.
	// +optional
	Audit *inputs.Audit `json:"audit"`
}
