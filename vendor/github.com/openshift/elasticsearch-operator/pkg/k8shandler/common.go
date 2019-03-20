package k8shandler

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"sort"
	"strconv"

	"github.com/openshift/elasticsearch-operator/pkg/utils"

	api "github.com/openshift/elasticsearch-operator/pkg/apis/elasticsearch/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// addOwnerRefToObject appends the desired OwnerReference to the object
func addOwnerRefToObject(o metav1.Object, r metav1.OwnerReference) {
	if (metav1.OwnerReference{}) != r {
		o.SetOwnerReferences(append(o.GetOwnerReferences(), r))
	}
}

func getImage(commonImage string) string {
	image := commonImage
	if image == "" {
		image = elasticsearchDefaultImage
	}
	return image
}

func getNodeRoleMap(node api.ElasticsearchNode) map[api.ElasticsearchNodeRole]bool {
	isClient := false
	isData := false
	isMaster := false

	for _, role := range node.Roles {
		if role == api.ElasticsearchRoleClient {
			isClient = true
		}

		if role == api.ElasticsearchRoleData {
			isData = true
		}

		if role == api.ElasticsearchRoleMaster {
			isMaster = true
		}
	}
	return map[api.ElasticsearchNodeRole]bool{
		api.ElasticsearchRoleClient: isClient,
		api.ElasticsearchRoleData:   isData,
		api.ElasticsearchRoleMaster: isMaster,
	}
}

// getOwnerRef returns an owner reference set as the vault cluster CR
func getOwnerRef(v *api.Elasticsearch) metav1.OwnerReference {
	trueVar := true
	return metav1.OwnerReference{
		APIVersion: api.SchemeGroupVersion.String(),
		Kind:       v.Kind,
		Name:       v.Name,
		UID:        v.UID,
		Controller: &trueVar,
	}
}

func isOnlyClientNode(node api.ElasticsearchNode) bool {
	for _, role := range node.Roles {
		if role != api.ElasticsearchRoleClient {
			return false
		}
	}

	return true
}

func isClientNode(node api.ElasticsearchNode) bool {
	for _, role := range node.Roles {
		if role == api.ElasticsearchRoleClient {
			return true
		}
	}

	return false
}

func isOnlyMasterNode(node api.ElasticsearchNode) bool {
	for _, role := range node.Roles {
		if role != api.ElasticsearchRoleMaster {
			return false
		}
	}

	return true
}

func isMasterNode(node api.ElasticsearchNode) bool {
	for _, role := range node.Roles {
		if role == api.ElasticsearchRoleMaster {
			return true
		}
	}

	return false
}

func isDataNode(node api.ElasticsearchNode) bool {
	for _, role := range node.Roles {
		if role == api.ElasticsearchRoleData {
			return true
		}
	}

	return false
}

func newAffinity(roleMap map[api.ElasticsearchNodeRole]bool) *v1.Affinity {

	labelSelectorReqs := []metav1.LabelSelectorRequirement{}
	if roleMap[api.ElasticsearchRoleClient] {
		labelSelectorReqs = append(labelSelectorReqs, metav1.LabelSelectorRequirement{
			Key:      "es-node-client",
			Operator: metav1.LabelSelectorOpIn,
			Values:   []string{"true"},
		})
	}
	if roleMap[api.ElasticsearchRoleData] {
		labelSelectorReqs = append(labelSelectorReqs, metav1.LabelSelectorRequirement{
			Key:      "es-node-data",
			Operator: metav1.LabelSelectorOpIn,
			Values:   []string{"true"},
		})
	}
	if roleMap[api.ElasticsearchRoleMaster] {
		labelSelectorReqs = append(labelSelectorReqs, metav1.LabelSelectorRequirement{
			Key:      "es-node-master",
			Operator: metav1.LabelSelectorOpIn,
			Values:   []string{"true"},
		})
	}

	return &v1.Affinity{
		PodAntiAffinity: &v1.PodAntiAffinity{
			PreferredDuringSchedulingIgnoredDuringExecution: []v1.WeightedPodAffinityTerm{
				{
					Weight: 100,
					PodAffinityTerm: v1.PodAffinityTerm{
						LabelSelector: &metav1.LabelSelector{
							MatchExpressions: labelSelectorReqs,
						},
						TopologyKey: "kubernetes.io/hostname",
					},
				},
			},
		},
	}
}

func newElasticsearchContainer(imageName string, envVars []v1.EnvVar, resourceRequirements v1.ResourceRequirements) v1.Container {

	return v1.Container{
		Name:            "elasticsearch",
		Image:           imageName,
		ImagePullPolicy: "IfNotPresent",
		Env:             envVars,
		Ports: []v1.ContainerPort{
			v1.ContainerPort{
				Name:          "cluster",
				ContainerPort: 9300,
				Protocol:      v1.ProtocolTCP,
			},
			v1.ContainerPort{
				Name:          "restapi",
				ContainerPort: 9200,
				Protocol:      v1.ProtocolTCP,
			},
		},
		ReadinessProbe: &v1.Probe{
			TimeoutSeconds:      30,
			InitialDelaySeconds: 10,
			PeriodSeconds:       5,
			Handler: v1.Handler{
				Exec: &v1.ExecAction{
					Command: []string{
						"/usr/share/elasticsearch/probe/readiness.sh",
					},
				},
			},
		},
		VolumeMounts: []v1.VolumeMount{
			v1.VolumeMount{
				Name:      "elasticsearch-storage",
				MountPath: "/elasticsearch/persistent",
			},
			v1.VolumeMount{
				Name:      "elasticsearch-config",
				MountPath: elasticsearchConfigPath,
			},
			v1.VolumeMount{
				Name:      "certificates",
				MountPath: elasticsearchCertsPath,
			},
		},
		Resources: resourceRequirements,
	}
}

func newProxyContainer(imageName, clusterName string) (v1.Container, error) {
	proxyCookieSecret, err := utils.RandStringBase64(16)
	if err != nil {
		return v1.Container{}, err
	}
	container := v1.Container{
		Name:            "proxy",
		Image:           imageName,
		ImagePullPolicy: "IfNotPresent",
		Ports: []v1.ContainerPort{
			v1.ContainerPort{
				Name:          "metrics",
				ContainerPort: 60000,
				Protocol:      v1.ProtocolTCP,
			},
		},
		VolumeMounts: []v1.VolumeMount{
			v1.VolumeMount{
				Name:      fmt.Sprintf("%s-%s", clusterName, "metrics"),
				MountPath: "/etc/proxy/secrets",
			},
			v1.VolumeMount{
				Name:      "certificates",
				MountPath: "/etc/proxy/elasticsearch",
			},
		},
		Args: []string{
			"--https-address=:60000",
			"--provider=openshift",
			"--upstream=https://127.0.0.1:9200",
			"--tls-cert=/etc/proxy/secrets/tls.crt",
			"--tls-key=/etc/proxy/secrets/tls.key",
			"--upstream-ca=/etc/proxy/elasticsearch/admin-ca",
			"--openshift-service-account=elasticsearch",
			`-openshift-sar={"resource": "namespaces", "verb": "get"}`,
			`-openshift-delegate-urls={"/": {"resource": "namespaces", "verb": "get"}}`,
			"--pass-user-bearer-token",
			fmt.Sprintf("--cookie-secret=%s", proxyCookieSecret),
		},
	}
	return container, nil
}

func newEnvVars(nodeName, clusterName, instanceRam string, roleMap map[api.ElasticsearchNodeRole]bool) []v1.EnvVar {

	return []v1.EnvVar{
		v1.EnvVar{
			Name:  "DC_NAME",
			Value: nodeName,
		},
		v1.EnvVar{
			Name: "NAMESPACE",
			ValueFrom: &v1.EnvVarSource{
				FieldRef: &v1.ObjectFieldSelector{
					FieldPath: "metadata.namespace",
				},
			},
		},
		v1.EnvVar{
			Name:  "KUBERNETES_TRUST_CERT",
			Value: "true",
		},
		v1.EnvVar{
			Name:  "SERVICE_DNS",
			Value: fmt.Sprintf("%s-cluster", clusterName),
		},
		v1.EnvVar{
			Name:  "CLUSTER_NAME",
			Value: clusterName,
		},
		v1.EnvVar{
			Name:  "INSTANCE_RAM",
			Value: instanceRam,
		},
		v1.EnvVar{
			Name:  "HEAP_DUMP_LOCATION",
			Value: heapDumpLocation,
		},
		v1.EnvVar{
			Name:  "RECOVER_AFTER_TIME",
			Value: "5m",
		},
		v1.EnvVar{
			Name:  "READINESS_PROBE_TIMEOUT",
			Value: "30",
		},
		v1.EnvVar{
			Name:  "POD_LABEL",
			Value: fmt.Sprintf("cluster=%s", clusterName),
		},
		v1.EnvVar{
			Name:  "IS_MASTER",
			Value: strconv.FormatBool(roleMap[api.ElasticsearchRoleMaster]),
		},
		v1.EnvVar{
			Name:  "HAS_DATA",
			Value: strconv.FormatBool(roleMap[api.ElasticsearchRoleData]),
		},
	}
}

func newLabels(clusterName string, roleMap map[api.ElasticsearchNodeRole]bool) map[string]string {

	return map[string]string{
		"es-node-client":                   strconv.FormatBool(roleMap[api.ElasticsearchRoleClient]),
		"es-node-data":                     strconv.FormatBool(roleMap[api.ElasticsearchRoleData]),
		"es-node-master":                   strconv.FormatBool(roleMap[api.ElasticsearchRoleMaster]),
		"cluster-name":                     clusterName,
		"component":                        clusterName,
		"tuned.openshift.io/elasticsearch": "true",
	}
}

func newLabelSelector(clusterName string, roleMap map[api.ElasticsearchNodeRole]bool) map[string]string {

	return map[string]string{
		"es-node-client": strconv.FormatBool(roleMap[api.ElasticsearchRoleClient]),
		"es-node-data":   strconv.FormatBool(roleMap[api.ElasticsearchRoleData]),
		"es-node-master": strconv.FormatBool(roleMap[api.ElasticsearchRoleMaster]),
		"cluster-name":   clusterName,
	}
}

func newPodTemplateSpec(nodeName, clusterName, namespace string, node api.ElasticsearchNode, commonSpec api.ElasticsearchNodeSpec, labels map[string]string, roleMap map[api.ElasticsearchNodeRole]bool) v1.PodTemplateSpec {

	resourceRequirements := newResourceRequirements(node.Resources, commonSpec.Resources)
	proxyImage := utils.LookupEnvWithDefault("PROXY_IMAGE", "quay.io/openshift/origin-oauth-proxy:latest")
	proxyContainer, _ := newProxyContainer(proxyImage, clusterName)

	return v1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Labels: labels,
		},
		Spec: v1.PodSpec{
			Affinity: newAffinity(roleMap),
			Containers: []v1.Container{
				newElasticsearchContainer(
					getImage(commonSpec.Image),
					newEnvVars(nodeName, clusterName, resourceRequirements.Limits.Memory().String(), roleMap),
					resourceRequirements,
				),
				proxyContainer,
			},
			NodeSelector:       node.NodeSelector,
			ServiceAccountName: clusterName,
			Volumes:            newVolumes(clusterName, nodeName, namespace, node),
		},
	}
}

func newResourceRequirements(nodeResRequirements, commonResRequirements v1.ResourceRequirements) v1.ResourceRequirements {
	limitCPU := nodeResRequirements.Limits.Cpu()
	if limitCPU.IsZero() {
		if commonResRequirements.Limits.Cpu().IsZero() {
			CPU, _ := resource.ParseQuantity(defaultCPULimit)
			limitCPU = &CPU
		} else {
			limitCPU = commonResRequirements.Limits.Cpu()
		}
	}
	limitMem := nodeResRequirements.Limits.Memory()
	if limitMem.IsZero() {
		if commonResRequirements.Limits.Memory().IsZero() {
			Mem, _ := resource.ParseQuantity(defaultMemoryLimit)
			limitMem = &Mem
		} else {
			limitMem = commonResRequirements.Limits.Memory()
		}

	}
	requestCPU := nodeResRequirements.Requests.Cpu()
	if requestCPU.IsZero() {
		if commonResRequirements.Requests.Cpu().IsZero() {
			CPU, _ := resource.ParseQuantity(defaultCPURequest)
			requestCPU = &CPU
		} else {
			requestCPU = commonResRequirements.Requests.Cpu()
		}
	}
	requestMem := nodeResRequirements.Requests.Memory()
	if requestMem.IsZero() {
		if commonResRequirements.Requests.Memory().IsZero() {
			Mem, _ := resource.ParseQuantity(defaultMemoryRequest)
			requestMem = &Mem
		} else {
			requestMem = commonResRequirements.Requests.Memory()
		}
	}

	return v1.ResourceRequirements{
		Limits: v1.ResourceList{
			"cpu":    *limitCPU,
			"memory": *limitMem,
		},
		Requests: v1.ResourceList{
			"cpu":    *requestCPU,
			"memory": *requestMem,
		},
	}
}

func newVolumes(clusterName, nodeName, namespace string, node api.ElasticsearchNode) []v1.Volume {
	return []v1.Volume{
		v1.Volume{
			Name: "elasticsearch-config",
			VolumeSource: v1.VolumeSource{
				ConfigMap: &v1.ConfigMapVolumeSource{
					LocalObjectReference: v1.LocalObjectReference{
						Name: clusterName,
					},
				},
			},
		},
		v1.Volume{
			Name:         "elasticsearch-storage",
			VolumeSource: newVolumeSource(clusterName, nodeName, namespace, node),
		},
		v1.Volume{
			Name: "certificates",
			VolumeSource: v1.VolumeSource{
				Secret: &v1.SecretVolumeSource{
					SecretName: clusterName,
				},
			},
		},
		v1.Volume{
			Name: fmt.Sprintf("%s-%s", clusterName, "metrics"),
			VolumeSource: v1.VolumeSource{
				Secret: &v1.SecretVolumeSource{
					SecretName: fmt.Sprintf("%s-%s", clusterName, "metrics"),
				},
			},
		},
	}
}

func newVolumeSource(clusterName, nodeName, namespace string, node api.ElasticsearchNode) v1.VolumeSource {

	specVol := node.Storage
	volSource := v1.VolumeSource{}

	switch {
	case specVol.StorageClassName != nil && specVol.Size != nil:
		claimName := fmt.Sprintf("%s-%s", clusterName, nodeName)
		volSource.PersistentVolumeClaim = &v1.PersistentVolumeClaimVolumeSource{
			ClaimName: claimName,
		}

		volSpec := v1.PersistentVolumeClaimSpec{
			AccessModes: []v1.PersistentVolumeAccessMode{
				v1.ReadWriteOnce,
			},
			Resources: v1.ResourceRequirements{
				Requests: v1.ResourceList{
					v1.ResourceStorage: *specVol.Size,
				},
			},
			StorageClassName: specVol.StorageClassName,
		}

		err := createOrUpdatePersistentVolumeClaim(volSpec, claimName, namespace)
		if err != nil {
			logrus.Errorf("Unable to create PersistentVolumeClaim: %v", err)
		}

	case specVol.Size != nil:
		volSource.EmptyDir = &v1.EmptyDirVolumeSource{
			SizeLimit: specVol.Size,
		}

	default:
		volSource.EmptyDir = &v1.EmptyDirVolumeSource{}
	}

	return volSource
}

func sortDataHashKeys(dataHash map[string][32]byte) []string {
	keys := []string{}
	for key, _ := range dataHash {
		keys = append(keys, key)
	}

	sort.Strings(keys)

	return keys
}
