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
	"strings"
	"text/template"

	"k8s.io/apimachinery/pkg/util/sets"
)

var (
	funcMap = template.FuncMap{
		"indent":  indent,
		"include": include,
	}
)

//IncludeTemplate allows root templates to include partials by reusing
//the root template object.  Callers can provide hints for use in temporarily changing
//behavior of the included template
type IncludeTemplate interface {
	Template() *template.Template
	SetHints(hints []string)
	Hints() sets.String
}

//include a named template using the given binding and hints
func include(name string, context IncludeTemplate, hints ...string) (string, error) {
	origHints := context.Hints()
	context.SetHints(hints)
	var out bytes.Buffer
	if err := context.Template().ExecuteTemplate(&out, name, context); err != nil {
		context.SetHints(origHints.List())
		return "", err
	}
	context.SetHints(origHints.List())
	return out.String(), nil
}

//indent helper function to prefix each line of the output by N spaces
func indent(length int, in string) string {
	pad := strings.Repeat(" ", length)
	return pad + strings.Replace(in, "\n", "\n"+pad, -1)
}
