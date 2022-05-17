package functional

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/ViaQ/logerr/v2/log"
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	"github.com/openshift/cluster-logging-operator/test/helpers/kafka"
)

const (
	// Kafka deployment definitions
	ImageRemoteKafka            = kafka.KafkaImage
	ImageRemoteKafkaInit        = kafka.KafkaInitUtilsImage
	kafkaBrokerContainerName    = "broker"
	kafkaZookeeperContainerName = "zookeeper"

	kafkaInsidePort         = int32(9093)
	kafkaOutsidePort        = int32(9094)
	kafkaJMXPort            = int32(5555)
	zookeeperDeploymentName = "zookeeper"

	zookeeperClientPort         = int32(2181)
	zookeeperPeerPort           = int32(2888)
	zookeeperLeaderElectionPort = int32(3888)
)

func (f *CollectorFunctionalFramework) addKafkaOutput(b *runtime.PodBuilder, output logging.OutputSpec) error {
	logger := log.NewLogger("output-kafka-testing")

	logger.V(2).Info("Adding kafka output", "name", output.Name)
	name := strings.ToLower(output.Name)

	logger.V(2).Info("Standing up Kafka instance", "name", name)
	//steps to deploy kafka single node cluster a. Deploy Zookeeper b. Deploy Broker c. Deploy kafka Consumer

	//step a
	// Deploy Zookeeper steps : create configmap, create container, create zookeeper service

	//zookeeperm - configmap with zookeeperDeploymentName being "zookeeper", in b.Pod.Namespace
	zookeepercm := kafka.NewZookeeperConfigMapFunctionalTestPod(b.Pod.Namespace)
	logger.V(2).Info("Creating zookeeper configmap", "namespace", zookeepercm.Namespace, "name", zookeepercm.Name)
	if err := f.Test.Client.Create(zookeepercm); err != nil {
		return err
	}

	//init container for zookeeper
	b.AddInitContainer("init-config0", ImageRemoteKafkaInit).
		AddVolumeMount("configmapkafka", "/etc/kafka-configmap", "", false).
		AddVolumeMount("configkafka", "/etc/kafka", "", false).
		AddVolumeMount("zookeeper-data", "/var/lib/zookeeper", "", false).
		WithCmdArgs([]string{"./bin/bash", "/etc/kafka-configmap/init.sh"}).
		End()

	//standup container running zookeeper
	logger.V(2).Info("Adding container", "name", name)
	b.AddContainer(kafkaZookeeperContainerName, ImageRemoteKafka).
		AddEnvVar("KAFKA_LOG4J_OPTS", "-Dlog4j.configuration=file:/etc/kafka/log4j.properties").
		AddContainerPort("client", zookeeperClientPort).
		AddContainerPort("peer", zookeeperPeerPort).
		AddContainerPort("leader-election", zookeeperLeaderElectionPort).
		AddVolumeMount("configmapkafka", "/etc/kafka-configmap", "", false).
		AddVolumeMount("configkafka", "/etc/kafka", "", false).
		AddVolumeMount("zookeeper-data", "/var/lib/zookeeper", "", false).
		WithCmd([]string{"./bin/zookeeper-server-start.sh", "/etc/kafka/zookeeper.properties"}).
		End().
		AddConfigMapVolume("configmapkafka", zookeeperDeploymentName).
		AddEmptyDirVolume("configkafka").
		AddEmptyDirVolume("zookeeper-logs").
		AddEmptyDirVolume("zookeeper-data")

	///////////////////////////////////

	//step b
	// Deploy Broker steps : create configmap, create container, create broker service

	// configmap for broker with DeploymentName=kafka, b.Pod.Namespace, and data specific to broker
	brokercm := kafka.NewBrokerConfigMapFunctionalTestPod(b.Pod.Namespace)
	logger.V(2).Info("Creating Broker ConfigMap", "namespace", brokercm.Namespace, "name", brokercm.Name)
	if err := f.Test.Client.Create(brokercm); err != nil {
		return err
	}

	//Doing below at the time of creating spec for fluentd container now as tls.key, tls.crts need to be mounted on fluentd pod

	//add broker secret containing keys,certs for secure connection in the below fluentd container

	b.AddSecretVolume("kafka", kafka.DeploymentName).
		GetContainer(constants.CollectorName).
		AddVolumeMount("kafka", "/var/run/ocp-collector/secrets/kafka", "", true)

	//standup pod with container running broker
	b.AddInitContainer("init-config1", ImageRemoteKafkaInit).
		AddEnvVar("NODE_NAME", "functional-test-node").
		AddEnvVar("POD_NAME", "functional").
		AddEnvVar("POD_NAMESPACE", b.Pod.Namespace).
		AddEnvVar("ADVERTISE_ADDR", fmt.Sprintf("%s.%s.svc.cluster.local", kafka.DeploymentName, b.Pod.Namespace)).
		WithCmdArgs([]string{"/bin/bash", "/etc/kafka-configmap/init.sh"}).
		AddVolumeMount("brokerconfig", "/etc/kafka-configmap", "", false).
		AddVolumeMount("configkafka", "/etc/kafka", "", false).
		AddVolumeMount("extensions", "/opt/kafka/libs/extensions", "", false).
		End()

	cmdCreateTopicAndDeployBroker := "./bin/kafka-server-start.sh /etc/kafka-configmap/server.properties"

	logger.V(2).Info("Adding container", "name", name)
	b.AddContainer(kafkaBrokerContainerName, ImageRemoteKafka).
		AddEnvVar("CLASSPATH", "/opt/kafka/libs/extensions/*").
		AddEnvVar("KAFKA_LOG4J_OPTS", "-Dlog4j.configuration=file:/etc/kafka/log4j.properties").
		AddEnvVar("JMX_PORT", strconv.Itoa(int(kafkaJMXPort))).
		AddContainerPort("inside", kafkaInsidePort).
		AddContainerPort("outside", kafkaOutsidePort).
		AddContainerPort("jmx", kafkaJMXPort).
		WithCmd([]string{"/bin/bash", "-c", cmdCreateTopicAndDeployBroker}).
		AddVolumeMount("brokerconfig", "/etc/kafka-configmap", "", false).
		AddVolumeMount("configkafka", "/etc/kafka", "", false).
		AddVolumeMount("brokerlogs", "/opt/kafka/logs", "", false).
		AddVolumeMount("kafka", "/var/run/ocp-collector/secrets/kafka", "", false).
		AddEnvVar("extensions", "/opt/kafka/libs/extensions").
		AddEnvVar("data", "/var/lib/kafka/data").
		End().
		AddConfigMapVolume("brokerconfig", kafka.DeploymentName).
		AddEmptyDirVolume("brokerlogs").
		AddEmptyDirVolume("extensions")

	/////////////////////////////////////////////
	//step c
	topics := []string{kafka.AppLogsTopic}

	for _, topic := range topics {
		containername := kafka.ConsumerNameForTopic(topic)
		// Deploy consumer  - reference implementation followed from the e2e kafka test as below
		//consumerapp := kafka.NewKafkaConsumerDeployment(OpenshiftLoggingNS, topic)

		logger.V(2).Info("Creating a Topic and Deploying Consumer app")
		//create topic and deploy consumer app
		cmdCreateTopic := fmt.Sprintf(`./bin/kafka-topics.sh --zookeeper localhost:2181 --create --if-not-exists --topic %s --partitions 1 --replication-factor 1 ;`, topic)
		cmdRunConsumer := fmt.Sprintf(`./bin/kafka-console-consumer.sh --bootstrap-server %s --topic %s --from-beginning > /shared/consumed.logs ;`, "localhost:9092", topic)
		cmdSlice := []string{"sleep 120;", cmdCreateTopic, cmdRunConsumer}
		cmdJ := strings.Join(cmdSlice, "")
		cmdCreateTopicAndDeployConsumer := []string{"/bin/bash", "-c", cmdJ}

		b.AddContainer(containername, ImageRemoteKafka).
			WithCmd(cmdCreateTopicAndDeployConsumer).
			AddVolumeMount("brokerconfig", "/etc/kafka-configmap", "", false).
			AddVolumeMount("kafka", "/var/run/ocp-collector/secrets/kafka", "", false).
			AddVolumeMount("shared", "/shared", "", false).
			End().
			AddEmptyDirVolume("shared")

	}

	return nil
}
