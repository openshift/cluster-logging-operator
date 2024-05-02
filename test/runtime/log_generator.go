package runtime

import (
	"fmt"
	"time"

	"github.com/openshift/cluster-logging-operator/internal/utils"

	"github.com/openshift/cluster-logging-operator/internal/runtime"

	corev1 "k8s.io/api/core/v1"
)

// NewLogGenerator creates a pod that will print `count` lines to stdout, waiting for
// `delay` between each line.  Lines are of the form "<timestamp> [n] `message`"
// where n is the number of lines output so far. Once done printing the pod will
// be idle but will not exit until deleted.
// If count <= 0, print lines until killed.
func NewLogGenerator(namespace, name string, msgCount int, delay time.Duration, message string) *corev1.Pod {
	return NewMultiContainerLogGenerator(namespace, name, msgCount, delay, message, 1, map[string]string{})
}

func NewMultiContainerLogGenerator(namespace, name string, msgCount int, delay time.Duration, message string, containerCount int, labels map[string]string) *corev1.Pod {
	condition := "true"
	if msgCount > 0 {
		condition = fmt.Sprintf("[ $i -lt %v ]", msgCount)
	}
	cmd := fmt.Sprintf(`i=0; while %v; do echo "$(date) [ $i ]: %v"; i=$((i+1)); sleep %f; done; sleep infinity`, condition, message, delay.Seconds())
	containers := []corev1.Container{}
	containerName := name
	for i := 0; i < containerCount; i++ {
		if containerCount > 1 {
			containerName = fmt.Sprintf("%s-%d", name, i)
		}
		containers = append(containers, corev1.Container{
			Name:    containerName,
			Image:   "quay.io/quay/busybox",
			Command: []string{"sh", "-c", cmd},
			SecurityContext: &corev1.SecurityContext{
				AllowPrivilegeEscalation: utils.GetPtr(false),
				Capabilities: &corev1.Capabilities{
					Drop: []corev1.Capability{"ALL"},
				},
			},
		})
	}
	l := runtime.NewPod(namespace, name, containers...)
	l.Spec.RestartPolicy = corev1.RestartPolicyNever
	l.Spec.SecurityContext = &corev1.PodSecurityContext{
		RunAsNonRoot: utils.GetPtr(true),
		SeccompProfile: &corev1.SeccompProfile{
			Type: corev1.SeccompProfileTypeRuntimeDefault,
		},
	}
	for k, v := range labels {
		l.Labels[k] = v
	}
	return l
}

// NewCURLLogGenerator creates a pod that will cURL `count` lines to an endpoint, waiting for
// `delay` between each line.  Lines are of the form "<timestamp> [n] `message`"
// where n is the number of lines output so far. Once done printing the pod will
// be idle but will not exit until deleted.
// If count <= 0, print lines until killed.
func NewCURLLogGenerator(namespace, name, endpoint string, count int, delay time.Duration, message string) *corev1.Pod {
	condition := "true"
	if count > 0 {
		condition = fmt.Sprintf("[ $i -lt %v ]", count)
	}
	cmd := fmt.Sprintf(`sleep 15; i=0; while %v; do timestamp=$(date); message="{\"log_message\": \"[$i]: %s\", \"log_type\": \"audit\", \"@timestamp\": \"$timestamp\"}"; curl -ksv -X POST -H "Content-Type: application/json" -d "$message" %s; i=$((i+1)); sleep %f; done; sleep infinity`, condition, message, endpoint, delay.Seconds())
	l := runtime.NewPod(namespace, name, corev1.Container{
		Name:    name,
		Image:   "quay.io/curl/curl",
		Command: []string{"sh", "-c", cmd},
		SecurityContext: &corev1.SecurityContext{
			AllowPrivilegeEscalation: utils.GetPtr(false),
			Capabilities: &corev1.Capabilities{
				Drop: []corev1.Capability{"ALL"},
			},
		},
	})
	l.Spec.RestartPolicy = corev1.RestartPolicyNever
	l.Spec.SecurityContext = &corev1.PodSecurityContext{
		RunAsNonRoot: utils.GetPtr(true),
		SeccompProfile: &corev1.SeccompProfile{
			Type: corev1.SeccompProfileTypeRuntimeDefault,
		},
	}
	return l
}

// NewSocatPod creates pods with socat software which allow advance network call e.g. syslog message
func NewSocatPod(namespace, name, forwarderName string, labels map[string]string) *corev1.Pod {
	var containers []corev1.Container
	containerName := name

	containers = append(containers, corev1.Container{
		Name:    containerName,
		Image:   "quay.io/openshift-logging/alpine-socat:1.8.0.0",
		Command: []string{"sh", "-c", "sleep infinity"},
		SecurityContext: &corev1.SecurityContext{
			AllowPrivilegeEscalation: utils.GetPtr(false),
			Capabilities: &corev1.Capabilities{
				Drop: []corev1.Capability{"ALL"},
			},
		},
		VolumeMounts: []corev1.VolumeMount{{
			Name:      fmt.Sprintf("%s-syslog", forwarderName),
			ReadOnly:  true,
			MountPath: "/etc/collector/syslog",
		}},
	})

	pod := runtime.NewPod(namespace, name, containers...)
	pod.Spec.RestartPolicy = corev1.RestartPolicyNever
	pod.Spec.SecurityContext = &corev1.PodSecurityContext{
		RunAsNonRoot: utils.GetPtr(true),
		SeccompProfile: &corev1.SeccompProfile{
			Type: corev1.SeccompProfileTypeRuntimeDefault,
		},
	}
	for k, v := range labels {
		pod.Labels[k] = v
	}
	pod.Spec.Volumes = append(pod.Spec.Volumes, corev1.Volume{
		Name: fmt.Sprintf("%s-syslog", forwarderName), VolumeSource: corev1.VolumeSource{
			Secret: &corev1.SecretVolumeSource{
				SecretName: fmt.Sprintf("%s-syslog", forwarderName),
			},
		},
	})
	return pod
}
