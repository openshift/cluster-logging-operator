package splunk

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	. "github.com/onsi/ginkgo"
	loggingv1 "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/test/client"
	framework "github.com/openshift/cluster-logging-operator/test/framework/e2e"
	"github.com/openshift/cluster-logging-operator/test/helpers"
	"github.com/openshift/cluster-logging-operator/test/helpers/oc"
	"github.com/openshift/cluster-logging-operator/test/helpers/types"
	"github.com/openshift/cluster-logging-operator/test/runtime"
	"github.com/stretchr/testify/require"
	"golang.org/x/sync/errgroup"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strings"
	"testing"
	"time"
)

const (
	splunkReceiverName = "splunk-receiver"
	message            = "To be, or not to be, that is the question"
	logGenerator       = "log-generator"
)

var (
	e2e    = framework.NewE2ETestFramework()
	secret = corev1.Secret{

		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: corev1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "vector-splunk-secret",
			Namespace: constants.OpenshiftNS,
		},
		Type: "Opaque",
		Data: map[string][]byte{
			"hecToken": framework.HecToken,
		},
	}
	spec = loggingv1.ClusterLogForwarderSpec{
		Outputs: []loggingv1.OutputSpec{{
			Name: splunkReceiverName,
			Type: loggingv1.OutputTypeSplunk,
			URL:  framework.SplunkEndpoint.String(),
			OutputTypeSpec: loggingv1.OutputTypeSpec{
				Splunk: &loggingv1.Splunk{},
			},
			Secret: &loggingv1.OutputSecretSpec{
				Name: "vector-splunk-secret",
			},
		}},
		Pipelines: []loggingv1.PipelineSpec{
			{
				Name:       "test-app",
				InputRefs:  []string{loggingv1.InputNameApplication, loggingv1.InputNameInfrastructure},
				OutputRefs: []string{splunkReceiverName},
			},
		},
	}
)

func TestLogForwardingToSplunkWithVector(t *testing.T) {
	_, err := e2e.KubeClient.CoreV1().Secrets(constants.OpenshiftNS).Create(context.TODO(), &secret, metav1.CreateOptions{})
	if err != nil {
		if apierrors.IsAlreadyExists(err) {
			sOpts := metav1.UpdateOptions{}
			_, err := e2e.KubeClient.CoreV1().Secrets(constants.OpenshiftNS).Update(context.TODO(), &secret, sOpts)
			if err != nil {
				Fail(fmt.Sprintf("Unable to create secret for Vector: %v", err))
			}
		} else {
			fmt.Println(err)
			Fail(fmt.Sprintf("Unable to create secret for Vector: %v", err))
		}
	}
	if _, err := e2e.DeploySplunk(); err != nil {
		Fail(fmt.Sprintf("Unable to deploy splunk receiver: %v", err))
	}
	cl := runtime.NewClusterLogging()
	cl.Spec.Collection.Type = loggingv1.LogCollectionTypeVector
	cl.Spec.Collection.CollectorSpec = loggingv1.CollectorSpec{}
	clf := runtime.NewClusterLogForwarder()
	clf.Spec = spec
	testLogForwardingToSplunk(t, cl, clf)

}

func testLogForwardingToSplunk(t *testing.T, cl *loggingv1.ClusterLogging, clf *loggingv1.ClusterLogForwarder) {
	c := client.ForTest(t)
	defer e2e.Cleanup()
	gen := runtime.NewLogGenerator(c.NS.Name, logGenerator, 3, 1, message)
	clf.Spec.Outputs[0].URL = framework.SplunkEndpoint.String()

	var g errgroup.Group
	g.Go(func() error { return c.Recreate(cl) })
	defer func(r *loggingv1.ClusterLogging) { _ = c.Delete(r) }(cl)
	g.Go(func() error { return c.Recreate(clf) })
	defer func(r *loggingv1.ClusterLogForwarder) { _ = c.Delete(r) }(clf)
	g.Go(func() error { return c.Create(gen) })
	require.NoError(t, g.Wait())
	require.NoError(t, c.WaitFor(clf, client.ClusterLogForwarderReady))
	require.NoError(t, e2e.WaitFor(helpers.ComponentTypeCollector))
	time.Sleep(1 * time.Minute)
	splunkPod, err := oc.Get().WithNamespace(constants.OpenshiftNS).Pod().Selector("app.kubernetes.io/component=" + framework.SplunkStandalone).OutputJsonpath("{.items[*].metadata.name}").Run()
	fmt.Println(splunkPod)
	require.NoError(t, err)
	require.NotEmpty(t, splunkPod)

	err, logs := readApplicationLogs(t, splunkPod)
	require.NoError(t, err)
	require.Equal(t, 3, len(logs))
	for _, log := range logs {
		require.Contains(t, log.Message, message)
		require.Equal(t, log.Kubernetes.ContainerName, logGenerator)
		require.Equal(t, log.LogType, "application")
	}

	err, infraLogs := readInfraLogs(t, splunkPod)
	require.NoError(t, err)
	require.Equal(t, 5, len(infraLogs))
	for _, infraLog := range infraLogs {
		require.Equal(t, infraLog.LogType, "infrastructure")
	}
}

func readApplicationLogs(t *testing.T, podName string) (error, []types.ApplicationLog) {
	cmd := fmt.Sprintf("/opt/splunk/bin/splunk search \"%s\" -auth \"admin:%s\"", message, framework.AdminPassword)
	output, err := oc.Exec().WithNamespace(constants.OpenshiftNS).Pod(podName).Container(framework.Splunk).WithCmd("/bin/sh", "-c", cmd).Run()
	require.NotEmpty(t, output)
	require.NoError(t, err)
	output = "[" + strings.Replace(output, "}\n{", "},{", -1) + "]"
	var logs []types.ApplicationLog
	dec := json.NewDecoder(bytes.NewBufferString(output))
	err = dec.Decode(&logs)
	require.NoError(t, err)
	return err, logs
}

func readInfraLogs(t *testing.T, podName string) (error, []types.InfraLog) {
	cmd := fmt.Sprintf("/opt/splunk/bin/splunk search 'log_type=infrastructure' -maxout 5 -auth \"admin:%s\"", framework.AdminPassword)
	output, err := oc.Exec().WithNamespace(constants.OpenshiftNS).Pod(podName).Container(framework.Splunk).WithCmd("/bin/sh", "-c", cmd).Run()
	require.NotEmpty(t, output)
	require.NoError(t, err)
	output = "[" + strings.Replace(output, "}\n{", "},{", -1) + "]"
	var logs []types.InfraLog
	dec := json.NewDecoder(bytes.NewBufferString(output))
	err = dec.Decode(&logs)
	require.NoError(t, err)
	return err, logs
}
