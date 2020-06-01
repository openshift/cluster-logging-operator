package helpers

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	apps "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"

	"github.com/openshift/cluster-logging-operator/pkg/k8shandler"
	"github.com/openshift/cluster-logging-operator/pkg/logger"
	"github.com/openshift/cluster-logging-operator/pkg/utils"
)

type fluentReceiverLogStore struct {
	deployment     *apps.Deployment
	tc             *E2ETestFramework
	pipelineSecret *corev1.Secret
}

const (
	receiverName             = "fluent-receiver"
	secureFluentConfTemplate = `
<system>
	@log_level info
</system>
<source>
  @type forward
  <transport tls>
	  ca_cert_path /etc/fluentd/secrets/ca-bundle.crt
	  ca_private_key_path /etc/fluentd/secrets/ca.key
  </transport>
  <security>
	shared_key fluent-receiver
	self_hostname fluent-receiver
  </security>
</source>
<match *_default_** **_kube-*_** **_openshift-*_** **_openshift_** journal.** system.var.log**>
  @type file
  append true
  path /tmp/infra.logs
  symlink_path /tmp/infra-logs
</match>
<match kubernetes.**>
  @type file
  append true
  path /tmp/app.logs
  symlink_path /tmp/app-logs
  </match>
  <match linux-audit.log** k8s-audit.log** openshift-audit.log**>
  @type file
  append true
  path /tmp/audit.logs
  symlink_path /tmp/audit-logs
</match>
<match **>
	@type stdout
</match>
	`
	unsecureFluentConf = `
<system>
	@log_level warn
</system>
<source>
  @type forward
</source>
<match *_default_** **_kube-*_** **_openshift-*_** **_openshift_** journal.** system.var.log**>
  @type file
  append true
  path /tmp/infra.logs
  symlink_path /tmp/infra-logs
  </match>
  <match kubernetes.**>
  @type file
  append true
  path /tmp/app.logs
  symlink_path /tmp/app-logs
  </match>
  <match linux-audit.log** k8s-audit.log** openshift-audit.log**>
  @type file
  append true
  path /tmp/audit.logs
  symlink_path /tmp/audit-logs
</match>
<match **>
	@type stdout
</match>
	`
)

func (fluent *fluentReceiverLogStore) hasLogs(file string, timeToWait time.Duration) (bool, error) {
	options := metav1.ListOptions{
		LabelSelector: "component=fluent-receiver",
	}
	pods, err := fluent.tc.KubeClient.CoreV1().Pods(OpenshiftLoggingNS).List(options)
	if err != nil {
		return false, err
	}
	if len(pods.Items) == 0 {
		return false, errors.New("No pods found for fluent receiver")
	}
	logger.Debugf("Pod %s", pods.Items[0].Name)
	cmd := fmt.Sprintf("ls %s | wc -l", file)

	err = wait.Poll(defaultRetryInterval, timeToWait, func() (done bool, err error) {
		output, err := fluent.tc.PodExec(OpenshiftLoggingNS, pods.Items[0].Name, "fluent-receiver", []string{"bash", "-c", cmd})
		if err != nil {
			return false, err
		}
		value, err := strconv.Atoi(strings.TrimSpace(output))
		if err != nil {
			logger.Debugf("Error parsing output: %s", output)
			return false, nil
		}
		return value > 0, nil
	})
	if err == wait.ErrWaitTimeout {
		return false, err
	}
	return true, err
}

func (fluent *fluentReceiverLogStore) logs(file string, timeToWait time.Duration) (string, error) {
	options := metav1.ListOptions{
		LabelSelector: "component=fluent-receiver",
	}
	pods, err := fluent.tc.KubeClient.CoreV1().Pods(OpenshiftLoggingNS).List(options)
	if err != nil {
		return "", err
	}
	if len(pods.Items) == 0 {
		return "", errors.New("No pods found for fluent receiver")
	}
	logger.Debugf("Pod %s", pods.Items[0].Name)
	cmd := fmt.Sprintf("cat %s", file)
	result := ""
	err = wait.Poll(defaultRetryInterval, timeToWait, func() (done bool, err error) {
		if result, err = fluent.tc.PodExec(OpenshiftLoggingNS, pods.Items[0].Name, "fluent-receiver", []string{"bash", "-c", cmd}); err != nil {
			return false, err
		}
		return true, nil
	})
	if err == wait.ErrWaitTimeout {
		return "", err
	}
	return result, nil
}

func (fluent *fluentReceiverLogStore) ApplicationLogs(timeToWait time.Duration) (logs, error) {
	fl, err := fluent.logs("/tmp/app-logs", timeToWait)
	if err != nil {
		return nil, err
	}
	return ParseLogs(fl)
}

func (fluent fluentReceiverLogStore) HasInfraStructureLogs(timeToWait time.Duration) (bool, error) {
	return fluent.hasLogs("/tmp/infra.logs", timeToWait)
}
func (fluent *fluentReceiverLogStore) HasApplicationLogs(timeToWait time.Duration) (bool, error) {
	return fluent.hasLogs("/tmp/app.logs", timeToWait)
}

func (fluent *fluentReceiverLogStore) HasAuditLogs(timeToWait time.Duration) (bool, error) {
	return fluent.hasLogs("/tmp/audit.logs", timeToWait)
}

func (es *fluentReceiverLogStore) GrepLogs(expr string, timeToWait time.Duration) (string, error) {
	return "Not Found", fmt.Errorf("Not implemented")
}

func (fluent *fluentReceiverLogStore) ClusterLocalEndpoint() string {
	panic("Not implemented")
}

func (tc *E2ETestFramework) createServiceAccount() (serviceAccount *corev1.ServiceAccount, err error) {
	serviceAccount = k8shandler.NewServiceAccount("fluent-receiver", OpenshiftLoggingNS)
	if serviceAccount, err = tc.KubeClient.Core().ServiceAccounts(OpenshiftLoggingNS).Create(serviceAccount); err != nil {
		return nil, err
	}
	tc.AddCleanup(func() error {
		return tc.KubeClient.Core().ServiceAccounts(OpenshiftLoggingNS).Delete(serviceAccount.Name, nil)
	})
	return serviceAccount, nil
}

func (tc *E2ETestFramework) createRbac(name string) (err error) {
	saRole := k8shandler.NewRole(
		name,
		OpenshiftLoggingNS,
		k8shandler.NewPolicyRules(
			k8shandler.NewPolicyRule(
				[]string{"security.openshift.io"},
				[]string{"securitycontextconstraints"},
				[]string{"privileged"},
				[]string{"use"},
			),
		),
	)
	if _, err = tc.KubeClient.Rbac().Roles(OpenshiftLoggingNS).Create(saRole); err != nil {
		return err
	}
	tc.AddCleanup(func() error {
		return tc.KubeClient.Rbac().Roles(OpenshiftLoggingNS).Delete(name, nil)
	})
	subject := k8shandler.NewSubject(
		"ServiceAccount",
		name,
	)
	subject.APIGroup = ""
	roleBinding := k8shandler.NewRoleBinding(
		name,
		OpenshiftLoggingNS,
		saRole.Name,
		k8shandler.NewSubjects(
			subject,
		),
	)
	if _, err = tc.KubeClient.Rbac().RoleBindings(OpenshiftLoggingNS).Create(roleBinding); err != nil {
		return err
	}
	tc.AddCleanup(func() error {
		return tc.KubeClient.Rbac().RoleBindings(OpenshiftLoggingNS).Delete(name, nil)
	})
	return nil
}

func (tc *E2ETestFramework) DeployFluentdReceiver(rootDir string, secure bool) (deployment *apps.Deployment, err error) {
	logStore := &fluentReceiverLogStore{
		tc: tc,
	}
	serviceAccount, err := tc.createServiceAccount()
	if err != nil {
		return nil, err
	}
	if err := tc.createRbac(receiverName); err != nil {
		return nil, err
	}
	container := corev1.Container{
		Name:            receiverName,
		Image:           "quay.io/openshift/origin-logging-fluentd:latest",
		ImagePullPolicy: corev1.PullAlways,
		Args:            []string{"fluentd", "-c", "/fluentd/etc/fluent.conf"},
		VolumeMounts: []corev1.VolumeMount{
			{
				Name:      "config",
				ReadOnly:  true,
				MountPath: "/fluentd/etc",
			},
		},
		SecurityContext: &corev1.SecurityContext{
			Privileged: utils.GetBool(true),
		},
	}
	podSpec := corev1.PodSpec{
		Containers: []corev1.Container{container},
		Volumes: []corev1.Volume{
			{
				Name: "config", VolumeSource: corev1.VolumeSource{
					ConfigMap: &corev1.ConfigMapVolumeSource{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: container.Name,
						},
					},
				},
			},
		},
		ServiceAccountName: serviceAccount.Name,
	}

	fluentConf := unsecureFluentConf
	if secure {
		fluentConf = secureFluentConfTemplate
		otherConf := map[string][]byte{
			"shared_key": []byte("my_shared_key"),
		}
		if logStore.pipelineSecret, err = tc.CreatePipelineSecret(rootDir, receiverName, receiverName, otherConf); err != nil {
			return nil, err
		}
		tc.AddCleanup(func() error {
			return tc.KubeClient.Core().Secrets(OpenshiftLoggingNS).Delete(receiverName, nil)
		})
		container.VolumeMounts = append(container.VolumeMounts, corev1.VolumeMount{
			Name:      "certs",
			ReadOnly:  true,
			MountPath: "/etc/fluentd/secrets",
		})
		podSpec.Containers = []corev1.Container{container}
		logStore.pipelineSecret.Data["shared_key"] = []byte("fluent-receiver")
		if logStore.pipelineSecret, err = tc.KubeClient.Core().Secrets(OpenshiftLoggingNS).Update(logStore.pipelineSecret); err != nil {
			return nil, err
		}
		podSpec.Volumes = append(podSpec.Volumes, corev1.Volume{
			Name: "certs", VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: logStore.pipelineSecret.Name,
				},
			},
		})
	}

	config := k8shandler.NewConfigMap(container.Name, OpenshiftLoggingNS, map[string]string{
		"fluent.conf": fluentConf,
	})
	config, err = tc.KubeClient.Core().ConfigMaps(OpenshiftLoggingNS).Create(config)
	if err != nil {
		return nil, err
	}
	tc.AddCleanup(func() error {
		return tc.KubeClient.Core().ConfigMaps(OpenshiftLoggingNS).Delete(config.Name, nil)
	})

	fluentDeployment := k8shandler.NewDeployment(
		container.Name,
		OpenshiftLoggingNS,
		container.Name,
		serviceAccount.Name,
		podSpec,
	)

	fluentDeployment, err = tc.KubeClient.Apps().Deployments(OpenshiftLoggingNS).Create(fluentDeployment)
	if err != nil {
		return nil, err
	}
	service := k8shandler.NewService(
		serviceAccount.Name,
		OpenshiftLoggingNS,
		serviceAccount.Name,
		[]corev1.ServicePort{
			{
				Port: 24224,
			},
		},
	)
	tc.AddCleanup(func() error {
		var zerograce int64
		deleteopts := metav1.DeleteOptions{
			GracePeriodSeconds: &zerograce,
		}
		return tc.KubeClient.AppsV1().Deployments(OpenshiftLoggingNS).Delete(fluentDeployment.Name, &deleteopts)
	})
	service, err = tc.KubeClient.Core().Services(OpenshiftLoggingNS).Create(service)
	if err != nil {
		return nil, err
	}
	tc.AddCleanup(func() error {
		return tc.KubeClient.Core().Services(OpenshiftLoggingNS).Delete(service.Name, nil)
	})
	logStore.deployment = fluentDeployment
	tc.LogStore = logStore
	return fluentDeployment, tc.waitForDeployment(OpenshiftLoggingNS, fluentDeployment.Name, defaultRetryInterval, defaultTimeout)
}
