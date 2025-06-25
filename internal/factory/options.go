package factory

import (
	log "github.com/ViaQ/logerr/v2/log/static"
	"github.com/openshift/cluster-logging-operator/internal/generator/framework"
	"github.com/openshift/cluster-logging-operator/internal/utils"
)

func IncludesKubeCacheOption(op framework.Options) bool {
	_, found := utils.GetOption(op, framework.UseKubeCacheOption, "")
	log.V(3).Info("Kube caching option enabled")
	return found
}
