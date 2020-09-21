package central

import (
	"bytes"

	"github.com/openshift/cluster-logging-operator/pkg/utils"
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/yaml"
)

const (
	normalizerImageName = "fluentd"
)

var (
	serviceProto            core.Service
	statefulSetProto        apps.StatefulSet
	collectorDaemonSetProto apps.DaemonSet
	deploymentProto         apps.Deployment
)

func init() {
	service := &core.Service{}
	dec := yaml.NewYAMLOrJSONDecoder(bytes.NewReader([]byte(serviceTemplate)), len(serviceTemplate))
	if err := dec.Decode(&service); err != nil {
		panic(err)
	}
	serviceProto = *service

	obj := &apps.StatefulSet{}
	dec = yaml.NewYAMLOrJSONDecoder(bytes.NewReader([]byte(statefulSetTemplate)), len(statefulSetTemplate))
	if err := dec.Decode(&obj); err != nil {
		panic(err)
	}
	statefulSetProto = *obj

	ds := &apps.DaemonSet{}
	dec = yaml.NewYAMLOrJSONDecoder(bytes.NewReader([]byte(collectorDaemonSetTemplate)), len(collectorDaemonSetTemplate))
	if err := dec.Decode(&ds); err != nil {
		panic(err)
	}
	collectorDaemonSetProto = *ds

	dp := &apps.Deployment{}
	dec = yaml.NewYAMLOrJSONDecoder(bytes.NewReader([]byte(deploymentTemplate)), len(deploymentTemplate))
	if err := dec.Decode(&dp); err != nil {
		panic(err)
	}
	deploymentProto = *dp
}
func NewCollector() *apps.DaemonSet {
	image := "quay.io/jcantril/fluent-bit:v1.5.2-rh"
	//tolerations
	//nodeselectors
	//resources
	ds := collectorDaemonSetProto.DeepCopy()
	ds.Spec.Template.Spec.Containers[0].Image = image
	return ds
}

func NewService() *core.Service {
	return serviceProto.DeepCopy()
}

func NewNormalizer() *apps.Deployment {
	deployment := deploymentProto.DeepCopy()
	image := utils.GetComponentImage(normalizerImageName)
	deployment.Spec.Template.Spec.Containers[0].Image = image
	deployment.Spec.Replicas = utils.GetInt32(4)
	//tolerations
	//nodeselectors
	//resources
	return deployment
}

const serviceTemplate = `
apiVersion: v1
kind: Service
metadata:
  name: normalizer
  namespace: openshift-logging
  annotations:
    service.alpha.openshift.io/serving-cert-secret-name: fluentd-metrics
  labels:
    component: normalizer
    logging-infra: normalizer
    provider: openshift
spec:
  ports:
  - port: 24224
    name: forward
  selector:
    component: normalizer
    logging-infra: normalizer
    provider: openshift
`
const deploymentTemplate = `
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    component: normalizer
    logging-infra: normalizer
    provider: openshift
  name: normalizer
  namespace: openshift-logging
spec:
  revisionHistoryLimit: 10
  selector:
    matchLabels:
      component: normalizer
      logging-infra: normalizer
      provider: openshift
  serviceName: normalizer
  replicas: 1
  template:
    metadata:
      annotations:
        scheduler.alpha.kubernetes.io/critical-pod: ""
      creationTimestamp: null
      labels:
        component: normalizer
        logging-infra: normalizer
        provider: openshift
      name: normalizer
    spec:
      containers:
      - env:
        - name: NODE_NAME
          valueFrom:
            fieldRef:
              apiVersion: v1
              fieldPath: spec.nodeName
        - name: ALLOWED_PERCENT_OF_DISK
          value: "100"
        - name: K8S_HOST_URL
          value: https://kubernetes.default.svc
        - name: METRICS_CERT
          value: /etc/fluent/metrics/tls.crt
        - name: METRICS_KEY
          value: /etc/fluent/metrics/tls.key
        - name: NODE_IPV4
          valueFrom:
            fieldRef:
              apiVersion: v1
              fieldPath: status.hostIP
        - name: POD_IP
          valueFrom:
            fieldRef:
              apiVersion: v1
              fieldPath: status.podIP
        image: registry.svc.ci.openshift.org/ocp/4.7:logging-fluentd
        imagePullPolicy: Always
        name: fluentd
        ports:
        - containerPort: 24231
          name: metrics
          protocol: TCP
        - containerPort: 24224
          name: forward
          protocol: TCP
        resources:
          requests:
            cpu: 100m
            memory: 256Mi
        terminationMessagePath: /dev/termination-log
        terminationMessagePolicy: File
        volumeMounts:
        - mountPath: /etc/fluent/configs.d/user
          name: config
          readOnly: true
        - mountPath: /opt/app-root/src/run.sh
          name: entrypoint
          readOnly: true
          subPath: run.sh
        - mountPath: /etc/fluent/keys
          name: certs
          readOnly: true
        - mountPath: /var/lib/fluentd
          name: filebufferstorage
        - mountPath: /etc/fluent/metrics
          name: collector-metrics
        - mountPath: /var/run/ocp-collector/secrets/fluentd
          name: certs
      dnsPolicy: ClusterFirst
      nodeSelector:
        kubernetes.io/os: linux
      priorityClassName: cluster-logging
      restartPolicy: Always
      schedulerName: default-scheduler
      securityContext: {}
      serviceAccount: logcollector
      serviceAccountName: logcollector
      terminationGracePeriodSeconds: 10
      tolerations:
      - effect: NoSchedule
        key: node-role.kubernetes.io/master
        operator: Exists
      - effect: NoSchedule
        key: node.kubernetes.io/disk-pressure
        operator: Exists
      volumes:
      - configMap:
          defaultMode: 420
          name: normalizer
        name: config
      - configMap:
          defaultMode: 420
          name: normalizer
        name: entrypoint
      - name: certs
        secret:
          defaultMode: 420
          secretName: fluentd
      - name: collector-metrics
        secret:
          defaultMode: 420
          secretName: fluentd-metrics
      - name: filebufferstorage
        emptyDir: {}
  templateGeneration: 6
  updateStrategy:
    rollingUpdate:
      maxUnavailable: 1
    type: RollingUpdate
`
const statefulSetTemplate = `
apiVersion: apps/v1
kind: StatefulSet
metadata:
  labels:
    component: normalizer
    logging-infra: normalizer
    provider: openshift
  name: normalizer
  namespace: openshift-logging
spec:
  revisionHistoryLimit: 10
  selector:
    matchLabels:
      component: normalizer
      logging-infra: normalizer
      provider: openshift
  serviceName: normalizer
  replicas: 1
  template:
    metadata:
      annotations:
        scheduler.alpha.kubernetes.io/critical-pod: ""
      creationTimestamp: null
      labels:
        component: normalizer
        logging-infra: normalizer
        provider: openshift
      name: normalizer
    spec:
      containers:
      - env:
        - name: NODE_NAME
          valueFrom:
            fieldRef:
              apiVersion: v1
              fieldPath: spec.nodeName
        - name: ALLOWED_PERCENT_OF_DISK
          value: "100"
        - name: K8S_HOST_URL
          value: https://kubernetes.default.svc.cluster.local
        - name: METRICS_CERT
          value: /etc/fluent/metrics/tls.crt
        - name: METRICS_KEY
          value: /etc/fluent/metrics/tls.key
        - name: BUFFER_QUEUE_LIMIT
          value: "32"
        - name: BUFFER_SIZE_LIMIT
          value: 8m
        - name: FILE_BUFFER_LIMIT
          value: 256Mi
        - name: FLUENTD_CPU_LIMIT
          valueFrom:
            resourceFieldRef:
              containerName: fluentd
              divisor: "0"
              resource: limits.cpu
        - name: FLUENTD_MEMORY_LIMIT
          valueFrom:
            resourceFieldRef:
              containerName: fluentd
              divisor: "0"
              resource: limits.memory
        - name: NODE_IPV4
          valueFrom:
            fieldRef:
              apiVersion: v1
              fieldPath: status.hostIP
        - name: POD_IP
          valueFrom:
            fieldRef:
              apiVersion: v1
              fieldPath: status.podIP
        - name: CDM_KEEP_EMPTY_FIELDS
          value: message
        image: registry.svc.ci.openshift.org/ocp/4.7:logging-fluentd
        imagePullPolicy: Always
        name: fluentd
        ports:
        - containerPort: 24231
          name: metrics
          protocol: TCP
        - containerPort: 24224
          name: forward
          protocol: TCP
        resources:
          requests:
            cpu: 100m
            memory: 256Mi
        terminationMessagePath: /dev/termination-log
        terminationMessagePolicy: File
        volumeMounts:
        - mountPath: /etc/fluent/configs.d/user
          name: config
          readOnly: true
        - mountPath: /opt/app-root/src/run.sh
          name: entrypoint
          readOnly: true
          subPath: run.sh
        - mountPath: /etc/fluent/keys
          name: certs
          readOnly: true
        - mountPath: /var/lib/fluentd
          name: filebufferstorage
        - mountPath: /etc/fluent/metrics
          name: collector-metrics
        - mountPath: /var/run/ocp-collector/secrets/fluentd
          name: certs
      dnsPolicy: ClusterFirst
      nodeSelector:
        kubernetes.io/os: linux
      priorityClassName: cluster-logging
      restartPolicy: Always
      schedulerName: default-scheduler
      securityContext: {}
      serviceAccount: logcollector
      serviceAccountName: logcollector
      terminationGracePeriodSeconds: 10
      tolerations:
      - effect: NoSchedule
        key: node-role.kubernetes.io/master
        operator: Exists
      - effect: NoSchedule
        key: node.kubernetes.io/disk-pressure
        operator: Exists
      volumes:
      - configMap:
          defaultMode: 420
          name: normalizer
        name: config
      - configMap:
          defaultMode: 420
          name: normalizer
        name: entrypoint
      - name: certs
        secret:
          defaultMode: 420
          secretName: fluentd
      - name: collector-metrics
        secret:
          defaultMode: 420
          secretName: fluentd-metrics
  volumeClaimTemplates:
  - metadata:
      name: filebufferstorage
    spec:
      accessModes: [ "ReadWriteOnce" ]
      resources:
        requests:
          storage: 40Gi
  templateGeneration: 6
  updateStrategy:
    rollingUpdate:
      maxUnavailable: 1
    type: RollingUpdate
`
const collectorDaemonSetTemplate = `
apiVersion: apps/v1
kind: DaemonSet
metadata:
  labels:
    component: collector
    logging-infra: collector
    provider: openshift
  name: collector
  namespace: openshift-logging
spec:
  selector:
    matchLabels:
      component: collector
      logging-infra: collector
      provider: openshift
  template:
    metadata:
      annotations:
        scheduler.alpha.kubernetes.io/critical-pod: ""
      creationTimestamp: null
      labels:
        component: collector
        logging-infra: collector
        provider: openshift
      name: fluentbit
    spec:
      containers:
      - name: fluentbit
        env:
        - name: POD_IP
          valueFrom:
            fieldRef:
              apiVersion: v1
              fieldPath: status.podIP
        image: quay.io/jcantril/fluent-bit:v1.5.2-rh
        imagePullPolicy: IfNotPresent
        volumeMounts:
        - mountPath: /var/log
          name: varlog
        - mountPath: /etc/fluent-bit
          name: fluentbit-conf
        securityContext:
          privileged: true
      dnsPolicy: ClusterFirst
      nodeSelector:
        kubernetes.io/os: linux
      priorityClassName: cluster-logging
      restartPolicy: Always
      schedulerName: default-scheduler
      securityContext: {}
      serviceAccount: logcollector
      serviceAccountName: logcollector
      terminationGracePeriodSeconds: 10
      tolerations:
      - effect: NoSchedule
        key: node-role.kubernetes.io/master
        operator: Exists
      - effect: NoSchedule
        key: node.kubernetes.io/disk-pressure
        operator: Exists
      volumes:
      - configMap:
          defaultMode: 420
          name: fluentbit
        name: fluentbit-conf
      - hostPath:
          path: /var/log
          type: ""
        name: varlog
      - configMap:
          defaultMode: 420
          items:
          - key: ca-bundle.crt
            path: tls-ca-bundle.pem
          name: fluentd-trusted-ca-bundle
        name: fluentd-trusted-ca-bundle
  updateStrategy:
    rollingUpdate:
      maxUnavailable: 1
    type: RollingUpdate
`
