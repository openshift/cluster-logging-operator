package helpers

import (
	logging "github.com/openshift/cluster-logging-operator/api/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/generator/framework"
)

func SetTLSProfileOptions(o logging.OutputSpec, op framework.Options) {
	op[framework.MinTLSVersion], op[framework.Ciphers] = func() (string, string) {
		if o.Name == logging.OutputNameDefault && o.Type == logging.OutputTypeElasticsearch {
			return "", ""
		} else {
			return op.TLSProfileInfo(o, ",")
		}
	}()
}
