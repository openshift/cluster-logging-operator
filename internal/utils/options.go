package utils

import (
	"sort"

	"github.com/openshift/cluster-logging-operator/internal/utils/sets"
)

// Options is a map of Options used to customize the config generation. E.g. Debugging, legacy config generation
type Options map[string]interface{}

// NoOptions is used to pass empty options
var NoOptions = Options{}

// Has takes a key and returns true if it exists
func (o Options) Has(key string) bool {
	_, found := o[key]
	return found
}

// Set the option to the given value and return the options
func (o Options) Set(key string, value interface{}) Options {
	o[key] = value
	return o
}

func (o Options) GetStringSet(name string) []string {
	entry, found := o[name]
	if !found {
		return []string{}
	}
	value := sets.NewString(entry.([]string)...).List()
	sort.Strings(value)
	return value
}

func (o Options) AddToStringSet(name string, values ...string) {
	entry, found := o[name]
	if !found {
		entry = []string{}
	}
	value := entry.([]string)
	o[name] = append(value, values...)
}

// Update sets the named options with the provided value processing it using the provided function if not nil
func Update[T any](options Options, name string, value T, updater func(T) T) {
	if updater == nil {
		updater = func(value T) T {
			return value
		}
	}
	if existingValue, found := options[name]; found {
		options[name] = updater(existingValue.(T))
	} else {
		options[name] = value
	}
}

// GetOption from the named value from the list of options and convert it as needed
func GetOption[T any](options Options, name string, ifNotFound T) (T, bool) {
	value, found := options[name]
	if !found {
		return ifNotFound, false
	}
	return value.(T), found
}
