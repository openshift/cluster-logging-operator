module github.com/openshift/cluster-logging-operator

go 1.14

// Pinned to kubernetes-1.18.3
require (
	cloud.google.com/go v0.54.0 // indirect
	github.com/ViaQ/logerr v1.0.10
	github.com/aws/aws-sdk-go-v2 v1.9.0
	github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs v1.7.0
	github.com/coreos/prometheus-operator v0.38.1-0.20200424145508-7e176fda06cc
	github.com/go-logr/logr v0.2.1
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/google/go-cmp v0.5.6
	github.com/googleapis/gnostic v0.4.0 // indirect
	github.com/onsi/ginkgo v1.12.1
	github.com/onsi/gomega v1.10.1
	github.com/openshift/api v0.0.0-20200602204738-768b7001fe69
	github.com/openshift/elasticsearch-operator v0.0.0-20210215150408-c19b2867ac70
	github.com/pavel-v-chernykh/keystore-go/v4 v4.1.0
	github.com/pkg/errors v0.9.1
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.5.1
	golang.org/x/net v0.0.0-20201021035429-f5854403a974
	golang.org/x/sync v0.0.0-20201020160332-67f06af15bc9
	gopkg.in/yaml.v2 v2.3.0
	k8s.io/api v0.18.8
	k8s.io/apimachinery v0.18.8
	k8s.io/client-go v12.0.0+incompatible
	sigs.k8s.io/controller-runtime v0.6.3
	sigs.k8s.io/yaml v1.2.0
)

replace (
	github.com/Azure/go-autorest => github.com/Azure/go-autorest/autorest/adal v0.9.12 // Required by OLM
	github.com/gogo/protobuf => github.com/gogo/protobuf v1.3.2
	k8s.io/client-go => k8s.io/client-go v0.18.3 // Required by prometheus-operator
)
