package azuremonitor

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/openshift/cluster-logging-operator/internal/runtime"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	"github.com/openshift/cluster-logging-operator/test/framework/functional"
	"github.com/openshift/cluster-logging-operator/test/helpers/oc"
	"github.com/openshift/cluster-logging-operator/test/helpers/types"
	v1 "k8s.io/api/core/v1"
)

//go:embed azure-http-data-collector-api.json
var AzureHttpDataCollectorApi string

const (
	Mockoon          = "mockoon"
	Port             = 3000
	azureApiJsonFile = "azure-http-data-collector-api.json"
	image            = "quay.io/openshift-logging/mockoon-cli:6.2.0"
	data             = "data"
)

type MockoonLog struct {
	App             string      `json:"app,omitempty"`
	EnvironmentName string      `json:"environmentName,omitempty"`
	EnvironmentUUID string      `json:"environmentUUID,omitempty"`
	Level           string      `json:"level,omitempty"`
	Message         string      `json:"message,omitempty"`
	RequestMethod   string      `json:"requestMethod,omitempty"`
	RequestPath     string      `json:"requestPath,omitempty"`
	RequestProxied  bool        `json:"requestProxied,omitempty"`
	ResponseStatus  int         `json:"responseStatus,omitempty"`
	Timestamp       time.Time   `json:"timestamp,omitempty"`
	Transaction     Transaction `json:"transaction,omitempty"`
}
type Headers struct {
	Key   string `json:"key,omitempty"`
	Value string `json:"value,omitempty"`
}
type QueryParams struct {
	APIVersion string `json:"api-version,omitempty"`
}
type Request struct {
	Body        string      `json:"body,omitempty"`
	Headers     []Headers   `json:"headers,omitempty"`
	Method      string      `json:"method,omitempty"`
	Params      []any       `json:"params,omitempty"`
	Query       string      `json:"query,omitempty"`
	QueryParams QueryParams `json:"queryParams,omitempty"`
	Route       string      `json:"route,omitempty"`
	URLPath     string      `json:"urlPath,omitempty"`
}
type Response struct {
	Body          string    `json:"body,omitempty"`
	Headers       []Headers `json:"headers,omitempty"`
	StatusCode    int       `json:"statusCode,omitempty"`
	StatusMessage string    `json:"statusMessage,omitempty"`
}
type Transaction struct {
	Proxied           bool     `json:"proxied,omitempty"`
	Request           Request  `json:"request,omitempty"`
	Response          Response `json:"response,omitempty"`
	RouteResponseUUID string   `json:"routeResponseUUID,omitempty"`
	RouteUUID         string   `json:"routeUUID,omitempty"`
}

func NewMockoonVisitor(pb *runtime.PodBuilder, azureAltHost string, framework *functional.CollectorFunctionalFramework) error {
	configMap := runtime.NewConfigMap(framework.Namespace, Mockoon, map[string]string{})
	runtime.NewConfigMapBuilder(configMap).Add(azureApiJsonFile, AzureHttpDataCollectorApi)
	if err := framework.Test.Create(configMap); err != nil {
		return err
	}

	hostAlias := v1.HostAlias{
		IP:        "127.0.0.1",
		Hostnames: []string{azureAltHost},
	}

	mountPath := "/data"
	pb.AddConfigMapVolume(data, Mockoon).
		AddHostAlias(hostAlias).
		AddContainer(Mockoon, image).
		AddContainerPort(Mockoon, Port).
		WithCmdArgs([]string{
			fmt.Sprintf("--data=%s/%s", mountPath, azureApiJsonFile),
			"--log-transaction",
		}).AddVolumeMount(data, mountPath, "", true).End()

	return nil
}

func ReadApplicationLogFromMockoon(framework *functional.CollectorFunctionalFramework) ([]types.ApplicationLog, error) {
	output, err := oc.Literal().From("oc logs -n %s pod/%s -c %s", framework.Test.NS.Name, framework.Name, Mockoon).Run()
	if err != nil {
		return nil, err
	}
	return extractStructuredLogs(output, "application")
}

// parse output and extract structured log by given log type
func extractStructuredLogs(output, logType string) ([]types.ApplicationLog, error) {
	var logs []MockoonLog

	// Preprocess JSON output to make it valid JSON array
	output = "[" + strings.Replace(output, "}\n{", "},{", -1) + "]"

	dec := json.NewDecoder(bytes.NewBufferString(output))
	if err := dec.Decode(&logs); err != nil {
		return nil, fmt.Errorf("error decoding logs: %v", err)
	}

	// Parse application logs
	var appLogs []types.ApplicationLog
	for _, log := range logs {
		if log.RequestMethod == "POST" && log.ResponseStatus == 200 {
			appLog := log.Transaction.Request.Body
			var tmp []types.ApplicationLog
			if err := types.ParseLogsFrom(utils.ToJsonLogs([]string{appLog}), &tmp, false); err != nil {
				// Log error but continue processing other logs
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
