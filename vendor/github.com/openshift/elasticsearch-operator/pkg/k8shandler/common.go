package k8shandler

import (
	"fmt"
	"sort"
	"strconv"

	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/openshift/elasticsearch-operator/pkg/utils"
	"sigs.k8s.io/controller-runtime/pkg/client"

	api "github.com/openshift/elasticsearch-operator/pkg/apis/logging/v1"
)

// addOwnerRefToObject appends the desired OwnerReference to the object
// deprecated in favor of Elasticsearch#AddOwnerRefTo
func addOwnerRefToObject(o metav1.Object, r metav1.OwnerReference) {
	if (metav1.OwnerReference{}) != r {
		o.SetOwnerReferences(append(o.GetOwnerReferences(), r))
	}
}

func getImage(commonImage string) string {
	image := commonImage
	if image == "" {
		image = utils.LookupEnvWithDefault("ELASTICSEARCH_IMAGE", elasticsearchDefaultImage)
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
		Kind:       "Elasticsearch",
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

	cpuLimit, err := resource.ParseQuantity("100m")
	if err != nil {
		return v1.Container{}, err
	}

	memoryLimit, err := resource.ParseQuantity("64Mi")
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
			"--http-address=:4180",
			"--https-address=:60000",
			"--provider=openshift",
			"--upstream=https://localhost:9200",
			"--tls-cert=/etc/proxy/secrets/tls.crt",
			"--tls-key=/etc/proxy/secrets/tls.key",
			"--upstream-ca=/etc/proxy/elasticsearch/admin-ca",
			"--openshift-service-account=elasticsearch",
			`-openshift-sar={"resource": "namespaces", "verb": "get"}`,
			`-openshift-delegate-urls={"/": {"resource": "namespaces", "verb": "get"}}`,
			"--pass-user-bearer-token",
			fmt.Sprintf("--cookie-secret=%s", proxyCookieSecret),
		},
		Resources: v1.ResourceRequirements{
			Limits: v1.ResourceList{
				"memory": memoryLimit,
			},
			Requests: v1.ResourceList{
				"cpu":    cpuLimit,
				"memory": memoryLimit,
			},
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

// TODO: add isChanged check for labels and label selector
func newLabels(clusterName, nodeName string, roleMap map[api.ElasticsearchNodeRole]bool) map[string]string {

	return map[string]string{
		"es-node-client":                   strconv.FormatBool(roleMap[api.ElasticsearchRoleClient]),
		"es-node-data":                     strconv.FormatBool(roleMap[api.ElasticsearchRoleData]),
		"es-node-master":                   strconv.FormatBool(roleMap[api.ElasticsearchRoleMaster]),
		"cluster-name":                     clusterName,
		"component":                        "elasticsearch",
		"tuned.openshift.io/elasticsearch": "true",
		"node-name":                        nodeName,
	}
}

func newLabelSelector(clusterName, nodeName string, roleMap map[api.ElasticsearchNodeRole]bool) map[string]string {

	return map[string]string{
		"es-node-client": strconv.FormatBool(roleMap[api.ElasticsearchRoleClient]),
		"es-node-data":   strconv.FormatBool(roleMap[api.ElasticsearchRoleData]),
		"es-node-master": strconv.FormatBool(roleMap[api.ElasticsearchRoleMaster]),
		"cluster-name":   clusterName,
		"node-name":      nodeName,
	}
}

func newPodTemplateSpec(nodeName, clusterName, namespace string, node api.ElasticsearchNode, commonSpec api.ElasticsearchNodeSpec, labels map[string]string, roleMap map[api.ElasticsearchNodeRole]bool, client client.Client) v1.PodTemplateSpec {

	resourceRequirements := newResourceRequirements(node.Resources, commonSpec.Resources)
	proxyImage := utils.LookupEnvWithDefault("PROXY_IMAGE", "quay.io/openshift/origin-oauth-proxy:latest")
	proxyContainer, _ := newProxyContainer(proxyImage, clusterName)

	selectors := mergeSelectors(node.NodeSelector, commonSpec.NodeSelector)
	// We want to make sure the pod ends up allocated on linux node. Thus we make sure the
	// linux node selectors is always present. See LOG-411
	selectors = utils.EnsureLinuxNodeSelector(selectors)

	tolerations := appendTolerations(node.Tolerations, commonSpec.Tolerations)
	tolerations = appendTolerations(tolerations, []v1.Toleration{
		v1.Toleration{
			Key:      "node.kubernetes.io/disk-pressure",
			Operator: v1.TolerationOpExists,
			Effect:   v1.TaintEffectNoSchedule,
		},
	})

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
			NodeSelector:       selectors,
			ServiceAccountName: clusterName,
			Volumes:            newVolumes(clusterName, nodeName, namespace, node, client),
			Tolerations:        tolerations,
		},
	}
}

func newResourceRequirements(nodeResRequirements, commonResRequirements v1.ResourceRequirements) v1.ResourceRequirements {
	// if only one resource (cpu or memory) is specified as a limit/request use it for the other value as well instead of
	//  using the defaults.

	var requestMem *resource.Quantity
	var limitMem *resource.Quantity
	var requestCPU *resource.Quantity
	var limitCPU *resource.Quantity

	// first check if either limit or resource is left off
	// Mem
	nodeLimitMem := nodeResRequirements.Limits.Memory()
	nodeRequestMem := nodeResRequirements.Requests.Memory()
	commonLimitMem := commonResRequirements.Limits.Memory()
	commonRequestMem := commonResRequirements.Requests.Memory()

	if commonRequestMem.IsZero() && commonLimitMem.IsZero() {
		// no common memory settings
		if nodeRequestMem.IsZero() && nodeLimitMem.IsZero() {
			// no node settings, use defaults
			lMem, _ := resource.ParseQuantity(defaultMemoryLimit)
			limitMem = &lMem

			rMem, _ := resource.ParseQuantity(defaultMemoryRequest)
			requestMem = &rMem
		} else {
			// either one is not zero or both aren't zero but common is empty
			if nodeRequestMem.IsZero() {
				// request is zero use limit for both
				requestMem = nodeLimitMem
				limitMem = nodeLimitMem
			} else {
				if nodeLimitMem.IsZero() {
					// limit is zero use request for both
					requestMem = nodeRequestMem
					limitMem = nodeRequestMem
				} else {
					// both aren't zero
					requestMem = nodeRequestMem
					limitMem = nodeLimitMem
				}
			}
		}
	} else {
		// either one is not zero or both aren't zero (common)

		//check node for override
		if nodeRequestMem.IsZero() {
			// no node request mem, check that common has it
			if commonRequestMem.IsZero() {
				requestMem = commonLimitMem
			} else {
				requestMem = commonRequestMem
			}
		} else {
			requestMem = nodeRequestMem
		}

		if nodeLimitMem.IsZero() {
			// no node request mem, check that common has it
			if commonLimitMem.IsZero() {
				limitMem = commonRequestMem
			} else {
				limitMem = commonLimitMem
			}
		} else {
			limitMem = nodeLimitMem
		}
	}

	// CPU
	nodeLimitCPU := nodeResRequirements.Limits.Cpu()
	nodeRequestCPU := nodeResRequirements.Requests.Cpu()
	commonLimitCPU := commonResRequirements.Limits.Cpu()
	commonRequestCPU := commonResRequirements.Requests.Cpu()

	if commonRequestCPU.IsZero() && commonLimitCPU.IsZero() {
		// no common memory settings
		if nodeRequestCPU.IsZero() && nodeLimitCPU.IsZero() {
			// no node settings, use defaults
			rCPU, _ := resource.ParseQuantity(defaultCPURequest)
			requestCPU = &rCPU
		} else {
			// either one is not zero or both aren't zero but common is empty
			if nodeRequestCPU.IsZero() {
				// request is zero use limit for both
				requestCPU = nodeLimitCPU
				limitCPU = nodeLimitCPU
			} else {
				if nodeLimitCPU.IsZero() {
					// limit is zero use request for both
					requestCPU = nodeRequestCPU
				} else {
					// both aren't zero
					requestCPU = nodeRequestCPU
					limitCPU = nodeLimitCPU
				}
			}
		}
	} else {
		// either one is not zero or both aren't zero (common)

		//check node for override
		if nodeRequestCPU.IsZero() {
			// no node request mem, check that common has it
			if commonRequestCPU.IsZero() {
				requestCPU = commonLimitCPU
			} else {
				requestCPU = commonRequestCPU
			}
		} else {
			requestCPU = nodeRequestCPU
		}

		if nodeLimitCPU.IsZero() {
			// no node request mem, check that common has it
			if !commonLimitCPU.IsZero() {
				limitCPU = commonLimitCPU
			}
		} else {
			limitCPU = nodeLimitCPU
		}
	}

	if limitCPU == nil {
		return v1.ResourceRequirements{
			Limits: v1.ResourceList{
				"memory": *limitMem,
			},
			Requests: v1.ResourceList{
				"cpu":    *requestCPU,
				"memory": *requestMem,
			},
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

func newVolumes(clusterName, nodeName, namespace string, node api.ElasticsearchNode, client client.Client) []v1.Volume {
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
			VolumeSource: newVolumeSource(clusterName, nodeName, namespace, node, client),
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

func newVolumeSource(clusterName, nodeName, namespace string, node api.ElasticsearchNode, client client.Client) v1.VolumeSource {

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

		err := createOrUpdatePersistentVolumeClaim(volSpec, claimName, namespace, client)
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
