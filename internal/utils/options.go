package utils

// Options is a map of Options used to customize the config generation. E.g. Debugging, legacy config generation
type Options map[string]interface{}

// NoOptions is used to pass empty options
var NoOptions = Options{}

// Has takes a key and returns true if it exists
func (o Options) Has(key string) bool {
	_, found := o[key]
	return found
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
