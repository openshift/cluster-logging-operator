package functional

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"text/template"

	log "github.com/ViaQ/logerr/v2/log/static"
	logging "github.com/openshift/cluster-logging-operator/api/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	"github.com/openshift/cluster-logging-operator/test/helpers/oc"
	"github.com/openshift/cluster-logging-operator/test/helpers/rand"
	"k8s.io/apimachinery/pkg/util/wait"
)

const (
	SplunkImage        = "quay.io/openshift-logging/splunk:9.0.0"
	SplunkHecPort      = 8088
	SplunkIndexName    = "fooIndex"
	SplunkIndexKeyName = "log_type"
	SplunkDefaultIndex = "main"
)

var (
	HecToken      = rand.Word(16)
	AdminPassword = rand.Word(16)

	configTemplateName = "splunkserver"
	ConfigTemplate     = fmt.Sprintf(`
splunk:
  hec:
    ssl: false
    token: "{{ string .Token }}"
  password: "{{ string .Password }}"
  idxc_secret: "{{ string .IdxcSecret }}"
  shc_secret: "{{ string .SHCSecret }}"
  conf:
    - key: indexes
      value:
        directory: /opt/splunk/etc/system/local/
        content:
          %s:
            homePath: $SPLUNK_DB/%s/db
            coldPath: $SPLUNK_DB/%s/colddb
            thawedPath: $SPLUNK_DB/%s/thaweddb
          %s:
            homePath: $SPLUNK_DB/%s/db
            coldPath: $SPLUNK_DB/%s/colddb
            thawedPath: $SPLUNK_DB/%s/thaweddb
`, SplunkIndexName, SplunkIndexName, SplunkIndexName, SplunkIndexName, logging.InputNameApplication, logging.InputNameApplication, logging.InputNameApplication, logging.InputNameApplication)
	SplunkEndpointHTTP = fmt.Sprintf("http://localhost:%d", SplunkHecPort)
)

func (f *CollectorFunctionalFramework) AddSplunkOutput(b *runtime.PodBuilder, output logging.OutputSpec) error {
	data, err := GenerateConfigmapData()
	if err != nil {
		return err
	}
	config := runtime.NewConfigMap(b.Pod.Namespace, logging.OutputTypeSplunk, data)
	log.V(2).Info("Creating configmap", "namespace", config.Namespace, "name", config.Name)
	if err := f.Test.Client.Create(config); err != nil {
		return err
	}
	cb := b.AddContainer(logging.OutputTypeSplunk, SplunkImage).
		AddContainerPort("http-splunkweb", 8000).
		AddContainerPort("http-hec", SplunkHecPort).
		AddContainerPort("https-splunkd", 8089).
		AddContainerPort("tcp-s2s", 9097).
		AddEnvVar("SPLUNK_DECLARATIVE_ADMIN_PASSWORD", "true").
		AddEnvVar("SPLUNK_DEFAULTS_URL", "/mnt/splunk-secrets/default.yml").
		AddEnvVar("SPLUNK_DECLARATIVE_ADMIN_PASSWORD", "true").
		AddEnvVar("SPLUNK_HOME_OWNERSHIP_ENFORCEMENT", "false").
		AddEnvVar("SPLUNK_ROLE", "splunk_standalone").
		AddEnvVar("SPLUNK_START_ARGS", "--accept-license").
		AddVolumeMount(config.Name, "/mnt/splunk-secrets", "", true).
		AddVolumeMount("optvar", "/opt/splunk/var", "", false).
		AddVolumeMount("optetc", "/opt/splunk/etc", "", false).
		WithPrivilege()

	cb.End()
	b.AddConfigMapVolume(config.Name, config.Name)
	b.AddEmptyDirVolume("optvar")
	b.AddEmptyDirVolume("optetc")
	return nil
}

func GenerateConfigmapData() (data map[string]string, err error) {
	b := &bytes.Buffer{}
	t := template.Must(
		template.New(configTemplateName).
			Funcs(template.FuncMap{
				"string": func(arg []byte) string {
					return string(arg)
				},
			}).
			Parse(ConfigTemplate),
	)
	if err = t.Execute(b,
		struct {
			Token        []byte
			Password     []byte
			Pass4SymmKey []byte
			IdxcSecret   []byte
			SHCSecret    []byte
		}{
			Token:        HecToken,
			Password:     AdminPassword,
			Pass4SymmKey: []byte("o4a9itWyG1YECvxpyVV9faUO"),
			IdxcSecret:   []byte("5oPyAqIlod4sxH1Xk7fZpNe4"),
			SHCSecret:    []byte("77mwFNOSUzmQLG9EGa2ZVEFq"),
		},
	); err != nil {
		log.V(3).Error(err, "Error executing template")
		return data, err
	}
	data = map[string]string{
		"default.yml": b.String(),
	}

	return data, nil
}

func (f *CollectorFunctionalFramework) ReadLogsByTypeFromSplunk(namespace, name, logType string) (results []string, err error) {
	var output string
	cmd := fmt.Sprintf(`/opt/splunk/bin/splunk search log_type=%s -auth "admin:%s"`, logType, AdminPassword)
	err = wait.PollUntilContextTimeout(context.TODO(), defaultRetryInterval, f.GetMaxReadDuration(), true, func(cxt context.Context) (done bool, err error) {
		output, err = oc.Exec().WithNamespace(namespace).Pod(name).Container(logging.OutputTypeSplunk).WithCmd("/bin/sh", "-c", cmd).Run()
		if output == "" || err != nil {
			return false, err
		}
		return true, nil
	})

	if err != nil {
		return nil, err
	}

	results = strings.Split(output, "\n")
	return results, nil
}

func (f *CollectorFunctionalFramework) ReadAppLogsByIndexFromSplunk(namespace, name, index string) (results []string, err error) {
	var output string
	cmd := fmt.Sprintf(`/opt/splunk/bin/splunk search index=%s -auth "admin:%s"`, index, AdminPassword)
	err = wait.PollUntilContextTimeout(context.TODO(), defaultRetryInterval, f.GetMaxReadDuration(), true, func(cxt context.Context) (done bool, err error) {
		output, err = oc.Exec().WithNamespace(namespace).Pod(name).Container(logging.OutputTypeSplunk).WithCmd("/bin/sh", "-c", cmd).Run()
		if output == "" || err != nil {
			return false, err
		}
		return true, nil
	})

	if err != nil {
		return nil, err
	}

	results = strings.Split(output, "\n")
	return results, nil
}
