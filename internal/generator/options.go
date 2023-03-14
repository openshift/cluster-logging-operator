package generator

// Options is a map of Options used to customize the config generation. E.g. Debugging, legacy config generation
type Options map[string]interface{}

// NoOptions is used to pass empty options
var NoOptions = Options{}

// Has takes a key and returns true if it exists
func (o Options) Has(key string) bool {
	_, found := o[key]
	return found
}
