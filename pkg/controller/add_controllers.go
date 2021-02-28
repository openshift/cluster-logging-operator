package controller

import (
	"github.com/openshift/cluster-logging-operator/pkg/controller/clusterlogging"
	"github.com/openshift/cluster-logging-operator/pkg/controller/forwarding"
	"github.com/openshift/cluster-logging-operator/pkg/controller/trustedcabundle"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs,
		clusterlogging.Add,
		forwarding.Add,
		trustedcabundle.Add)
}
