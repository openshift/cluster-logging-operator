package toml

import (
	"bytes"
	"fmt"

	"github.com/pelletier/go-toml"
)

func MustMarshal(v interface{}) string {
	out, err := Marshal(v)
	if err != nil {
		panic(err)
	}
	return out
}

func Marshal(v interface{}) (string, error) {
	out := new(bytes.Buffer)
	encoder := toml.NewEncoder(out).Indentation("").Order(toml.OrderPreserve)
	if err := encoder.Encode(v); err != nil {
		return "", err
	}
	return out.String(), nil
}

func Unmarshal(s string, v interface{}) error {
	return toml.Unmarshal([]byte(s), v)
}
func MustUnmarshal(s string, v interface{}) {
	if err := Unmarshal(s, v); err != nil {
		panic(fmt.Sprintf("Error unmarshalling toml: %v\n%s\n", err, s))
	}
}
