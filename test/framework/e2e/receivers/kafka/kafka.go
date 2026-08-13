package kafka

import (
	"context"
	"fmt"
	"strings"
	"time"

	clolog "github.com/ViaQ/logerr/v2/log/static"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	testclient "github.com/openshift/cluster-logging-operator/test/client"
	"github.com/openshift/cluster-logging-operator/test/framework"
	"github.com/openshift/cluster-logging-operator/test/framework/e2e/receivers/elasticsearch"
	"github.com/openshift/cluster-logging-operator/test/helpers/kafka"
	"github.com/openshift/cluster-logging-operator/test/helpers/types"
	"github.com/pkg/errors"
	apps "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
)

type Receiver struct {
	framework.Test
	testClient *testclient.Client
	app        *apps.StatefulSet
	topics     []string
}

// New creates a kafka cluster named 'kafka' in the openshift-logging namespace
func New(test framework.Test, topics ...string) *Receiver {
	return &Receiver{
		Test:   test,
		topics: topics,
	}
}

func consumeLogs(rcv *Receiver, inputName string) (types.Logs, error) {
	topic := kafka.TopicForInputName(rcv.topics, inputName)
	name := kafka.ConsumerNameForTopic(topic)

	options := metav1.ListOptions{
		LabelSelector: fmt.Sprintf("component=%s", name),
	}
	pods, err := rcv.Client().CoreV1().Pods(constants.OpenshiftNS).List(context.TODO(), options)
	if err != nil {
		return nil, err
	}
	if len(pods.Items) == 0 {
		return nil, fmt.Errorf("no pods found for %s", name)
	}

	cmd := "tail -n 5000 /shared/consumed.logs"
	stdout, err := rcv.PodExec(constants.OpenshiftNS, pods.Items[0].Name, name, []string{"bash", "-c", cmd})
	if err != nil {
		return nil, err
	}

	// Hack Teach kafka-console-consumer to output a proper json array
	out := "[" + strings.TrimRight(strings.ReplaceAll(stdout, "\n", ","), ",") + "]"
	logs, err := types.ParseLogs(out)
	if err != nil {
		return nil, types.ErrParse
	}

	return logs, nil
}

func (r *Receiver) ApplicationLogs(_ time.Duration) (types.Logs, error) {
	logs, err := consumeLogs(r, string(obs.InputTypeApplication))
	if err != nil {
		return nil, fmt.Errorf("failed to read consumed application logs: %s", err)
	}
	return logs.ByIndex(elasticsearch.ProjectIndexPrefix), nil
}

func (r *Receiver) HasInfraStructureLogs(timeout time.Duration) (bool, error) {
	return hasLogs(r, obs.InputTypeInfrastructure, elasticsearch.InfraIndexPrefix, timeout)
}

func (r *Receiver) HasApplicationLogs(timeout time.Duration) (bool, error) {
	return hasLogs(r, obs.InputTypeApplication, elasticsearch.ProjectIndexPrefix, timeout)
}

func (r *Receiver) HasAuditLogs(timeout time.Duration) (bool, error) {
	return hasLogs(r, obs.InputTypeAudit, elasticsearch.AuditIndexPrefix, timeout)
}

func hasLogs(r *Receiver, inputType obs.InputType, prefix string, timeout time.Duration) (bool, error) {
	err := wait.PollUntilContextTimeout(context.TODO(), framework.DefaultRetryInterval, timeout, true, func(cxt context.Context) (done bool, err error) {
		logs, err := consumeLogs(r, string(inputType))
		if err != nil {
			if errors.Is(err, types.ErrParse) {
				clolog.Error(err, "check the test artifact.", "inputType", inputType)
				// return error here else loop will keep on parsing
				return false, err
			}
			clolog.Error(err, "unable to fetch audit logs", "inputType", inputType)
			return false, nil
		}
		l := logs.ByIndex(prefix)
		if l.NonEmpty() {
			clolog.Info("found logs", "inputType", inputType)
		} else {
			clolog.Info("could not find logs", "inputType", inputType)
		}
		return l.NonEmpty(), nil
	})
	return true, err
}

func (r *Receiver) GrepLogs(_ string, _ time.Duration) (string, error) {
	return "Not Found", fmt.Errorf("not implemented")
}

func (r *Receiver) RetrieveLogs() (map[string]string, error) {
	return nil, fmt.Errorf("not implemented")
}

func (r *Receiver) ClusterLocalEndpoint() string {
	return kafka.ClusterLocalEndpoint(constants.OpenshiftNS)
}

func (r *Receiver) Name() string {
	return kafka.DeploymentName
}

func (r *Receiver) Deploy() (err error) {
	if err = r.createZookeeper(); err != nil {
		return err
	}

	r.app, err = r.createKafkaBroker()
	if err != nil {
		return err
	}

	if err = r.createKafkaConsumers(); err != nil {
		return err
	}

	return nil
}

func (r *Receiver) createKafkaBroker() (*apps.StatefulSet, error) {
	if err := r.createKafkaBrokerRBAC(); err != nil {
		return nil, err
	}

	if err := r.createKafkaBrokerConfigMap(); err != nil {
		return nil, err
	}

	if err := r.createKafkaBrokerSecret(); err != nil {
		return nil, err
	}

	if err := r.createKafkaBrokerService(); err != nil {
		return nil, err
	}

	app, err := r.createKafkaBrokerStatefulSet()
	if err != nil {
		return nil, err
	}

	return app, nil
}

func (r *Receiver) createZookeeper() error {
	if err := r.createZookeeperConfigMap(); err != nil {
		return err
	}

	if _, err := r.createZookeeperStatefulSet(); err != nil {
		return err
	}

	if err := r.createZookeeperService(); err != nil {
		return err
	}

	return nil
}

func (r *Receiver) createKafkaConsumers() (err error) {
	for _, topic := range r.topics {
		app := kafka.NewKafkaConsumerDeployment(constants.OpenshiftNS, topic)

		r.AddCleanup(func() error {
			return r.testClient.Delete(app)
		})

		if err = r.testClient.Create(app); err != nil {
			return err
		}

		if err = r.testClient.WaitFor(app, framework.NewDeploymentWaitCondition(app)); err != nil {
			return err
		}

	}
	return nil
}

func (r *Receiver) createKafkaBrokerStatefulSet() (app *apps.StatefulSet, err error) {
	app = kafka.NewBrokerStatefuleSet(constants.OpenshiftNS)

	r.AddCleanup(func() error {
		return r.testClient.Delete(app)
	})

	if err = r.testClient.Create(app); err != nil {
		return app, err
	}

	return app, r.testClient.WaitFor(app, framework.NewStatefulSetWaitCondition(app))
}

func (r *Receiver) createZookeeperStatefulSet() (app *apps.StatefulSet, err error) {
	app = kafka.NewZookeeperStatefuleSet(constants.OpenshiftNS)

	r.AddCleanup(func() error {
		return r.testClient.Delete(app)
	})

	if err = r.testClient.Create(app); err != nil {
		return app, err
	}

	return app, r.testClient.WaitFor(app, framework.NewStatefulSetWaitCondition(app))
}

func (r *Receiver) createKafkaBrokerService() (err error) {
	svc := kafka.NewBrokerService(constants.OpenshiftNS)

	r.AddCleanup(func() error {
		return r.testClient.Delete(svc)
	})

	if err = r.testClient.Create(svc); err != nil {
		return err
	}

	return nil
}

func (r *Receiver) createZookeeperService() (err error) {
	svc := kafka.NewZookeeperService(constants.OpenshiftNS)

	r.AddCleanup(func() error {
		return r.testClient.Delete(svc)
	})

	if err = r.testClient.Create(svc); err != nil {
		return err
	}

	return nil
}

func (r *Receiver) createKafkaBrokerRBAC() (err error) {
	cr, crb := kafka.NewBrokerRBAC(constants.OpenshiftNS)

	r.AddCleanup(func() error {
		return r.testClient.Delete(cr)
	})

	if err = r.testClient.Create(cr); err != nil {
		return err
	}

	r.AddCleanup(func() error {
		return r.testClient.Delete(crb)
	})

	if err = r.testClient.Create(crb); err != nil {
		return err
	}
	return nil
}

func (r *Receiver) createKafkaBrokerConfigMap() (err error) {
	cm := kafka.NewBrokerConfigMap(constants.OpenshiftNS)

	r.AddCleanup(func() error {
		return r.testClient.Delete(cm)
	})

	if err = r.testClient.Create(cm); err != nil {
		return err
	}

	return nil
}

func (r *Receiver) createKafkaBrokerSecret() (err error) {
	s := kafka.NewBrokerSecret(constants.OpenshiftNS)

	r.AddCleanup(func() error {
		return r.testClient.Delete(s)
	})

	if err = r.testClient.Create(s); err != nil {
		return err
	}

	return nil
}

func (r *Receiver) createZookeeperConfigMap() (err error) {
	cm := kafka.NewZookeeperConfigMap(constants.OpenshiftNS)

	r.AddCleanup(func() error {
		return r.testClient.Delete(cm)
	})

	if err = r.testClient.Create(cm); err != nil {
		return err
	}

	return nil
}
