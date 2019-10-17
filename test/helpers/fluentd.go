package helpers

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/openshift/cluster-logging-operator/pkg/k8shandler"
	"github.com/openshift/cluster-logging-operator/pkg/logger"
	"github.com/openshift/cluster-logging-operator/pkg/utils"
	apps "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
)

type fluentReceiverLogStore struct {
	deployment *apps.Deployment
	tc         *E2ETestFramework
}

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
	return true, err
}

func (fluent *fluentReceiverLogStore) HasInfraStructureLogs(timeToWait time.Duration) (bool, error) {
	return fluent.hasLogs("/tmp/infra.logs", timeToWait)
}
func (fluent *fluentReceiverLogStore) HasApplicationStructureLogs(timeToWait time.Duration) (bool, error) {
	return fluent.hasLogs("/tmp/app.logs", timeToWait)
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

func (tc *E2ETestFramework) DeployFluendReceiver() (deployment *apps.Deployment, pipelineSecret *corev1.Secret, err error) {
	serviceAccount, err := tc.createServiceAccount()
	if err != nil {
		return nil, nil, err
	}
	if err := tc.createRbac(serviceAccount.Name); err != nil {
		return nil, nil, err
	}
	fluentConf := `
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
</match>
<match kubernetes.**>
  @type file
  append true
  path /tmp/app.logs
</match>
<match **>
	@type stdout
</match>
	`
	container := corev1.Container{
		Name:            serviceAccount.Name,
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
	config := k8shandler.NewConfigMap(container.Name, OpenshiftLoggingNS, map[string]string{
		"fluent.conf": fluentConf,
	})
	config, err = tc.KubeClient.Core().ConfigMaps(OpenshiftLoggingNS).Create(config)
	if err != nil {
		return nil, nil, err
	}
	tc.AddCleanup(func() error {
		return tc.KubeClient.Core().ConfigMaps(OpenshiftLoggingNS).Delete(config.Name, nil)
	})

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
	fluentDeployment := k8shandler.NewDeployment(
		container.Name,
		OpenshiftLoggingNS,
		container.Name,
		serviceAccount.Name,
		podSpec,
	)

	fluentDeployment, err = tc.KubeClient.Apps().Deployments(OpenshiftLoggingNS).Create(fluentDeployment)
	if err != nil {
		return nil, nil, err
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
		return tc.KubeClient.Apps().Deployments(OpenshiftLoggingNS).Delete(fluentDeployment.Name, nil)
	})
	service, err = tc.KubeClient.Core().Services(OpenshiftLoggingNS).Create(service)
	if err != nil {
		return nil, nil, err
	}
	tc.AddCleanup(func() error {
		return tc.KubeClient.Core().Services(OpenshiftLoggingNS).Delete(service.Name, nil)
	})
	tc.LogStore = &fluentReceiverLogStore{
		deployment: fluentDeployment,
		tc:         tc,
	}
	return fluentDeployment, nil, tc.waitForDeployment(OpenshiftLoggingNS, fluentDeployment.Name, defaultRetryInterval, defaultTimeout)
}
