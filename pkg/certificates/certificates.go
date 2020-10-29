package certificates

import (
	"fmt"
	"os/exec"

	"github.com/openshift/cluster-logging-operator/pkg/logger"
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
		logger.Infof("cert_generation output: %s", result)
	}
	logger.Tracef("Returning err: %v , updated: %v", err, updated)
	return err, updated
}
