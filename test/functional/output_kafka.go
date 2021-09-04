package functional

import (
	"github.com/ViaQ/logerr/log"
	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/test/helpers/kafka"
	"github.com/openshift/cluster-logging-operator/test/runtime"
	"strconv"
	"strings"
)

const (
	// Kafka deployment definitions
	ImageRemoteKafka            = kafka.KafkaImage
	ImageRemoteKafkaInit        = kafka.KafkaInitUtilsImage
	OpenshiftLoggingNS          = "openshift-logging"
	kafkaBrokerContainerName    = "broker"
	kafkaBrokerComponent        = "kafka"
	kafkaBrokerProvider         = "openshift"
	kafkaNodeReader             = "kafka-node-reader"
	kafkaNodeReaderBinding      = "kafka-node-reader-binding"
	kafkaInsidePort             = int32(9093)
	kafkaOutsidePort            = int32(9094)
	kafkaJMXPort                = int32(5555)
       	zookeeperDeploymentName     = "zookeeper"
	zookeeperComponent          = "zookeeper"
	zookeeperProvider           = "openshift"
	zookeeperClientPort         = int32(2181)
	zookeeperPeerPort           = int32(2888)
	zookeeperLeaderElectionPort = int32(3888)
)

func (f *FluentdFunctionalFramework) addKafkaOutput(b *runtime.PodBuilder, output logging.OutputSpec) error {
	log.V(2).Info("Adding kafka output", "name", output.Name)
	name := strings.ToLower(output.Name)

	log.V(2).Info("Standing up Kafka instance", "name", name)
	//steps to deploy kafka single node cluster a. Deploy Zookeeper b. Deploy Broker c. Deploy kafka Consumer

	//step a
	// Deploy Zookeeper steps : create configmap, create container, create zookeeper service
  
        //zookeeperm - configmap with zookeeperDeploymentName being "zookeeper", in b.Pod.Namespace
	zookeepercm := kafka.NewZookeeperConfigMap(b.Pod.Namespace)
	log.V(2).Info("Creating zookeeper configmap", "namespace", zookeepercm.Namespace, "name", zookeepercm.Name)
	if err := f.Test.Client.Create(zookeepercm); err != nil {
		return err
	}

        //init container for zookeep
        b.AddInitContainerZookeeper("init-config",RemoteImageKafkaInit).
               AddVolumeMount("configmapkafka", "/etc/kafka-configmap").
               AddVolumeMount("configkafka","/etc/kafka").
               AddVolumeMount("datazookeeper","/var/lib/zookeeper").
               WithCmdArgs([]string{"/bin/bash","/etc/kafka-configmap/init.sh"})
               
	//standup container running zookeeper
	log.V(2).Info("Adding container", "name", name)
	b.AddContainer("zookeeper", ImageRemoteKafka).
                AddEnvVar("KAFKA_LOG4J_OPTS","-Dlog4j.configuration=file:/etc/kafka/log4j.properties")
		AddVolumeMount("configmap", "/etc/kafka-configmap", "", false).
		AddVolumeMount("config", "/etc/kafka", "", false).
		AddVolumeMount("data", "/var/lib/zookeeper", "", false).
		AddContainerPort("client", zookeeperClientPort).
		AddContainerPort("peer", zookeeperPeerPort).
		AddContainerPort("leader-election", zookeeperLeaderElectionPort).
                AddVolumeMount("configmapkafka", "/etc/kafka-configmap").
                AddVolumeMount("configkafka","/etc/kafka").
                AddVolumeMount("datazookeeper","/var/lib/zookeeper").
		WithCmdArgs([]string{"./bin/zookeeper-server-start.sh", "/etc/kafka/zookeeper.properties"}).
		End().
		AddConfigMapVolume("configmapkafka", zookeeperDeploymentName).
		AddEmptyDirVolume("configkafka").
		AddEmptyDirVolume("zookeeperlogs").
		AddEmptyDirVolume("data")
                

	///////////////////////////////////

	//step b
	// Deploy Broker steps : create configmap, create container, create broker service

       // configmap for broker with DeploymentName=kafka, b.Pod.Namespace, and data specific to broker
	brokercm := kafka.NewBrokerConfigMap(b.Pod.Namespace)
	log.V(2).Info("Creating Broker ConfigMap", "namespace", brokercm.Namespace, "name", brokercm.Name)
	if err := f.Test.Client.Create(brokercm); err != nil {
		return err
	}

        brokersecret := kafka.NewBrokerSecret(b.Pod.Namespace)
	if err := f.Test.Client.Create(brokersecret); err != nil {
		return err
	}
	//standup pod with container running broker
        b.AddInitContainerBroker("init-config",RemoteImageKafkaInit).
               AddEnvVarFromEnvVarSourceNode("NODE_NAME").
               AddEnvVarFromEnvVarSourcePod("POD_NAME").
               AddEnvVarFromEnvVarSourceNamespace("POD_NAMESPACE").
               AddEnvVar("ADVERTISE_ADDR",fmt.Sprintf("%s.%s.svc.cluster.local", DeploymentName, namespace)).
               WithCmdArgs([]string{"/bin/bash","/etc/kafka-configmap/init.sh"}).
               AddVolumeMount("brokerconfig", "/etc/kafka-configmap").
               AddVolumeMount("configkafka","/etc/kafka").
               AddVolumeMount("extensions","/opt/kafka/libs/extensions").


	log.V(2).Info("Adding container", "name", name)
	b.AddContainer(kafkaBrokerContainerName, ImageRemoteKafka).
                AddEnvVar("CLASSPATH","/opt/kafka/libs/extensions/*")
                AddEnvVar("KAFKA_LOG4J_OPTS","-Dlog4j.configuration=file:/etc/kafka/log4j.properties")
                AddEnvVar("JMX_PORT",strconv.Itoa(kafkaJMXPort))
		AddContainerPort("inside", kafkaInsidePort).
		AddContainerPort("outside", kafkaOutsidePort).
		AddContainerPort("inside", kafkaJMXPort).
		WithCmdArgs([]string{"./bin/kafka-server-start.sh", "/etc/kafka/server.properties"}).
		AddVolumeMount("brokerconfig", "/etc/kafka-configmap", "", false).
		AddVolumeMount("configkafka", "/etc/kafka", "", false).
		AddVolumeMount("brokerlogs", "/opt/kafka/logs", "", false).
		AddEnvVar("extensions", "/opt/kafka/libs/extensions").
		AddEnvVar("data", "/var/lib/kafka/data").
		End().
		AddConfigMapVolume("brokerconfig",DeploymentName).
                AddSecretVolume("brokercerts",DeploymentName).
                AddEmptyDirVolume("brokerlogs").
                AddEmptyDirVolume("configkafka").
                AddEmptyDirVolume("extensions")
                

	/////////////////////////////////////////////
	//step c
	topics := []string{kafka.DefaultTopic}

	for _, topic := range topics {
                containername := kafka.ConsumerNameForTopic(topic)
		// Deploy consumer
		//consumerapp := kafka.NewKafkaConsumerDeployment(OpenshiftLoggingNS, topic)

                b.AddInitContainer("topic-create",ImageRemoteKafka).
                    WithCmdArgs([]string{"./bin/kafka-topics.sh",
					"--zookeeper",
					"zookeeper.openshift-logging.svc.cluster.local:2181",
					"--create",
					"--if-not-exists",
					"--topic",
					topic,
					"--partitions",
					"1",
					"--replication-factor",
					"1",
				})
                b.AddContainer(containername, ImageRemoteKafka).
                     WithCmdArg([]string{"/bin/bash","-ce",fmt.Sprintf(
						`./bin/kafka-console-consumer.sh --bootstrap-server %s --topic %s --from-beginning --consumer.config /etc/kafka-configmap/client.properties > /shared/consumed.logs`,
						ClusterLocalEndpoint(namespace),
						topic,
					)}).
                     AddVolumeMount("brokerconfig", "/etc/kafka-configmap").
                     AddVolumeMount("brokercert", "/etc/kafka-certs").
                     AddVolumeMount("shared", "/shared").
                     End.
                     AddConfigMapVolume("brokerconfig",DeploymentName).
                     AddSecretNameVolume("brokercert", DeploymentName).
                     AddEmptyDirVolume("shared")
                


		log.V(2).Info("Creating Broker Consumer app", "namespace", consumerapp.Namespace, "name", consumerapp.Name)
		if err := f.Test.Client.Create(consumerapp); err != nil {
			return err
		}
	}

	return nil
}
