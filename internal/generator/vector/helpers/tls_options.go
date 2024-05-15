package helpers

import (
	logging "github.com/openshift/cluster-logging-operator/api/logging/v1"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/generator/framework"
)

// TODO: Can we replace this in the generators by moving to a migration and setting before generation?
func SetTLSProfileOptions(o obs.OutputSpec, op framework.Options) {
	op[framework.MinTLSVersion], op[framework.Ciphers] = func() (string, string) {
		if o.Name == logging.OutputNameDefault && o.Type == obs.OutputTypeElasticsearch {
			return "", ""
		} else {
			return op.TLSProfileInfo(o, ",")
		}
	}()
}
