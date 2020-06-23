package helpers

import (
	"fmt"
	"strings"
	"time"

	loggingv1 "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/pkg/logger"
	"github.com/openshift/cluster-logging-operator/test/helpers/kafka"
	apps "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
)

type kafkaReceiver struct {
	tc     *E2ETestFramework
	app    *apps.StatefulSet
	topics []string
}

func (kr *kafkaReceiver) ApplicationLogs(timeToWait time.Duration) (string, error) {
	logs, err := kr.tc.consumedLogs(kr.app.Name, loggingv1.InputNameApplication)
	if err != nil {
		return "", fmt.Errorf("Failed to read consumed application logs: %s", err)
	}
	// FIXME(alanconway) return only app- logs here?
	return logs, nil
}

func (kr *kafkaReceiver) HasInfraStructureLogs(timeout time.Duration) (bool, error) {
	err := wait.Poll(defaultRetryInterval, timeout, func() (done bool, err error) {
		logs, err := kr.tc.consumedLogs(kr.app.Name, loggingv1.InputNameInfrastructure)
		if err != nil {
			return false, err
		}
		// FIXME(alanconway) proper check for infra- in logs here?
		return strings.Contains(logs, InfraIndexPrefix), nil
	})
	return true, err
}

func (kr *kafkaReceiver) HasApplicationLogs(timeout time.Duration) (bool, error) {
	err := wait.Poll(defaultRetryInterval, timeout, func() (done bool, err error) {
		logs, err := kr.tc.consumedLogs(kr.app.Name, loggingv1.InputNameApplication)
		if err != nil {
			return false, err
		}
		// FIXME(alanconway) proper check for infra- in logs here?
		return strings.Contains(logs, ProjectIndexPrefix), nil
	})
	return true, err
}

func (kr *kafkaReceiver) HasAuditLogs(timeout time.Duration) (bool, error) {
	err := wait.Poll(defaultRetryInterval, timeout, func() (done bool, err error) {
		logs, err := kr.tc.consumedLogs(kr.app.Name, loggingv1.InputNameAudit)
		if err != nil {
			return false, err
		}
		// FIXME(alanconway) proper check for audit- in logs here?
		return strings.Contains(logs, AuditIndexPrefix), nil
	})
	return true, err
}

func (kr *kafkaReceiver) GrepLogs(expr string, timeToWait time.Duration) (string, error) {
	return "Not Found", fmt.Errorf("Not implemented")
}

func (tc *E2ETestFramework) DeployKafkaReceiver(topics []string) (*apps.StatefulSet, error) {
	if err := tc.createZookeeper(); err != nil {
		return nil, err
	}

	app, err := tc.createKafkaBroker()
	if err != nil {
		return nil, err
	}

	receiver := &kafkaReceiver{
		tc:     tc,
		app:    app,
		topics: topics,
	}
	tc.LogStore = receiver

	if err := tc.createKafkaConsumers(receiver); err != nil {
		return nil, err
	}

	return app, nil
}

func (tc *E2ETestFramework) consumedLogs(rcvName, inputName string) (string, error) {
	rcv := tc.LogStore.(*kafkaReceiver)
	topic := kafka.TopicForInputName(rcv.topics, inputName)
	name := kafka.ConsumerNameForTopic(topic)

	options := metav1.ListOptions{
		LabelSelector: fmt.Sprintf("component=%s", name),
	}
	pods, err := tc.KubeClient.CoreV1().Pods(OpenshiftLoggingNS).List(options)
	if err != nil {
		return "", err
	}
	if len(pods.Items) == 0 {
		return "", fmt.Errorf("No pods found for %s", name)
	}

	logger.Debugf("Pod %s", pods.Items[0].Name)
	stdout, err := tc.PodExec(OpenshiftLoggingNS, pods.Items[0].Name, name, []string{"cat", "/shared/consumed.logs"})
	if err != nil {
		return "", err
	}

	// Hack Teach kafka-console-consumer to output a proper json array
	out := "[" + strings.TrimRight(strings.Replace(stdout, "\n", ",", -1), ",") + "]"
	return out, nil
}

func (tc *E2ETestFramework) createKafkaBroker() (*apps.StatefulSet, error) {
	if err := tc.createKafkaBrokerRBAC(); err != nil {
		return nil, err
	}

	if err := tc.createKafkaBrokerConfigMap(); err != nil {
		return nil, err
	}

	if err := tc.createKafkaBrokerService(); err != nil {
		return nil, err
	}

	app, err := tc.createKafkaBrokerStatefulSet()
	if err != nil {
		return nil, err
	}

	return app, nil
}

func (tc *E2ETestFramework) createZookeeper() error {
	if err := tc.createZookeeperConfigMap(); err != nil {
		return err
	}

	if _, err := tc.createZookeeperStatefulSet(); err != nil {
		return err
	}

	if err := tc.createZookeeperService(); err != nil {
		return err
	}

	return nil
}

func (tc *E2ETestFramework) createKafkaConsumers(rcv *kafkaReceiver) error {
	for _, topic := range rcv.topics {
		app := kafka.NewKafkaConsumerDeployment(OpenshiftLoggingNS, topic)

		tc.AddCleanup(func() error {
			var zerograce int64
			return tc.KubeClient.Apps().Deployments(OpenshiftLoggingNS).Delete(app.GetName(), metav1.NewDeleteOptions(zerograce))
		})

		app, err := tc.KubeClient.Apps().Deployments(OpenshiftLoggingNS).Create(app)
		if err != nil {
			return err
		}

		if err := tc.waitForDeployment(OpenshiftLoggingNS, app.GetName(), defaultRetryInterval, defaultTimeout); err != nil {
			return err
		}
	}

	return err
}

func (tc *E2ETestFramework) createKafkaBrokerStatefulSet() (*apps.StatefulSet, error) {
	app := kafka.NewBrokerStatefuleSet(OpenshiftLoggingNS)

	tc.AddCleanup(func() error {
		var zerograce int64
		return tc.KubeClient.Apps().StatefulSets(OpenshiftLoggingNS).Delete(app.GetName(), metav1.NewDeleteOptions(zerograce))
	})

	_, err := tc.KubeClient.Apps().StatefulSets(OpenshiftLoggingNS).Create(app)
	if err != nil {
		return nil, err
	}

	return app, tc.waitForStatefulSet(OpenshiftLoggingNS, app.GetName(), defaultRetryInterval, defaultTimeout)
}

func (tc *E2ETestFramework) createZookeeperStatefulSet() (*apps.StatefulSet, error) {
	app := kafka.NewZookeeperStatefuleSet(OpenshiftLoggingNS)

	tc.AddCleanup(func() error {
		var zerograce int64
		return tc.KubeClient.Apps().StatefulSets(OpenshiftLoggingNS).Delete(app.GetName(), metav1.NewDeleteOptions(zerograce))
	})

	app, err := tc.KubeClient.Apps().StatefulSets(OpenshiftLoggingNS).Create(app)
	if err != nil {
		return nil, err
	}

	return app, tc.waitForStatefulSet(OpenshiftLoggingNS, app.GetName(), defaultRetryInterval, defaultTimeout)
}

func (tc *E2ETestFramework) createKafkaBrokerService() error {
	svc := kafka.NewBrokerService(OpenshiftLoggingNS)

	tc.AddCleanup(func() error {
		var zerograce int64
		return tc.KubeClient.CoreV1().Services(OpenshiftLoggingNS).Delete(svc.GetName(), metav1.NewDeleteOptions(zerograce))
	})

	if _, err := tc.KubeClient.CoreV1().Services(OpenshiftLoggingNS).Create(svc); err != nil {
		return err
	}

	return nil
}

func (tc *E2ETestFramework) createZookeeperService() error {
	svc := kafka.NewZookeeperService(OpenshiftLoggingNS)

	tc.AddCleanup(func() error {
		var zerograce int64
		return tc.KubeClient.CoreV1().Services(OpenshiftLoggingNS).Delete(svc.GetName(), metav1.NewDeleteOptions(zerograce))
	})

	if _, err := tc.KubeClient.CoreV1().Services(OpenshiftLoggingNS).Create(svc); err != nil {
		return err
	}

	return nil
}

func (tc *E2ETestFramework) createKafkaBrokerRBAC() error {
	cr, crb := kafka.NewBrokerRBAC(OpenshiftLoggingNS)

	tc.AddCleanup(func() error {
		var zerograce int64
		return tc.KubeClient.Rbac().ClusterRoles().Delete(cr.GetName(), metav1.NewDeleteOptions(zerograce))
	})

	if _, err := tc.KubeClient.Rbac().ClusterRoles().Create(cr); err != nil {
		return err
	}

	tc.AddCleanup(func() error {
		var zerograce int64
		return tc.KubeClient.Rbac().ClusterRoleBindings().Delete(crb.GetName(), metav1.NewDeleteOptions(zerograce))
	})

	if _, err := tc.KubeClient.Rbac().ClusterRoleBindings().Create(crb); err != nil {
		return err
	}
	return nil
}

func (tc *E2ETestFramework) createKafkaBrokerConfigMap() error {
	cm := kafka.NewBrokerConfigMap(OpenshiftLoggingNS)

	tc.AddCleanup(func() error {
		var zerograce int64
		return tc.KubeClient.Core().ConfigMaps(OpenshiftLoggingNS).Delete(cm.GetName(), metav1.NewDeleteOptions(zerograce))
	})

	if _, err := tc.KubeClient.Core().ConfigMaps(OpenshiftLoggingNS).Create(cm); err != nil {
		return err
	}

	return nil
}

func (tc *E2ETestFramework) createZookeeperConfigMap() error {
	cm := kafka.NewZookeeperConfigMap(OpenshiftLoggingNS)

	tc.AddCleanup(func() error {
		var zerograce int64
		return tc.KubeClient.Core().ConfigMaps(OpenshiftLoggingNS).Delete(cm.GetName(), metav1.NewDeleteOptions(zerograce))
	})

	if _, err := tc.KubeClient.Core().ConfigMaps(OpenshiftLoggingNS).Create(cm); err != nil {
		return err
	}

	return nil
}
