package auth

import (
	"context"
	"fmt"

	security "github.com/openshift/api/security/v1"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
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

	AllowedCapabilities = []corev1.Capability{
		"DAC_READ_SEARCH",
	}

	DesiredSCCVolumes = []security.FSType{"configMap", "secret", "emptyDir", "projected", "hostPath"}
)

func NewSCC() *security.SecurityContextConstraints {
	scc := runtime.NewSCC(sccName)
	scc.AllowPrivilegedContainer = false // Forbids privileged containers
	scc.RequiredDropCapabilities = RequiredDropCapabilities
	scc.AllowedCapabilities = AllowedCapabilities
	scc.AllowHostDirVolumePlugin = true
	scc.Volumes = DesiredSCCVolumes
	scc.DefaultAllowPrivilegeEscalation = new(false)
	scc.AllowPrivilegeEscalation = new(false)
	scc.RunAsUser = security.RunAsUserStrategyOptions{
		Type: security.RunAsUserStrategyRunAsAny,
	}
	scc.SELinuxContext = security.SELinuxContextStrategyOptions{
		Type: security.SELinuxStrategyRunAsAny, // Permits spc_t on init container
	}
	scc.ReadOnlyRootFilesystem = true
	scc.ForbiddenSysctls = []string{"*"}
	scc.SeccompProfiles = []string{
		"runtime/default",
	}
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
