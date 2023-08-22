package collector

import (
	"context"
	"fmt"
	log "github.com/ViaQ/logerr/v2/log/static"
	"github.com/openshift/cluster-logging-operator/internal/reconcile"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	"github.com/openshift/cluster-logging-operator/internal/tls"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ReconcileDaemonset reconciles a daemonset specifically for the collector defined by the factory
func (f *Factory) ReconcileDaemonset(er record.EventRecorder, k8sClient client.Client, namespace string, owner metav1.OwnerReference) error {
	trustedCABundle, trustHash := GetTrustedCABundle(k8sClient, namespace, f.ResourceNames.CaTrustBundle)
	f.TrustedCAHash = trustHash
	tlsProfile, _ := tls.FetchAPIServerTlsProfile(k8sClient)
	desired := f.NewDaemonSet(namespace, f.ResourceNames.DaemonSetName(), trustedCABundle, tls.GetClusterTLSProfileSpec(tlsProfile))
	utils.AddOwnerRefToObject(desired, owner)
	return reconcile.DaemonSet(er, k8sClient, desired)
}

func Remove(k8sClient client.Client, namespace, name string) (err error) {
	log.V(3).Info("Removing collector", "namespace", namespace, "name", name)
	ds := runtime.NewDaemonSet(namespace, name)
	if err = k8sClient.Delete(context.TODO(), ds); err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("Failure deleting daemonset %s/%s: %v", namespace, name, err)
	}
	return nil
}
