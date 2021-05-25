package helpers

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/openshift/cluster-logging-operator/test/helpers/types"

	clolog "github.com/ViaQ/logerr/log"
	loggingv1 "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
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

func (kr *kafkaReceiver) ApplicationLogs(timeToWait time.Duration) (types.Logs, error) {
	logs, err := kr.tc.consumedLogs(kr.app.Name, loggingv1.InputNameApplication)
	if err != nil {
		return nil, fmt.Errorf("Failed to read consumed application logs: %s", err)
	}
	return logs.ByIndex(ProjectIndexPrefix), nil
}

func (kr *kafkaReceiver) HasInfraStructureLogs(timeout time.Duration) (bool, error) {
	err := wait.PollImmediate(defaultRetryInterval, timeout, func() (done bool, err error) {
		logs, err := kr.tc.consumedLogs(kr.app.Name, loggingv1.InputNameInfrastructure)
		if err != nil {
			if err == types.ErrParse {
				clolog.Error(err, "error occurred while parsing fetched infra logs from kafka topic. Please check the collected test artifact.")
				// return error here else loop will keep on parsing
				return false, err
			}
			clolog.Error(err, "error occurred while fetching infra logs")
			return false, nil
		}
		l := logs.ByIndex(InfraIndexPrefix)
		if l.NonEmpty() {
			clolog.Info("found infra logs")
		} else {
			clolog.Info("could not find infra logs")
		}
		return l.NonEmpty(), nil
	})
	return true, err
}

func (kr *kafkaReceiver) HasApplicationLogs(timeout time.Duration) (bool, error) {
	err := wait.PollImmediate(defaultRetryInterval, timeout, func() (done bool, err error) {
		logs, err := kr.tc.consumedLogs(kr.app.Name, loggingv1.InputNameApplication)
		if err != nil {
			if err == types.ErrParse {
				clolog.Error(err, "error occurred while parsing fetched application logs from kafka topic. Please check the collected test artifact.")
				// return error here else loop will keep on parsing
				return false, err
			}
			clolog.Error(err, "error occurred while fetching application logs")
			return false, nil
		}
		l := logs.ByIndex(ProjectIndexPrefix)
		if l.NonEmpty() {
			clolog.Info("found app logs")
		} else {
			clolog.Info("could not find app logs")
		}
		return l.NonEmpty(), nil
	})
	return true, err
}

func (kr *kafkaReceiver) HasAuditLogs(timeout time.Duration) (bool, error) {
	err := wait.Poll(defaultRetryInterval, timeout, func() (done bool, err error) {
		logs, err := kr.tc.consumedLogs(kr.app.Name, loggingv1.InputNameAudit)
		if err != nil {
			if err == types.ErrParse {
				clolog.Error(err, "error occurred while parsing fetched audit logs from kafka topic. Please check the collected test artifact.")
				// return error here else loop will keep on parsing
				return false, err
			}
			clolog.Error(err, "error occurred while fetching audit logs")
			return false, nil
		}
		l := logs.ByIndex(AuditIndexPrefix)
		if l.NonEmpty() {
			clolog.Info("found audit logs")
		} else {
			clolog.Info("could not find audit logs")
		}
		return l.NonEmpty(), nil
	})
	return true, err
}

func (kr *kafkaReceiver) GrepLogs(expr string, timeToWait time.Duration) (string, error) {
	return "Not Found", fmt.Errorf("Not implemented")
}

func (kr *kafkaReceiver) RetrieveLogs() (map[string]string, error) {
	return nil, fmt.Errorf("Not implemented")
}

func (kr *kafkaReceiver) ClusterLocalEndpoint() string {
	return kafka.ClusterLocalEndpoint(OpenshiftLoggingNS)
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
	tc.LogStores[app.Name] = receiver

	if err := tc.createKafkaConsumers(receiver); err != nil {
		return nil, err
	}

	return app, nil
}

func (tc *E2ETestFramework) consumedLogs(rcvName, inputName string) (types.Logs, error) {
	rcv := tc.LogStores[rcvName].(*kafkaReceiver)
	topic := kafka.TopicForInputName(rcv.topics, inputName)
	name := kafka.ConsumerNameForTopic(topic)

	options := metav1.ListOptions{
		LabelSelector: fmt.Sprintf("component=%s", name),
	}
	pods, err := tc.KubeClient.CoreV1().Pods(OpenshiftLoggingNS).List(context.TODO(), options)
	if err != nil {
		return nil, err
	}
	if len(pods.Items) == 0 {
		return nil, fmt.Errorf("No pods found for %s", name)
	}

	cmd := "tail -n 5000 /shared/consumed.logs"
	stdout, err := tc.PodExec(OpenshiftLoggingNS, pods.Items[0].Name, name, []string{"bash", "-c", cmd})
	if err != nil {
		return nil, err
	}

	// Hack Teach kafka-console-consumer to output a proper json array
	out := "[" + strings.TrimRight(strings.Replace(stdout, "\n", ",", -1), ",") + "]"
	logs, err := types.ParseLogs(out)
	if err != nil {
		return nil, types.ErrParse
	}

	return logs, nil
}

func (tc *E2ETestFramework) createKafkaBroker() (*apps.StatefulSet, error) {
	if err := tc.createKafkaBrokerRBAC(); err != nil {
		return nil, err
	}

	if err := tc.createKafkaBrokerConfigMap(); err != nil {
		return nil, err
	}

	if err := tc.createKafkaBrokerSecret(); err != nil {
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
			opts := metav1.DeleteOptions{
				GracePeriodSeconds: &zerograce,
			}
			return tc.KubeClient.AppsV1().Deployments(OpenshiftLoggingNS).Delete(context.TODO(), app.GetName(), opts)
		})

		opts := metav1.CreateOptions{}
		app, err := tc.KubeClient.AppsV1().Deployments(OpenshiftLoggingNS).Create(context.TODO(), app, opts)
		if err != nil {
			return err
		}

		if err := tc.waitForDeployment(OpenshiftLoggingNS, app.GetName(), defaultRetryInterval, defaultTimeout); err != nil {
			return err
		}
	}
	return nil
}

func (tc *E2ETestFramework) createKafkaBrokerStatefulSet() (*apps.StatefulSet, error) {
	app := kafka.NewBrokerStatefuleSet(OpenshiftLoggingNS)

	tc.AddCleanup(func() error {
		var zerograce int64
		opts := metav1.DeleteOptions{
			GracePeriodSeconds: &zerograce,
		}
		return tc.KubeClient.AppsV1().StatefulSets(OpenshiftLoggingNS).Delete(context.TODO(), app.GetName(), opts)
	})

	opts := metav1.CreateOptions{}
	_, err := tc.KubeClient.AppsV1().StatefulSets(OpenshiftLoggingNS).Create(context.TODO(), app, opts)
	if err != nil {
		return nil, err
	}

	return app, tc.waitForStatefulSet(OpenshiftLoggingNS, app.GetName(), defaultRetryInterval, defaultTimeout)
}

func (tc *E2ETestFramework) createZookeeperStatefulSet() (*apps.StatefulSet, error) {
	app := kafka.NewZookeeperStatefuleSet(OpenshiftLoggingNS)

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

func (tc *E2ETestFramework) createKafkaBrokerService() error {
	svc := kafka.NewBrokerService(OpenshiftLoggingNS)

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

func (tc *E2ETestFramework) createZookeeperService() error {
	svc := kafka.NewZookeeperService(OpenshiftLoggingNS)

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

func (tc *E2ETestFramework) createKafkaBrokerRBAC() error {
	cr, crb := kafka.NewBrokerRBAC(OpenshiftLoggingNS)

	tc.AddCleanup(func() error {
		var zerograce int64
		opts := metav1.DeleteOptions{
			GracePeriodSeconds: &zerograce,
		}
		return tc.KubeClient.RbacV1().ClusterRoles().Delete(context.TODO(), cr.GetName(), opts)
	})

	opts := metav1.CreateOptions{}
	if _, err := tc.KubeClient.RbacV1().ClusterRoles().Create(context.TODO(), cr, opts); err != nil {
		return err
	}

	tc.AddCleanup(func() error {
		var zerograce int64
		opts := metav1.DeleteOptions{
			GracePeriodSeconds: &zerograce,
		}
		return tc.KubeClient.RbacV1().ClusterRoleBindings().Delete(context.TODO(), crb.GetName(), opts)
	})

	opts = metav1.CreateOptions{}
	if _, err := tc.KubeClient.RbacV1().ClusterRoleBindings().Create(context.TODO(), crb, opts); err != nil {
		return err
	}
	return nil
}

func (tc *E2ETestFramework) createKafkaBrokerConfigMap() error {
	cm := kafka.NewBrokerConfigMap(OpenshiftLoggingNS)

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

func (tc *E2ETestFramework) createKafkaBrokerSecret() error {
	s := kafka.NewBrokerSecret(OpenshiftLoggingNS)

	tc.AddCleanup(func() error {
		var zerograce int64
		opts := metav1.DeleteOptions{
			GracePeriodSeconds: &zerograce,
		}
		return tc.KubeClient.CoreV1().Secrets(OpenshiftLoggingNS).Delete(context.TODO(), s.GetName(), opts)
	})

	opts := metav1.CreateOptions{}
	if _, err := tc.KubeClient.CoreV1().Secrets(OpenshiftLoggingNS).Create(context.TODO(), s, opts); err != nil {
		return err
	}

	return nil
}

func (tc *E2ETestFramework) createZookeeperConfigMap() error {
	cm := kafka.NewZookeeperConfigMap(OpenshiftLoggingNS)

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
