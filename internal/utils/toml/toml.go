package toml

import (
	"bytes"

	toml "github.com/pelletier/go-toml"
)

func MustMarshal(v interface{}) string {
	out := new(bytes.Buffer)
	encoder := toml.NewEncoder(out).Indentation("")
	if err := encoder.Encode(v); err != nil {
		panic(err)
	}
	return out.String()
}
