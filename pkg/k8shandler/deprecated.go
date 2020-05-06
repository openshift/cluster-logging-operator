package k8shandler

import (
	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/pkg/utils"
)

// TODO(alanconway) Support for backwards-compatibility, everything this file
// should be eliminated when the TP API is no longer supported.

const (
	//PreviewForwardingAnnotation enable preview instance with value "enabled"
	PreviewForwardingAnnotation = "clusterlogging.openshift.io/logforwardingtechpreview"
	// UseOldRemoteSyslogPlugin use old syslog plugin (docebo/fluent-plugin-remote-syslog)
	UseOldRemoteSyslogPlugin = "clusterlogging.openshift.io/useoldremotesyslogplugin"
)

// IsPreviewForwardingEnabled check if the tech-preview annotation is enabled.
// For use in upgrade code to check if an existing preview deployment needs to be upgraded.
func IsPreviewForwardingEnabled(cluster *logging.ClusterLogging) bool {
	if value, _ := utils.GetAnnotation(PreviewForwardingAnnotation, cluster.ObjectMeta); value == "enabled" {
		return true
	}
	return false
}

// Avoid "unused" lint error -  not used now but will be needed for upgrade code.
var _ = IsPreviewForwardingEnabled
