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

type syslogReceiverLogStore struct {
	deployment *apps.Deployment
	tc         *E2ETestFramework
}

const (
	syslogReceiverName = "syslog-receiver"
	tcpSyslogConf      = `
# Provides TCP syslog reception
# for parameters see http://www.rsyslog.com/doc/imtcp.html
module(load="imtcp") # needs to be done just once
input(type="imtcp" port="24224" ruleset="test")

#### RULES ####
ruleset(name="test"){
    action(type="omfile" file="/var/log/infra.log")
}
	`
	udpSyslogConf = `
# Provides UDP syslog reception
# for parameters see http://www.rsyslog.com/doc/imudp.html
module(load="imudp") # needs to be done just once
input(type="imudp" port="24224" ruleset="test")

#### RULES ####
ruleset(name="test"){
    action(type="omfile" file="/var/log/infra.log")
}
	`
)

func (syslog *syslogReceiverLogStore) hasLogs(file string, timeToWait time.Duration) (bool, error) {
	options := metav1.ListOptions{
		LabelSelector: "component=syslog-receiver",
	}
	pods, err := syslog.tc.KubeClient.CoreV1().Pods(OpenshiftLoggingNS).List(options)
	if err != nil {
		return false, err
	}
	if len(pods.Items) == 0 {
		return false, errors.New("No pods found for syslog receiver")
	}
	logger.Debugf("Pod %s", pods.Items[0].Name)
	cmd := fmt.Sprintf("ls %s | wc -l", file)

	err = wait.Poll(defaultRetryInterval, timeToWait, func() (done bool, err error) {
		output, err := syslog.tc.PodExec(OpenshiftLoggingNS, pods.Items[0].Name, "syslog-receiver", []string{"bash", "-c", cmd})
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

func (syslog *syslogReceiverLogStore) HasInfraStructureLogs(timeToWait time.Duration) (bool, error) {
	return syslog.hasLogs("/var/log/infra.log", timeToWait)
}

func (syslog *syslogReceiverLogStore) HasApplicationLogs(timeToWait time.Duration) (bool, error) {
	return false, fmt.Errorf("Not implemented")
}

func (syslog *syslogReceiverLogStore) HasAuditLogs(timeToWait time.Duration) (bool, error) {
	return false, fmt.Errorf("Not implemented")
}

func (tc *E2ETestFramework) createSyslogServiceAccount() (serviceAccount *corev1.ServiceAccount, err error) {
	serviceAccount = k8shandler.NewServiceAccount("syslog-receiver", OpenshiftLoggingNS)
	if serviceAccount, err = tc.KubeClient.Core().ServiceAccounts(OpenshiftLoggingNS).Create(serviceAccount); err != nil {
		return nil, err
	}
	tc.AddCleanup(func() error {
		return tc.KubeClient.Core().ServiceAccounts(OpenshiftLoggingNS).Delete(serviceAccount.Name, nil)
	})
	return serviceAccount, nil
}

func (tc *E2ETestFramework) CreateLegacySyslogConfigMap(namespace, conf string) (err error) {
	fluentdConfigMap := k8shandler.NewConfigMap(
		"syslog",
		namespace,
		map[string]string{
			"syslog.conf": conf,
		},
	)

	if fluentdConfigMap, err = tc.KubeClient.Core().ConfigMaps(namespace).Create(fluentdConfigMap); err != nil {
		return err
	}
	tc.AddCleanup(func() error {
		return tc.KubeClient.Core().ConfigMaps(namespace).Delete(fluentdConfigMap.Name, nil)
	})
	return nil
}

func (tc *E2ETestFramework) createSyslogRbac(name string) (err error) {
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

func (tc *E2ETestFramework) DeploySyslogReceiver(protocol corev1.Protocol) (deployment *apps.Deployment, err error) {
	logStore := &syslogReceiverLogStore{
		tc: tc,
	}
	serviceAccount, err := tc.createSyslogServiceAccount()
	if err != nil {
		return nil, err
	}
	if err := tc.createSyslogRbac(syslogReceiverName); err != nil {
		return nil, err
	}
	container := corev1.Container{
		Name:            syslogReceiverName,
		Image:           "quay.io/openshift/origin-logging-rsyslog:latest",
		ImagePullPolicy: corev1.PullAlways,
		Args:            []string{"rsyslogd", "-n", "-f", "/rsyslog/etc/rsyslog.conf"},
		VolumeMounts: []corev1.VolumeMount{
			{
				Name:      "config",
				ReadOnly:  true,
				MountPath: "/rsyslog/etc",
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

	var rsyslogConf string
	switch {
	case protocol == corev1.ProtocolUDP:
		rsyslogConf = udpSyslogConf

	default:
		rsyslogConf = tcpSyslogConf
	}

	config := k8shandler.NewConfigMap(container.Name, OpenshiftLoggingNS, map[string]string{
		"rsyslog.conf": rsyslogConf,
	})
	config, err = tc.KubeClient.Core().ConfigMaps(OpenshiftLoggingNS).Create(config)
	if err != nil {
		return nil, err
	}
	tc.AddCleanup(func() error {
		return tc.KubeClient.Core().ConfigMaps(OpenshiftLoggingNS).Delete(config.Name, nil)
	})

	syslogDeployment := k8shandler.NewDeployment(
		container.Name,
		OpenshiftLoggingNS,
		container.Name,
		serviceAccount.Name,
		podSpec,
	)

	syslogDeployment, err = tc.KubeClient.Apps().Deployments(OpenshiftLoggingNS).Create(syslogDeployment)
	if err != nil {
		return nil, err
	}
	service := k8shandler.NewService(
		serviceAccount.Name,
		OpenshiftLoggingNS,
		serviceAccount.Name,
		[]corev1.ServicePort{
			{
				Protocol: protocol,
				Port:     24224,
			},
		},
	)
	tc.AddCleanup(func() error {
		var zerograce int64
		deleteopts := metav1.DeleteOptions{
			GracePeriodSeconds: &zerograce,
		}
		return tc.KubeClient.AppsV1().Deployments(OpenshiftLoggingNS).Delete(syslogDeployment.Name, &deleteopts)
	})
	service, err = tc.KubeClient.Core().Services(OpenshiftLoggingNS).Create(service)
	if err != nil {
		return nil, err
	}
	tc.AddCleanup(func() error {
		return tc.KubeClient.Core().Services(OpenshiftLoggingNS).Delete(service.Name, nil)
	})
	logStore.deployment = syslogDeployment
	tc.LogStore = logStore
	return syslogDeployment, tc.waitForDeployment(OpenshiftLoggingNS, syslogDeployment.Name, defaultRetryInterval, defaultTimeout)
}
