package source

import (
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api"
	sourcesfile "github.com/openshift/cluster-logging-operator/internal/generator/vector/api/sources/file"
)

const (
	GlobalMinimumCooldown = 15000
	IgnoreOlderSecs       = 3600
	MaxLineBytes          = 3145728
	MaxReadBytes          = 262144
	RotateWaitSecs        = 5
)

type HostAuditLog struct {
	api.Config
}

func NewHostAuditLog(id string) HostAuditLog {
	f := sourcesfile.New("/var/log/audit/audit.log")
	f.HostKey = "hostname"
	f.GlobalMinimumCooldownMilliSeconds = GlobalMinimumCooldown
	f.IgnoreOlderSecs = IgnoreOlderSecs
	f.MaxLineBytes = MaxLineBytes
	f.MaxReadBytes = MaxReadBytes
	f.RotateWaitSecs = RotateWaitSecs
	return HostAuditLog{
		Config: api.Config{
			Sources: map[string]interface{}{
				id: f,
			},
		},
	}
}

type OpenshiftAuditLog struct {
	api.Config
}

func NewOpenshiftAuditLog(id string) OpenshiftAuditLog {
	f := sourcesfile.New("/var/log/oauth-apiserver/audit.log", "/var/log/openshift-apiserver/audit.log", "/var/log/oauth-server/audit.log")
	f.HostKey = "hostname"
	f.GlobalMinimumCooldownMilliSeconds = GlobalMinimumCooldown
	f.IgnoreOlderSecs = IgnoreOlderSecs
	f.MaxLineBytes = MaxLineBytes
	f.MaxReadBytes = MaxReadBytes
	f.RotateWaitSecs = RotateWaitSecs
	return OpenshiftAuditLog{
		Config: api.Config{
			Sources: map[string]interface{}{
				id: f,
			},
		},
	}
}

type K8sAuditLog struct {
	api.Config
}

func NewK8sAuditLog(id string) K8sAuditLog {
	f := sourcesfile.New("/var/log/kube-apiserver/audit.log")
	f.HostKey = "hostname"
	f.GlobalMinimumCooldownMilliSeconds = GlobalMinimumCooldown
	f.IgnoreOlderSecs = IgnoreOlderSecs
	f.MaxLineBytes = MaxLineBytes
	f.MaxReadBytes = MaxReadBytes
	f.RotateWaitSecs = RotateWaitSecs
	return K8sAuditLog{
		Config: api.Config{
			Sources: map[string]interface{}{
				id: f,
			},
		},
	}
}

type OVNAuditLog struct {
	api.Config
}

func NewOVNAuditLog(id string) K8sAuditLog {
	f := sourcesfile.New("/var/log/ovn/acl-audit-log.log")
	f.HostKey = "hostname"
	f.GlobalMinimumCooldownMilliSeconds = GlobalMinimumCooldown
	f.IgnoreOlderSecs = IgnoreOlderSecs
	f.MaxLineBytes = MaxLineBytes
	f.MaxReadBytes = MaxReadBytes
	f.RotateWaitSecs = RotateWaitSecs
	return K8sAuditLog{
		Config: api.Config{
			Sources: map[string]interface{}{
				id: f,
			},
		},
	}
}
