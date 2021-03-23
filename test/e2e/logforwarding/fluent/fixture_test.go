package fluent_test

import (
	loggingv1 "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/test/client"
	"github.com/openshift/cluster-logging-operator/test/helpers/fluentd"
	. "github.com/openshift/cluster-logging-operator/test/matchers"
	"github.com/openshift/cluster-logging-operator/test/runtime"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apimachinery/pkg/watch"
	"time"
)

type Fixture struct {
	ClusterLogging      *loggingv1.ClusterLogging
	ClusterLogForwarder *loggingv1.ClusterLogForwarder
	Receiver            *fluentd.Receiver
	LogGenerator        *corev1.Pod
}

// NewTest returns a new test, the clf and receiver are not yet configured.
func NewFixture(namespace, message string) *Fixture {
	return &Fixture{
		ClusterLogging:      runtime.NewClusterLogging(),
		ClusterLogForwarder: runtime.NewClusterLogForwarder(),
		Receiver:            fluentd.NewReceiver(namespace, "receiver"),
		LogGenerator:        runtime.NewLogGenerator(namespace, "log-generator", 1000, 0, message),
	}
}

// Create resources, wait for them to be ready.
func (f *Fixture) Create(c *client.Client) {
	ExpectOK(c.Remove(f.ClusterLogging))
	ExpectOK(c.Remove(f.ClusterLogForwarder))
	ExpectOK(c.RemoveSync(&appsv1.DaemonSet{
		TypeMeta: metav1.TypeMeta{
			Kind:       "DaemonSet",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "fluentd",
			Namespace: f.ClusterLogging.Namespace,
		},
	}))
	f.cleanFluentDBuffers(c)
	ExpectOK(c.Recreate(f.ClusterLogging))
	ExpectOK(c.Recreate(f.ClusterLogForwarder))
	ExpectOK(c.WaitFor(f.ClusterLogForwarder, client.ClusterLogForwarderReady))
	ExpectOK(c.WaitFor(
		&appsv1.DaemonSet{
			TypeMeta: metav1.TypeMeta{
				Kind:       "DaemonSet",
				APIVersion: "apps/v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "fluentd",
				Namespace: f.ClusterLogging.Namespace,
			},
		},
		func(e watch.Event) (bool, error) {
			ds := e.Object.(*appsv1.DaemonSet)
			return ds.Status.NumberUnavailable == 0 &&
				ds.Status.CurrentNumberScheduled == ds.Status.DesiredNumberScheduled &&
				ds.Status.CurrentNumberScheduled == ds.Status.NumberReady &&
				ds.Status.CurrentNumberScheduled == ds.Status.NumberAvailable, nil
		},
	))
	ExpectOK(c.Create(f.LogGenerator))
	ExpectOK(c.WaitFor(f.LogGenerator, client.PodRunning))
	ExpectOK(f.Receiver.Create(c))
}

func (f *Fixture) cleanFluentDBuffers(c *client.Client) {
	h := corev1.HostPathDirectory
	p := true
	ds := &appsv1.DaemonSet{
		TypeMeta: metav1.TypeMeta{
			Kind:       "DaemonSet",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "clean-fluentd-buffers",
			Namespace: "default",
		},
		Spec: appsv1.DaemonSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"name": "clean-fluentd-buffers",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"name": "clean-fluentd-buffers",
					},
				},
				Spec: corev1.PodSpec{
					Tolerations: []corev1.Toleration{
						{
							Key:    "node-role.kubernetes.io/master",
							Effect: corev1.TaintEffectNoSchedule,
						},
					},
					InitContainers: []corev1.Container{
						{
							Name:  "clean-fluentd-buffers",
							Image: "docker.io/library/busybox:latest",
							Args:  []string{"sh", "-c", "rm -rf /fluentd-buffers/** || rm /logs/audit/audit.log.pos || rm /logs/kube-apiserver/audit.log.pos || rm /logs/es-containers.log.pos"},
							SecurityContext: &corev1.SecurityContext{
								Privileged: &p,
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "fluentd-buffers",
									MountPath: "/fluentd-buffers",
								},
								{
									Name:      "logs",
									MountPath: "/logs",
								},
							},
						},
					},
					Containers: []corev1.Container{
						{
							Name:  "pause",
							Image: "centos:centos7",
							Args:  []string{"sh", "-c", "echo done!!!!"},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "fluentd-buffers",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/var/lib/fluentd",
									Type: &h,
								},
							},
						},
						{
							Name: "logs",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/var/log",
									Type: &h,
								},
							},
						},
					},
				},
			},
		},
	}

	ExpectOK(c.Recreate(ds))
	_ = wait.PollImmediate(time.Second*10, time.Minute*5, func() (bool, error) {
		ExpectOK(c.Get(ds))
		return ds.Status.DesiredNumberScheduled == ds.Status.CurrentNumberScheduled, nil
	})
	ExpectOK(c.RemoveSync(ds))
}
