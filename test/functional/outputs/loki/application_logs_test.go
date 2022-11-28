//go:build fluentd
// +build fluentd

package loki

import (
	"fmt"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	"github.com/openshift/cluster-logging-operator/test/framework/functional"
	"testing"
	"time"

	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/test/helpers/loki"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// NOTE: This test demonstrates how Ginkgo-like nesting & naming can be
// accomplished using the standard go testing package plus testify require/assert.
//
// The test uses the Ginkgo-like name it() to hammer home the point,
// that isn't a requirement for tests in general.
func TestLokiOutput(t *testing.T) {
	var (
		f      *functional.CollectorFunctionalFramework
		l      *loki.Receiver
		tsTime = time.Now()
		ts     = functional.CRIOTime(tsTime)
	)

	// testCase does common setup/teardown around a testFunc
	testCase := func(name string, lokiSpec *logging.Loki, testFunc func(t *testing.T)) {
		t.Helper()
		t.Run(name, func(t *testing.T) {
			f = functional.NewFluentdFunctionalFrameworkForTest(t)
			defer f.Cleanup()

			// Start a Loki server
			l = loki.NewReceiver(f.Namespace, "loki-server")
			require.NoError(t, l.Create(f.Test.Client))

			// Set up the common template forwarder configuration.
			f.Forwarder.Spec.Outputs = append(f.Forwarder.Spec.Outputs,
				logging.OutputSpec{
					Name: "loki",
					Type: "loki",
					URL:  l.InternalURL("").String(),
					OutputTypeSpec: logging.OutputTypeSpec{
						Loki: lokiSpec,
					},
				})
			f.Forwarder.Spec.Pipelines = append(f.Forwarder.Spec.Pipelines,
				logging.PipelineSpec{
					OutputRefs: []string{"loki"},
					InputRefs:  []string{"application"},
					Labels:     map[string]string{"logging": "logging-value"},
				})
			// Deploy the framework with labels needed by tests.
			require.NoError(t, f.DeployWithVisitor(func(p *runtime.PodBuilder) error {
				p.AddLabels(map[string]string{"k8s": "k8s-value"})
				return nil
			}))
			testFunc(t)
		})
	}

	testCase("forwards application logs with default labels", nil, func(t *testing.T) {
		msg := functional.NewFullCRIOLogMessage(ts, "application log message")
		require.NoError(t, f.WriteMessagesToApplicationLog(msg, 3))

		query := fmt.Sprintf(`{kubernetes_namespace_name=%q, kubernetes_pod_name=%q}`, f.Namespace, f.Name)
		r, err := l.QueryUntil(query, "", 3)
		assert.NoError(t, err)

		// Check expected Loki labels
		labels := r[0].Stream
		delete(labels, "fluentd_thread") // Added by loki plugin.
		want := map[string]string{
			"log_type":                  "application",
			"kubernetes_host":           functional.FunctionalNodeName,
			"kubernetes_namespace_name": f.Namespace,
			"kubernetes_pod_name":       f.Name,
			"kubernetes_container_name": f.Pod.Spec.Containers[0].Name,
		}
		assert.Equal(t, want, labels)

		// Check expected log records
		records := r[0].Records()
		assert.Len(t, records, 3, "expected 3 log records")
		for _, record := range records {
			assert.Equal(t, "application", record["log_type"])
			assert.Equal(t, "application log message", record["message"])
			k := record["kubernetes"].(map[string]interface{})
			assert.Equal(t, f.Namespace, k["namespace_name"], k)
			assert.Equal(t, f.Name, k["pod_name"], k)
			// Timestamp will not match exactly, some sub-second digits are truncated.
			recordTime, err := time.Parse(time.RFC3339Nano, record["@timestamp"].(string))
			assert.NoError(t, err)
			assert.WithinDuration(t, tsTime, recordTime, time.Millisecond)
			k8sLabels := k["labels"].(map[string]interface{})
			assert.Equal(t, "k8s-value", k8sLabels["k8s"], k8sLabels)
		}
	})

	testCase(
		"forwards application logs with custom Loki labels",
		&logging.Loki{LabelKeys: []string{
			"kubernetes.labels.k8s",
			"openshift.labels.logging",
			"kubernetes.container_name",
		}},
		func(t *testing.T) {
			msg := functional.NewFullCRIOLogMessage(ts, "application log message")
			require.NoError(t, f.WriteMessagesToApplicationLog(msg, 1))

			// Verify we can query by Loki labels
			query := fmt.Sprintf(`{kubernetes_labels_k8s=%q, openshift_labels_logging=%q}`, "k8s-value", "logging-value")
			r, err := l.QueryUntil(query, "", 1)
			assert.NoError(t, err, query)
			records := r[0].Records()
			assert.Len(t, records, 1)
			assert.Equal(t, "application log message", records[0]["message"])

			want := map[string]string{
				"kubernetes_container_name": f.Pod.Spec.Containers[0].Name,
				"kubernetes_labels_k8s":     "k8s-value",
				"openshift_labels_logging":  "logging-value",
				"kubernetes_host":           functional.FunctionalNodeName,
			}
			labels := r[0].Stream
			delete(labels, "fluentd_thread") // Added by loki plugin.
			assert.Equal(t, want, labels)

		})

	for _, data := range []struct {
		name   string
		spec   *logging.Loki
		tenant func() string
	}{
		{"default tenant", &logging.Loki{}, func() string { return "" }},
		{"namespace tenant", &logging.Loki{TenantKey: "kubernetes.namespace_name"}, func() string { return f.Namespace }},
		{"k8s label tenant", &logging.Loki{TenantKey: "kubernetes.label.k8s"}, func() string { return "k8s-value" }},
		{"logging label tenant", &logging.Loki{TenantKey: "openshift.label.logging"}, func() string { return "openshift-value" }},
	} {
		name, spec, tenant := data.name, data.spec, data.tenant
		testCase(name, spec, func(t *testing.T) {
			msg := functional.NewFullCRIOLogMessage(time.Now().UTC().Format(time.RFC3339Nano), name)
			require.NoError(t, f.WriteMessagesToApplicationLog(msg, 1))
			query := fmt.Sprintf(`{kubernetes_namespace_name=%q, kubernetes_pod_name=%q}`, f.Namespace, f.Name)
			r, err := l.QueryUntil(query, tenant(), 1)
			assert.NoError(t, err, query)
			records := r[0].Records()
			assert.Len(t, records, 1)
			assert.Equal(t, name, records[0]["message"])
		})
	}
}
