// Licensed to Red Hat, Inc under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Red Hat, Inc licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package generators

import (
	"bytes"
	"fmt"
	"text/template"
)

//Generator is a template engine
type Generator struct {
	*template.Template
}

//New creates an instance of a template engine for a set of templates
func New(name string, addFunctions *template.FuncMap, templates ...string) (*Generator, error) {
	tmpl := template.New(name).Funcs(*addFunctions).Funcs(funcMap)
	var err error
	for i, s := range templates {
		tmpl, err = tmpl.Parse(s)
		if err != nil {
			return nil, fmt.Errorf("Error parsing %q template %v: %s", name, i, err)
		}
	}
	return &Generator{tmpl}, nil
}

//Execute the named template using the given data
func (gen *Generator) Execute(namedTemplate string, data interface{}) (string, error) {
	var out bytes.Buffer
	err := gen.Template.ExecuteTemplate(&out, namedTemplate, data)
	return out.String(), err
}
