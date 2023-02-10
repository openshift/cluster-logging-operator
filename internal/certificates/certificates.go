package certificates

import (
	"fmt"
	"math/rand"
	"os"
	"os/exec"

	"sigs.k8s.io/yaml"

	log "github.com/ViaQ/logerr/v2/log/static"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	"github.com/openshift/cluster-logging-operator/test"
)

func GenerateTestCertificates(logStoreName string) (err error, certsDir string) {
	certsDir = fmt.Sprintf("/tmp/clo-test-%d-%d", os.Getpid(), rand.Intn(10000)) //nolint:gosec
	log.V(3).Info("Generating Pipeline certificates for Log Store to certs dir", "logStoreName", logStoreName, "certsDir", certsDir)
	err, _, _ = GenerateCertificates(constants.WatchNamespace, test.GitRoot("scripts"), logStoreName, certsDir)
	return
}

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
