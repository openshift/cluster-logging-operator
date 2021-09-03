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
	zookeepercm := kafka.NewZookeeperConfigMap(OpenshiftLoggingNS)
	log.V(2).Info("Creating zookeeper configmap", "namespace", zookeepercm.Namespace, "name", zookeepercm.Name)
	if err := f.Test.Client.Create(zookeepercm); err != nil {
		return err
	}

	//standup container running zookeeper
	log.V(2).Info("Adding container", "name", name)
	b.AddContainer(name, ImageRemoteKafkaInit).
		AddVolumeMount("configmap", "/etc/kafka-configmap", "", false).
		AddVolumeMount("config", "/etc/kafka", "", false).
		AddVolumeMount("data", "/var/lib/zookeeper", "", false).
		AddContainerPort("client", zookeeperClientPort).
		AddContainerPort("peer", zookeeperPeerPort).
		AddContainerPort("leader-election", zookeeperLeaderElectionPort).
		WithCmdArgs([]string{"/bin/bash", "/etc/kafka-configmap/init.sh"}).
		WithPrivilege().
		End().
		AddConfigMapVolume(zookeepercm.Name, zookeepercm.Name)

	///////////////////////////////////

	//step b
	// Deploy Broker steps : create configmap, create container, create broker service

	brokercm := kafka.NewBrokerConfigMap(OpenshiftLoggingNS)
	log.V(2).Info("Creating Broker ConfigMap", "namespace", brokercm.Namespace, "name", brokercm.Name)
	if err := f.Test.Client.Create(brokercm); err != nil {
		return err
	}

	//standup pod with container running broker
	log.V(2).Info("Adding container", "name", name)
	b.AddContainer(name, ImageRemoteKafka).
		AddVolumeMount("brokerconfig", "/etc/kafka-configmap", "", false).
		AddVolumeMount("config", "/etc/kafka", "", false).
		AddVolumeMount("brokerlogs", "/opt/kafka/logs", "", false).
		AddEnvVar("extensions", "/opt/kafka/libs/extensions").
		AddEnvVar("data", "/var/lib/kafka/data").
		AddEnvVar("KAFKA_LOG4J_OPTS", "-Dlog4j.configuration=file:/etc/kafka/log4j.properties").
		AddEnvVar("JMX_PORT", strconv.Itoa(int(kafkaJMXPort))).
		AddContainerPort("inside", kafkaInsidePort).
		AddContainerPort("outside", kafkaOutsidePort).
		AddContainerPort("inside", kafkaJMXPort).
		WithCmdArgs([]string{"./bin/kafka-server-start.sh", "/etc/kafka/server.properties"}).
		WithPrivilege().
		End().
		AddConfigMapVolume(brokercm.Name, brokercm.Name)

	/////////////////////////////////////////////
	//step c
	topics := []string{kafka.DefaultTopic}

	for _, topic := range topics {
		// Deploy consumer
		consumerapp := kafka.NewKafkaConsumerDeployment(OpenshiftLoggingNS, topic)
		log.V(2).Info("Creating Broker Consumer app", "namespace", consumerapp.Namespace, "name", consumerapp.Name)
		if err := f.Test.Client.Create(consumerapp); err != nil {
			return err
		}
	}

	return nil
}
