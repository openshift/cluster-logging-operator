package input

import (
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/adapters"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api"
	sourcesfile "github.com/openshift/cluster-logging-operator/internal/generator/vector/api/sources"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api/types"
	v1 "github.com/openshift/cluster-logging-operator/internal/generator/vector/filter/openshift/viaq/v1"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
)

const (
	GlobalMinimumCooldown = 15000
	IgnoreOlderSecs       = 3600
	MaxLineBytes          = 3145728
	MaxReadBytes          = 262144
	RotateWaitSecs        = 5
)

func NewAuditAuditdSource(input *adapters.Input) (id string, _ types.Source, tfs api.Transforms) {
	tfs = api.Transforms{}
	id = helpers.MakeInputID(input.Name, "host")
	metaID := helpers.MakeID(id, "meta")
	f := sourcesfile.NewFile("/var/log/audit/audit.log")
	f.HostKey = "hostname"
	f.GlobalMinimumCooldownMilliSeconds = GlobalMinimumCooldown
	f.IgnoreOlderSecs = IgnoreOlderSecs
	f.MaxLineBytes = MaxLineBytes
	f.MaxReadBytes = MaxReadBytes
	f.RotateWaitSecs = RotateWaitSecs
	tfs.Add(metaID, NewInternalNormalization(obs.AuditSourceAuditd, obs.InputTypeAudit, id, v1.ParseHostAuditLogs))
	input.Ids = append(input.Ids, metaID)
	return id, f, tfs

}

func NewK8sAuditSource(input *adapters.Input) (id string, _ types.Source, tfs api.Transforms) {
	tfs = api.Transforms{}
	id = helpers.MakeInputID(input.Name, "kube")
	metaID := helpers.MakeID(id, "meta")
	f := sourcesfile.NewFile("/var/log/kube-apiserver/audit.log")
	f.HostKey = "hostname"
	f.GlobalMinimumCooldownMilliSeconds = GlobalMinimumCooldown
	f.IgnoreOlderSecs = IgnoreOlderSecs
	f.MaxLineBytes = MaxLineBytes
	f.MaxReadBytes = MaxReadBytes
	f.RotateWaitSecs = RotateWaitSecs
	tfs.Add(metaID, NewAuditInternalNormalization(obs.AuditSourceKube, id, true))
	input.Ids = append(input.Ids, metaID)
	return id, f, tfs
}

func NewOpenshiftAuditSource(input *adapters.Input) (id string, _ types.Source, tfs api.Transforms) {
	tfs = api.Transforms{}
	id = helpers.MakeInputID(input.Name, "openshift")
	metaID := helpers.MakeID(id, "meta")
	f := sourcesfile.NewFile("/var/log/oauth-apiserver/audit.log", "/var/log/openshift-apiserver/audit.log", "/var/log/oauth-server/audit.log")
	f.HostKey = "hostname"
	f.GlobalMinimumCooldownMilliSeconds = GlobalMinimumCooldown
	f.IgnoreOlderSecs = IgnoreOlderSecs
	f.MaxLineBytes = MaxLineBytes
	f.MaxReadBytes = MaxReadBytes
	f.RotateWaitSecs = RotateWaitSecs
	tfs.Add(metaID, NewAuditInternalNormalization(obs.AuditSourceOpenShift, id, true))
	input.Ids = append(input.Ids, metaID)
	return id, f, tfs
}

func NewOVNAuditSource(input *adapters.Input) (id string, _ types.Source, tfs api.Transforms) {
	tfs = api.Transforms{}
	id = helpers.MakeInputID(input.Name, "ovn")
	metaID := helpers.MakeID(id, "meta")
	f := sourcesfile.NewFile("/var/log/ovn/acl-audit-log.log")
	f.HostKey = "hostname"
	f.GlobalMinimumCooldownMilliSeconds = GlobalMinimumCooldown
	f.IgnoreOlderSecs = IgnoreOlderSecs
	f.MaxLineBytes = MaxLineBytes
	f.MaxReadBytes = MaxReadBytes
	f.RotateWaitSecs = RotateWaitSecs
	tfs.Add(metaID, NewInternalNormalization(obs.AuditSourceOVN, obs.InputTypeAudit, id))
	input.Ids = append(input.Ids, metaID)
	return id, f, tfs
}
