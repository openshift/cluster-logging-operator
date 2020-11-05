package certificates

import (
	"fmt"
	"os/exec"

	"github.com/ViaQ/logerr/log"
)

func GenerateCertificates(namespace, scriptsDir, logStoreName, workDir string) (err error) {
	script := fmt.Sprintf("%s/cert_generation.sh", scriptsDir)
	return RunCertificatesScript(namespace, logStoreName, workDir, script)
}

func RunCertificatesScript(namespace, logStoreName, workDir, script string) (err error) {
	log.V(2).Info("Running script", "script", script, "WORKING_DIR", workDir, "namespace", namespace, "logstorename", logStoreName)
	cmd := exec.Command(script, workDir, namespace, logStoreName)
	_, err = cmd.Output()
	return err
}
