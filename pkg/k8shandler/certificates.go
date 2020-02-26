package k8shandler

import (
	"fmt"
	"os/exec"

	"github.com/openshift/cluster-logging-operator/pkg/logger"
	"github.com/openshift/cluster-logging-operator/pkg/utils"
	"k8s.io/apimachinery/pkg/api/errors"

	"k8s.io/apimachinery/pkg/util/sets"
)

var deprecatedKeys = sets.NewString("app-ca", "app-key", "app-cert", "infra-ca", "infra-key", "infra-cert")

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
		"server-key":     "kibana-internal.key",
		"server-cert":    "kibana-internal.crt",
		"session-secret": "kibana-session-secret",
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
			logger.Infof("No secret data %q found. Please be aware but likely not an issue for deprecated keys: %v", key, deprecatedKeys.List())
		} else {
			logger.Warnf("No secret data %q found", key)

		}
		return nil
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
	if err = GenerateCertificates(clusterRequest.cluster.Namespace, ".", "elasticsearch", utils.DefaultWorkingDir); err != nil {
		return fmt.Errorf("Error running script: %v", err)
	}

	if err = clusterRequest.writeSecret(); err != nil {
		return
	}

	return nil
}

func GenerateCertificates(namespace, rootDir, logStoreName, workDir string) (err error) {
	script := fmt.Sprintf("%s/scripts/cert_generation.sh", rootDir)
	return RunCertificatesScript(namespace, logStoreName, workDir, script)
}

func RunCertificatesScript(namespace, logStoreName, workDir, script string) (err error) {
	logger.Debugf("Running script '%s %s %s %s'", script, workDir, namespace, logStoreName)
	cmd := exec.Command(script, workDir, namespace, logStoreName)
	result, err := cmd.Output()
	if logger.IsDebugEnabled() {
		logger.Debugf("cert_generation output: %s", string(result))
		logger.Debugf("err: %v", err)
	}
	return err
}
