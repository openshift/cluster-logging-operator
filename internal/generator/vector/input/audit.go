package input

import (
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	generator "github.com/openshift/cluster-logging-operator/internal/generator/framework"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api"
	sourcesfile "github.com/openshift/cluster-logging-operator/internal/generator/vector/api/sources"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/filter/openshift/viaq/v1"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
)

const (
	GlobalMinimumCooldown = 15000
	IgnoreOlderSecs       = 3600
	MaxLineBytes          = 3145728
	MaxReadBytes          = 262144
	RotateWaitSecs        = 5
)

func NewAuditAuditdSource(input obs.InputSpec, op generator.Options) ([]generator.Element, []string) {
	hostID := helpers.MakeInputID(input.Name, "host")
	metaID := helpers.MakeID(hostID, "meta")
	el := []generator.Element{
		api.NewConfig(func(c *api.Config) {
			f := sourcesfile.NewFile("/var/log/audit/audit.log")
			f.HostKey = "hostname"
			f.GlobalMinimumCooldownMilliSeconds = GlobalMinimumCooldown
			f.IgnoreOlderSecs = IgnoreOlderSecs
			f.MaxLineBytes = MaxLineBytes
			f.MaxReadBytes = MaxReadBytes
			f.RotateWaitSecs = RotateWaitSecs
			c.Sources[hostID] = f
		}),
		NewInternalNormalization(metaID, obs.AuditSourceAuditd, obs.InputTypeAudit, hostID, v1.ParseHostAuditLogs),
	}
	return el, []string{metaID}

}

func NewK8sAuditSource(input obs.InputSpec, op generator.Options) ([]generator.Element, []string) {
	id := helpers.MakeInputID(input.Name, "kube")
	metaID := helpers.MakeID(id, "meta")
	el := []generator.Element{
		api.NewConfig(func(c *api.Config) {
			f := sourcesfile.NewFile("/var/log/kube-apiserver/audit.log")
			f.HostKey = "hostname"
			f.GlobalMinimumCooldownMilliSeconds = GlobalMinimumCooldown
			f.IgnoreOlderSecs = IgnoreOlderSecs
			f.MaxLineBytes = MaxLineBytes
			f.MaxReadBytes = MaxReadBytes
			f.RotateWaitSecs = RotateWaitSecs
			c.Sources[id] = f
		}),
		NewAuditInternalNormalization(metaID, obs.AuditSourceKube, id, true),
	}
	return el, []string{metaID}
}

func NewOpenshiftAuditSource(input obs.InputSpec, op generator.Options) ([]generator.Element, []string) {
	id := helpers.MakeInputID(input.Name, "openshift")
	metaID := helpers.MakeID(id, "meta")
	el := []generator.Element{
		api.NewConfig(func(c *api.Config) {
			f := sourcesfile.NewFile("/var/log/oauth-apiserver/audit.log", "/var/log/openshift-apiserver/audit.log", "/var/log/oauth-server/audit.log")
			f.HostKey = "hostname"
			f.GlobalMinimumCooldownMilliSeconds = GlobalMinimumCooldown
			f.IgnoreOlderSecs = IgnoreOlderSecs
			f.MaxLineBytes = MaxLineBytes
			f.MaxReadBytes = MaxReadBytes
			f.RotateWaitSecs = RotateWaitSecs
			c.Sources[id] = f
		}),
		NewAuditInternalNormalization(metaID, obs.AuditSourceOpenShift, id, true),
	}
	return el, []string{metaID}
}

func NewOVNAuditSource(input obs.InputSpec, op generator.Options) ([]generator.Element, []string) {
	id := helpers.MakeInputID(input.Name, "ovn")
	metaID := helpers.MakeID(id, "meta")
	el := []generator.Element{
		api.NewConfig(func(c *api.Config) {
			f := sourcesfile.NewFile("/var/log/ovn/acl-audit-log.log")
			f.HostKey = "hostname"
			f.GlobalMinimumCooldownMilliSeconds = GlobalMinimumCooldown
			f.IgnoreOlderSecs = IgnoreOlderSecs
			f.MaxLineBytes = MaxLineBytes
			f.MaxReadBytes = MaxReadBytes
			f.RotateWaitSecs = RotateWaitSecs
			c.Sources[id] = f
		}),
		NewInternalNormalization(metaID, obs.AuditSourceOVN, obs.InputTypeAudit, id),
	}
	return el, []string{metaID}
}
