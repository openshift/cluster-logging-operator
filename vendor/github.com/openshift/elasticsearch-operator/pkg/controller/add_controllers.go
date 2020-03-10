package controller

import (
	"github.com/openshift/elasticsearch-operator/pkg/controller/elasticsearch"
	"github.com/openshift/elasticsearch-operator/pkg/controller/kibana"
	"github.com/openshift/elasticsearch-operator/pkg/controller/kibanasecret"
	"github.com/openshift/elasticsearch-operator/pkg/controller/proxyconfig"
	"github.com/openshift/elasticsearch-operator/pkg/controller/trustedcabundle"
)

func init() {
	AddToManagerFuncs = append(AddToManagerFuncs,
		kibana.Add,
		elasticsearch.Add,
		proxyconfig.Add,
		kibanasecret.Add,
		trustedcabundle.Add)
}
