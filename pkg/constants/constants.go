// Licensed to Red Hat, Inc under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Red Hat, Inc licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package constants

const (
	SingletonName = "instance"
	OpenshiftNS   = "openshift-logging"
	// global proxy / trusted ca bundle consts
	ProxyName                  = "cluster"
	TrustedCABundleKey         = "ca-bundle.crt"
	InjectTrustedCABundleLabel = "config.openshift.io/inject-trusted-cabundle"
	TrustedCABundleMountFile   = "tls-ca-bundle.pem"
	TrustedCABundleMountDir    = "/etc/pki/ca-trust/extracted/pem/"
	TrustedCABundleHashName    = "logging.openshift.io/hash"
	SecretHashPrefix           = "logging.openshift.io/"
	FluentdTrustedCAName       = "fluentd-trusted-ca-bundle"
	KibanaTrustedCAName        = "kibana-trusted-ca-bundle"
	// internal elasticsearch FQDN to prevent to connect to the global proxy
	ElasticsearchFQDN   = "elasticsearch.openshift-logging.svc.cluster.local"
	ElasticsearchName   = "elasticsearch"
	ElasticsearchPort   = "9200"
	FluentdName         = "fluentd"
	KibanaName          = "kibana"
	KibanaProxyName     = "kibana-proxy"
	CuratorName         = "curator"
	LogStoreURL         = "https://" + ElasticsearchFQDN + ":" + ElasticsearchPort
	MasterCASecretName  = "master-certs"
	CollectorSecretName = "fluentd"
)

var ReconcileForGlobalProxyList = []string{FluentdTrustedCAName}
