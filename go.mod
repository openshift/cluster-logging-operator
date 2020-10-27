module github.com/openshift/cluster-logging-operator

go 1.14

// Pinned to kubernetes-1.18.3
require (
	cloud.google.com/go v0.54.0 // indirect
	github.com/ViaQ/logerr v1.0.3
	github.com/coreos/prometheus-operator v0.38.1-0.20200424145508-7e176fda06cc
	github.com/emicklei/go-restful v2.12.0+incompatible // indirect
	github.com/go-openapi/spec v0.19.7 // indirect
	github.com/go-openapi/swag v0.19.8 // indirect
	github.com/gobuffalo/envy v1.9.0 // indirect
	github.com/google/go-cmp v0.4.0
	github.com/googleapis/gnostic v0.4.0 // indirect
	github.com/gophercloud/gophercloud v0.8.0 // indirect
	github.com/hashicorp/golang-lru v0.5.4 // indirect
	github.com/mailru/easyjson v0.7.1 // indirect
	github.com/onsi/ginkgo v1.12.0
	github.com/onsi/gomega v1.9.0
	github.com/openshift/api v0.0.0-20200602204738-768b7001fe69
	github.com/openshift/elasticsearch-operator v0.0.0-20200722044541-14fae5dcddfd
	github.com/operator-framework/operator-sdk v0.18.1
	github.com/prometheus/procfs v0.0.10 // indirect
	github.com/rogpeppe/go-internal v1.5.2 // indirect
	github.com/sirupsen/logrus v1.6.0
	golang.org/x/crypto v0.0.0-20200602180216-279210d13fed // indirect
	golang.org/x/sync v0.0.0-20200625203802-6e8e738ad208
	gopkg.in/yaml.v2 v2.2.8
	k8s.io/api v0.18.8
	k8s.io/apimachinery v0.18.8
	k8s.io/client-go v12.0.0+incompatible
	sigs.k8s.io/controller-runtime v0.6.0
	sigs.k8s.io/yaml v1.2.0
)

replace (
	github.com/Azure/go-autorest => github.com/Azure/go-autorest v13.3.2+incompatible // Required by OLM
	k8s.io/client-go => k8s.io/client-go v0.18.3 // Required by prometheus-operator
)
