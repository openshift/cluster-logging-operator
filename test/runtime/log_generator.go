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
func NewLogGenerator(namespace, name string, count int, delay time.Duration, message string) *corev1.Pod {
	condition := "true"
	if count > 0 {
		condition = fmt.Sprintf("[ $i -lt %v ]", count)
	}
	cmd := fmt.Sprintf(`i=0; while %v; do echo "$(date) [ $i ]: %v"; i=$((i+1)); sleep %f; done; sleep infinity`, condition, message, delay.Seconds())
	l := runtime.NewPod(namespace, name, corev1.Container{
		Name:    name,
		Image:   "quay.io/quay/busybox",
		Command: []string{"sh", "-c", cmd},
		SecurityContext: &corev1.SecurityContext{
			AllowPrivilegeEscalation: utils.GetBool(false),
			Capabilities: &corev1.Capabilities{
				Drop: []corev1.Capability{"ALL"},
			},
		},
	})
	l.Spec.RestartPolicy = corev1.RestartPolicyNever
	l.Spec.SecurityContext = &corev1.PodSecurityContext{
		RunAsNonRoot: utils.GetBool(true),
		SeccompProfile: &corev1.SeccompProfile{
			Type: corev1.SeccompProfileTypeRuntimeDefault,
		},
	}
	return l
}
