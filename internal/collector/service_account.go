package collector

import (
	"context"
	"fmt"

	security "github.com/openshift/api/security/v1"
	"github.com/openshift/cluster-logging-operator/internal/factory"
	"github.com/openshift/cluster-logging-operator/internal/reconcile"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const sccName = "log-collector-scc"

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

	DesiredSCCVolumes = []security.FSType{"configMap", "secret", "emptyDir", "projected"}
)

// ReconcileServiceAccount reconciles the serviceaccount specifically for a collector deployment
func ReconcileServiceAccount(er record.EventRecorder, k8sClient client.Client, namespace string, resNames *factory.ForwarderResourceNames, owner metav1.OwnerReference) (err error) {
	serviceAccount := runtime.NewServiceAccount(namespace, resNames.ServiceAccount)
	utils.AddOwnerRefToObject(serviceAccount, owner)
	serviceAccount.ObjectMeta.Finalizers = append(serviceAccount.ObjectMeta.Finalizers, metav1.FinalizerDeleteDependents)
	if serviceAccount, err = reconcile.ServiceAccount(er, k8sClient, serviceAccount); err != nil {
		return err
	}
	if err = reconcile.SecurityContextConstraints(k8sClient, NewSCC()); err != nil {
		return err
	}
	return reconcileServiceAccountTokenSecret(serviceAccount, k8sClient, namespace, resNames.ServiceAccountTokenSecret, owner)
}

func reconcileServiceAccountTokenSecret(serviceAccount *corev1.ServiceAccount, k8sClient client.Client, namespace, name string, owner metav1.OwnerReference) error {
	desired := runtime.NewSecret(namespace, name, map[string][]byte{})
	desired.Annotations = map[string]string{
		corev1.ServiceAccountNameKey: serviceAccount.Name,
		corev1.ServiceAccountUIDKey:  string(serviceAccount.UID),
	}
	desired.Type = corev1.SecretTypeServiceAccountToken
	utils.AddOwnerRefToObject(desired, owner)
	current := &corev1.Secret{}
	if err := k8sClient.Get(context.TODO(), client.ObjectKeyFromObject(desired), current); err == nil {
		accountName := desired.Annotations[corev1.ServiceAccountNameKey]
		accountUID := desired.Annotations[corev1.ServiceAccountUIDKey]
		if (accountName != serviceAccount.Name || accountUID != string(serviceAccount.UID)) &&
			!utils.HasSameOwner(current.OwnerReferences, desired.OwnerReferences) {
			// Delete secret, so that we can create a new one next loop
			if err := k8sClient.Delete(context.TODO(), current); err != nil {
				return nil
			}
			return fmt.Errorf("deleted stale secret: %s", name)
		}
		// Existing secret is up-to-date
		return nil
	} else if !errors.IsNotFound(err) {
		return fmt.Errorf("failed to get %s token secret: %w", name, err)
	}

	if err := k8sClient.Create(context.TODO(), desired); err != nil {
		return fmt.Errorf("failed to create %s token secret: %w", name, err)
	}

	return nil
}

func NewSCC() *security.SecurityContextConstraints {

	scc := runtime.NewSCC(sccName)
	scc.AllowPrivilegedContainer = false
	scc.RequiredDropCapabilities = RequiredDropCapabilities
	scc.AllowHostDirVolumePlugin = true
	scc.Volumes = DesiredSCCVolumes
	scc.DefaultAllowPrivilegeEscalation = utils.GetBool(false)
	scc.AllowPrivilegeEscalation = utils.GetBool(false)
	scc.RunAsUser = security.RunAsUserStrategyOptions{
		Type: security.RunAsUserStrategyRunAsAny,
	}
	scc.SELinuxContext = security.SELinuxContextStrategyOptions{
		Type: security.SELinuxStrategyRunAsAny,
	}
	scc.ReadOnlyRootFilesystem = true
	scc.ForbiddenSysctls = []string{"*"}
	scc.SeccompProfiles = []string{"runtime/default"}
	return scc
}
