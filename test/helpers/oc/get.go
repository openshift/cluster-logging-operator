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

package oc

import (
	"fmt"
	"strings"
)

// Getter is interface for collecting arguments for Get command
type Getter interface {
	Command

	// argument for option --config
	WithConfig(string) Getter
	// argument for option -n
	WithNamespace(string) Getter

	// sets the Get command to get pod
	Pod() Getter
	// argument for resource kind to get
	Resource(kind string, name string) Getter

	// argument for resource name to get
	Name(string) Getter
	// argument for resource selectors
	Selector(string) Getter

	// sets -o json output
	OutputJson() Getter
	// sets -o yaml output
	OutputYaml() Getter
	// argument for -o jsonpath output
	OutputJsonpath(string) Getter
}

type get struct {
	*runner
	namespace string

	kind     string
	name     string
	selector string

	output string
}

// Get creates an 'oc get' command
func Get() Getter {
	g := &get{
		runner: &runner{},
	}
	g.collectArgsFunc = g.args
	return g
}

func (g *get) WithConfig(cfg string) Getter {
	g.configPath = cfg
	return g
}

func (g *get) WithNamespace(namespace string) Getter {
	g.namespace = namespace
	return g
}

func (g *get) Pod() Getter {
	g.kind = "pod"
	return g
}

func (g *get) Name(name string) Getter {
	g.name = name
	return g
}

func (g *get) Selector(s string) Getter {
	g.selector = s
	return g
}

func (g *get) Resource(kind string, name string) Getter {
	g.kind = kind
	g.name = name
	return g
}

func (g *get) OutputJson() Getter {
	g.output = "-o json"
	return g
}

func (g *get) OutputYaml() Getter {
	g.output = "-o yaml"
	return g
}

func (g *get) OutputJsonpath(path string) Getter {
	g.output = fmt.Sprintf("-o jsonpath=%s", path)
	return g
}

func (g *get) args() []string {
	namespaceStr := ""
	if g.namespace != "" {
		namespaceStr = fmt.Sprintf("-n %s", g.namespace)
	}
	str := ""
	if g.selector != "" {
		str = fmt.Sprintf("-l %s", g.selector)
	} else if g.name != "" {
		str = g.name
	}
	return sanitizeArgs(fmt.Sprintf("%s get %s %s %s", namespaceStr, g.kind, str, g.output))
}

func (g *get) String() string {
	return fmt.Sprintf("%s %s", g.runner.String(), strings.Join(g.args(), " "))
}
