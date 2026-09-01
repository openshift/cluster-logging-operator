package e2e

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/test/helpers/certificate"
	"github.com/openshift/cluster-logging-operator/test/helpers/syslog"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/openshift/cluster-logging-operator/internal/factory"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	apierrors "k8s.io/apimachinery/pkg/api/errors"

	"github.com/openshift/cluster-logging-operator/test/helpers/types"

	apps "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"

	clolog "github.com/ViaQ/logerr/v2/log/static"
	"github.com/openshift/cluster-logging-operator/internal/utils"
)

type syslogReceiverLogStore struct {
	deployment *apps.Deployment
	tc         *E2ETestFramework
}

const (
	SyslogReceiverName = "syslog-receiver"
)

func (s *syslogReceiverLogStore) hasLogs(file string, timeToWait time.Duration) (bool, error) {
	options := metav1.ListOptions{
		LabelSelector: constants.LabelK8sComponent + "=" + s.deployment.Name,
	}
	clolog.V(3).Info("Listing syslog pods", "namespace", s.deployment.Namespace, "options", options)
	pods, err := s.tc.KubeClient.CoreV1().Pods(s.deployment.Namespace).List(context.TODO(), options)
	if err != nil {
		return false, err
	}
	if len(pods.Items) == 0 {
		return false, errors.New("no pods found for syslog receiver")
	}
	podName := pods.Items[0].Name
	cmd := fmt.Sprintf("ls %s | wc -l", file)
	clolog.V(3).Info("pod exec", "pod", podName, "cmd", cmd)
	err = wait.PollUntilContextTimeout(context.TODO(), 1*time.Second, timeToWait, true, func(cxt context.Context) (done bool, err error) {
		output, err := s.tc.PodExec(s.deployment.Namespace, podName, "syslog-receiver", []string{"bash", "-c", cmd})
		if err != nil {
			clolog.Error(err, "failed to fetch logs from syslog-receiver")
			return false, nil
		}
		clolog.V(3).Info("syslog-receiver pod exec", "pod", podName, "output", output)
		value, err := strconv.Atoi(strings.TrimSpace(output))
		if err != nil {
			clolog.V(2).Error(err, "Error parsing output", "output", output)
			return false, nil
		}
		return value > 0, nil
	})
	if wait.Interrupted(err) {
		return false, err
	}
	return true, err
}

func (s *syslogReceiverLogStore) grepLogs(expr string, logfile string, timeToWait time.Duration) (string, error) {
	NotFound := "No Found"
	options := metav1.ListOptions{
		LabelSelector: constants.LabelK8sComponent + "=" + s.deployment.Name,
	}
	pods, err := s.tc.KubeClient.CoreV1().Pods(s.deployment.Namespace).List(context.TODO(), options)
	if err != nil {
		return NotFound, err
	}
	if len(pods.Items) == 0 {
		return NotFound, errors.New("no pods found for syslog receiver")
	}
	clolog.V(3).Info("Pod", "PodName", pods.Items[0].Name)
	cmd := fmt.Sprintf(expr, logfile)
	clolog.V(3).Info("running expression", "expression", cmd)
	var value string

	err = wait.PollUntilContextTimeout(context.TODO(), 1*time.Second, timeToWait, true, func(cxt context.Context) (done bool, err error) {
		output, err := s.tc.PodExec(s.deployment.Namespace, pods.Items[0].Name, "syslog-receiver", []string{"bash", "-c", cmd})
		if err != nil {
			clolog.Error(err, "failed to fetch logs from syslog-receiver")
			return false, nil
		}
		value = strings.TrimSpace(output)
		return true, nil
	})
	if wait.Interrupted(err) {
		return NotFound, err
	}
	return value, nil
}

func (s *syslogReceiverLogStore) ApplicationLogs(timeToWait time.Duration) (types.Logs, error) {
	panic("Method not implemented")
}

func (s *syslogReceiverLogStore) HasInfraStructureLogs(timeToWait time.Duration) (bool, error) {
	return s.hasLogs("/tmp/infra.log", timeToWait)
}

func (s *syslogReceiverLogStore) HasApplicationLogs(timeToWait time.Duration) (bool, error) {
	return s.hasLogs("/tmp/app.log", timeToWait)
}

func (s *syslogReceiverLogStore) HasAuditLogs(timeToWait time.Duration) (bool, error) {
	return false, fmt.Errorf("not implemented")
}

func (s *syslogReceiverLogStore) GrepLogs(expr string, timeToWait time.Duration) (string, error) {
	return s.grepLogs(expr, "/tmp/infra.log", timeToWait)
}

func (s *syslogReceiverLogStore) RetrieveLogs() (map[string]string, error) {
	return nil, fmt.Errorf("not implemented")
}

func (s *syslogReceiverLogStore) ClusterLocalEndpoint() string {
	panic("not implemented")
}

func (tc *E2ETestFramework) CreateLegacySyslogConfigMap(namespace, conf string) (err error) {
	opts := metav1.CreateOptions{}
	fluentdConfigMap := runtime.NewConfigMap(
		namespace,
		"syslog",
		map[string]string{
			"syslog.conf": conf,
		},
	)

	if fluentdConfigMap, err = tc.KubeClient.CoreV1().ConfigMaps(namespace).Create(context.TODO(), fluentdConfigMap, opts); err != nil {
		return err
	}
	tc.AddCleanup(func() error {
		opts := metav1.DeleteOptions{}
		err := tc.KubeClient.CoreV1().ConfigMaps(namespace).Delete(context.TODO(), fluentdConfigMap.Name, opts)
		if apierrors.IsNotFound(err) {
			return nil
		}
		return err
	})
	return nil
}

func (tc *E2ETestFramework) createSyslogRbac(namespace, name string) (err error) {
	opts := metav1.CreateOptions{}
	saRole := runtime.NewRole(
		namespace,
		name,
		runtime.NewPolicyRules(
			runtime.NewPolicyRule(
				[]string{"security.openshift.io"},
				[]string{"securitycontextconstraints"},
				[]string{"privileged"},
				[]string{"use"},
			),
		)...,
	)

	if _, err = tc.KubeClient.RbacV1().Roles(namespace).Create(context.TODO(), saRole, opts); err != nil {
		return err
	}

	tc.AddCleanup(func() error {
		opts := metav1.DeleteOptions{}
		err := tc.KubeClient.RbacV1().Roles(namespace).Delete(context.TODO(), name, opts)
		if apierrors.IsNotFound(err) {
			return nil
		}
		return err
	})

	rbOpts := metav1.CreateOptions{}
	subject := runtime.NewSubject(
		"ServiceAccount",
		name,
	)

	subject.APIGroup = ""

	roleBinding := runtime.NewRoleBinding(
		namespace,
		name,
		rbacv1.RoleRef{
			Kind:     "Role",
			Name:     saRole.Name,
			APIGroup: rbacv1.GroupName,
		},
		runtime.NewSubjects(
			subject,
		)...,
	)

	if _, err = tc.KubeClient.RbacV1().RoleBindings(namespace).Create(context.TODO(), roleBinding, rbOpts); err != nil {
		return err
	}
	tc.AddCleanup(func() error {
		opts := metav1.DeleteOptions{}
		err := tc.KubeClient.RbacV1().RoleBindings(namespace).Delete(context.TODO(), name, opts)
		if apierrors.IsNotFound(err) {
			return nil
		}
		return err
	})
	return nil
}

func (tc *E2ETestFramework) DeploySyslogReceiver(namespace string, protocol corev1.Protocol, withTLS bool, rfc syslog.SyslogRfc) (deployment *apps.Deployment, err error) {
	logStore := &syslogReceiverLogStore{
		tc: tc,
	}
	serviceAccount, err := tc.createServiceAccount(namespace, SyslogReceiverName)
	if err != nil {
		return nil, err
	}
	if err := tc.createSyslogRbac(namespace, SyslogReceiverName); err != nil {
		return nil, err
	}
	container := corev1.Container{
		Name:            SyslogReceiverName,
		Image:           syslog.ImageRemoteSyslog,
		ImagePullPolicy: corev1.PullAlways,
		Command:         []string{"/usr/sbin/rsyslogd", "-i", "/tmp/rsyslog.pid", "-n", "-f", "/etc/rsyslog/rsyslog.conf"},
		VolumeMounts: []corev1.VolumeMount{
			{
				Name:      "config",
				ReadOnly:  true,
				MountPath: "/etc/rsyslog",
			},
		},
		SecurityContext: &corev1.SecurityContext{
			AllowPrivilegeEscalation: utils.GetPtr(false),
			Capabilities: &corev1.Capabilities{
				Drop: []corev1.Capability{"ALL"},
			},
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
			{
				Name: "log",
				VolumeSource: corev1.VolumeSource{
					EmptyDir: &corev1.EmptyDirVolumeSource{},
				},
			},
		},
		ServiceAccountName: serviceAccount.Name,
		SecurityContext: &corev1.PodSecurityContext{
			RunAsNonRoot: utils.GetPtr(true),
			SeccompProfile: &corev1.SeccompProfile{
				Type: corev1.SeccompProfileTypeRuntimeDefault,
			},
		},
	}

	var rsyslogConf string
	switch protocol {
	case corev1.ProtocolUDP:
		rsyslogConf = syslog.UdpSyslogInput

	default:
		rsyslogConf = syslog.TcpSyslogInput
	}

	if withTLS {
		switch protocol {
		case corev1.ProtocolUDP:
			rsyslogConf = syslog.UdpSyslogInputWithTLS

		default:
			rsyslogConf = syslog.TcpSyslogInputWithTLS
		}
		secret, err := tc.CreateSyslogReceiverSecrets(namespace, SyslogReceiverName, SyslogReceiverName)
		if err != nil {
			return nil, err
		}
		tc.AddCleanup(func() error {
			opts := metav1.DeleteOptions{}
			err := tc.KubeClient.CoreV1().Secrets(namespace).Delete(context.TODO(), SyslogReceiverName, opts)
			if apierrors.IsNotFound(err) {
				return nil
			}
			return err
		})
		container.VolumeMounts = append(container.VolumeMounts, corev1.VolumeMount{
			Name:      "certs",
			ReadOnly:  true,
			MountPath: "/rsyslog/etc/secrets",
		})
		podSpec.Containers = []corev1.Container{container}
		podSpec.Volumes = append(podSpec.Volumes, corev1.Volume{
			Name: "certs", VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: secret.Name,
				},
			},
		})
	}

	rsyslogConf = syslog.GenerateRsyslogConf(rsyslogConf, rfc)

	cOpts := metav1.CreateOptions{}
	config := runtime.NewConfigMap(namespace, container.Name, map[string]string{
		"rsyslog.conf": rsyslogConf,
	})
	config, err = tc.KubeClient.CoreV1().ConfigMaps(namespace).Create(context.TODO(), config, cOpts)
	if err != nil {
		return nil, err
	}
	tc.AddCleanup(func() error {
		opts := metav1.DeleteOptions{}
		err := tc.KubeClient.CoreV1().ConfigMaps(namespace).Delete(context.TODO(), config.Name, opts)
		if apierrors.IsNotFound(err) {
			return nil
		}
		return err
	})

	dOpts := metav1.CreateOptions{}
	syslogDeployment := factory.NewDeployment(
		namespace,
		container.Name,
		container.Name,
		serviceAccount.Name,
		1,
		podSpec,
	)
	syslogDeployment, err = tc.KubeClient.AppsV1().Deployments(namespace).Create(context.TODO(), syslogDeployment, dOpts)
	if err != nil {
		return nil, err
	}

	if _, err = tc.CreateSyslogService(syslogDeployment, protocol, corev1.ServiceTypeClusterIP); err != nil {
		return nil, err
	}

	tc.AddCleanup(func() error {
		var zerograce int64
		deleteopts := metav1.DeleteOptions{
			GracePeriodSeconds: &zerograce,
		}
		err := tc.KubeClient.AppsV1().Deployments(namespace).Delete(context.TODO(), syslogDeployment.Name, deleteopts)
		if apierrors.IsNotFound(err) {
			return nil
		}
		return err
	})

	logStore.deployment = syslogDeployment

	name := syslogDeployment.GetName()
	tc.LogStores[name] = logStore
	return syslogDeployment, tc.WaitForDeployment(namespace, syslogDeployment.Name, defaultRetryInterval, defaultTimeout)
}

func (tc *E2ETestFramework) CreateSyslogService(syslogDeployment *apps.Deployment, protocol corev1.Protocol, serviceType corev1.ServiceType) (service *corev1.Service, err error) {

	service = factory.NewService(
		syslogDeployment.Name,
		syslogDeployment.Namespace,
		syslogDeployment.Name,
		syslogDeployment.Name,
		[]corev1.ServicePort{
			{
				Name:       "udp",
				Protocol:   protocol,
				TargetPort: intstr.FromInt32(24224),
				Port:       514,
			},
			{
				Name:       "tcp",
				Protocol:   corev1.ProtocolTCP,
				TargetPort: intstr.FromInt32(24224),
				Port:       514,
			},
		},
		func(o runtime.Object) {
			runtime.SetCommonLabels(o, syslogDeployment.Name, syslogDeployment.Name, syslogDeployment.Name)
		},
	)
	service.Spec.Type = serviceType

	sOpts := metav1.CreateOptions{}
	service, err = tc.KubeClient.CoreV1().Services(syslogDeployment.Namespace).Create(context.TODO(), service, sOpts)
	if err != nil {
		return nil, err
	}
	tc.AddCleanup(func() error {
		opts := metav1.DeleteOptions{}
		err = tc.KubeClient.CoreV1().Services(syslogDeployment.Namespace).Delete(context.TODO(), service.Name, opts)
		if apierrors.IsNotFound(err) {
			return nil
		}
		return err
	})
	return service, nil
}

func (tc *E2ETestFramework) CreateSyslogReceiverSecrets(namespace, logStoreName, secretName string) (secret *corev1.Secret, err error) {
	ca := certificate.NewCA(nil, "Root CA") // Self-signed CA
	serverCert := certificate.NewCert(ca, "", logStoreName, fmt.Sprintf("%s.%s.svc", logStoreName, namespace))

	data := map[string][]byte{
		"tls.key":       serverCert.PrivateKeyPEM(),
		"tls.crt":       serverCert.CertificatePEM(),
		"ca-bundle.crt": ca.CertificatePEM(),
		"ca.key":        ca.PrivateKeyPEM(),
	}

	sOpts := metav1.CreateOptions{}
	secret = runtime.NewSecret(
		namespace,
		secretName,
		data,
	)
	clolog.V(3).Info("Creating secret for logStore", "secret", secret.Name, "logStore", logStoreName)
	if secret, err = tc.KubeClient.CoreV1().Secrets(namespace).Create(context.TODO(), secret, sOpts); err != nil {
		return nil, err
	}
	return secret, nil
}
