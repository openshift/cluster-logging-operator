package elements

import (
	"fmt"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	. "github.com/openshift/cluster-logging-operator/internal/generator"
)

type FluentConfig struct {
	StoreID string
	Field1  Element
	Field2  Element
	Field3  Element
	Type    string
	Field4  Element
}

func (c FluentConfig) Name() string {
	return "someTemplate"
}

func (c FluentConfig) Template() string {
	return `{{define "` + c.Name() + `"}}
@id {{.StoreID}}
{{kv .Field1 -}}
{{kv .Field2 -}}
{{kv .Field3 -}}
type {{.Type}}
{{kv .Field4 -}}
{{end}}
`
}

func TestKeyVal(t *testing.T) {
	tests := []struct {
		f    FluentConfig
		conf string
	}{
		{
			f: FluentConfig{
				StoreID: "teststore",
				Field1:  KV("field1", "value1"),
				Field2:  KV("field2", "value2"),
				Field3:  Nil,
				Type:    "Tcp",
				Field4:  Nil,
			},
			conf: `
@id teststore
field1 = value1
field2 = value2
type Tcp
`,
		},
		{
			f: FluentConfig{
				StoreID: "teststore",
				Field1:  Nil,
				Field2:  KV("field2", "value2"),
				Field3:  Nil,
				Type:    "Tcp",
				Field4:  Nil,
			},
			conf: `
@id teststore
field2 = value2
type Tcp
`,
		},
		{
			f: FluentConfig{
				StoreID: "teststore",
				Field1:  KV("field1", "value1"),
				Field2:  KV("field2", "value2"),
				Field3:  KV("field3", "value3"),
				Type:    "Tcp",
				Field4:  KV("field4", "value4"),
			},
			conf: `
@id teststore
field1 = value1
field2 = value2
field3 = value3
type Tcp
field4 = value4
`,
		},
		{
			f: FluentConfig{
				StoreID: "teststore",
				Field1:  Nil,
				Field2:  Nil,
				Field3:  Nil,
				Type:    "Tcp",
				Field4:  Nil,
			},
			conf: `
@id teststore
type Tcp
`,
		},
	}
	for _, tt := range tests {
		c, err := MakeGenerator().GenerateConf([]Element{
			tt.f,
		}...)
		if err != nil {
			t.Fail()
		}
		if c != strings.TrimSpace(tt.conf) {
			fmt.Println(cmp.Diff(c, strings.TrimSpace(tt.conf)))
			t.Fail()
		}
	}
}
