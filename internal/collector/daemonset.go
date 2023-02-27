package collector

import (
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/reconcile"
	"github.com/openshift/cluster-logging-operator/internal/tls"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ReconcileDaemonset reconciles a daemonset specifically for the collector defined by the factory
func (f *Factory) ReconcileDaemonset(er record.EventRecorder, k8sClient client.Client, namespace, name string, owner metav1.OwnerReference) error {
	trustedCABundle, trustHash := GetTrustedCABundle(k8sClient, namespace, constants.CollectorTrustedCAName)
	f.TrustedCAHash = trustHash
	tlsProfile, _ := tls.FetchAPIServerTlsProfile(k8sClient)
	desired := f.NewDaemonSet(namespace, name, trustedCABundle, tls.GetTLSProfileSpec(tlsProfile))
	utils.AddOwnerRefToObject(desired, owner)
	return reconcile.DaemonSet(er, k8sClient, desired)
}
