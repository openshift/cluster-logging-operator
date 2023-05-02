package flowcontrol

import (
	"fmt"
	"path/filepath"

	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/test/framework/functional"
)

var (
	LokiNsQuery     = `{kubernetes_namespace_name=~"%s"}`
	LokiAuditQuery  = `{log_type="audit"}`
	AllLogs         = `.+`
	MessageTemplate = "$(date -u +'%Y-%m-%dT%H:%M:%S.%N%:z') stdout F $msg "
	FileTemplate    = "%s/%s_%s_%s/%s/0.log"
	LogPath         = "/var/log/pods"
)

func WriteApplicationLogs(f *functional.CollectorFunctionalFramework, numOfLogs int) error {
	file := fmt.Sprintf(FileTemplate, LogPath, f.Pod.Namespace, f.Pod.Name, f.Pod.UID, constants.CollectorName)
	return WriteLogs(f, file, numOfLogs, 100)
}

func WriteLogs(f *functional.CollectorFunctionalFramework, filename string, numOfLogs, size int) error {
	logPath := filepath.Dir(filename)
	_, err := f.RunCommand(
		constants.CollectorName, "bash", "-c",
		fmt.Sprintf("bash -c 'mkdir -p %s;msg=$(cat /dev/urandom|tr -dc 'a-zA-Z0-9'|fold -w %d|head -n 1);for n in $(seq 1 %d);do echo %s >> %s; sleep %ds; done'", logPath, size, numOfLogs, MessageTemplate, filename, 1/numOfLogs))

	return err
}
