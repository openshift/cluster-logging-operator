package functional

import (
	"bufio"
	"context"
	"fmt"
	"strings"

	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"k8s.io/apimachinery/pkg/util/wait"

	"github.com/openshift/cluster-logging-operator/internal/generator/url"
	"github.com/openshift/cluster-logging-operator/internal/runtime"

	log "github.com/ViaQ/logerr/v2/log/static"
)

func (f *CollectorFunctionalFramework) AddVLOutput(b *runtime.PodBuilder, output obs.OutputSpec, args []string) error {
	log.V(2).Info("Adding output for victorialogs", "name", output.Name)
	name := strings.ToLower(output.Name)

	port := "9428"
	switch output.Type {
	case obs.OutputTypeElasticsearch:
		u, err := url.Parse(output.Elasticsearch.URL)
		if err != nil {
			return err
		}
		port = u.Port()
	case obs.OutputTypeHTTP:
		u, err := url.Parse(output.HTTP.URL)
		if err != nil {
			return err
		}
		port = u.Port()
	}

	log.V(2).Info("Adding container", "name", name)
	log.V(2).Info("Adding VictoriaLogs output container", "name", output.Type)

	cmdArgs := []string{
		"-httpListenAddr=:" + port,
		"-storageDataPath=/tmp/logs",
	}
	if len(args) > 0 {
		cmdArgs = append(cmdArgs, args...)
	}

	b.AddContainer(name, "quay.io/victoriametrics/victoria-logs:v1.34.0").
		AddRunAsUser(2000).
		WithCmdArgs(cmdArgs).
		End().
		AddContainer(fmt.Sprintf("%s-sidecar", name), "quay.io/curl/curl:8.16.0").
		WithCmd([]string{"/bin/sleep", "infinity"}).
		AddRunAsUser(2000).
		End()
	return nil
}

func (f *CollectorFunctionalFramework) GetLogsFromVL(outputName string, headers map[string]string, options ...Option) (results []string, err error) {
	port := "9428"
	if found, o := OptionsInclude("port", options); found {
		port = o.Value
	}

	pod := runtime.NewPod(f.Test.NS.Name, outputName)
	if err = f.Test.Get(pod); err != nil {
		pod = f.Pod
	}

	err = wait.PollUntilContextTimeout(context.TODO(), defaultRetryInterval, maxDuration, true, func(cxt context.Context) (done bool, err error) {
		curlArgs := `-X GET -H 'Content-Type: application/json'`
		for k, v := range headers {
			curlArgs += fmt.Sprintf(" -H '%s: %s'", k, v)
		}
		cmd := fmt.Sprintf(`curl localhost:%s/select/logsql/query?query='*' %s`, port, curlArgs)
		var result string
		result, err = f.RunCommandInPod(pod, fmt.Sprintf("%s-sidecar", outputName), "sh", "-c", cmd)
		if result != "" && err == nil {
			scanner := bufio.NewScanner(strings.NewReader(result))
			log.V(2).Info("results", "response", results)
			for scanner.Scan() {
				results = append(results, scanner.Text())
			}
			return true, nil
		}
		log.V(4).Info("Polling from VictoriaLogs", "err", err, "result", result)
		return false, nil
	})
	log.V(4).Info("Returning", "logs", results)
	return results, err
}
