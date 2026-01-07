package multilineexception

import (
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api/transforms"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api/types"
)

func New(inputs ...string) types.Transform {
	de := transforms.NewDetectExceptions([]transforms.LanguageType{transforms.LanguageTypeAll}, inputs...)
	de.GroupBy = []string{"._internal.kubernetes.namespace_name", "._internal.kubernetes.pod_name", "._internal.kubernetes.container_name", "._internal.kubernetes.pod_id", "_internal.kubernetes.io_stream"}
	de.ExpireAfterMs = 2000
	de.MultilineFlushIntervalMs = 1000
	de.MessageKey = "._internal.message"
	return de
}
