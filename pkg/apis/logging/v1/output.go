package v1

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1/outputs"
)

const OutputTypeDefault = "default"

var (
	outputNameToType map[string]reflect.Type
	outputTypeToName map[reflect.Type]string

	OutputTypeSyslog        string
	OutputTypeFluent        string
	OutputTypeElasticsearch string
)

func FindOutputType(name string) reflect.Type { return outputNameToType[name] }

func IsOutputTypeName(name string) bool {
	return FindOutputType(name) != nil || name == OutputTypeDefault
}

// OutputTypeName returns the name for an output type.
// v may be a reflect.type or an instance, e.g.
//     name := OutputTypeName(outputs.Syslog{})
//     name := OutputTypeName(reflect.TypeOf(outputs.Syslog{}))
func OutputTypeName(v interface{}) string {
	t, ok := v.(reflect.Type)
	if !ok {
		t = reflect.TypeOf(v)
	}
	return outputTypeToName[t]
}

// IsType returns true if name is a valid `output.type` value, i.e. it is
// the JSON field name of one of the options in the OutputType union.
func IsType(name string) bool { _, ok := outputNameToType[name]; return ok }

// Extract name/type pairs from the OutputType union class.
func init() {
	outputNameToType = map[string]reflect.Type{}
	outputTypeToName = map[reflect.Type]string{}

	ot := reflect.TypeOf(OutputTypeSpec{})
	for i := 0; i < ot.NumField(); i++ {
		f := ot.Field(i)
		ft := f.Type
		if ft.Kind() != reflect.Ptr || ft.Elem().Kind() != reflect.Struct {
			continue // Only pointer-to-struct
		}
		tag := f.Tag.Get("json")
		if tag == "" || tag == "-" {
			continue // Only JSON-named fields.
		}
		if i := strings.Index(tag, ","); i != -1 {
			tag = tag[:i]
		}
		if _, ok := outputNameToType[tag]; ok {
			panic(fmt.Errorf("%v has non-unique JSON field name %v", ot.Name(), tag))
		}
		outputNameToType[tag] = ft.Elem()
		outputTypeToName[ft.Elem()] = tag
	}
	OutputTypeSyslog = OutputTypeName(outputs.Syslog{})
	OutputTypeFluent = OutputTypeName(outputs.FluentForward{})
	OutputTypeElasticsearch = OutputTypeName(outputs.ElasticSearch{})
}
