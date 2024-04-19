package e2e

import (
	logging "github.com/openshift/cluster-logging-operator/api/logging/v1"
)

// CreateServiceAccountAndAuthorizeFor creates a serviceaccount and binds all collection roles to it
func (e2e *E2ETestFramework) CreateServiceAccountAndAuthorizeFor(forwarder *logging.ClusterLogForwarder) error {
	sa, err := e2e.BuildAuthorizationFor(forwarder.Namespace, forwarder.Name).
		AllowClusterRole("collect-application-logs").
		AllowClusterRole("collect-infrastructure-logs").
		AllowClusterRole("collect-audit-logs").
		Create()
	if err != nil {
		return err
	}
	forwarder.Spec.ServiceAccountName = sa.Name
	return nil
}
