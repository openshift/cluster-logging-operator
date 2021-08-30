package generator

// KeyVal is an Element which can be used to generate a <key value> line in config
// When used with 'kv' function, can be used to omit lines for which values are not set
// check tests for usage
type KeyVal struct {
	Key string
	Val string
}

func (kv KeyVal) Name() string {
	return "keyvalTemplate"
}

func (kv KeyVal) Template() string {
	return `{{define "` + kv.Name() + `" -}}
{{.Key}} {{.Val}}
{{end -}}`
}

func KV(k, v string) KeyVal {
	return KeyVal{
		Key: k,
		Val: v,
	}
}
