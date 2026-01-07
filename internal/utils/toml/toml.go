package toml

import (
	"bytes"
	"fmt"

	toml "github.com/pelletier/go-toml"
)

func MustMarshal(v interface{}) string {
	out := new(bytes.Buffer)
	encoder := toml.NewEncoder(out).Indentation("").Order(toml.OrderPreserve)
	if err := encoder.Encode(v); err != nil {
		panic(err)
	}
	return out.String()
}

func MustUnMarshal(s string, v interface{}) {
	if err := toml.NewDecoder(bytes.NewBufferString(s)).Decode(v); err != nil {
		panic(fmt.Sprintf("Error unmarshalling toml: %v\n%s\n", err, s))
	}
}
