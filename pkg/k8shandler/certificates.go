package k8shandler

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/openshift/cluster-logging-operator/pkg/utils"
	"k8s.io/apimachinery/pkg/api/errors"
)

// golang doesn't allow for const maps
var secretCertificates = map[string]map[string]string{
	"master-certs": map[string]string{
		"masterca":  "ca.crt",
		"masterkey": "ca.key",
	},
	"elasticsearch": map[string]string{
		"elasticsearch.key": "elasticsearch.key",
		"elasticsearch.crt": "elasticsearch.crt",
		"logging-es.key":    "logging-es.key",
		"logging-es.crt":    "logging-es.crt",
		"admin-key":         "system.admin.key",
		"admin-cert":        "system.admin.crt",
		"admin-ca":          "ca.crt",
	},
	"kibana": map[string]string{
		"ca":   "ca.crt",
		"key":  "system.logging.kibana.key",
		"cert": "system.logging.kibana.crt",
	},
	"kibana-proxy": map[string]string{
		"server-key":  "kibana-internal.key",
		"server-cert": "kibana-internal.crt",
	},
	"curator": map[string]string{
		"ca":       "ca.crt",
		"key":      "system.logging.curator.key",
		"cert":     "system.logging.curator.crt",
		"ops-ca":   "ca.crt",
		"ops-key":  "system.logging.curator.key",
		"ops-cert": "system.logging.curator.crt",
	},
	"fluentd": map[string]string{
		"app-ca":     "ca.crt",
		"app-key":    "system.logging.fluentd.key",
		"app-cert":   "system.logging.fluentd.crt",
		"infra-ca":   "ca.crt",
		"infra-key":  "system.logging.fluentd.key",
		"infra-cert": "system.logging.fluentd.crt",
	},
	"rsyslog": map[string]string{
		"app-ca":     "ca.crt",
		"app-key":    "system.logging.rsyslog.key",
		"app-cert":   "system.logging.rsyslog.crt",
		"infra-ca":   "ca.crt",
		"infra-key":  "system.logging.rsyslog.key",
		"infra-cert": "system.logging.rsyslog.crt",
	},
}

func (clusterRequest *ClusterLoggingRequest) extractSecretToFile(secretName string, key string, toFile string) (err error) {
	secret, err := clusterRequest.GetSecret(secretName)
	if err != nil {
		if errors.IsNotFound(err) {
			return err
		}
		return fmt.Errorf("Unable to extract secret %s to file: %v", secretName, err)
	}

	value, ok := secret.Data[key]

	// check to see if the map value exists
	if !ok {
		return fmt.Errorf("No secret data \"%s\" found", key)
	}

	return utils.WriteToWorkingDirFile(toFile, value)
}

func (clusterRequest *ClusterLoggingRequest) writeSecret() (err error) {

	secret := NewSecret(
		"master-certs",
		clusterRequest.cluster.Namespace,
		map[string][]byte{
			"masterca":  utils.GetWorkingDirFileContents("ca.crt"),
			"masterkey": utils.GetWorkingDirFileContents("ca.key"),
		})

	utils.AddOwnerRefToObject(secret, utils.AsOwner(clusterRequest.cluster))

	err = clusterRequest.CreateOrUpdateSecret(secret)
	if err != nil {
		return
	}

	return nil
}

func (clusterRequest *ClusterLoggingRequest) readSecrets() (err error) {

	for secretName, certMap := range secretCertificates {
		if err = clusterRequest.extractCertificates(secretName, certMap); err != nil {
			return
		}
	}

	return nil
}

func (clusterRequest *ClusterLoggingRequest) extractCertificates(secretName string, certs map[string]string) (err error) {

	for secretKey, certPath := range certs {
		if err = clusterRequest.extractSecretToFile(secretName, secretKey, certPath); err != nil {
			if errors.IsNotFound(err) {
				return nil
			}
			return
		}
	}

	return nil
}

//CreateOrUpdateCertificates for a cluster logging instance
func (clusterRequest *ClusterLoggingRequest) CreateOrUpdateCertificates() (err error) {

	// Pull master signing cert out from secret in logging.Spec.SecretName
	if err = clusterRequest.readSecrets(); err != nil {
		return
	}

	cmd := exec.Command("bash", "scripts/cert_generation.sh")
	cmd.Env = append(os.Environ(),
		"NAMESPACE="+clusterRequest.cluster.Namespace,
	)
	if err = cmd.Run(); err != nil {
		return fmt.Errorf("Error running script: %v", err)
	}

	if err = clusterRequest.writeSecret(); err != nil {
		return
	}

	return nil
}
