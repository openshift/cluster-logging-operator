package e2e

import (
	"context"
	"errors"
	"fmt"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	"strconv"
	"strings"
	"time"

	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/factory"
	fluentdhelpers "github.com/openshift/cluster-logging-operator/test/helpers/fluentd"
	"github.com/openshift/cluster-logging-operator/test/helpers/types"

	apps "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"

	clolog "github.com/ViaQ/logerr/v2/log/static"
	"github.com/openshift/cluster-logging-operator/internal/k8shandler"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	"github.com/openshift/cluster-logging-operator/test/helpers/oc"
)

type fluentReceiverLogStore struct {
	deployment     *apps.Deployment
	tc             *E2ETestFramework
	pipelineSecret *corev1.Secret
}

const (
	receiverName             = "fluent-receiver"
	FluentdSecretName        = receiverName
	FluentdSharedKey         = constants.SharedKey
	secureFluentConfTemplate = `
<system>
	log_level info
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
<match linux-audit.log** k8s-audit.log** openshift-audit.log** ovn-audit.log**>
  @type file
  append true
  path /tmp/audit.logs
  symlink_path /tmp/audit-logs
</match>
<match **>
	@type stdout
</match>
	`
	UnsecureFluentConf = `
<system>
	log_level trace
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
<match linux-audit.log** k8s-audit.log** openshift-audit.log** ovn-audit.log**>
  @type file
  append true
  path /tmp/audit.logs
  symlink_path /tmp/audit-logs
</match>
<match **>
	@type stdout
</match>
	`
	FluentConfHTTPWithTLS = `
<system>
  log_level info
</system>
<source>
  @type http
  port 24224
  bind 0.0.0.0
  body_size_limit 32m
  keepalive_timeout 10s
  # Headers are capitalized, and added with prefix "HTTP_"
  add_http_headers true
  add_remote_addr true
  <parse>
    @type json
  </parse>
  <transport tls>
	  ca_path /etc/fluentd/secrets/ca-bundle.crt
	  cert_path /etc/fluentd/secrets/tls.crt
	  private_key_path /etc/fluentd/secrets/tls.key
  </transport>
</source>

<match logs.application>
  @type file
  append true
  path /tmp/app.logs
  symlink_path /tmp/app-logs
</match>
<match logs.infrastructure>
  @type file
  append true
  path /tmp/infra.logs
  symlink_path /tmp/infra-logs
</match>
<match logs.audit>
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
	pods, err := fluent.tc.KubeClient.CoreV1().Pods(fluent.deployment.Namespace).List(context.TODO(), options)
	if err != nil {
		return false, err
	}
	if len(pods.Items) == 0 {
		return false, errors.New("No pods found for fluent receiver")
	}
	clolog.V(3).Info("Pod ", "PodName", pods.Items[0].Name)
	cmd := fmt.Sprintf("ls %s | wc -l", file)

	err = wait.PollUntilContextTimeout(context.TODO(), defaultRetryInterval, timeToWait, true, func(cxt context.Context) (done bool, err error) {
		output, err := fluent.tc.PodExec(fluent.deployment.Namespace, pods.Items[0].Name, "fluent-receiver", []string{"bash", "-c", cmd})
		if err != nil {
			clolog.Error(err, "Error polling fluent-receiver for logs")
			return false, nil
		}
		value, err := strconv.Atoi(strings.TrimSpace(output))
		if err != nil {
			clolog.V(3).Info("Error parsing output: ", "output", output)
			return false, nil
		}
		return value > 0, nil
	})
	if wait.Interrupted(err) {
		return false, err
	}
	return true, err
}

func (fluent *fluentReceiverLogStore) logs(file string, timeToWait time.Duration) (string, error) {
	options := metav1.ListOptions{
		LabelSelector: "component=fluent-receiver",
	}
	pods, err := fluent.tc.KubeClient.CoreV1().Pods(fluent.deployment.Namespace).List(context.TODO(), options)
	if err != nil {
		return "", err
	}
	if len(pods.Items) == 0 {
		return "", errors.New("No pods found for fluent receiver")
	}
	clolog.V(3).Info("Pod ", "PodName", pods.Items[0].Name)
	cmd := fmt.Sprintf("cat %s | awk -F '\t' '{print $3}'| head -n 1", file)
	result := ""
	err = wait.PollUntilContextTimeout(context.TODO(), defaultRetryInterval, timeToWait, true, func(cxt context.Context) (done bool, err error) {
		if result, err = fluent.tc.PodExec(fluent.deployment.Namespace, pods.Items[0].Name, "fluent-receiver", []string{"bash", "-c", cmd}); err != nil {
			clolog.Error(err, "Failed to fetch logs from fluent-receiver ")
			return false, nil
		}
		return true, nil
	})
	if wait.Interrupted(err) {
		return "", err
	}
	return result, nil
}

func (fluent *fluentReceiverLogStore) ApplicationLogs(timeToWait time.Duration) (types.Logs, error) {
	fl, err := fluent.logs("/tmp/app-logs", timeToWait)
	if err != nil {
		return nil, err
	}
	out := "[" + strings.TrimRight(strings.Replace(fl, "\n", ",", -1), ",") + "]"
	return types.ParseLogs(out)
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

func (fluent *fluentReceiverLogStore) GrepLogs(expr string, timeToWait time.Duration) (string, error) {
	return "Not Found", fmt.Errorf("Not implemented")
}

func (fluent *fluentReceiverLogStore) RetrieveLogs() (map[string]string, error) {
	result := map[string]string{
		"infra": "",
		"audit": "",
		"app":   "",
	}
	var err error
	for key := range result {
		var s string
		s, err = fluent.logs(fmt.Sprintf("/tmp/%s-logs", key), 30*time.Second)
		if err != nil {
			continue
		}
		result[key] = s
	}
	return result, err
}

func (fluent *fluentReceiverLogStore) ClusterLocalEndpoint() string {
	return fmt.Sprintf("https://%s.%s.svc:24224", fluent.deployment.Name, fluent.deployment.Namespace)
}

func (tc *E2ETestFramework) DeployFluentdReceiver(rootDir string, secure bool) (deployment *apps.Deployment, err error) {
	if secure {
		return tc.DeployFluentdReceiverWithConf(constants.OpenshiftNS, secure, secureFluentConfTemplate)
	}
	return tc.DeployFluentdReceiverWithConf(constants.OpenshiftNS, secure, UnsecureFluentConf)
}

func (tc *E2ETestFramework) DeployFluentdReceiverWithConf(namespace string, secure bool, fluentConf string) (deployment *apps.Deployment, err error) {
	logStore := &fluentReceiverLogStore{
		tc: tc,
	}
	serviceAccount, err := tc.createServiceAccount(namespace, "fluent-receiver")
	if err != nil {
		return nil, err
	}
	if err := tc.createRbac(namespace, receiverName); err != nil {
		return nil, err
	}
	container := corev1.Container{
		Name:            receiverName,
		Image:           fluentdhelpers.Image,
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
			Privileged: utils.GetPtr(true),
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

	if secure {
		otherConf := map[string][]byte{
			FluentdSharedKey: []byte("my_shared_key"),
		}
		if logStore.pipelineSecret, err = tc.CreatePipelineSecret(namespace, receiverName, FluentdSecretName, otherConf); err != nil {
			return nil, err
		}
		tc.AddCleanup(func() error {
			opts := metav1.DeleteOptions{}
			return tc.KubeClient.CoreV1().Secrets(namespace).Delete(context.TODO(), FluentdSecretName, opts)
		})
		container.VolumeMounts = append(container.VolumeMounts, corev1.VolumeMount{
			Name:      "certs",
			ReadOnly:  true,
			MountPath: "/etc/fluentd/secrets",
		})
		podSpec.Containers = []corev1.Container{container}
		logStore.pipelineSecret.Data[FluentdSharedKey] = []byte("fluent-receiver")

		opts := metav1.UpdateOptions{}
		if logStore.pipelineSecret, err = tc.KubeClient.CoreV1().Secrets(namespace).Update(context.TODO(), logStore.pipelineSecret, opts); err != nil {
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

	config := runtime.NewConfigMap(namespace, container.Name, map[string]string{
		"fluent.conf": fluentConf,
	})

	tc.AddCleanup(func() error {
		opts := metav1.DeleteOptions{}
		return tc.KubeClient.CoreV1().ConfigMaps(namespace).Delete(context.TODO(), config.Name, opts)
	})
	if err = tc.Test.Recreate(config); err != nil {
		return nil, err
	}

	fluentDeployment := k8shandler.NewDeployment(
		container.Name,
		namespace,
		container.Name,
		serviceAccount.Name,
		podSpec,
	)
	// Add instance label to pod spec template. Service now selects using instance name as well
	fluentDeployment.Spec.Template.Labels[constants.LabelK8sInstance] = serviceAccount.Name

	if err = tc.Test.Recreate(fluentDeployment); err != nil {
		return nil, err
	}

	service := factory.NewService(
		serviceAccount.Name,
		namespace,
		serviceAccount.Name,
		serviceAccount.Name,
		[]corev1.ServicePort{
			{
				Port: 24224,
			},
		},
	)

	tc.AddCleanup(func() error {
		_, err := oc.Exec().
			WithNamespace(namespace).
			WithPodGetter(oc.Get().
				WithNamespace(namespace).
				Pod().
				Selector("component=fluent-receiver").
				OutputJsonpath("{.items[0].metadata.name}")).
			Container("fluent-receiver").
			WithCmd("/bin/sh", "-c", "rm -rf /tmp/app.logs /tmp/app-logs").
			Run()
		return err
	})
	tc.AddCleanup(func() error {
		var zerograce int64
		opts := metav1.DeleteOptions{
			GracePeriodSeconds: &zerograce,
		}
		return tc.KubeClient.AppsV1().Deployments(namespace).Delete(context.TODO(), fluentDeployment.Name, opts)
	})

	tc.AddCleanup(func() error {
		opts := metav1.DeleteOptions{}
		return tc.KubeClient.CoreV1().Services(namespace).Delete(context.TODO(), service.Name, opts)
	})
	if err = tc.Test.Recreate(service); err != nil {
		return nil, err
	}

	logStore.deployment = fluentDeployment
	name := fluentDeployment.GetName()
	tc.LogStores[name] = logStore
	return fluentDeployment, tc.WaitForDeployment(namespace, fluentDeployment.Name, defaultRetryInterval, defaultTimeout)
}
