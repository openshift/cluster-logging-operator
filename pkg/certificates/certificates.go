package certificates

import (
	"fmt"
	"os/exec"

	"sigs.k8s.io/yaml"

	"github.com/ViaQ/logerr/log"
	"github.com/openshift/cluster-logging-operator/pkg/utils"
)

func GenerateCertificates(namespace, scriptsDir, logStoreName, workDir string) (err error, updated bool, output []interface{}) {
	script := fmt.Sprintf("%s/cert_generation.sh", scriptsDir)
	return RunCertificatesScript(namespace, logStoreName, workDir, script)
}

func RunCertificatesScript(namespace, logStoreName, workDir, script string) (err error, updated bool, output []interface{}) {
	updated = false
	log.V(3).Info("Running script", "script", script, "workDir", workDir, "namespace", namespace, "logStoreName", logStoreName)
	cmd := exec.Command(script, workDir, namespace, logStoreName)
	out, err := cmd.Output()
	result := string(out)
	// get error string from certificate generation script
	err = utils.WrapError(err)
	log.V(3).Info("Cert generation", "out", result, "err", err)
	if result != "" {
		updated = true
		output = readOutput(result)
	}
	log.V(3).Info("Returning", "err", err, "updated", updated)
	return err, updated, output
}

func readOutput(raw string) []interface{} {
	genLogs := []map[string]string{}
	err := yaml.Unmarshal([]byte(raw), &genLogs)
	if err != nil {
		log.Error(err, "Unable to unmarshal cert_generation to structured output")
		log.Info("cert_generation output", "output", raw)
		return nil
	}

	result := []interface{}{}
	for _, eventlog := range genLogs {
		keyvalues := []interface{}{}
		for k, v := range eventlog {
			keyvalues = append(keyvalues, k, v)
		}
		result = append(result, keyvalues)
		log.Info("cert_generation output", keyvalues...)
	}

	return result
}
