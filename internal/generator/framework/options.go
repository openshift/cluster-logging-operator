package framework

import (
	utils "github.com/openshift/cluster-logging-operator/internal/utils"
)

const (
	ClusterTLSProfileSpec = "tlsProfileSpec"

	MinTLSVersion = "minTLSVersion"
	Ciphers       = "ciphers"

	URL = "url"
)

// Options is a map of Options used to customize the config generation. E.g. Debugging, legacy config generation
// Deprecated
type Options = utils.Options

// NoOptions is used to pass empty options
var NoOptions = utils.NoOptions

// TODO: unit me with functional.Option
type Option struct {
	Name  string
	Value interface{}
}

func HasOption(name string, options []Option) (interface{}, bool) {
	for _, o := range options {
		if o.Name == name {
			return o.Value, true
		}
	}
	return nil, false
}
