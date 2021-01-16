package certificates

import (
	"fmt"
	"os/exec"

	"github.com/openshift/cluster-logging-operator/pkg/logger"
	"sigs.k8s.io/yaml"
)

func GenerateCertificates(namespace, scriptsDir, logStoreName, workDir string) (err error, updated bool) {
	script := fmt.Sprintf("%s/cert_generation.sh", scriptsDir)
	return RunCertificatesScript(namespace, logStoreName, workDir, script)
}

func RunCertificatesScript(namespace, logStoreName, workDir, script string) (err error, updated bool) {
	logger.Tracef("Running script %s workdir %s namespace %s logstore %s", script, workDir, namespace, logStoreName)
	cmd := exec.Command(script, workDir, namespace, logStoreName)
	out, err := cmd.Output()
	result := string(out)
	logger.Tracef("Cert generation result %s / err: %v", result, err)
	if result != "" {
		updated = true
		dumpLogs(result)
	}
	logger.Tracef("Returning err: %v , updated: %v", err, updated)
	return err, updated
}

func dumpLogs(raw string) {
	genLogs := []map[string]string{}
	err := yaml.Unmarshal([]byte(raw), &genLogs)
	if err == nil {
		for _, eventlog := range genLogs {
			logger.Infof("cert_generation output: %v", eventlog)
		}
		return
	}
	logger.Errorf("Unable to unmarshal cert_generation to structured output: %v", err)
	logger.Infof("cert_generation output: %s", raw)
}
