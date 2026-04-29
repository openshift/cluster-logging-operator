package datacollector

import (
	_ "embed"
	"fmt"

	"github.com/openshift/cluster-logging-operator/internal/runtime"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	"github.com/openshift/cluster-logging-operator/test/framework/functional"
	"github.com/openshift/cluster-logging-operator/test/helpers/mockoon"
	"github.com/openshift/cluster-logging-operator/test/helpers/oc"
	"github.com/openshift/cluster-logging-operator/test/helpers/types"
	v1 "k8s.io/api/core/v1"
)

//go:embed azure-http-data-collector-api.json
var apiConfig string

const (
	apiJsonFile = "azure-http-data-collector-api.json"
	SecretName  = "azure-secret"
)

func NewMockoonVisitor(pb *runtime.PodBuilder, azureAltHost string, framework *functional.CollectorFunctionalFramework) error {
	configMap := runtime.NewConfigMap(framework.Namespace, mockoon.ContainerName, map[string]string{})
	runtime.NewConfigMapBuilder(configMap).Add(apiJsonFile, apiConfig)
	if err := framework.Test.Create(configMap); err != nil {
		return err
	}

	hostAlias := v1.HostAlias{
		IP:        "127.0.0.1",
		Hostnames: []string{azureAltHost},
	}

	mountPath := "/data"
	pb.AddConfigMapVolume("data", mockoon.ContainerName).
		AddHostAlias(hostAlias).
		AddContainer(mockoon.ContainerName, mockoon.Image).
		AddContainerPort(mockoon.ContainerName, mockoon.Port).
		WithCmdArgs([]string{
			fmt.Sprintf("--data=%s/%s", mountPath, apiJsonFile),
			"--log-transaction",
		}).AddVolumeMount("data", mountPath, "", true).End()

	return nil
}

func ReadApplicationLog(framework *functional.CollectorFunctionalFramework) ([]types.ApplicationLog, error) {
	output, err := oc.Literal().From("oc logs -n %s pod/%s -c %s", framework.Test.NS.Name, framework.Name, mockoon.ContainerName).Run()
	if err != nil {
		return nil, err
	}
	return extractStructuredLogs(output, "application")
}

func extractStructuredLogs(output, logType string) ([]types.ApplicationLog, error) {
	logs, err := mockoon.DecodeLogs(output)
	if err != nil {
		return nil, err
	}

	var appLogs []types.ApplicationLog
	for _, log := range logs {
		if log.RequestMethod == "POST" && log.ResponseStatus == 200 {
			appLog := log.Transaction.Request.Body
			var tmp []types.ApplicationLog
			if err := types.ParseLogsFrom(utils.ToJsonLogs([]string{appLog}), &tmp, false); err != nil {
				fmt.Printf("error parsing log: %v\n", err)
				continue
			}

			for _, applicationLog := range tmp {
				if applicationLog.LogType == logType {
					appLogs = append(appLogs, applicationLog)
				}
			}
		}
	}

	return appLogs, nil
}
