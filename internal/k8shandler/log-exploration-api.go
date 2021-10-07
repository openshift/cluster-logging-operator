package k8shandler

import (
	monitoringv1 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
	route "github.com/openshift/api/route/v1"
	"github.com/openshift/cluster-logging-operator/internal/factory"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const (
	logAPIPort            = 8080
	explorationAPIName    = "exploration-api"
	explorationAPIMetrics = "prom-metrics"
)

func newExplorationAPIRoute(routeName, namespace, serviceName, componentName, loggingComponent string) *route.Route {
	return &route.Route{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Route",
			APIVersion: route.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      routeName,
			Namespace: namespace,
			Labels: map[string]string{
				"component":     componentName,
				"logging-infra": loggingComponent,
			},
		},
		Spec: route.RouteSpec{
			To: route.RouteTargetReference{
				Name: serviceName,
				Kind: "Service",
			},
		},
	}
}

func (clusterRequest *ClusterLoggingRequest) createExplorationAPIService() error {
	desired := factory.NewService(
		explorationAPIName,
		clusterRequest.Cluster.Namespace,
		explorationAPIName,
		[]v1.ServicePort{
			{
				Port:       logAPIPort,
				TargetPort: intstr.FromInt(logAPIPort),
				Name:       explorationAPIMetrics,
			},
		},
	)

	// Override the label from factory.NewService
	desired.Labels = map[string]string{
		"component": explorationAPIName,
	}

	utils.AddOwnerRefToObject(desired, utils.AsOwner(clusterRequest.Cluster))
	err := clusterRequest.Create(desired)
	return err
}

func (clusterRequest *ClusterLoggingRequest) createExplorationAPIServiceMonitor() error {
	cluster := clusterRequest.Cluster
	desired := NewServiceMonitor(explorationAPIName, cluster.Namespace)

	endpoint := monitoringv1.Endpoint{
		Port:   explorationAPIMetrics,
		Path:   "/metrics",
		Scheme: "http",
	}

	// match label used to select the log-exploration-api service object
	labelSelector := metav1.LabelSelector{
		MatchLabels: map[string]string{
			"component": explorationAPIName,
		},
	}

	desired.Spec = monitoringv1.ServiceMonitorSpec{
		JobLabel:  "monitor-logexplorationapi",
		Endpoints: []monitoringv1.Endpoint{endpoint},
		Selector:  labelSelector,
		NamespaceSelector: monitoringv1.NamespaceSelector{
			MatchNames: []string{cluster.Namespace},
		},
	}

	utils.AddOwnerRefToObject(desired, utils.AsOwner(cluster))

	err := clusterRequest.Create(desired)
	if err != nil {
		return err
	}
	return nil
}

func (clusterRequest *ClusterLoggingRequest) createExplorationAPIRoute() error {
	apiRoute := newExplorationAPIRoute(explorationAPIName, clusterRequest.Cluster.Namespace, explorationAPIName, explorationAPIName, explorationAPIName)
	utils.AddOwnerRefToObject(apiRoute, utils.AsOwner(clusterRequest.Cluster))
	err := clusterRequest.Create(apiRoute)
	return err
}

func (clusterRequest *ClusterLoggingRequest) createExplorationAPIDeployment() error {
	logApiPodSpec := newLogExplorationApiPodSpec()
	logApiDeployment := NewDeployment(explorationAPIName, clusterRequest.Cluster.Namespace, explorationAPIName, explorationAPIName, logApiPodSpec)
	utils.AddOwnerRefToObject(logApiDeployment, utils.AsOwner(clusterRequest.Cluster))
	err := clusterRequest.Create(logApiDeployment)
	if err != nil {
		return err
	}
	return nil
}

func newLogExplorationApiPodSpec() v1.PodSpec {
	resources := &v1.ResourceRequirements{
		Limits: v1.ResourceList{
			v1.ResourceMemory: defaultLoggingApiMemory,
			v1.ResourceCPU:    defaultLoggingApiCpuRequest,
		},
		Requests: v1.ResourceList{
			v1.ResourceMemory: defaultLoggingApiMemory,
			v1.ResourceCPU:    defaultLoggingApiCpuRequest,
		},
	}

	logExplorationApiContainer := NewContainer("logexplorationapi", "logApi", v1.PullIfNotPresent, *resources)

	logExplorationApiContainer.Ports = []v1.ContainerPort{
		{
			ContainerPort: logAPIPort,
		},
	}
	logExplorationApiContainer.Env = []v1.EnvVar{
		{Name: "ES_ADDR", Value: "https://elasticsearch.openshift-logging:9200"},
		{Name: "ES_CERT", Value: "/etc/openshift/elasticsearch/secret/tls.crt"},
		{Name: "ES_KEY", Value: "/etc/openshift/elasticsearch/secret/tls.key"},
		{Name: "ES_TLS", Value: "true"},
		{Name: "POD_IP", ValueFrom: &v1.EnvVarSource{FieldRef: &v1.ObjectFieldSelector{APIVersion: "v1", FieldPath: "status.podIP"}}},
	}

	logExplorationApiContainer.VolumeMounts = []v1.VolumeMount{
		{Name: "certificates", MountPath: "/etc/openshift/elasticsearch/secret"},
	}
	httpGetStructLivenessProbe := &v1.HTTPGetAction{
		Path: "/ready",
		Port: intstr.FromInt(logAPIPort),
	}
	httpGetStructReadinessProbe := &v1.HTTPGetAction{
		Path: "/health",
		Port: intstr.FromInt(logAPIPort),
	}
	handlerStructLivenessProbe := v1.Handler{
		HTTPGet: httpGetStructLivenessProbe,
	}

	handlerStructReadinessProbe := v1.Handler{
		HTTPGet: httpGetStructReadinessProbe,
	}
	logExplorationApiContainer.ReadinessProbe = &v1.Probe{
		Handler:             handlerStructReadinessProbe,
		InitialDelaySeconds: 3,
		PeriodSeconds:       3,
	}
	logExplorationApiContainer.LivenessProbe = &v1.Probe{
		Handler:             handlerStructLivenessProbe,
		InitialDelaySeconds: 10,
		PeriodSeconds:       3,
		FailureThreshold:    30,
	}

	logExplorationApiPodSpec := newLogApiPodSpec([]v1.Container{logExplorationApiContainer},
		[]v1.Volume{
			{Name: "certificates",
				VolumeSource: v1.VolumeSource{Secret: &v1.SecretVolumeSource{SecretName: "fluentd", DefaultMode: utils.GetInt32(420)}}},
		},
	)
	return logExplorationApiPodSpec

}

func newLogApiPodSpec(containers []v1.Container, volumes []v1.Volume) v1.PodSpec {
	return v1.PodSpec{
		Containers: containers,
		Volumes:    volumes,
	}
}

func (clusterRequest *ClusterLoggingRequest) CreateOrDeleteLogExplorationApi() error {
	if _, ok := clusterRequest.Cluster.Annotations["api-enabled"]; ok {
		if err := clusterRequest.createExplorationAPIDeployment(); err != nil {
			return err
		}
		if err := clusterRequest.createExplorationAPIService(); err != nil {
			return err
		}
		if err := clusterRequest.createExplorationAPIRoute(); err != nil {
			return err
		}
		if err := clusterRequest.createExplorationAPIServiceMonitor(); err != nil {
			return err
		}
	} else {
		if err := clusterRequest.RemoveDeployment(explorationAPIName); err != nil {
			return err
		}
		if err := clusterRequest.RemoveService(explorationAPIName); err != nil {
			return err
		}
		if err := clusterRequest.RemoveRoute(explorationAPIName); err != nil {
			return err
		}
		if err := clusterRequest.RemoveServiceMonitor(explorationAPIName); err != nil {
			return err
		}
	}
	return nil
}
