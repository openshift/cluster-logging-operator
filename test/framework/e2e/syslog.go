package e2e

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/openshift/cluster-logging-operator/test/helpers/certificate"
	rbacv1 "k8s.io/api/rbac/v1"

	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/factory"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	apierrors "k8s.io/apimachinery/pkg/api/errors"

	"github.com/openshift/cluster-logging-operator/test/helpers/types"

	apps "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"

	clolog "github.com/ViaQ/logerr/v2/log/static"
	"github.com/openshift/cluster-logging-operator/internal/k8shandler"
	"github.com/openshift/cluster-logging-operator/internal/utils"
)

type syslogReceiverLogStore struct {
	deployment *apps.Deployment
	tc         *E2ETestFramework
}

const (
	SyslogReceiverName = "syslog-receiver"
	ImageRemoteSyslog  = "registry.redhat.io/rhel8/rsyslog:8.7-9"
)

// SyslogRfc type is the rfc used for sending syslog
type SyslogRfc int

const (
	// RFC3164 rfc3164
	RFC3164 SyslogRfc = iota
	// RFC5424 rfc5424
	RFC5424
	// RFC3164RFC5424 either rfc3164 or rfc5424
	RFC3164RFC5424
)

func MustParseRFC(rfc string) SyslogRfc {
	switch strings.ToUpper(rfc) {
	case "RFC3164":
		return RFC3164
	case "RFC5424":
		return RFC5424
	case "RFC3164 or RFC5424":
		return RFC3164RFC5424
	}
	log.Fatal("Unable to parse RFC", "rfc", rfc)
	return 0
}

func (e SyslogRfc) String() string {
	switch e {
	case RFC3164:
		return "RFC3164"
	case RFC5424:
		return "RFC5424"
	case RFC3164RFC5424:
		return "RFC3164 or RFC5424"
	default:
		return "Unknown rfc"
	}
}

func GenerateRsyslogConf(conf string, rfc SyslogRfc) string {
	switch rfc {
	case RFC5424:
		return strings.Join([]string{conf, RuleSetRfc5424}, "\n")
	case RFC3164:
		return strings.Join([]string{conf, RuleSetRfc3164}, "\n")
	case RFC3164RFC5424:
		return strings.Join([]string{conf, RuleSetRfc3164Rfc5424}, "\n")
	}
	return "Invalid Conf"
}

func (syslog *syslogReceiverLogStore) hasLogs(file string, timeToWait time.Duration) (bool, error) {
	options := metav1.ListOptions{
		LabelSelector: "component=syslog-receiver",
	}
	pods, err := syslog.tc.KubeClient.CoreV1().Pods(constants.OpenshiftNS).List(context.TODO(), options)
	if err != nil {
		return false, err
	}
	if len(pods.Items) == 0 {
		return false, errors.New("No pods found for syslog receiver")
	}
	podName := pods.Items[0].Name
	cmd := fmt.Sprintf("ls %s | wc -l", file)
	err = wait.Poll(defaultRetryInterval, timeToWait, func() (done bool, err error) {
		output, err := syslog.tc.PodExec(constants.OpenshiftNS, podName, "syslog-receiver", []string{"bash", "-c", cmd})
		if err != nil {
			clolog.Error(err, "failed to fetch logs from syslog-receiver")
			return false, nil
		}
		value, err := strconv.Atoi(strings.TrimSpace(output))
		if err != nil {
			clolog.V(2).Error(err, "Error parsing output", "output", output)
			return false, nil
		}
		return value > 0, nil
	})
	if err == wait.ErrWaitTimeout {
		return false, err
	}
	return true, err
}

func (syslog *syslogReceiverLogStore) grepLogs(expr string, logfile string, timeToWait time.Duration) (string, error) {
	NotFound := "No Found"
	options := metav1.ListOptions{
		LabelSelector: "component=syslog-receiver",
	}
	pods, err := syslog.tc.KubeClient.CoreV1().Pods(constants.WatchNamespace).List(context.TODO(), options)
	if err != nil {
		return NotFound, err
	}
	if len(pods.Items) == 0 {
		return NotFound, errors.New("No pods found for syslog receiver")
	}
	clolog.V(3).Info("Pod", "PodName", pods.Items[0].Name)
	cmd := fmt.Sprintf(expr, logfile)
	clolog.V(3).Info("running expression", "expression", cmd)
	var value string

	err = wait.Poll(defaultRetryInterval, timeToWait, func() (bool, error) {
		output, err := syslog.tc.PodExec(constants.WatchNamespace, pods.Items[0].Name, "syslog-receiver", []string{"bash", "-c", cmd})
		if err != nil {
			clolog.Error(err, "failed to fetch logs from syslog-receiver")
			return false, nil
		}
		value = strings.TrimSpace(output)
		return true, nil
	})
	if err == wait.ErrWaitTimeout {
		return NotFound, err
	}
	return value, nil
}

func (syslog *syslogReceiverLogStore) ApplicationLogs(timeToWait time.Duration) (types.Logs, error) {
	panic("Method not implemented")
}

func (syslog *syslogReceiverLogStore) HasInfraStructureLogs(timeToWait time.Duration) (bool, error) {
	return syslog.hasLogs("/tmp/infra.log", timeToWait)
}

func (syslog *syslogReceiverLogStore) HasApplicationLogs(timeToWait time.Duration) (bool, error) {
	return false, fmt.Errorf("Not implemented")
}

func (syslog *syslogReceiverLogStore) HasAuditLogs(timeToWait time.Duration) (bool, error) {
	return false, fmt.Errorf("Not implemented")
}

func (syslog *syslogReceiverLogStore) GrepLogs(expr string, timeToWait time.Duration) (string, error) {
	return syslog.grepLogs(expr, "/tmp/infra.log", timeToWait)
}

func (syslog *syslogReceiverLogStore) RetrieveLogs() (map[string]string, error) {
	return nil, fmt.Errorf("Not implemented")
}

func (syslog *syslogReceiverLogStore) ClusterLocalEndpoint() string {
	panic("Not implemented")
}

func (tc *E2ETestFramework) createSyslogServiceAccount() (serviceAccount *corev1.ServiceAccount, err error) {
	opts := metav1.CreateOptions{}
	serviceAccount = runtime.NewServiceAccount(constants.WatchNamespace, "syslog-receiver")
	if serviceAccount, err = tc.KubeClient.CoreV1().ServiceAccounts(constants.WatchNamespace).Create(context.TODO(), serviceAccount, opts); err != nil {
		return nil, err
	}
	tc.AddCleanup(func() error {
		opts := metav1.DeleteOptions{}
		err := tc.KubeClient.CoreV1().ServiceAccounts(constants.WatchNamespace).Delete(context.TODO(), serviceAccount.Name, opts)
		if apierrors.IsNotFound(err) {
			return nil
		}
		return err
	})
	return serviceAccount, nil
}

func (tc *E2ETestFramework) CreateLegacySyslogConfigMap(namespace, conf string) (err error) {
	opts := metav1.CreateOptions{}
	fluentdConfigMap := k8shandler.NewConfigMap(
		"syslog",
		namespace,
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

func (tc *E2ETestFramework) createSyslogRbac(name string) (err error) {
	opts := metav1.CreateOptions{}
	saRole := runtime.NewRole(
		constants.WatchNamespace,
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

	if _, err = tc.KubeClient.RbacV1().Roles(constants.WatchNamespace).Create(context.TODO(), saRole, opts); err != nil {
		return err
	}

	tc.AddCleanup(func() error {
		opts := metav1.DeleteOptions{}
		err := tc.KubeClient.RbacV1().Roles(constants.WatchNamespace).Delete(context.TODO(), name, opts)
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
		constants.WatchNamespace,
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

	if _, err = tc.KubeClient.RbacV1().RoleBindings(constants.WatchNamespace).Create(context.TODO(), roleBinding, rbOpts); err != nil {
		return err
	}
	tc.AddCleanup(func() error {
		opts := metav1.DeleteOptions{}
		err := tc.KubeClient.RbacV1().RoleBindings(constants.WatchNamespace).Delete(context.TODO(), name, opts)
		if apierrors.IsNotFound(err) {
			return nil
		}
		return err
	})
	return nil
}

func (tc *E2ETestFramework) DeploySyslogReceiver(testDir string, protocol corev1.Protocol, withTLS bool, rfc SyslogRfc) (deployment *apps.Deployment, err error) {
	logStore := &syslogReceiverLogStore{
		tc: tc,
	}
	serviceAccount, err := tc.createSyslogServiceAccount()
	if err != nil {
		return nil, err
	}
	if err := tc.createSyslogRbac(SyslogReceiverName); err != nil {
		return nil, err
	}
	container := corev1.Container{
		Name:            SyslogReceiverName,
		Image:           ImageRemoteSyslog,
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
			AllowPrivilegeEscalation: utils.GetBool(false),
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
		},
		ServiceAccountName: serviceAccount.Name,
		SecurityContext: &corev1.PodSecurityContext{
			RunAsNonRoot: utils.GetBool(true),
			SeccompProfile: &corev1.SeccompProfile{
				Type: corev1.SeccompProfileTypeRuntimeDefault,
			},
		},
	}

	var rsyslogConf string
	switch {
	case protocol == corev1.ProtocolUDP:
		rsyslogConf = UdpSyslogInput

	default:
		rsyslogConf = TcpSyslogInput
	}

	if withTLS {
		switch {
		case protocol == corev1.ProtocolUDP:
			rsyslogConf = UdpSyslogInputWithTLS

		default:
			rsyslogConf = TcpSyslogInputWithTLS
		}
		secret, err := tc.CreateSyslogReceiverSecrets(testDir, SyslogReceiverName, SyslogReceiverName)
		if err != nil {
			return nil, err
		}
		tc.AddCleanup(func() error {
			opts := metav1.DeleteOptions{}
			err := tc.KubeClient.CoreV1().Secrets(constants.WatchNamespace).Delete(context.TODO(), SyslogReceiverName, opts)
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
					SecretName: secret.ObjectMeta.Name,
				},
			},
		})
	}

	rsyslogConf = GenerateRsyslogConf(rsyslogConf, rfc)

	cOpts := metav1.CreateOptions{}
	config := k8shandler.NewConfigMap(container.Name, constants.WatchNamespace, map[string]string{
		"rsyslog.conf": rsyslogConf,
	})
	config, err = tc.KubeClient.CoreV1().ConfigMaps(constants.WatchNamespace).Create(context.TODO(), config, cOpts)
	if err != nil {
		return nil, err
	}
	tc.AddCleanup(func() error {
		opts := metav1.DeleteOptions{}
		err := tc.KubeClient.CoreV1().ConfigMaps(constants.WatchNamespace).Delete(context.TODO(), config.Name, opts)
		if apierrors.IsNotFound(err) {
			return nil
		}
		return err
	})

	dOpts := metav1.CreateOptions{}
	syslogDeployment := k8shandler.NewDeployment(
		container.Name,
		constants.WatchNamespace,
		container.Name,
		serviceAccount.Name,
		podSpec,
	)

	syslogDeployment, err = tc.KubeClient.AppsV1().Deployments(constants.WatchNamespace).Create(context.TODO(), syslogDeployment, dOpts)
	if err != nil {
		return nil, err
	}
	service := factory.NewService(
		serviceAccount.Name,
		constants.WatchNamespace,
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
		err := tc.KubeClient.AppsV1().Deployments(constants.WatchNamespace).Delete(context.TODO(), syslogDeployment.Name, deleteopts)
		if apierrors.IsNotFound(err) {
			return nil
		}
		return err
	})

	sOpts := metav1.CreateOptions{}
	service, err = tc.KubeClient.CoreV1().Services(constants.WatchNamespace).Create(context.TODO(), service, sOpts)
	if err != nil {
		return nil, err
	}
	tc.AddCleanup(func() error {
		opts := metav1.DeleteOptions{}
		err := tc.KubeClient.CoreV1().Services(constants.WatchNamespace).Delete(context.TODO(), service.Name, opts)
		if apierrors.IsNotFound(err) {
			return nil
		}
		return err
	})
	logStore.deployment = syslogDeployment

	name := syslogDeployment.GetName()
	tc.LogStores[name] = logStore
	return syslogDeployment, tc.waitForDeployment(constants.WatchNamespace, syslogDeployment.Name, defaultRetryInterval, defaultTimeout)
}

func (tc *E2ETestFramework) CreateSyslogReceiverSecrets(testDir, logStoreName, secretName string) (secret *corev1.Secret, err error) {
	ca := certificate.NewCA(nil, "Root CA") // Self-signed CA
	serverCert := certificate.NewCert(ca, "", logStoreName, fmt.Sprintf("%s.%s.svc", logStoreName, constants.WatchNamespace))

	data := map[string][]byte{
		"tls.key":       serverCert.PrivateKeyPEM(),
		"tls.crt":       serverCert.CertificatePEM(),
		"ca-bundle.crt": ca.CertificatePEM(),
		"ca.key":        ca.PrivateKeyPEM(),
	}

	sOpts := metav1.CreateOptions{}
	secret = k8shandler.NewSecret(
		secretName,
		constants.WatchNamespace,
		data,
	)
	clolog.V(3).Info("Creating secret for logStore", "secret", secret.Name, "logStore", logStoreName)
	if secret, err = tc.KubeClient.CoreV1().Secrets(constants.WatchNamespace).Create(context.TODO(), secret, sOpts); err != nil {
		return nil, err
	}
	return secret, nil
}
