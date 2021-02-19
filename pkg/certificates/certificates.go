package certificates

import (
	"fmt"
	"os/exec"

	"github.com/ViaQ/logerr/log"
	"sigs.k8s.io/yaml"
)

func GenerateCertificates(namespace, scriptsDir, logStoreName, workDir string) (err error, updated bool) {
	script := fmt.Sprintf("%s/cert_generation.sh", scriptsDir)
	return RunCertificatesScript(namespace, logStoreName, workDir, script)
}

func RunCertificatesScript(namespace, logStoreName, workDir, script string) (err error, updated bool) {
	log.V(3).Info("Running script", "script", script, "workDir", workDir, "namespace", namespace, "logstorename", logStoreName)
	cmd := exec.Command(script, workDir, namespace, logStoreName)
	out, err := cmd.Output()
	result := string(out)
	log.V(3).Info("Cert generation result", "result", result, "error", err)
	if result != "" {
		updated = true
		dumpLogs(result)
	}
	log.V(3).Info("Returning", "error", err, "updated", updated)
	return err, updated
}

func dumpLogs(raw string) {
	genLogs := []map[string]string{}
	err := yaml.Unmarshal([]byte(raw), &genLogs)
	if err == nil {
		for _, eventlog := range genLogs {
			log.Info("cert_generation output", eventlog)
		}
		return
	}
	log.Error(err, "Unable to unmarshal cert_generation to structured output")
	log.Info("cert_generation", "raw_output", raw)
}
