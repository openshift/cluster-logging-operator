package controller

import (
	"github.com/openshift/elasticsearch-operator/pkg/controller/elasticsearch"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, elasticsearch.Add)
}
