package helpers

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/openshift/cluster-logging-operator/test/helpers/types"

	apps "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"

	clolog "github.com/ViaQ/logerr/log"
	"github.com/openshift/cluster-logging-operator/pkg/factory"
	"github.com/openshift/cluster-logging-operator/pkg/k8shandler"
	"github.com/openshift/cluster-logging-operator/pkg/utils"
)

type syslogReceiverLogStore struct {
	deployment *apps.Deployment
	tc         *E2ETestFramework
}

const (
	SyslogReceiverName = "syslog-receiver"
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
	pods, err := syslog.tc.KubeClient.CoreV1().Pods(OpenshiftLoggingNS).List(context.TODO(), options)
	if err != nil {
		return false, err
	}
	if len(pods.Items) == 0 {
		return false, errors.New("No pods found for syslog receiver")
	}
	podName := pods.Items[0].Name
	cmd := fmt.Sprintf("ls %s | wc -l", file)

	err = wait.Poll(defaultRetryInterval, timeToWait, func() (done bool, err error) {
		output, err := syslog.tc.PodExec(OpenshiftLoggingNS, podName, "syslog-receiver", []string{"bash", "-c", cmd})
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
	pods, err := syslog.tc.KubeClient.CoreV1().Pods(OpenshiftLoggingNS).List(context.TODO(), options)
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
		output, err := syslog.tc.PodExec(OpenshiftLoggingNS, pods.Items[0].Name, "syslog-receiver", []string{"bash", "-c", cmd})
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
	return syslog.hasLogs("/var/log/infra.log", timeToWait)
}

func (syslog *syslogReceiverLogStore) HasApplicationLogs(timeToWait time.Duration) (bool, error) {
	return false, fmt.Errorf("Not implemented")
}

func (syslog *syslogReceiverLogStore) HasAuditLogs(timeToWait time.Duration) (bool, error) {
	return false, fmt.Errorf("Not implemented")
}

func (syslog *syslogReceiverLogStore) GrepLogs(expr string, timeToWait time.Duration) (string, error) {
	return syslog.grepLogs(expr, "/var/log/infra.log", timeToWait)
}

func (syslog *syslogReceiverLogStore) RetrieveLogs() (map[string]string, error) {
	return nil, fmt.Errorf("Not implemented")
}

func (syslog *syslogReceiverLogStore) ClusterLocalEndpoint() string {
	panic("Not implemented")
}

func (tc *E2ETestFramework) createSyslogServiceAccount() (serviceAccount *corev1.ServiceAccount, err error) {
	opts := metav1.CreateOptions{}
	serviceAccount = k8shandler.NewServiceAccount("syslog-receiver", OpenshiftLoggingNS)
	if serviceAccount, err = tc.KubeClient.CoreV1().ServiceAccounts(OpenshiftLoggingNS).Create(context.TODO(), serviceAccount, opts); err != nil {
		return nil, err
	}
	tc.AddCleanup(func() error {
		opts := metav1.DeleteOptions{}
		return tc.KubeClient.CoreV1().ServiceAccounts(OpenshiftLoggingNS).Delete(context.TODO(), serviceAccount.Name, opts)
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
		return tc.KubeClient.CoreV1().ConfigMaps(namespace).Delete(context.TODO(), fluentdConfigMap.Name, opts)
	})
	return nil
}

func (tc *E2ETestFramework) createSyslogRbac(name string) (err error) {
	opts := metav1.CreateOptions{}
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

	if _, err = tc.KubeClient.RbacV1().Roles(OpenshiftLoggingNS).Create(context.TODO(), saRole, opts); err != nil {
		return err
	}

	tc.AddCleanup(func() error {
		opts := metav1.DeleteOptions{}
		return tc.KubeClient.RbacV1().Roles(OpenshiftLoggingNS).Delete(context.TODO(), name, opts)
	})

	rbOpts := metav1.CreateOptions{}
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

	if _, err = tc.KubeClient.RbacV1().RoleBindings(OpenshiftLoggingNS).Create(context.TODO(), roleBinding, rbOpts); err != nil {
		return err
	}
	tc.AddCleanup(func() error {
		opts := metav1.DeleteOptions{}
		return tc.KubeClient.RbacV1().RoleBindings(OpenshiftLoggingNS).Delete(context.TODO(), name, opts)
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
			return tc.KubeClient.CoreV1().Secrets(OpenshiftLoggingNS).Delete(context.TODO(), SyslogReceiverName, opts)
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
	config := k8shandler.NewConfigMap(container.Name, OpenshiftLoggingNS, map[string]string{
		"rsyslog.conf": rsyslogConf,
	})
	config, err = tc.KubeClient.CoreV1().ConfigMaps(OpenshiftLoggingNS).Create(context.TODO(), config, cOpts)
	if err != nil {
		return nil, err
	}
	tc.AddCleanup(func() error {
		opts := metav1.DeleteOptions{}
		return tc.KubeClient.CoreV1().ConfigMaps(OpenshiftLoggingNS).Delete(context.TODO(), config.Name, opts)
	})

	dOpts := metav1.CreateOptions{}
	syslogDeployment := k8shandler.NewDeployment(
		container.Name,
		OpenshiftLoggingNS,
		container.Name,
		serviceAccount.Name,
		podSpec,
	)

	syslogDeployment, err = tc.KubeClient.AppsV1().Deployments(OpenshiftLoggingNS).Create(context.TODO(), syslogDeployment, dOpts)
	if err != nil {
		return nil, err
	}
	service := factory.NewService(
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
		return tc.KubeClient.AppsV1().Deployments(OpenshiftLoggingNS).Delete(context.TODO(), syslogDeployment.Name, deleteopts)
	})

	sOpts := metav1.CreateOptions{}
	service, err = tc.KubeClient.CoreV1().Services(OpenshiftLoggingNS).Create(context.TODO(), service, sOpts)
	if err != nil {
		return nil, err
	}
	tc.AddCleanup(func() error {
		opts := metav1.DeleteOptions{}
		return tc.KubeClient.CoreV1().Services(OpenshiftLoggingNS).Delete(context.TODO(), service.Name, opts)
	})
	logStore.deployment = syslogDeployment

	name := syslogDeployment.GetName()
	tc.LogStores[name] = logStore
	return syslogDeployment, tc.waitForDeployment(OpenshiftLoggingNS, syslogDeployment.Name, defaultRetryInterval, defaultTimeout)
}

func (tc *E2ETestFramework) CreateSyslogReceiverSecrets(testDir, logStoreName, secretName string) (*corev1.Secret, error) {
	workingDir := fmt.Sprintf("/tmp/clo-test-%d", rand.Intn(10000))
	clolog.V(3).Info("Generating Pipeline certificates for", "rsyslog-receiver", workingDir)
	if _, err := os.Stat(workingDir); os.IsNotExist(err) {
		if err = os.MkdirAll(workingDir, 0766); err != nil {
			return nil, err
		}
	}
	if err := os.Setenv("WORKING_DIR", workingDir); err != nil {
		return nil, err
	}
	script := fmt.Sprintf("%s/syslog_cert_generation.sh", testDir)
	clolog.Info("Running script '%s %s %s %s'", "script", script, "workingdir", workingDir, "namespace", OpenshiftLoggingNS, "logStore", logStoreName)
	cmd := exec.Command(script, workingDir, OpenshiftLoggingNS, logStoreName)
	result, err := cmd.Output()

	if clolog.V(3).Enabled() {
		clolog.V(3).Info("cert_generation :", "output", string(result))
	}
	if err != nil {
		clolog.V(3).Error(err, "Error:")
	}

	data := map[string][]byte{
		"tls.key":       utils.GetWorkingDirFileContents("syslog-server.key"),
		"tls.crt":       utils.GetWorkingDirFileContents("syslog-server.crt"),
		"ca-bundle.crt": utils.GetWorkingDirFileContents("ca-syslog.crt"),
		"ca.key":        utils.GetWorkingDirFileContents("ca-syslog.key"),
	}

	sOpts := metav1.CreateOptions{}
	secret := k8shandler.NewSecret(
		secretName,
		OpenshiftLoggingNS,
		data,
	)
	clolog.V(3).Info("Creating secret for logStore", "secret", secret.Name, "logStore", logStoreName)
	if secret, err = tc.KubeClient.CoreV1().Secrets(OpenshiftLoggingNS).Create(context.TODO(), secret, sOpts); err != nil {
		return nil, err
	}
	return secret, nil
}
