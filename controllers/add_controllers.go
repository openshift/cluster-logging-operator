package controllers

import (
	"github.com/openshift/cluster-logging-operator/controllers/clusterlogging"
	"github.com/openshift/cluster-logging-operator/controllers/forwarding"
	"github.com/openshift/cluster-logging-operator/controllers/trustedcabundle"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs,
		clusterlogging.Add,
		forwarding.Add,
		trustedcabundle.Add)
}
