module github.com/openshift/cluster-logging-operator

go 1.16

// Pinned to kubernetes-1.18.3
require (
	cloud.google.com/go v0.83.0 // indirect
	github.com/Azure/go-autorest/autorest v0.11.19 // indirect
	github.com/Azure/go-autorest/autorest/adal v0.9.14 // indirect
	github.com/ViaQ/logerr v1.0.10
	github.com/aws/aws-sdk-go-v2 v1.9.0
	github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs v1.7.0
	github.com/coreos/prometheus-operator v0.38.1-0.20200424145508-7e176fda06cc
	github.com/go-logr/logr v0.4.0
	github.com/google/go-cmp v0.5.6
	github.com/onsi/ginkgo v1.16.4
	github.com/onsi/gomega v1.13.0
	github.com/openshift/api v0.0.0-20210713130143-be21c6cb1bea
	github.com/openshift/elasticsearch-operator v0.0.0-20210825140237-c3774595747e
	github.com/pavel-v-chernykh/keystore-go/v4 v4.1.0
	github.com/pkg/errors v0.9.1
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.7.0
	go.uber.org/atomic v1.8.0 // indirect
	golang.org/x/crypto v0.0.0-20210616213533-5ff15b29337e // indirect
	golang.org/x/net v0.0.0-20210610132358-84b48f89b13b
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c
	golang.org/x/tools v0.1.3 // indirect
	gopkg.in/yaml.v2 v2.4.0
	k8s.io/api v0.21.2
	k8s.io/apimachinery v0.21.2
	k8s.io/client-go v12.0.0+incompatible
	k8s.io/klog/v2 v2.9.0 // indirect
	k8s.io/utils v0.0.0-20210629042839-4a2b36d8d73f // indirect
	sigs.k8s.io/controller-runtime v0.9.2
	sigs.k8s.io/yaml v1.2.0
)

replace k8s.io/client-go => k8s.io/client-go v0.21.2 // Required by prometheus-operator
