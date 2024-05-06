package tls

import (
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/generator/framework"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/common"
)

type TLSConf struct {
	*common.TLSConf
}

func New(id string, tls *obs.OutputTLSSpec, secrets helpers.Secrets, op framework.Options) common.TLSConf {
	conf := common.TLSConf{
		ComponentID:  id,
		NeedsEnabled: tls != nil,
	}
	if tls != nil {
		conf.CAFilePath = ConfigMapOrSecretPath(tls.CA)
		conf.CertPath = ConfigMapOrSecretPath(tls.Certificate)
		conf.KeyPath = SecretPath(tls.Key)
		conf.PassPhrase = secrets.AsString(tls.KeyPassphrase)
		conf.InsecureSkipVerify = tls.InsecureSkipVerify

	}
	conf.SetTLSProfileFromOptions(op)

	return conf
}

func ConfigMapOrSecretPath(resource *obs.ConfigMapOrSecretKey) string {
	if resource == nil {
		return ""
	}
	if resource.Secret != nil {
		return helpers.SecretPath(resource.Secret.Name, resource.Key)
	} else if resource.ConfigMap != nil {
		return helpers.AuthPath(resource.ConfigMap.Name, resource.Key)
	}
	return ""
}

func SecretPath(resource *obs.SecretKey) string {
	if resource == nil || resource.Secret == nil {
		return ""
	}
	return helpers.SecretPath(resource.Secret.Name, resource.Key)
}
