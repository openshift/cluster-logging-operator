package auth

import (
	"context"
	"fmt"

	security "github.com/openshift/api/security/v1"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	client "sigs.k8s.io/controller-runtime/pkg/client"
)

const sccName = "logging-scc"

var (
	RequiredDropCapabilities = []corev1.Capability{
		"CHOWN",
		"DAC_OVERRIDE",
		"FSETID",
		"FOWNER",
		"SETGID",
		"SETUID",
		"SETPCAP",
		"NET_BIND_SERVICE",
		"KILL",
	}

	DesiredSCCVolumes = []security.FSType{"configMap", "secret", "emptyDir", "projected", "hostPath"}
)

func NewSCC() *security.SecurityContextConstraints {

	scc := runtime.NewSCC(sccName)
	scc.AllowPrivilegedContainer = true // Allow init container to run privileged for data directory setup
	scc.RequiredDropCapabilities = RequiredDropCapabilities
	scc.AllowHostDirVolumePlugin = true
	scc.Volumes = DesiredSCCVolumes
	scc.DefaultAllowPrivilegeEscalation = utils.GetPtr(true)  // Allow privilege escalation for init container
	scc.AllowPrivilegeEscalation = utils.GetPtr(true)
	scc.RunAsUser = security.RunAsUserStrategyOptions{
		Type: security.RunAsUserStrategyRunAsAny, // Allow any UID (0 for init, 1000 for collector)
	}
	scc.SELinuxContext = security.SELinuxContextStrategyOptions{
		Type: security.SELinuxStrategyRunAsAny, // Allow any SELinux type (spc_t for init, container_logwriter_t for collector)
	}
	scc.ReadOnlyRootFilesystem = true
	scc.ForbiddenSysctls = []string{"*"}
	scc.SeccompProfiles = []string{"runtime/default"}
	return scc
}

func RemoveSecurityContextConstraint(k8sClient client.Client, sccName string) error {
	scc := runtime.NewSCC(sccName)

	err := k8sClient.Delete(context.TODO(), scc)
	if err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("failure deleting %v security context constraint %v", sccName, err)
	}
	return nil
}
