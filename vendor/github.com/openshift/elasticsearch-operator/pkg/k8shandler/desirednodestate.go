package k8shandler

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	v1alpha1 "github.com/openshift/elasticsearch-operator/pkg/apis/elasticsearch/v1alpha1"
	"github.com/operator-framework/operator-sdk/pkg/sdk"
	"github.com/sirupsen/logrus"
	apps "k8s.io/api/apps/v1"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/util/retry"
)

const (
	elasticsearchCertsPath    = "/etc/openshift/elasticsearch/secret"
	elasticsearchConfigPath   = "/usr/share/java/elasticsearch/config"
	elasticsearchDefaultImage = "docker.io/openshift/origin-logging-elasticsearch5"
	heapDumpLocation          = "/elasticsearch/persistent/heapdump.hprof"
)

type nodeState struct {
	Desired desiredNodeState
	Actual  actualNodeState
}

type desiredNodeState struct {
	ClusterName        string
	Namespace          string
	DeployName         string
	Roles              []v1alpha1.ElasticsearchNodeRole
	ESNodeSpec         v1alpha1.ElasticsearchNode
	Image              string
	SecretName         string
	NodeNum            int32
	ReplicaNum         int32
	ServiceAccountName string
	ConfigMapName      string
	Labels             map[string]string
	MasterNum          int32
	DataNum            int32
	EnvVars            []v1.EnvVar
	Paused             bool
}

type actualNodeState struct {
	StatefulSet *apps.StatefulSet
	Deployment  *apps.Deployment
	ReplicaSet  *apps.ReplicaSet
	Pod         *v1.Pod
}

func constructNodeSpec(dpl *v1alpha1.Elasticsearch, esNode v1alpha1.ElasticsearchNode, configMapName, serviceAccountName string, nodeNum int32, replicaNum int32, masterNum, dataNum int32) (desiredNodeState, error) {
	nodeCfg := desiredNodeState{
		ClusterName:        dpl.Name,
		Namespace:          dpl.Namespace,
		Roles:              esNode.Roles,
		ESNodeSpec:         esNode,
		SecretName:         v1alpha1.SecretName,
		NodeNum:            nodeNum,
		ReplicaNum:         replicaNum,
		ServiceAccountName: serviceAccountName,
		ConfigMapName:      configMapName,
		Labels:             dpl.Labels,
		MasterNum:          masterNum,
		DataNum:            dataNum,
		Paused:             true,
	}
	deployName, err := constructDeployName(dpl.Name, esNode.Roles, nodeNum, replicaNum)
	if err != nil {
		return nodeCfg, err
	}
	nodeCfg.DeployName = deployName
	nodeCfg.EnvVars = nodeCfg.getEnvVars()

	nodeCfg.ESNodeSpec.Resources = getResourceRequirements(dpl.Spec.Spec.Resources, esNode.Resources)
	nodeCfg.Image = getImage(dpl.Spec.Spec.Image)
	return nodeCfg, nil
}

func constructDeployName(name string, roles []v1alpha1.ElasticsearchNodeRole, nodeNum int32, replicaNum int32) (string, error) {
	if len(roles) == 0 {
		return "", fmt.Errorf("No node roles specified for a node in cluster %s", name)
	}
	if len(roles) == 1 && roles[0] == "master" {
		return fmt.Sprintf("%s-master-%d", name, nodeNum), nil
	}
	var nodeType []string
	for _, role := range roles {
		if role != "client" && role != "data" && role != "master" {
			return "", fmt.Errorf("Unknown node's role: %s", role)
		}
		nodeType = append(nodeType, string(role))
	}

	sort.Strings(nodeType)

	return fmt.Sprintf("%s-%s-%d-%d", name, strings.Join(nodeType, ""), nodeNum, replicaNum), nil
}

// getReplicas returns the desired number of replicas in the deployment/statefulset
// if this is a data deployment, we always want to create separate deployment per replica
// so we'll return 1. if this is not a data node, we can simply scale existing replica.
func (cfg *desiredNodeState) getReplicas() int32 {
	if cfg.isNodeData() {
		return 1
	}
	return cfg.ESNodeSpec.NodeCount
}

func (cfg *desiredNodeState) isNodeMaster() bool {
	for _, role := range cfg.Roles {
		if role == "master" {
			return true
		}
	}
	return false
}

func (cfg *desiredNodeState) isNodePureMaster() bool {
	if len(cfg.Roles) > 1 {
		return false
	}
	for _, role := range cfg.Roles {
		if role != v1alpha1.ElasticsearchRoleMaster {
			return false
		}
	}
	return true
}

func (cfg *desiredNodeState) isNodeData() bool {
	for _, role := range cfg.Roles {
		if role == "data" {
			return true
		}
	}
	return false
}

func (cfg *desiredNodeState) isNodeClient() bool {
	for _, role := range cfg.Roles {
		if role == "client" {
			return true
		}
	}
	return false
}

func (cfg *desiredNodeState) getLabels() map[string]string {
	labels := cfg.Labels
	if labels == nil {
		labels = make(map[string]string)
	}
	labels["es-node-client"] = strconv.FormatBool(cfg.isNodeClient())
	labels["es-node-data"] = strconv.FormatBool(cfg.isNodeData())
	labels["es-node-master"] = strconv.FormatBool(cfg.isNodeMaster())
	labels["cluster-name"] = cfg.ClusterName
	labels["component"] = cfg.ClusterName
	labels["tuned.openshift.io/elasticsearch"] = "true"
	return labels
}

func (cfg *desiredNodeState) getLabelSelector() map[string]string {
	return map[string]string{
		"es-node-client": strconv.FormatBool(cfg.isNodeClient()),
		"es-node-data":   strconv.FormatBool(cfg.isNodeData()),
		"es-node-master": strconv.FormatBool(cfg.isNodeMaster()),
		"cluster-name":   cfg.ClusterName,
	}
}

func (cfg *desiredNodeState) getNode() NodeTypeInterface {
	if cfg.isNodeData() {
		return NewDeploymentNode(cfg.DeployName, cfg.Namespace)
	}
	return NewStatefulSetNode(cfg.DeployName, cfg.Namespace)
}

func (cfg *desiredNodeState) CreateNode(owner metav1.OwnerReference) error {
	node := cfg.getNode()
	err := node.query()
	if err != nil {
		// Node's resource doesn't exist, we can construct one
		cfg.Paused = false
		logrus.Infof("Constructing new resource %v", cfg.DeployName)
		dep, err := node.constructNodeResource(cfg, owner)
		if err != nil {
			return fmt.Errorf("Could not construct node resource: %v", err)
		}
		err = sdk.Create(dep)
		if err != nil && !errors.IsAlreadyExists(err) {
			return fmt.Errorf("Could not create node resource: %v", err)
		}
	}
	return nil
}

func (cfg *desiredNodeState) PauseNode(owner metav1.OwnerReference) error {
	node := cfg.getNode()
	err := node.query()

	// TODO: what is allowed to be changed in the StatefulSet ?
	// Validate Elasticsearch cluster parameters
	needsPause, err := node.needsPause(cfg)
	if err != nil {
		return fmt.Errorf("Failed to see if the node resource is different from what's needed: %v", err)
	}

	if needsPause {
		// set it back to being paused
		cfg.Paused = true
		nretries := -1
		retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
			nretries++
			if getErr := node.query(); getErr != nil {
				logrus.Debugf("Could not get Elasticsearch node resource %v: %v", cfg.DeployName, getErr)
				return getErr
			}
			dep, updateErr := node.constructNodeResource(cfg, owner)
			if updateErr != nil {
				return fmt.Errorf("Could not construct node resource %v for update: %v", cfg.DeployName, updateErr)
			}
			if nretries == 0 {
				logrus.Infof("Updating node resource to be paused again %v", cfg.DeployName)
			}
			if updateErr = sdk.Update(dep); updateErr != nil {
				logrus.Debugf("Failed to update node resource %v: %v", cfg.DeployName, updateErr)
			}
			return updateErr
		})
		if retryErr != nil {
			return fmt.Errorf("Error: could not update status for Elasticsearch %v after %v retries: %v", cfg.DeployName, nretries, retryErr)
		}
		logrus.Debugf("Updated Elasticsearch %v after %v retries", cfg.DeployName, nretries)
	}
	return nil
}

func (cfg *desiredNodeState) UpdateNode(owner metav1.OwnerReference) error {
	node := cfg.getNode()
	if err := node.query(); err != nil {
		return err
	}

	// TODO: what is allowed to be changed in the StatefulSet ?
	// Validate Elasticsearch cluster parameters
	diff := node.isUpdateNeeded(cfg)

	if diff {
		cfg.Paused = false
		nretries := -1
		retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
			nretries++
			if getErr := node.query(); getErr != nil {
				logrus.Debugf("Could not get Elasticsearch node resource %v: %v", cfg.DeployName, getErr)
				return getErr
			}
			dep, updateErr := node.constructNodeResource(cfg, owner)
			if updateErr != nil {
				return fmt.Errorf("Could not construct node resource %v for update: %v", cfg.DeployName, updateErr)
			}
			logrus.Infof("Updating node resource %v", cfg.DeployName)
			if updateErr = sdk.Update(dep); updateErr != nil {
				logrus.Debugf("Failed to update node resource %v: %v", cfg.DeployName, updateErr)
			}
			return updateErr
		})
		if retryErr != nil {
			return fmt.Errorf("Error: could not update status for Elasticsearch %v after %v retries: %v", cfg.DeployName, nretries, retryErr)
		}
		logrus.Debugf("Updated Elasticsearch %v after %v retries", cfg.DeployName, nretries)

		currentRevision, err := node.getRevision(cfg)
		if err != nil {
			return err
		}

		err = wait.Poll(time.Second*1, time.Second*10, func() (done bool, err error) {
			if getErr := node.query(); getErr != nil {
				logrus.Debugf("Could not get Elasticsearch node resource %v: %v", cfg.DeployName, getErr)
				return false, getErr
			}

			if awaitingRollout, _ := node.awaitingRollout(cfg, currentRevision); !awaitingRollout {
				err = cfg.PauseNode(owner)
				return err == nil, err
			}

			return false, nil
		})
		return err
	}
	return nil
}

func (cfg *desiredNodeState) IsPauseNeeded() bool {
	// FIXME: to be refactored. query() must not exist here, since
	// we already have information in clusterState
	node := cfg.getNode()
	err := node.query()
	if err != nil {
		// resource doesn't exist, so the pause is not needed
		return false
	}

	unpaused, err := node.needsPause(cfg)
	if err != nil {
		logrus.Errorf("Failed to obtain if there is a pause needed for resource: %v", err)
		return false
	}

	return unpaused
}

func (cfg *desiredNodeState) IsUpdateNeeded() bool {
	// FIXME: to be refactored. query() must not exist here, since
	// we already have information in clusterState
	node := cfg.getNode()
	err := node.query()
	if err != nil {
		// resource doesn't exist, so the update is needed
		return true
	}
	return node.isUpdateNeeded(cfg)
}

func (node *nodeState) setStatefulSet(statefulSet apps.StatefulSet) {
	node.Actual.StatefulSet = &statefulSet
}

func (node *nodeState) setDeployment(deployment apps.Deployment) {
	node.Actual.Deployment = &deployment
}

func (node *nodeState) setReplicaSet(replicaSet apps.ReplicaSet) {
	node.Actual.ReplicaSet = &replicaSet
}

func (node *nodeState) setPod(pod v1.Pod) {
	node.Actual.Pod = &pod
}

func (cfg *desiredNodeState) getAffinity() v1.Affinity {
	labelSelectorReqs := []metav1.LabelSelectorRequirement{}
	if cfg.isNodeClient() {
		labelSelectorReqs = append(labelSelectorReqs, metav1.LabelSelectorRequirement{
			Key:      "es-node-client",
			Operator: metav1.LabelSelectorOpIn,
			Values:   []string{"true"},
		})
	}
	if cfg.isNodeData() {
		labelSelectorReqs = append(labelSelectorReqs, metav1.LabelSelectorRequirement{
			Key:      "es-node-data",
			Operator: metav1.LabelSelectorOpIn,
			Values:   []string{"true"},
		})
	}
	if cfg.isNodeMaster() {
		labelSelectorReqs = append(labelSelectorReqs, metav1.LabelSelectorRequirement{
			Key:      "es-node-master",
			Operator: metav1.LabelSelectorOpIn,
			Values:   []string{"true"},
		})
	}

	return v1.Affinity{
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

func (cfg *desiredNodeState) getEnvVars() []v1.EnvVar {
	return []v1.EnvVar{
		v1.EnvVar{
			Name:  "DC_NAME",
			Value: cfg.DeployName,
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
			Value: fmt.Sprintf("%s-cluster", cfg.ClusterName),
		},
		v1.EnvVar{
			Name:  "CLUSTER_NAME",
			Value: cfg.ClusterName,
		},
		v1.EnvVar{
			Name:  "INSTANCE_RAM",
			Value: cfg.getInstanceRAM(),
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
			Value: fmt.Sprintf("cluster=%s", cfg.ClusterName),
		},
		v1.EnvVar{
			Name:  "IS_MASTER",
			Value: strconv.FormatBool(cfg.isNodeMaster()),
		},
		v1.EnvVar{
			Name:  "HAS_DATA",
			Value: strconv.FormatBool(cfg.isNodeData()),
		},
	}
}

func (cfg *desiredNodeState) getInstanceRAM() string {
	memory := cfg.ESNodeSpec.Resources.Limits.Memory()
	if !memory.IsZero() {
		return memory.String()
	}
	return defaultMemoryLimit
}

func (cfg *desiredNodeState) getESContainer() v1.Container {
	probe := getReadinessProbe()
	return v1.Container{
		Name:            "elasticsearch",
		Image:           cfg.Image,
		ImagePullPolicy: "IfNotPresent",
		Env:             cfg.getEnvVars(),
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
		ReadinessProbe: &probe,
		VolumeMounts:   cfg.getVolumeMounts(),
		Resources:      cfg.ESNodeSpec.Resources,
	}
}

func (cfg *desiredNodeState) getVolumeMounts() []v1.VolumeMount {
	return []v1.VolumeMount{
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
	}
}

// generateMasterPVC method builds PVC for pure master nodes to be used in
// volumeClaimTemplate in StatefulSet spec
func (cfg *desiredNodeState) generateMasterPVC() (v1.PersistentVolumeClaim, bool, error) {
	specVol := cfg.ESNodeSpec.Storage
	if &specVol.StorageClassName != nil &&
		specVol.Size != nil {
		pvc := v1.PersistentVolumeClaim{
			ObjectMeta: metav1.ObjectMeta{
				Name:   "elasticsearch-storage",
				Labels: cfg.getLabels(),
			},
			Spec: v1.PersistentVolumeClaimSpec{
				AccessModes: []v1.PersistentVolumeAccessMode{
					v1.ReadWriteOnce,
				},
				Resources: v1.ResourceRequirements{
					Requests: v1.ResourceList{
						v1.ResourceStorage: *specVol.Size,
					},
				},
				StorageClassName: &specVol.StorageClassName,
			},
		}
		return pvc, true, nil
	} else if (specVol == v1alpha1.ElasticsearchStorageSpec{}) {
		return v1.PersistentVolumeClaim{}, false, nil
	}

	return v1.PersistentVolumeClaim{}, false, fmt.Errorf("Unsupported volume configuration for master in cluster %s", cfg.ClusterName)
}

func (cfg *desiredNodeState) generatePersistentStorage() v1.VolumeSource {
	volSource := v1.VolumeSource{}
	specVol := cfg.ESNodeSpec.Storage

	switch {
	/*
		case specVol.PersistentVolumeClaim != nil:
		volSource.PersistentVolumeClaim = specVol.PersistentVolumeClaim
	*/

	case &specVol.StorageClassName != nil && specVol.Size != nil:
		claimName := fmt.Sprintf("%s-%s", cfg.ClusterName, cfg.DeployName)
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
			StorageClassName: &specVol.StorageClassName,
		}

		err := createOrUpdatePersistentVolumeClaim(volSpec, claimName, cfg.Namespace)
		if err != nil {
			logrus.Errorf("Unable to create PersistentVolumeClaim: %v", err)
		}

	default:
		logrus.Debugf("Defaulting volume source to emptyDir for node %q", cfg.DeployName)
		volSource.EmptyDir = &v1.EmptyDirVolumeSource{}
	}

	return volSource
}

func (cfg *desiredNodeState) getVolumes() []v1.Volume {
	vols := []v1.Volume{
		v1.Volume{
			Name: "elasticsearch-config",
			VolumeSource: v1.VolumeSource{
				ConfigMap: &v1.ConfigMapVolumeSource{
					LocalObjectReference: v1.LocalObjectReference{
						Name: cfg.ConfigMapName,
					},
				},
			},
		},
	}

	if !cfg.isNodePureMaster() {
		vols = append(vols, v1.Volume{
			Name:         "elasticsearch-storage",
			VolumeSource: cfg.generatePersistentStorage(),
		})
	}

	secretName := cfg.SecretName
	if cfg.SecretName == "" {
		secretName = cfg.ClusterName
	}
	vols = append(vols, v1.Volume{
		Name: "certificates",
		VolumeSource: v1.VolumeSource{
			Secret: &v1.SecretVolumeSource{
				SecretName: secretName,
			},
		},
	})
	return vols
}

func (cfg *desiredNodeState) getSelector() (map[string]string, bool) {
	if len(cfg.ESNodeSpec.NodeSelector) == 0 {
		return nil, false
	}
	return cfg.ESNodeSpec.NodeSelector, true
}

func (actualState *actualNodeState) isStatusUpdateNeeded(nodesInStatus v1alpha1.ElasticsearchStatus) bool {
	if actualState.Deployment == nil {
		return false
	}
	for _, node := range nodesInStatus.Nodes {
		if actualState.Deployment.Name == node.DeploymentName {
			if actualState.ReplicaSet == nil {
				return false
			}
			// This is the proper item in the array of node statuses
			if actualState.ReplicaSet.Name != node.ReplicaSetName {
				return true
			}

			if actualState.Pod == nil {
				return false
			}

			if actualState.Pod.Name != node.PodName || string(actualState.Pod.Status.Phase) != node.Status {
				return true
			}
			return false

		}
	}

	// no corresponding nodes in status
	return true
}

func (cfg *desiredNodeState) constructPodTemplateSpec() v1.PodTemplateSpec {
	affinity := cfg.getAffinity()

	template := v1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Labels: cfg.getLabels(),
		},
		Spec: v1.PodSpec{
			Affinity: &affinity,
			Containers: []v1.Container{
				cfg.getESContainer(),
			},
			Volumes: cfg.getVolumes(),
			// ImagePullSecrets: TemplateImagePullSecrets(imagePullSecrets),
			ServiceAccountName: cfg.ServiceAccountName,
		},
	}
	nodeSelector, ok := cfg.getSelector()
	if ok {
		template.Spec.NodeSelector = nodeSelector
	}
	return template
}
