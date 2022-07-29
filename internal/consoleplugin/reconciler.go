package consoleplugin

import (
	"context"
	"fmt"

	log "github.com/ViaQ/logerr/v2/log/static"
	consolev1alpha1 "github.com/openshift/api/console/v1alpha1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/apimachinery/pkg/util/intstr"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func init() {
	utilruntime.Must(consolev1alpha1.AddToScheme(scheme.Scheme))
}

// Reconciler reconciles the console plugin state with the desired configuration
type Reconciler struct {
	Config
	c client.Client

	consolePlugin consolev1alpha1.ConsolePlugin
	configMap     corev1.ConfigMap
	deployment    appv1.Deployment
	service       corev1.Service
}

// NewReconciler creates a Reconciler using client for config.
func NewReconciler(c client.Client, cf Config) *Reconciler {
	r := &Reconciler{Config: cf, c: c}
	_ = r.each(func(m mutable) error {
		if m.o == &r.consolePlugin {
			runtime.Initialize(m.o, "", r.Name) // Plugin is Cluster scope
		} else {
			runtime.Initialize(m.o, r.Namespace(), r.Name)
		}
		return nil
	})
	return r
}

// Reconcile creates or updates cluster objects to match config.
func (r *Reconciler) Reconcile(ctx context.Context) error {
	modified := false
	// Call CreateOrUpdate for each object.
	err := r.each(func(m mutable) error {
		return retry.RetryOnConflict(retry.DefaultBackoff, func() error {
			result, err := controllerutil.CreateOrUpdate(ctx, r.c, m.o, m.mutate)
			if err == nil && result != controllerutil.OperationResultNone {
				modified = true
				log.V(2).Info("reconciled", "object", runtime.ID(m.o), "action", result)
			}
			return err
		})
	})
	if err != nil {
		log.Error(err, "reconciling console", "plugin", runtime.ID(&r.consolePlugin))
		_ = r.Delete(ctx) // Clear out any partial setup
		return err
	}
	if modified {
		log.Info("reconciled console", "plugin", runtime.ID(&r.consolePlugin))
	}
	return nil
}

// Delete the consoleplugin and related objects.
func (r *Reconciler) Delete(ctx context.Context) error {
	var errs []error // Collect errors, don't stop on first.
	_ = r.each(func(m mutable) error {
		err := r.c.Delete(ctx, m.o)
		if err != nil && !apierrors.IsNotFound(err) {
			errs = append(errs, err)
			log.Error(err, "deleting console", "object", runtime.ID(m.o))
		}
		return nil // Don't stop on first error.
	})
	return utilerrors.NewAggregate(errs)
}

// each calls f for each object. Stops on first error and returns it.
func (r *Reconciler) each(f func(m mutable) error) error {
	for _, m := range []mutable{
		{&r.consolePlugin, r.mutateConsolePlugin},
		{&r.configMap, r.mutateConfigMap},
		{&r.deployment, r.mutateDeployment},
		{&r.service, r.mutateService},
	} {
		if err := f(m); err != nil {
			return err
		}
	}
	return nil
}

// mutable is an object and a mutate function that sets the desired state on the object.
// Suitable to be passed to controllerutil.CreateOrUpdate
type mutable struct {
	o      client.Object
	mutate controllerutil.MutateFn
}

// mutateCommon sets common labels for all objects.
func (r *Reconciler) mutateCommon(o client.Object) {
	o.SetLabels(map[string]string{
		constants.LabelApp:          r.Name,
		constants.LabelK8sName:      r.Name,
		constants.LabelK8sCreatedBy: r.CreatedBy(),
	})
}

// mutateOwned calls mutateCommon and also sets owner for owned objects.
func (r *Reconciler) mutateOwned(o client.Object) error {
	r.mutateCommon(o)
	return controllerutil.SetControllerReference(r.Owner, o, r.c.Scheme())
}

func (r *Reconciler) mutateConsolePlugin() error {
	o := &r.consolePlugin
	o.Spec = consolev1alpha1.ConsolePluginSpec{
		DisplayName: "Logging Console Plugin",
		Service: consolev1alpha1.ConsolePluginService{
			Name:      r.Name,
			Namespace: r.Namespace(),
			BasePath:  "/",
			Port:      r.nginxPort(),
		},
		Proxy: []consolev1alpha1.ConsolePluginProxy{{
			Type:      "Service",
			Alias:     "backend",
			Authorize: true,
			Service: consolev1alpha1.ConsolePluginProxyServiceConfig{
				Name:      r.LokiService,
				Namespace: r.Namespace(),
				Port:      r.LokiPort,
			},
		}},
	}
	r.mutateCommon(o)
	return nil
}

func (r *Reconciler) mutateConfigMap() error {
	o := &r.configMap
	o.Data = map[string]string{
		"nginx.conf": fmt.Sprintf(`
				error_log /dev/stdout info;
				events {}
				http {
					access_log         /dev/stdout;
					include            /etc/nginx/mime.types;
					default_type       application/octet-stream;
					keepalive_timeout  65;
					server {
						listen              %v ssl;
						ssl_certificate     /var/serving-cert/tls.crt;
						ssl_certificate_key /var/serving-cert/tls.key;
						root                /usr/share/nginx/html;
					}
				}
				`, r.nginxPort())}
	return r.mutateOwned(o)
}

// selector map used by service and deployment.
func (r *Reconciler) selector() map[string]string {
	return map[string]string{constants.LabelK8sName: r.Name, constants.LabelK8sCreatedBy: r.CreatedBy()}
}

func (r *Reconciler) mutateService() error {
	o := &r.service
	o.ObjectMeta.Annotations = map[string]string{constants.AnnotationServingCertSecretName: r.Name}
	// Don't replace Spec entirely it may contain immutable values like ClusterIP if we are updating.
	o.Spec.Selector = r.selector()
	o.Spec.Ports = []corev1.ServicePort{{
		Name:       fmt.Sprintf("%v-tcp", r.nginxPort()),
		Protocol:   "TCP",
		Port:       r.nginxPort(),
		TargetPort: intstr.IntOrString{IntVal: r.nginxPort()},
	}}
	return r.mutateOwned(o)
}

func (r *Reconciler) mutateDeployment() error {
	o := &r.deployment
	o.Spec = appv1.DeploymentSpec{
		Replicas: utils.GetInt32(1),
		Selector: &metav1.LabelSelector{MatchLabels: r.selector()},
		Template: corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{Labels: r.selector()},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name:  r.Name,
						Image: r.Image,
						Ports: []corev1.ContainerPort{
							{
								ContainerPort: 9443,
								Protocol:      "TCP",
							},
						},
						VolumeMounts: []corev1.VolumeMount{
							{
								Name:      "serving-cert",
								ReadOnly:  true,
								MountPath: "/var/serving-cert",
							},
							{
								Name:      "nginx-conf",
								ReadOnly:  true,
								MountPath: "/etc/nginx/nginx.conf",
								SubPath:   "nginx.conf",
							},
						},
					},
				},
				Volumes: []corev1.Volume{
					{
						Name: "serving-cert",
						VolumeSource: corev1.VolumeSource{
							Secret: &corev1.SecretVolumeSource{
								SecretName:  r.Name,
								DefaultMode: r.defaultMode(),
							},
						},
					},
					{
						Name: "nginx-conf",
						VolumeSource: corev1.VolumeSource{
							ConfigMap: &corev1.ConfigMapVolumeSource{
								LocalObjectReference: corev1.LocalObjectReference{Name: r.Name},
								DefaultMode:          r.defaultMode(),
							},
						},
					},
				},
				RestartPolicy: "Always",
				DNSPolicy:     "ClusterFirst",
			},
		},
		Strategy: appv1.DeploymentStrategy{
			Type: "RollingUpdate",
			RollingUpdate: &appv1.RollingUpdateDeployment{
				MaxUnavailable: &intstr.IntOrString{
					Type:   intstr.Type(1),
					StrVal: "25%",
				},
				MaxSurge: &intstr.IntOrString{
					Type:   intstr.Type(1),
					StrVal: "25%",
				},
			},
		},
	}
	return r.mutateOwned(o)
}
