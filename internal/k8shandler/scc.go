package k8shandler

import (
	. "github.com/openshift/api/security/v1"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/utils"
)

const LogCollectorSCCName = "log-collector-scc"

func NewSCC(name string) *SecurityContextConstraints {
	scc := SecurityContextConstraints{
		TypeMeta: metav1.TypeMeta{
			Kind:       "SecurityContextConstraints",
			APIVersion: SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Labels: map[string]string{
				"app.kubernetes.io/name":       name,
				"app.kubernetes.io/component":  constants.CollectorName,
				"app.kubernetes.io/created-by": constants.ClusterLoggingOperator,
				"app.kubernetes.io/managed-by": constants.ClusterLoggingOperator,
			},
		},
		AllowPrivilegedContainer: false,
		RequiredDropCapabilities: []corev1.Capability{
			"CHOWN",
			"DAC_OVERRIDE",
			"FSETID",
			"FOWNER",
			"SETGID",
			"SETUID",
			"SETPCAP",
			"NET_BIND_SERVICE",
			"KILL",
		},
		AllowHostDirVolumePlugin:        true,
		Volumes:                         []FSType{"configMap", "secret", "emptyDir", "projected"},
		DefaultAllowPrivilegeEscalation: utils.GetBool(false),
		AllowPrivilegeEscalation:        utils.GetBool(false),
		RunAsUser: RunAsUserStrategyOptions{
			Type: RunAsUserStrategyRunAsAny,
		},
		SELinuxContext: SELinuxContextStrategyOptions{
			Type: SELinuxStrategyRunAsAny,
		},
		ReadOnlyRootFilesystem: true,
		ForbiddenSysctls:       []string{"*"},
	}
	return &scc
}
