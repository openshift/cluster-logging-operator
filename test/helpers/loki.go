package helpers

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/openshift/cluster-logging-operator/pkg/logger"
	"github.com/openshift/cluster-logging-operator/test/helpers/loki"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
)

type lokiReceiver struct {
	app *appsv1.StatefulSet
	tc  *E2ETestFramework
}

func (lr *lokiReceiver) ApplicationLogs(timeout time.Duration) (logs, error) {
	rcvName := lr.app.GetName()
	res, err := lr.tc.lokiLogs(rcvName, ProjectIndexPrefix)
	if err != nil {
		return nil, err
	}
	return ParseLogs(res.ToString())
}

func (lr *lokiReceiver) HasInfraStructureLogs(timeout time.Duration) (bool, error) {
	err := wait.Poll(defaultRetryInterval, timeout, func() (done bool, err error) {
		rcvName := lr.app.GetName()
		res, err := lr.tc.lokiLogs(rcvName, InfraIndexPrefix)
		if err != nil {
			return false, err
		}
		return res.NonEmpty(), nil
	})
	return true, err
}

func (lr *lokiReceiver) HasApplicationLogs(timeout time.Duration) (bool, error) {
	err := wait.Poll(defaultRetryInterval, timeout, func() (done bool, err error) {
		rcvName := lr.app.GetName()
		res, err := lr.tc.lokiLogs(rcvName, ProjectIndexPrefix)
		if err != nil {
			return false, err
		}
		return res.NonEmpty(), nil
	})
	return true, err
}

func (lr *lokiReceiver) HasAuditLogs(timeout time.Duration) (bool, error) {
	err := wait.Poll(defaultRetryInterval, timeout, func() (done bool, err error) {
		rcvName := lr.app.GetName()
		res, err := lr.tc.lokiLogs(rcvName, AuditIndexPrefix)
		if err != nil {
			return false, err
		}
		return res.NonEmpty(), nil
	})
	return true, err
}

func (kr *lokiReceiver) GrepLogs(expr string, timeToWait time.Duration) (string, error) {
	return "Not Found", fmt.Errorf("Not implemented")
}

func (kr *lokiReceiver) ClusterLocalEndpoint() string {
	return loki.ClusterLocalEndpoint(OpenshiftLoggingNS)

}

func (tc *E2ETestFramework) lokiLogs(rcvName, indexName string) (*loki.Response, error) {
	opts := metav1.GetOptions{}
	pod, err := tc.KubeClient.CoreV1().Pods(OpenshiftLoggingNS).Get(context.TODO(), loki.QuerierName, opts)
	if err != nil {
		return nil, err
	}
	logger.Debugf("Pod %s", pod.GetName())

	indexName = fmt.Sprintf("%swrite", indexName)
	cmd := []string{"/bin/sh", "/data/loki_util", tc.LogStores[rcvName].ClusterLocalEndpoint(), indexName}
	stdout, err := tc.PodExec(OpenshiftLoggingNS, loki.QuerierName, loki.QuerierName, cmd)
	if err != nil {
		fmt.Println(strings.Join(cmd, " "))
		return nil, err
	}

	res, err := loki.Parse(stdout)
	if err != nil {
		fmt.Println(strings.Join(cmd, " "))
		return nil, err
	}

	return &res, nil
}

func (tc *E2ETestFramework) DeployLokiReceiver() (*appsv1.StatefulSet, error) {
	if err := tc.createLokiConfigMap(); err != nil {
		return nil, err
	}

	app, err := tc.createLokiStatefulSet()
	if err != nil {
		return nil, err
	}

	if err := tc.createLokiService(); err != nil {
		return nil, err
	}

	if err := tc.createLokiQuerier(); err != nil {
		return nil, err
	}

	receiver := &lokiReceiver{
		tc:  tc,
		app: app,
	}
	tc.LogStores[app.GetName()] = receiver

	return app, nil
}

func (tc *E2ETestFramework) createLokiConfigMap() error {
	cm := loki.NewConfigMap(OpenshiftLoggingNS)

	tc.AddCleanup(func() error {
		var zerograce int64
		opts := metav1.DeleteOptions{
			GracePeriodSeconds: &zerograce,
		}
		return tc.KubeClient.CoreV1().ConfigMaps(OpenshiftLoggingNS).Delete(context.TODO(), cm.GetName(), opts)
	})

	opts := metav1.CreateOptions{}
	if _, err := tc.KubeClient.CoreV1().ConfigMaps(OpenshiftLoggingNS).Create(context.TODO(), cm, opts); err != nil {
		return err
	}

	return nil
}

func (tc *E2ETestFramework) createLokiStatefulSet() (*appsv1.StatefulSet, error) {
	app := loki.NewStatefulSet(OpenshiftLoggingNS)

	tc.AddCleanup(func() error {
		var zerograce int64
		opts := metav1.DeleteOptions{
			GracePeriodSeconds: &zerograce,
		}
		return tc.KubeClient.AppsV1().StatefulSets(OpenshiftLoggingNS).Delete(context.TODO(), app.GetName(), opts)
	})

	opts := metav1.CreateOptions{}
	app, err := tc.KubeClient.AppsV1().StatefulSets(OpenshiftLoggingNS).Create(context.TODO(), app, opts)
	if err != nil {
		return nil, err
	}

	return app, tc.waitForStatefulSet(OpenshiftLoggingNS, app.GetName(), defaultRetryInterval, defaultTimeout)
}

func (tc *E2ETestFramework) createLokiService() error {
	svc := loki.NewService(OpenshiftLoggingNS)

	tc.AddCleanup(func() error {
		var zerograce int64
		opts := metav1.DeleteOptions{
			GracePeriodSeconds: &zerograce,
		}
		return tc.KubeClient.CoreV1().Services(OpenshiftLoggingNS).Delete(context.TODO(), svc.GetName(), opts)
	})

	opts := metav1.CreateOptions{}
	if _, err := tc.KubeClient.CoreV1().Services(OpenshiftLoggingNS).Create(context.TODO(), svc, opts); err != nil {
		return err
	}

	return nil
}

func (tc *E2ETestFramework) createLokiQuerier() error {
	cm := loki.NewQuerierConfigMap(OpenshiftLoggingNS)
	tc.AddCleanup(func() error {
		opts := metav1.DeleteOptions{}
		return tc.KubeClient.CoreV1().ConfigMaps(OpenshiftLoggingNS).Delete(context.TODO(), cm.GetName(), opts)
	})

	opts := metav1.CreateOptions{}
	_, err := tc.KubeClient.CoreV1().ConfigMaps(OpenshiftLoggingNS).Create(context.TODO(), cm, opts)
	if err != nil {
		return err
	}

	pod := loki.NewQuerierPod(OpenshiftLoggingNS)
	tc.AddCleanup(func() error {
		opts := metav1.DeleteOptions{}
		return tc.KubeClient.AppsV1().Deployments(OpenshiftLoggingNS).Delete(context.TODO(), pod.GetName(), opts)
	})

	opts = metav1.CreateOptions{}
	pod, err = tc.KubeClient.CoreV1().Pods(OpenshiftLoggingNS).Create(context.TODO(), pod, opts)
	if err != nil {
		return err
	}
	return tc.waitForPod(OpenshiftLoggingNS, loki.QuerierName, defaultRetryInterval, defaultTimeout)
}
