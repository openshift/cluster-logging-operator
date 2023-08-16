package loader

import (
	"context"
	"fmt"

	log "github.com/ViaQ/logerr/v2/log/static"
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/factory"
	"github.com/openshift/cluster-logging-operator/internal/migrations"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	"github.com/openshift/cluster-logging-operator/internal/validations/clusterlogforwarder"
	"github.com/openshift/cluster-logging-operator/internal/validations/clusterlogging"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// FetchClusterLogging, migrate and validate
func FetchClusterLogging(k8sClient client.Client, namespace, name string, skipMigrations bool) (clusterLogging logging.ClusterLogging, err error, migrationMessages []logging.Condition) {
	key := types.NamespacedName{Name: name, Namespace: namespace}
	proto := runtime.NewClusterLogging(namespace, name)
	if err := k8sClient.Get(context.TODO(), key, proto); err != nil {
		return logging.ClusterLogging{}, err, nil
	}
	// Do not modify cached copy
	clusterLogging = *proto.DeepCopy()
	if skipMigrations {
		return clusterLogging, nil, nil
	}
	// TODO Drop migration upon introduction of v2
	clusterLogging.Spec, migrationMessages = migrations.MigrateClusterLogging(clusterLogging.Spec)
	if err = clusterlogging.Validate(clusterLogging, k8sClient, map[string]bool{}); err != nil {
		return clusterLogging, err, migrationMessages
	}
	return clusterLogging, nil, migrationMessages
}

// FetchClusterLogForwarder, migrate and validate
func FetchClusterLogForwarder(k8sClient client.Client, namespace, name string, isClfReconcile bool, fetchClusterLogging func() logging.ClusterLogging) (forwarder logging.ClusterLogForwarder, err error, status *logging.ClusterLogForwarderStatus) {
	key := types.NamespacedName{Name: name, Namespace: namespace}
	proto := runtime.NewClusterLogForwarder(namespace, name)
	if err = k8sClient.Get(context.TODO(), key, proto); err != nil {
		if !apierrors.IsNotFound(err) {
			log.Error(err, "Encountered unexpected error getting", "forwarder", key)
			return logging.ClusterLogForwarder{}, err, nil
		}

		// CLF Custom resource deleted, needed to stop reconciliation
		if isClfReconcile {
			log.Info(fmt.Sprintf("clusterLogForwarder: %s/%s was deleted", namespace, name))
			return logging.ClusterLogForwarder{}, err, nil
		}
	}

	// Do not modify cached copy
	forwarder = *proto.DeepCopy()
	internalLogStoreSecret := factory.GenerateResourceNames(forwarder).InternalLogStoreSecret
	saTokenSecret := factory.GenerateResourceNames(forwarder).ServiceAccountTokenSecret
	// TODO Drop migration upon introduction of v2
	extras := map[string]bool{}
	forwarder.Spec, extras = migrations.MigrateClusterLogForwarderSpec(namespace, name, forwarder.Spec, fetchClusterLogging().Spec.LogStore, extras, internalLogStoreSecret, saTokenSecret)

	extras[constants.ClusterLoggingAvailable] = (fetchClusterLogging().Name != "")
	if err, status = clusterlogforwarder.Validate(forwarder, k8sClient, extras); err != nil {
		return forwarder, err, status
	}

	return forwarder, nil, status
}
