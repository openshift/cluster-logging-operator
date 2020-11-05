package k8shandler

import (
	"fmt"
	"os/exec"

	"github.com/ViaQ/logerr/log"
	"github.com/openshift/cluster-logging-operator/pkg/certificates"
	"github.com/openshift/cluster-logging-operator/pkg/constants"
	"github.com/openshift/cluster-logging-operator/pkg/utils"
	"k8s.io/apimachinery/pkg/api/errors"

	"k8s.io/apimachinery/pkg/util/sets"
)

var deprecatedKeys = sets.NewString("app-ca", "app-key", "app-cert", "infra-ca", "infra-key", "infra-cert")

// golang doesn't allow for const maps
var secretCertificates = map[string]map[string]string{
	"master-certs": {
		"masterca":  "ca.crt",
		"masterkey": "ca.key",
	},
	"elasticsearch": {
		"elasticsearch.key": "elasticsearch.key",
		"elasticsearch.crt": "elasticsearch.crt",
		"logging-es.key":    "logging-es.key",
		"logging-es.crt":    "logging-es.crt",
		"admin-key":         "system.admin.key",
		"admin-cert":        "system.admin.crt",
		"admin-ca":          "ca.crt",
	},
	"kibana": {
		"ca":   "ca.crt",
		"key":  "system.logging.kibana.key",
		"cert": "system.logging.kibana.crt",
	},
	"kibana-proxy": {
		"server-key":     "kibana-internal.key",
		"server-cert":    "kibana-internal.crt",
		"session-secret": "kibana-session-secret",
	},
	"curator": {
		"ca":       "ca.crt",
		"key":      "system.logging.curator.key",
		"cert":     "system.logging.curator.crt",
		"ops-ca":   "ca.crt",
		"ops-key":  "system.logging.curator.key",
		"ops-cert": "system.logging.curator.crt",
	},
	"fluentd": {
		"ca-bundle.crt": "ca.crt",
		"tls.key":       "system.logging.fluentd.key",
		"tls.crt":       "system.logging.fluentd.crt",
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
		if deprecatedKeys.Has(key) {
			log.Info("No secret data found. Please be aware but likely not an issue for deprecated keys", "key", key, "deprecatedKeys", deprecatedKeys.List())
		} else {
			log.Info("No secret data found", key)

		}
		return nil
	}

	return utils.WriteToWorkingDirFile(toFile, value)
}

func (clusterRequest *ClusterLoggingRequest) writeSecret() (err error) {

	secret := NewSecret(
		constants.MasterCASecretName,
		clusterRequest.Cluster.Namespace,
		map[string][]byte{
			"masterca":  utils.GetWorkingDirFileContents("ca.crt"),
			"masterkey": utils.GetWorkingDirFileContents("ca.key"),
		})

	utils.AddOwnerRefToObject(secret, utils.AsOwner(clusterRequest.Cluster))

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

	scriptsDir := utils.GetScriptsDir()
	if err = certificates.GenerateCertificates(clusterRequest.Cluster.Namespace, scriptsDir, "elasticsearch", utils.DefaultWorkingDir); err != nil {
		return fmt.Errorf("Error running script: %v", err)
	}

	if err = clusterRequest.writeSecret(); err != nil {
		return
	}

	return nil
}

func GenerateCertificates(namespace, scriptsDir, logStoreName, workDir string) (err error) {
	script := fmt.Sprintf("%s/cert_generation.sh", scriptsDir)
	return RunCertificatesScript(namespace, logStoreName, workDir, script)
}

func RunCertificatesScript(namespace, logStoreName, workDir, script string) (err error) {
	log.V(3).Info("Running script", "script", script, "workDir", workDir, "namespace", namespace, "logStoreName", logStoreName)
	cmd := exec.Command(script, workDir, namespace, logStoreName)
	result, err := cmd.Output()

	if log.V(3).Enabled() {
		log.V(3).Info("cert_generation output", "output", string(result))
	}
	if err != nil {
		log.V(2).Error(err, "Error")
	}
	return err
}
