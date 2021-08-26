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
	//zookeepercm := kafka.NewZookeeperConfigMap(OpenshiftLoggingNS)
        zookeepercm := runtime.NewConfigMap(b.Pod.Namespace, name, map[string]string{
		"init.sh":              kafka.InitZookeeperScript,
		"zookeeper.properties": kafka.ZookeeperProperties,
		"log4j.properties":     kafka.ZookeeperLog4JProperties,
        })
	log.V(2).Info("Creating zookeeper configmap", "namespace", zookeepercm.Namespace, "name", zookeepercm.Name)
	if err := f.Test.Client.Create(zookeepercm); err != nil {
		return err
	}
	log.V(2).Info("Created zookeepercm",zookeepercm)

	//to standup container running zookeeper : first build initContainer
	log.V(2).Info("Adding now container for zookeeper")
	 b.AddInitContainerzookeeper(name,ImageRemoteKafkaInit)

	 b.AddContainer("zookeeper", ImageRemoteKafka).
		AddVolumeMount("configkafka","/etc/kafka","",false).
		AddVolumeMount("zookeeperlogs","/opt/kafka/logs","",false).
		AddVolumeMount("datazookeeper","/var/lib/zookeeper","",false).
		AddContainerPort("client", zookeeperClientPort).
		AddContainerPort("peer", zookeeperPeerPort).
		AddContainerPort("leader-election", zookeeperLeaderElectionPort).
		WithCmdArgs([]string{"./bin/zookeeper-server-start.sh", "/etc/kafka/zookeeper.properties"}).
		AddEnvVar("KAFKA_LOG4J_OPTS","-Dlog4j.configuration=file:/etc/kafka/log4j.properties").
		WithPrivilege().
		End().
		AddConfigMapVolume(zookeepercm.Name, zookeepercm.Name).
		AddConfigMapVolume("configmap",zookeepercm.Name).//required by the AddInitContainerzookeeper()
	    AddEmptyDirVolume("configkafka").
	    AddEmptyDirVolume("zookeeperlogs").
	    AddEmptyDirVolume("datazookeeper")




	///////////////////////////////////

	//step b
	// Deploy Broker steps : create configmap, create container, create broker service

	//brokercm := kafka.NewBrokerConfigMap(OpenshiftLoggingNS)
        brokercm := runtime.NewConfigMap(b.Pod.Namespace, "brokerkafka", map[string]string{
		"init.sh":           kafka.InitKafkaScript,
		"server.properties": kafka.ServerProperties,
		"client.properties": kafka.ClientProperties,
		"log4j.properties":  kafka.Log4jProperties,
        })

        brokersecret := kafka.NewBrokerSecret(b.Pod.Namespace)

	log.V(2).Info("Creating Broker ConfigMap", "namespace", brokercm.Namespace, "name", brokercm.Name)
	if err := f.Test.Client.Create(brokercm); err != nil {
		return err
	}
	log.V(2).Info("Creating Broker Secret", "namespace", brokersecret.Namespace, "name", brokersecret.Name)
	if err := f.Test.Client.Create(brokersecret); err != nil {
		return err
	}

	//standup pod with container running broker

	log.V(2).Info("Adding container broker")
	b.AddInitContainerbroker("init-config",ImageRemoteKafkaInit,brokercm.Namespace, brokercm.Name)

	b.AddContainer("broker", ImageRemoteKafka).
		AddEnvVar("CLASSPATH","/opt/kafka/libs/extensions/*").
		AddEnvVar("KAFKA_LOG4J_OPTS", "-Dlog4j.configuration=file:/etc/kafka/log4j.properties").
		AddEnvVar("JMX_PORT", strconv.Itoa(int(kafkaJMXPort))).
		AddContainerPort("inside", kafkaInsidePort).
		AddContainerPort("outside", kafkaOutsidePort).
		AddContainerPort("jmx", kafkaJMXPort).
		AddVolumeMount("brokerconfig","/etc/kafka-configmap","",false).
		AddVolumeMount("brokercerts","/etc/kafka-certs","",false).
		AddVolumeMount("configkafka","/etc/kafka","",false).
		AddVolumeMount("brokerlogs","/opt/kafka/logs","",false).
		AddVolumeMount("extensions","/opt/kafka/libs/extensions","",false).
		AddVolumeMount("datakafka","/var/lib/kafka/data","",false).
		WithCmdArgs([]string{"./bin/kafka-server-start.sh", "/etc/kafka/server.properties"}).
		WithPrivilege().
		End().
		AddConfigMapVolume(brokercm.Name, brokercm.Name).
		AddConfigMapVolume("brokerconfig","brokerkafka").
		AddSecretVolume("brokercerts","kafka").
	    AddEmptyDirVolume("brokerlogs").
	    AddEmptyDirVolume("extensions").
	    AddEmptyDirVolume("datakafka")

	/////////////////////////////////////////////
	//step c
	topics := []string{kafka.DefaultTopic}

	for _, topic := range topics {
		// Deploy consumer
		consumerapp := kafka.NewKafkaConsumerDeployment(b.Pod.Namespace, topic)
		log.V(2).Info("Creating Broker Consumer app", "namespace", consumerapp.Namespace, "name", consumerapp.Name)
		if err := f.Test.Client.Create(consumerapp); err != nil {
			return err
		}
	}

	return nil
}
