package certificates

import (
	"fmt"
	"os/exec"

	"github.com/ViaQ/logerr/log"
)

func GenerateCertificates(namespace, scriptsDir, logStoreName, workDir string) (err error, updated bool) {
	script := fmt.Sprintf("%s/cert_generation.sh", scriptsDir)
	return RunCertificatesScript(namespace, logStoreName, workDir, script)
}

func RunCertificatesScript(namespace, logStoreName, workDir, script string) (err error, updated bool) {
	updated = false
	log.V(3).Info("Running script", "script", script, "workDir", workDir, "namespace", namespace, "logStoreName", logStoreName)
	cmd := exec.Command(script, workDir, namespace, logStoreName)
	out, err := cmd.Output()
	result := string(out)
	log.V(3).Info("Cert generation", "out", result, "err", err)
	if result != "" {
		updated = true
		log.Info("cert_generation output", "output", result)
	}
	log.V(3).Info("Returning", "err", err, "updated", updated)
	return err, updated
}
