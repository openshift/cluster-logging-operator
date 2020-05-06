package v1

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1/outputs"

	sets "k8s.io/apimachinery/pkg/util/sets"
)

var ReservedOutputNames = sets.NewString(OutputNameDefault)

func IsOutputTypeName(s string) bool {
	_, ok := goNames[s]
	return ok || ReservedOutputNames.Has(s)
}

// Output defines a destination for log messages.
type OutputSpec struct {
	// Name used to refer to the output from a `pipeline`.
	//
	// +required
	Name string `json:"name"`

	// Type of output plugin, for example 'syslog'
	//
	// +required
	Type string `json:"type"`

	// URL to send log messages to.
	//
	// Must be an absolute URL, with a scheme. Valid URL schemes depend on `type`.
	// Special schemes 'tcp', 'udp' and 'tls' are used for output types that don't
	// define their own URL scheme.  Example:
	//
	//     { type: syslog, url: tls://syslog.example.com:1234 }
	//
	// TLS with server authentication is enabled by the URL scheme, for
	// example 'tls' or 'https'.  See `secret` for TLS client authentication.
	//
	// +optional
	URL string `json:"url"`

	// OutputTypeSpec provides optional extra configuration that is specific to the
	// output `type`
	//
	// +optional
	OutputTypeSpec `json:",inline"`

	// Secret for secure communication.
	// Secrets must be stored in the namespace containing the cluster logging operator.
	//
	// Client-authenticated TLS is enabled if the secret contains keys `tls.crt`,
	// `tls.key` and `ca.crt`. Output types with password authentication will use
	// keys `password` and `username`, not the exposed 'username@password' part of
	// the `url`.
	//
	// +optional
	Secret *OutputSecretSpec `json:"secret,omitempty"`

	// Insecure must be true for intentionally insecure outputs.
	// Has no function other than a marker to help avoid configuration mistakes.
	//
	// +optional
	Insecure bool `json:"insecure,omitempty"`
}

// OutputSecretSpec is a secret reference containing name only, no namespace.
type OutputSecretSpec struct {
	// Name of a secret in the namespace configured for log forwarder secrets.
	//
	// +required
	Name string `json:"name"`
}

// OutputTypeSpec is a union of optional additional configuration specific to an
// output type.
type OutputTypeSpec struct {
	// +optional
	Syslog *outputs.Syslog `json:"syslog,omitempty"`
	// +optional
	FluentForward *outputs.FluentForward `json:"fluentForward,omitempty"`
	// +optional
	ElasticSearch *outputs.ElasticSearch `json:"elasticsearch,omitempty"`
}

// OutputTypeHandler has methods for each of the valid output types.
// They receive the output type spec field (possibly nil) and
// return a validation error.
//
type OutputTypeHandler interface {
	Syslog(*outputs.Syslog) error
	FluentForward(*outputs.FluentForward) error
	ElasticSearch(*outputs.ElasticSearch) error
}

// HandleType validates spec.Type and spec.OutputType,
// then calls the relevant handler method with the OutputType
// pointer, which may be nil.
//
func (spec OutputSpec) HandleType(h OutputTypeHandler) error {
	if !IsOutputTypeName(spec.Type) {
		return fmt.Errorf("not a valid output type: '%s'", spec.Type)
	}
	// Call handler method with OutputSpec field value
	goName := goNames[spec.Type]
	args := []reflect.Value{reflect.ValueOf(spec).FieldByName(goName)}
	result := reflect.ValueOf(h).MethodByName(goName).Call(args)[0].Interface()
	err, _ := result.(error)
	return err
}

var goNames = map[string]string{}

func init() {
	otsType := reflect.TypeOf(OutputTypeSpec{})
	for i := 0; i < otsType.NumField(); i++ {
		f := otsType.Field(i)
		tags := strings.Split(f.Tag.Get("json"), ",")
		if len(tags) > 0 && tags[0] != "-" && tags[0] != "" {
			goNames[tags[0]] = f.Name
		}
	}
}

// Output type and name constants.
const (
	OutputTypeElasticsearch = "elasticsearch"
	OutputTypeFluentForward = "fluentForward"
	OutputTypeSyslog        = "syslog"

	OutputNameDefault = "default" // Default log store.
)
