package functional

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/test/helpers/oc"
	testruntime "github.com/openshift/cluster-logging-operator/test/runtime"
	"k8s.io/apimachinery/pkg/util/wait"

	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/generator/url"
	"github.com/openshift/cluster-logging-operator/internal/runtime"

	log "github.com/ViaQ/logerr/v2/log/static"
)

type ElasticsearchVersion int

const (
	ElasticsearchVersion6 ElasticsearchVersion = 6
	ElasticsearchVersion7 ElasticsearchVersion = 7
	ElasticsearchVersion8 ElasticsearchVersion = 8
	ElasticPassword                            = "default-password"
	ElasticUsername                            = "elastic"
)

var (
	logTypeIndexMap  = map[string]string{}
	esVersionToImage = map[ElasticsearchVersion]string{
		ElasticsearchVersion6: "quay.io/openshift-logging/elasticsearch:6.8.23",
		ElasticsearchVersion7: "quay.io/openshift-logging/elasticsearch:7.17.28",
		ElasticsearchVersion8: "quay.io/openshift-logging/elasticsearch:8.17.5",
	}
)

func init() {
	for _, t := range obs.InputTypes {
		logTypeIndexMap[string(t)] = fmt.Sprintf("%v-write", t)
	}
}

func (f *CollectorFunctionalFramework) AddESOutput(version ElasticsearchVersion, b *runtime.PodBuilder, output obs.OutputSpec, envVars map[string]string) error {
	log.V(2).Info("Adding elasticsearch output", "name", output.Name, "version", version)
	name := strings.ToLower(output.Name)

	esURL, err := url.Parse(output.Elasticsearch.URL)
	if err != nil {
		return err
	}
	transportPort, portError := determineTransportPort(esURL.Port())
	if portError != nil {
		return portError
	}

	log.V(2).Info("Adding container", "name", name)
	log.V(2).Info("Adding ElasticSearch output container", "name", obs.OutputTypeElasticsearch)

	esCont := b.AddContainer(name, esVersionToImage[version]).
		AddEnvVar("discovery.type", "single-node").
		AddEnvVar("http.port", esURL.Port()).
		AddEnvVar("transport.port", transportPort).
		AddEnvVar("HOME", "/tmp").
		AddEnvVar("ES_JAVA_OPTS", "-Xms256m -Xmx256m").
		AddRunAsUser(2000).
		AddEnvVar("xpack.security.enabled", "false")

	for k, v := range envVars {
		esCont.AddEnvVar(k, v)
	}

	esCont.End()
	return nil
}

func determineTransportPort(port string) (string, error) {
	iPort, err := strconv.Atoi(port)
	if err != nil {
		return "", err
	}
	iPort = iPort + 100
	return strconv.Itoa(iPort), nil
}

func (f *CollectorFunctionalFramework) AddESOutputWithBasicSecurity(password string, b *runtime.PodBuilder, output obs.OutputSpec) error {
	env := map[string]string{
		"xpack.security.enabled": "true",
		"ELASTIC_PASSWORD":       password,
	}

	return f.AddESOutput(ElasticsearchVersion8, b, output, env)
}

func (f *CollectorFunctionalFramework) AddESOutputWithTokenSecurity(b *runtime.PodBuilder, output obs.OutputSpec) error {
	envVars := map[string]string{
		"xpack.security.enabled":             "true",
		"xpack.security.authc.token.enabled": "true",
		"xpack.license.self_generated.type":  "trial",
		"ELASTIC_PASSWORD":                   ElasticPassword,
	}

	return f.AddESOutput(ElasticsearchVersion8, b, output, envVars)
}

func (f *CollectorFunctionalFramework) GetLogsFromElasticSearch(outputName string, outputLogType string, options ...Option) (results []string, err error) {
	index, ok := logTypeIndexMap[outputLogType]
	if !ok {
		return []string{}, fmt.Errorf("can't find log of type %s", outputLogType)
	}
	return f.GetLogsFromElasticSearchIndex(outputName, index, options...)
}

func getAuth(options []Option) string {
	var (
		username string
		password string
	)

	// Check for token
	if found, o := OptionsInclude("token", options); found {
		return fmt.Sprintf(`-H "Authorization: Bearer %s"`, o.Value)
	}

	// Check for basic username/password auth
	if found, o := OptionsInclude("username", options); found {
		username = o.Value
	}
	if found, o := OptionsInclude("password", options); found {
		password = o.Value
	}

	if username != "" && password != "" {
		return fmt.Sprintf(`-u '%s':'%s'`, username, password)
	}

	return ""
}

func (f *CollectorFunctionalFramework) GetLogsFromElasticSearchIndex(outputName string, index string, options ...Option) (results []string, err error) {
	port := "9200"
	if found, o := OptionsInclude("port", options); found {
		port = o.Value
	}

	pod := runtime.NewPod(f.Test.NS.Name, outputName)
	if err = f.Test.Get(pod); err != nil {
		pod = f.Pod
	}

	err = wait.PollUntilContextTimeout(context.TODO(), defaultRetryInterval, maxDuration, true, func(cxt context.Context) (done bool, err error) {
		cmd := fmt.Sprintf(`curl -X GET "localhost:%s/%s/_search?pretty" %s -H "Content-Type: application/json"  -d'{"query": {"match_all": {}}}'`,
			port, index, getAuth(options))
		var result string
		result, err = f.RunCommandInPod(pod, outputName, "bash", "-c", cmd)
		if result != "" && err == nil {
			var elasticResult map[string]interface{}
			log.V(2).Info("results", "response", result)
			err = json.Unmarshal([]byte(result), &elasticResult)
			if err == nil {
				if elasticResult["timed_out"] == false {
					rawHits, ok := elasticResult["hits"]
					if !ok {
						return false, fmt.Errorf("no hits found")
					}
					resultHits := rawHits.(map[string]interface{})
					total, ok := resultHits["total"].(map[string]interface{})
					if ok {
						if int(total["value"].(float64)) == 0 {
							return false, nil
						}
					} else {
						if resultHits["total"].(float64) == 0 {
							return false, nil
						}
					}
					hits := resultHits["hits"].([]interface{})
					for i := 0; i < len(hits); i++ {
						hit := hits[i].(map[string]interface{})
						jsonHit, err := json.Marshal(hit["_source"])
						if err == nil {
							results = append(results, string(jsonHit))
						} else {
							log.V(4).Info("Marshall failed", "err", err)
						}
					}
					return true, nil
				}
			} else {
				log.V(4).Info("Unmarshall failed", "err", err)
			}
		}
		log.V(4).Info("Polling from ElasticSearch", "err", err, "result", result)
		return false, nil
	})
	log.V(4).Info("Returning", "logs", results)
	return results, err
}

func (f *CollectorFunctionalFramework) GenerateESAccessToken(container string, options ...Option) (token string, err error) {
	log.V(2).Info(fmt.Sprintf("generating %s token", container))
	var (
		auth string
	)

	port := "9200"
	if found, o := OptionsInclude("port", options); found {
		port = o.Value
	}

	auth = fmt.Sprintf(`-u '%s':'%s'`, ElasticUsername, ElasticPassword)

	esPod := runtime.NewPod(f.Test.NS.Name, string(obs.OutputTypeElasticsearch))
	if err := f.Test.Get(esPod); err != nil {
		return "", err
	}

	err = wait.PollUntilContextTimeout(context.TODO(), defaultRetryInterval, maxDuration, true, func(cxt context.Context) (done bool, err error) {
		cmd := fmt.Sprintf(`curl %s -X POST "localhost:%s/_security/oauth2/token" -H 'Content-Type: application/json' -d'{
"grant_type": "password",
"username": "%s",
"password": "%s"
}'`, auth, port, ElasticUsername, ElasticPassword)
		var result string
		result, err = testruntime.ExecOc(esPod, strings.ToLower(container), "bash", "-c", cmd)
		if result != "" && err == nil {
			var elasticResult map[string]interface{}
			log.V(2).Info("results", "response", result)
			err = json.Unmarshal([]byte(result), &elasticResult)
			if err == nil {
				if elasticResult["access_token"] != "" {
					hit, err := json.Marshal(elasticResult["access_token"])
					if err == nil {
						token = string(hit)
					}
					return true, nil
				}
			} else {
				log.V(4).Info("Unmarshall failed", "err", err)
			}
		}
		log.V(4).Info("Polling from ElasticSearch", "err", err, "token", token)
		return false, nil
	})
	log.V(4).Info("Returning", "token", token)
	return token, err
}

func (f *CollectorFunctionalFramework) DeployESTokenPodWithService() (err error) {
	out := f.Forwarder.Spec.Outputs[0]
	log.V(2).Info(fmt.Sprintf("deploying pod and service for %s", out.Name))
	esPod := runtime.NewPod(f.Test.NS.Name, out.Name)

	esPodLabels := f.Labels
	esPodLabels[constants.LabelK8sComponent] = out.Name

	b := runtime.NewPodBuilder(esPod).
		WithLabels(esPodLabels)

	if err = f.AddESOutputWithTokenSecurity(b, out); err != nil {
		return err
	}

	if err = f.Test.Create(esPod); err != nil {
		return err
	}

	if err = oc.Literal().From("oc wait -n %s pod/%s --timeout=120s --for=condition=Ready", f.Test.NS.Name, out.Name).Output(); err != nil {
		return err
	}

	if err = f.Test.Get(esPod); err != nil {
		return err
	}

	if err = f.DeployESService(out.Name, esPodLabels); err != nil {
		return err
	}

	return err
}

func (f *CollectorFunctionalFramework) DeployESService(name string, labels map[string]string) error {
	log.V(2).Info(fmt.Sprintf("creating service for %s", name))
	service := runtime.NewService(f.Test.NS.Name, name)
	runtime.NewServiceBuilder(service).
		AddServicePort(9200, 9200).
		WithSelector(labels)

	if err := f.Test.Create(service); err != nil {
		return err
	}
	log.V(2).Info("waiting for service endpoints to be ready")
	err := wait.PollUntilContextTimeout(context.TODO(), time.Second*2, time.Second*10, true, func(cxt context.Context) (done bool, err error) {
		ips, err := oc.Get().WithNamespace(f.Test.NS.Name).Resource("endpoints", name).OutputJsonpath("{.subsets[*].addresses[*].ip}").Run()
		if err != nil {
			return false, nil
		}
		// if there are IPs in the service endpoint, the service is available
		if strings.TrimSpace(ips) != "" {
			return true, nil
		}
		return false, nil
	})

	if err != nil {
		return fmt.Errorf("service could not be started")
	}

	return nil
}
