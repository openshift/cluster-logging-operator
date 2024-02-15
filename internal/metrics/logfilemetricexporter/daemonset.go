package logfilemetricexporter

import (
	loggingv1alpha1 "github.com/openshift/cluster-logging-operator/api/logging/v1alpha1"
	"github.com/openshift/cluster-logging-operator/internal/reconcile"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	"github.com/openshift/cluster-logging-operator/internal/tls"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ReconcileDaemonset reconciles the daemonset for the LogFileMetricExporter
func ReconcileDaemonset(exporter loggingv1alpha1.LogFileMetricExporter,
	er record.EventRecorder,
	k8sClient client.Client,
	namespace,
	name string,
	owner metav1.OwnerReference, visitors ...func(o runtime.Object)) error {

	tlsProfile, _ := tls.FetchAPIServerTlsProfile(k8sClient)
	desired := NewDaemonSet(exporter, namespace, name, tls.GetClusterTLSProfileSpec(tlsProfile), visitors...)
	utils.AddOwnerRefToObject(desired, owner)
	return reconcile.DaemonSet(er, k8sClient, desired)
}
