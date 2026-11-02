package api

import "github.com/openshift/cluster-logging-operator/internal/generator/impl/otelc/api/types"

type Extensions map[string]types.Extension

func (m *Extensions) Add(ext types.Extension) {
	(*m)[ext.ID()] = ext
}

func (m *Extensions) Merge(other Extensions) {
	for _, e := range other {
		(*m)[e.ID()] = e
	}
}
