package controller

import (
	"github.com/openshift/cluster-logging-operator/pkg/controller/clusterlogging"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, clusterlogging.Add)
}
