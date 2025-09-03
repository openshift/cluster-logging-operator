// Code generated for package testdata by go-bindata DO NOT EDIT. (@generated)
// sources:
// testdata/UIPlugin/UIPlugin.yaml
// testdata/eventrouter/eventrouter.yaml
// testdata/external-log-stores/cert_generation.sh
// testdata/external-log-stores/elasticsearch/6/http/no_user/configmap.yaml
// testdata/external-log-stores/elasticsearch/6/http/no_user/deployment.yaml
// testdata/external-log-stores/elasticsearch/6/http/user_auth/configmap.yaml
// testdata/external-log-stores/elasticsearch/6/http/user_auth/deployment.yaml
// testdata/external-log-stores/elasticsearch/6/https/no_user/configmap.yaml
// testdata/external-log-stores/elasticsearch/6/https/no_user/deployment.yaml
// testdata/external-log-stores/elasticsearch/6/https/user_auth/configmap.yaml
// testdata/external-log-stores/elasticsearch/6/https/user_auth/deployment.yaml
// testdata/external-log-stores/elasticsearch/7/http/no_user/configmap.yaml
// testdata/external-log-stores/elasticsearch/7/http/no_user/deployment.yaml
// testdata/external-log-stores/elasticsearch/7/http/user_auth/configmap.yaml
// testdata/external-log-stores/elasticsearch/7/http/user_auth/deployment.yaml
// testdata/external-log-stores/elasticsearch/7/https/no_user/configmap.yaml
// testdata/external-log-stores/elasticsearch/7/https/no_user/deployment.yaml
// testdata/external-log-stores/elasticsearch/7/https/user_auth/configmap.yaml
// testdata/external-log-stores/elasticsearch/7/https/user_auth/deployment.yaml
// testdata/external-log-stores/elasticsearch/8/http/no_user/configmap.yaml
// testdata/external-log-stores/elasticsearch/8/http/no_user/deployment.yaml
// testdata/external-log-stores/elasticsearch/8/http/user_auth/configmap.yaml
// testdata/external-log-stores/elasticsearch/8/http/user_auth/deployment.yaml
// testdata/external-log-stores/elasticsearch/8/https/no_user/configmap.yaml
// testdata/external-log-stores/elasticsearch/8/https/no_user/deployment.yaml
// testdata/external-log-stores/elasticsearch/8/https/user_auth/configmap.yaml
// testdata/external-log-stores/elasticsearch/8/https/user_auth/deployment.yaml
// testdata/external-log-stores/fluentd/insecure/configmap.yaml
// testdata/external-log-stores/fluentd/insecure/deployment.yaml
// testdata/external-log-stores/fluentd/insecure/http-configmap.yaml
// testdata/external-log-stores/fluentd/secure/cm-mtls-share.yaml
// testdata/external-log-stores/fluentd/secure/cm-mtls.yaml
// testdata/external-log-stores/fluentd/secure/cm-serverauth-share.yaml
// testdata/external-log-stores/fluentd/secure/cm-serverauth.yaml
// testdata/external-log-stores/fluentd/secure/deployment.yaml
// testdata/external-log-stores/fluentd/secure/http-cm-mtls.yaml
// testdata/external-log-stores/fluentd/secure/http-cm-serverauth.yaml
// testdata/external-log-stores/kafka/amqstreams/kafka-no-auth-cluster.yaml
// testdata/external-log-stores/kafka/amqstreams/kafka-no-auth-consumer-job.yaml
// testdata/external-log-stores/kafka/amqstreams/kafka-sasl-cluster.yaml
// testdata/external-log-stores/kafka/amqstreams/kafka-sasl-consumer-job.yaml
// testdata/external-log-stores/kafka/amqstreams/kafka-sasl-consumers-config.yaml
// testdata/external-log-stores/kafka/amqstreams/kafka-sasl-user.yaml
// testdata/external-log-stores/kafka/amqstreams/kafka-topic.yaml
// testdata/external-log-stores/kafka/cert_generation.sh
// testdata/external-log-stores/kafka/kafka-rbac.yaml
// testdata/external-log-stores/kafka/kafka-svc.yaml
// testdata/external-log-stores/kafka/plaintext-ssl/consumer-configmap.yaml
// testdata/external-log-stores/kafka/plaintext-ssl/kafka-configmap.yaml
// testdata/external-log-stores/kafka/plaintext-ssl/kafka-consumer-deployment.yaml
// testdata/external-log-stores/kafka/plaintext-ssl/kafka-statefulset.yaml
// testdata/external-log-stores/kafka/sasl-plaintext/consumer-configmap.yaml
// testdata/external-log-stores/kafka/sasl-plaintext/kafka-configmap.yaml
// testdata/external-log-stores/kafka/sasl-plaintext/kafka-consumer-deployment.yaml
// testdata/external-log-stores/kafka/sasl-plaintext/kafka-statefulset.yaml
// testdata/external-log-stores/kafka/sasl-ssl/consumer-configmap.yaml
// testdata/external-log-stores/kafka/sasl-ssl/kafka-configmap.yaml
// testdata/external-log-stores/kafka/sasl-ssl/kafka-consumer-deployment.yaml
// testdata/external-log-stores/kafka/sasl-ssl/kafka-statefulset.yaml
// testdata/external-log-stores/kafka/zookeeper/configmap-ssl.yaml
// testdata/external-log-stores/kafka/zookeeper/configmap.yaml
// testdata/external-log-stores/kafka/zookeeper/zookeeper-statefulset.yaml
// testdata/external-log-stores/kafka/zookeeper/zookeeper-svc.yaml
// testdata/external-log-stores/loki/loki-configmap.yaml
// testdata/external-log-stores/loki/loki-deployment.yaml
// testdata/external-log-stores/otel/otel-collector.yaml
// testdata/external-log-stores/rsyslog/insecure/configmap.yaml
// testdata/external-log-stores/rsyslog/insecure/deployment.yaml
// testdata/external-log-stores/rsyslog/insecure/svc.yaml
// testdata/external-log-stores/rsyslog/secure/configmap.yaml
// testdata/external-log-stores/rsyslog/secure/deployment.yaml
// testdata/external-log-stores/rsyslog/secure/svc.yaml
// testdata/external-log-stores/splunk/route-edge_splunk_template.yaml
// testdata/external-log-stores/splunk/route-passthrough_splunk_template.yaml
// testdata/external-log-stores/splunk/secret_splunk_template.yaml
// testdata/external-log-stores/splunk/secret_tls_passphrase_splunk_template.yaml
// testdata/external-log-stores/splunk/secret_tls_splunk_template.yaml
// testdata/external-log-stores/splunk/statefulset_splunk-8.2_template.yaml
// testdata/external-log-stores/splunk/statefulset_splunk-9.0_template.yaml
// testdata/generatelog/42981.yaml
// testdata/generatelog/container_json_log_template.json
// testdata/generatelog/container_json_log_template_unannoted.json
// testdata/generatelog/container_non_json_log_template.json
// testdata/generatelog/logging-performance-app-generator.json
// testdata/generatelog/multi_container_json_log_template.yaml
// testdata/generatelog/multiline-error-log.yaml
// testdata/logfilemetricexporter/lfme.yaml
// testdata/loki-log-alerts/cluster-monitoring-config.yaml
// testdata/loki-log-alerts/loki-app-alerting-rule-template.yaml
// testdata/loki-log-alerts/loki-app-recording-rule-template.yaml
// testdata/loki-log-alerts/loki-infra-alerting-rule-template.yaml
// testdata/loki-log-alerts/loki-infra-recording-rule-template.yaml
// testdata/loki-log-alerts/user-workload-monitoring-config.yaml
// testdata/lokistack/fine-grained-access-roles.yaml
// testdata/lokistack/lokistack-simple-ipv6-tls.yaml
// testdata/lokistack/lokistack-simple-ipv6.yaml
// testdata/lokistack/lokistack-simple-tls.yaml
// testdata/lokistack/lokistack-simple.yaml
// testdata/minIO/deploy.yaml
// testdata/observability.openshift.io_clusterlogforwarder/48593.yaml
// testdata/observability.openshift.io_clusterlogforwarder/67421.yaml
// testdata/observability.openshift.io_clusterlogforwarder/68318.yaml
// testdata/observability.openshift.io_clusterlogforwarder/71049.yaml
// testdata/observability.openshift.io_clusterlogforwarder/71749.yaml
// testdata/observability.openshift.io_clusterlogforwarder/affinity-81397.yaml
// testdata/observability.openshift.io_clusterlogforwarder/affinity-81398.yaml
// testdata/observability.openshift.io_clusterlogforwarder/audit-policy.yaml
// testdata/observability.openshift.io_clusterlogforwarder/azureMonitor-min-opts.yaml
// testdata/observability.openshift.io_clusterlogforwarder/azureMonitor.yaml
// testdata/observability.openshift.io_clusterlogforwarder/clf-external-loki-with-secret-tenantKey.yaml
// testdata/observability.openshift.io_clusterlogforwarder/clf-external-loki-with-secret.yaml
// testdata/observability.openshift.io_clusterlogforwarder/clf-kafka-multi-brokers.yaml
// testdata/observability.openshift.io_clusterlogforwarder/clf-kafka-multi-topics.yaml
// testdata/observability.openshift.io_clusterlogforwarder/clf-kafka-no-auth.yaml
// testdata/observability.openshift.io_clusterlogforwarder/clf-kafka-sasl-ssl.yaml
// testdata/observability.openshift.io_clusterlogforwarder/clf-kafka-with-auth.yaml
// testdata/observability.openshift.io_clusterlogforwarder/cloudwatch-accessKey.yaml
// testdata/observability.openshift.io_clusterlogforwarder/cloudwatch-iamRole.yaml
// testdata/observability.openshift.io_clusterlogforwarder/cloudwatch-multiple-iamRole.yaml
// testdata/observability.openshift.io_clusterlogforwarder/elasticsearch-https.yaml
// testdata/observability.openshift.io_clusterlogforwarder/elasticsearch-mtls.yaml
// testdata/observability.openshift.io_clusterlogforwarder/elasticsearch-userauth-https.yaml
// testdata/observability.openshift.io_clusterlogforwarder/elasticsearch-userauth-mtls.yaml
// testdata/observability.openshift.io_clusterlogforwarder/elasticsearch-userauth.yaml
// testdata/observability.openshift.io_clusterlogforwarder/elasticsearch.yaml
// testdata/observability.openshift.io_clusterlogforwarder/google-cloud-logging-multi-logids.yaml
// testdata/observability.openshift.io_clusterlogforwarder/googleCloudLogging.yaml
// testdata/observability.openshift.io_clusterlogforwarder/http-output-85490.yaml
// testdata/observability.openshift.io_clusterlogforwarder/http-output.yaml
// testdata/observability.openshift.io_clusterlogforwarder/https-61567.yaml
// testdata/observability.openshift.io_clusterlogforwarder/https-output-ca.yaml
// testdata/observability.openshift.io_clusterlogforwarder/https-output-mtls.yaml
// testdata/observability.openshift.io_clusterlogforwarder/https-output-skipverify.yaml
// testdata/observability.openshift.io_clusterlogforwarder/httpserver-to-httpoutput.yaml
// testdata/observability.openshift.io_clusterlogforwarder/httpserver-to-splunk.yaml
// testdata/observability.openshift.io_clusterlogforwarder/loki.yaml
// testdata/observability.openshift.io_clusterlogforwarder/lokistack-with-labelkeys.yaml
// testdata/observability.openshift.io_clusterlogforwarder/lokistack.yaml
// testdata/observability.openshift.io_clusterlogforwarder/lokistack_gateway_https_secret.yaml
// testdata/observability.openshift.io_clusterlogforwarder/otlp-lokistack.yaml
// testdata/observability.openshift.io_clusterlogforwarder/otlp.yaml
// testdata/observability.openshift.io_clusterlogforwarder/rsyslog-mtls.yaml
// testdata/observability.openshift.io_clusterlogforwarder/rsyslog-serverAuth.yaml
// testdata/observability.openshift.io_clusterlogforwarder/splunk-mtls-passphrase.yaml
// testdata/observability.openshift.io_clusterlogforwarder/splunk-mtls.yaml
// testdata/observability.openshift.io_clusterlogforwarder/splunk-serveronly.yaml
// testdata/observability.openshift.io_clusterlogforwarder/splunk.yaml
// testdata/observability.openshift.io_clusterlogforwarder/syslog-75317.yaml
// testdata/observability.openshift.io_clusterlogforwarder/syslog-75431.yaml
// testdata/observability.openshift.io_clusterlogforwarder/syslog-81512.yaml
// testdata/observability.openshift.io_clusterlogforwarder/syslog-selected-ns.yaml
// testdata/observability.openshift.io_clusterlogforwarder/syslog.yaml
// testdata/odf/objectBucketClaim.yaml
// testdata/prometheus-k8s-rbac.yaml
// testdata/rapidast/customscan.policy
// testdata/rapidast/data_rapidastconfig_logging_v1.yaml
// testdata/rapidast/data_rapidastconfig_logging_v1alpha1.yaml
// testdata/rapidast/data_rapidastconfig_loki_v1.yaml
// testdata/rapidast/data_rapidastconfig_observability_v1.yaml
// testdata/rapidast/job_rapidast.yaml
// testdata/subscription/allnamespace-og.yaml
// testdata/subscription/catsrc.yaml
// testdata/subscription/namespace.yaml
// testdata/subscription/singlenamespace-og.yaml
// testdata/subscription/sub-template.yaml
package testdata

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type asset struct {
	bytes []byte
	info  os.FileInfo
}

type bindataFileInfo struct {
	name    string
	size    int64
	mode    os.FileMode
	modTime time.Time
}

// Name return file name
func (fi bindataFileInfo) Name() string {
	return fi.name
}

// Size return file size
func (fi bindataFileInfo) Size() int64 {
	return fi.size
}

// Mode return file mode
func (fi bindataFileInfo) Mode() os.FileMode {
	return fi.mode
}

// Mode return file modify time
func (fi bindataFileInfo) ModTime() time.Time {
	return fi.modTime
}

// IsDir return file whether a directory
func (fi bindataFileInfo) IsDir() bool {
	return fi.mode&os.ModeDir != 0
}

// Sys return file is sys mode
func (fi bindataFileInfo) Sys() interface{} {
	return nil
}

var _testdataUipluginUipluginYaml = []byte(`kind: Template
apiVersion: template.openshift.io/v1
metadata:
  name: subscription-template
objects:
- apiVersion: observability.openshift.io/v1alpha1
  kind: UIPlugin
  metadata:
    name: logging
  spec:
    logging:
      logsLimit: 50
      lokiStack:
        name: ${LOKISTACK_NAME}
    type: Logging
parameters:
  - name: LOKISTACK_NAME
    value: logging-loki
`)

func testdataUipluginUipluginYamlBytes() ([]byte, error) {
	return _testdataUipluginUipluginYaml, nil
}

func testdataUipluginUipluginYaml() (*asset, error) {
	bytes, err := testdataUipluginUipluginYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "testdata/UIPlugin/UIPlugin.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testdataEventrouterEventrouterYaml = []byte(`kind: Template
apiVersion: template.openshift.io/v1
metadata:
  name: eventrouter-template
  annotations:
    description: "A pod forwarding kubernetes events to OpenShift Logging stack."
    tags: "events,EFK,logging,cluster-logging"
objects:
  - kind: ServiceAccount
    apiVersion: v1
    metadata:
      name: ${NAME}
      namespace: ${NAMESPACE}
  - kind: ClusterRole
    apiVersion: v1
    metadata:
      name: ${NAME}-reader
    rules:
    - apiGroups: [""]
      resources: ["events"]
      verbs: ["get", "watch", "list"]
  - kind: ClusterRoleBinding
    apiVersion: v1
    metadata:
      name: ${NAME}-reader-binding
    subjects:
    - kind: ServiceAccount
      name: ${NAME}
      namespace: ${NAMESPACE}
    roleRef:
      kind: ClusterRole
      name: ${NAME}-reader
  - kind: ConfigMap
    apiVersion: v1
    metadata:
      name: ${NAME}
      namespace: ${NAMESPACE}
    data:
      config.json: |-
        {
          "sink": "stdout"
        }
  - kind: Deployment
    apiVersion: apps/v1
    metadata:
      name: ${NAME}
      namespace: ${NAMESPACE}
      labels:
        component: "eventrouter"
        logging-infra: "eventrouter"
        provider: "openshift"
    spec:
      selector:
        matchLabels:
          component: "eventrouter"
          logging-infra: "eventrouter"
          provider: "openshift"
      replicas: 1
      template:
        metadata:
          labels:
            component: "eventrouter"
            logging-infra: "eventrouter"
            provider: "openshift"
          name: ${NAME}
        spec:
          serviceAccount: ${NAME}
          containers:
            - name: kube-eventrouter
              image: ${IMAGE}
              imagePullPolicy: IfNotPresent
              resources:
                requests:
                  cpu: ${CPU}
                  memory: ${MEMORY}
              volumeMounts:
              - name: config-volume
                mountPath: /etc/eventrouter
          volumes:
            - name: config-volume
              configMap:
                name: ${NAME}
parameters:
  - name: IMAGE
    displayName: Image
    value: "brew.registry.redhat.io/rh-osbs/openshift-logging-eventrouter-rhel9:v0.4.0"
  - name: CPU
    displayName: CPU
    value: "100m"
  - name: MEMORY
    displayName: Memory
    value: "128Mi"
  - name: NAMESPACE
    displayName: Namespace
    value: "openshift-logging"
  - name: NAME
    value: eventrouter
    displayName: Event Router name
`)

func testdataEventrouterEventrouterYamlBytes() ([]byte, error) {
	return _testdataEventrouterEventrouterYaml, nil
}

func testdataEventrouterEventrouterYaml() (*asset, error) {
	bytes, err := testdataEventrouterEventrouterYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "testdata/eventrouter/eventrouter.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testdataExternalLogStoresCert_generationSh = []byte(`#! /bin/bash

WORKING_DIR=${1:-/tmp/_working_dir}
NAMESPACE=${2:-openshift-logging}
CA_PATH=${CA_PATH:-$WORKING_DIR/ca.crt}
LOG_STORE=${3:-elasticsearch}
BASE_DOMAIN=${4:-}
PASS_PHRASE=${5:-}
REGENERATE_NEEDED=0

function init_cert_files() {

  if [ ! -f ${WORKING_DIR}/ca.db ]; then
    touch ${WORKING_DIR}/ca.db
  fi

  if [ ! -f ${WORKING_DIR}/ca.serial.txt ]; then
    echo 00 >${WORKING_DIR}/ca.serial.txt
  fi
}

function generate_signing_ca() {
  if [ ! -f ${WORKING_DIR}/ca.crt ] || [ ! -f ${WORKING_DIR}/ca.key ] || ! openssl x509 -checkend 0 -noout -in ${WORKING_DIR}/ca.crt; then
    openssl req -x509 \
      -new \
      -newkey rsa:4096 \
      -keyout ${WORKING_DIR}/ca.key \
      -nodes \
      -days 1825 \
      -out ${WORKING_DIR}/ca.crt \
      -subj "/CN=openshift-cluster-logging-signer"

    REGENERATE_NEEDED=1
  fi
}

function create_signing_conf() {
  cat <<EOF >"${WORKING_DIR}/signing.conf"
# Simple Signing CA

# The [default] section contains global constants that can be referred to from
# the entire configuration file. It may also hold settings pertaining to more
# than one openssl command.

[ default ]
dir                     = ${WORKING_DIR}               # Top dir

# The next part of the configuration file is used by the openssl req command.
# It defines the CA's key pair, its DN, and the desired extensions for the CA
# certificate.

[ req ]
default_bits            = 4096                  # RSA key size
encrypt_key             = yes                   # Protect private key
default_md              = sha512                # MD to use
utf8                    = yes                   # Input is UTF-8
string_mask             = utf8only              # Emit UTF-8 strings
prompt                  = no                    # Don't prompt for DN
distinguished_name      = ca_dn                 # DN section
req_extensions          = ca_reqext             # Desired extensions

[ ca_dn ]
0.domainComponent       = "io"
1.domainComponent       = "openshift"
organizationName        = "OpenShift Origin"
organizationalUnitName  = "Logging Signing CA"
commonName              = "Logging Signing CA"

[ ca_reqext ]
keyUsage                = critical,keyCertSign,cRLSign
basicConstraints        = critical,CA:true,pathlen:0
subjectKeyIdentifier    = hash

# The remainder of the configuration file is used by the openssl ca command.
# The CA section defines the locations of CA assets, as well as the policies
# applying to the CA.

[ ca ]
default_ca              = signing_ca            # The default CA section

[ signing_ca ]
certificate             = \$dir/ca.crt       # The CA cert
private_key             = \$dir/ca.key # CA private key
new_certs_dir           = \$dir/           # Certificate archive
serial                  = \$dir/ca.serial.txt # Serial number file
crlnumber               = \$dir/ca.crl.srl # CRL number file
database                = \$dir/ca.db # Index file
unique_subject          = no                    # Require unique subject
default_days            = 730                   # How long to certify for
default_md              = sha512                # MD to use
policy                  = any_pol             # Default naming policy
email_in_dn             = no                    # Add email to cert DN
preserve                = no                    # Keep passed DN ordering
name_opt                = ca_default            # Subject DN display options
cert_opt                = ca_default            # Certificate display options
copy_extensions         = copy                  # Copy extensions from CSR
x509_extensions         = client_ext             # Default cert extensions
default_crl_days        = 7                     # How long before next CRL
crl_extensions          = crl_ext               # CRL extensions

# Naming policies control which parts of a DN end up in the certificate and
# under what circumstances certification should be denied.

[ match_pol ]
domainComponent         = match                 # Must match 'simple.org'
organizationName        = match                 # Must match 'Simple Inc'
organizationalUnitName  = optional              # Included if present
commonName              = supplied              # Must be present

[ any_pol ]
domainComponent         = optional
countryName             = optional
stateOrProvinceName     = optional
localityName            = optional
organizationName        = optional
organizationalUnitName  = optional
commonName              = optional
emailAddress            = optional

# Certificate extensions define what types of certificates the CA is able to
# create.

[ client_ext ]
keyUsage                = critical,digitalSignature,keyEncipherment
basicConstraints        = CA:false
extendedKeyUsage        = clientAuth
subjectKeyIdentifier    = hash
authorityKeyIdentifier  = keyid

[ server_ext ]
keyUsage                = critical,digitalSignature,keyEncipherment
basicConstraints        = CA:false
extendedKeyUsage        = serverAuth,clientAuth
subjectKeyIdentifier    = hash
authorityKeyIdentifier  = keyid

# CRL extensions exist solely to point to the CA certificate that has issued
# the CRL.

[ crl_ext ]
authorityKeyIdentifier  = keyid
EOF
}

function sign_cert() {
  local component=$1

  openssl ca \
    -in ${WORKING_DIR}/${component}.csr \
    -notext \
    -out ${WORKING_DIR}/${component}.crt \
    -config ${WORKING_DIR}/signing.conf \
    -extensions v3_req \
    -batch \
    -extensions server_ext
}

function generate_cert_config() {
  local component=$1
  local extensions=${2:-}

  if [ "$extensions" != "" ]; then
    cat <<EOF >"${WORKING_DIR}/${component}.conf"
[ req ]
default_bits = 4096
prompt = no
encrypt_key = yes
default_md = sha512
distinguished_name = dn
req_extensions = req_ext
[ dn ]
CN = ${component}
OU = OpenShift
O = Logging
[ req_ext ]
subjectAltName = ${extensions}
EOF
  else
    cat <<EOF >"${WORKING_DIR}/${component}.conf"
[ req ]
default_bits = 4096
prompt = no
encrypt_key = yes
default_md = sha512
distinguished_name = dn
[ dn ]
CN = ${component}
OU = OpenShift
O = Logging
EOF
  fi
}

function generate_request() {
  local component=$1

  if [[ "$component" == "server" ]] || [[ -z "$PASS_PHRASE" ]]; then
    openssl req -new \
      -out ${WORKING_DIR}/${component}.csr \
      -newkey rsa:4096 \
      -keyout ${WORKING_DIR}/${component}.key \
      -config ${WORKING_DIR}/${component}.conf \
      -days 712 \
      -nodes
  else
    openssl req -new \
      -passout pass:"$PASS_PHRASE" \
      -out ${WORKING_DIR}/${component}.csr \
      -newkey rsa:4096 \
      -keyout ${WORKING_DIR}/${component}.key.pem \
      -config ${WORKING_DIR}/${component}.conf \
      -days 712
    # use pkcs8 for client key to avoid htting issue in FIPS cluster
    openssl pkcs8 -passin pass:"$PASS_PHRASE" -in ${WORKING_DIR}/${component}.key.pem -topk8 -nocrypt -passout pass:"$PASS_PHRASE" -out ${WORKING_DIR}/${component}.key
  fi

}

function generate_certs() {
  local component=$1
  local extensions=${2:-}

  if [ $REGENERATE_NEEDED = 1 ] || [ ! -f ${WORKING_DIR}/${component}.crt ] || ! openssl x509 -checkend 0 -noout -in ${WORKING_DIR}/${component}.crt; then
    generate_cert_config $component $extensions
    generate_request $component
    sign_cert $component
  fi
}

function generate_extensions() {
  local add_oid=$1
  local add_localhost=$2
  shift
  shift
  local cert_names=$@

  extension_names=""
  extension_index=1
  local use_comma=0

  if [ "$add_localhost" == "true" ]; then
    extension_names="IP.1:127.0.0.1,DNS.1:localhost"
    extension_index=2
    use_comma=1
  fi

  for name in ${cert_names//,/}; do
    if [ $use_comma = 1 ]; then
      extension_names="${extension_names},DNS.${extension_index}:${name}"
    else
      extension_names="DNS.${extension_index}:${name}"
      use_comma=1
    fi
    extension_index=$((extension_index + 1))
  done

  if [ "$add_oid" == "true" ]; then
    extension_names="${extension_names},RID.1:1.2.3.4.5.5"
  fi

  if [ ! -z "$BASE_DOMAIN" ]; then
    extension_names="${extension_names},DNS.${extension_index}:${LOG_STORE}-${NAMESPACE}.${BASE_DOMAIN}"
  fi

  echo "$extension_names"
}

if [ ! -d $WORKING_DIR ]; then
  mkdir -p $WORKING_DIR
fi

generate_signing_ca
init_cert_files
create_signing_conf

generate_certs 'server' "$(generate_extensions false true $LOG_STORE{,-cluster}{,.${NAMESPACE}.svc}{,.cluster.local})"
generate_certs 'client' "$(generate_extensions false false $LOG_STORE{,.${NAMESPACE}.svc}{,.cluster.local})"
`)

func testdataExternalLogStoresCert_generationShBytes() ([]byte, error) {
	return _testdataExternalLogStoresCert_generationSh, nil
}

func testdataExternalLogStoresCert_generationSh() (*asset, error) {
	bytes, err := testdataExternalLogStoresCert_generationShBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "testdata/external-log-stores/cert_generation.sh", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testdataExternalLogStoresElasticsearch6HttpNo_userConfigmapYaml = []byte(`kind: Template
apiVersion: template.openshift.io/v1
metadata:
  name: external-log-store-template
objects:
- apiVersion: v1
  data:
    elasticsearch.yml: |
      node.name:  ${NAME}
      cluster.name: ${NAME}
      discovery.zen.minimum_master_nodes: 1
      network.host: 0.0.0.0
      http.port: 9200
      http.host: 0.0.0.0
      transport.host: 127.0.0.1
      discovery.type: single-node
      xpack.security.enabled: false
      xpack.security.authc.api_key.enabled: false
      xpack.monitoring.enabled : false
      xpack.license.self_generated.type: basic
      xpack.security.http.ssl.enabled: false
      xpack.ml.enabled: ${MACHINE_LEARNING}
  kind: ConfigMap
  metadata:
    name: ${NAME}
    namespace: ${NAMESPACE}
parameters:
- name: NAME
  value: elasticsearch-server
- name: NAMESPACE
  value: openshift-logging
- name: MACHINE_LEARNING
  value: "true"
`)

func testdataExternalLogStoresElasticsearch6HttpNo_userConfigmapYamlBytes() ([]byte, error) {
	return _testdataExternalLogStoresElasticsearch6HttpNo_userConfigmapYaml, nil
}

func testdataExternalLogStoresElasticsearch6HttpNo_userConfigmapYaml() (*asset, error) {
	bytes, err := testdataExternalLogStoresElasticsearch6HttpNo_userConfigmapYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "testdata/external-log-stores/elasticsearch/6/http/no_user/configmap.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testdataExternalLogStoresElasticsearch6HttpNo_userDeploymentYaml = []byte(`kind: Template
apiVersion: template.openshift.io/v1
metadata:
  name: external-log-store-template
objects:
- apiVersion: apps/v1
  kind: Deployment
  metadata:
    labels:
      app: ${NAME}
    name: ${NAME}
    namespace: ${NAMESPACE}
  spec:
    replicas: 1
    revisionHistoryLimit: 10
    selector:
      matchLabels:
        app: ${NAME}
    strategy:
      activeDeadlineSeconds: 21600
      resources: {}
      rollingParams:
        intervalSeconds: 1
        maxSurge: 25%
        maxUnavailable: 25%
        timeoutSeconds: 600
        updatePeriodSeconds: 1
      type: Recreate
    template:
      metadata:
        labels:
          app: ${NAME}
      spec:
        securityContext:
          runAsNonRoot: true
          seccompProfile:
            type: RuntimeDefault
        containers:
        - image: quay.io/openshifttest/elasticsearch@sha256:206dea14c8a2c8a4d408808a08e2b4dc932218b45aae6147ba000fa08cc7251a
          imagePullPolicy: IfNotPresent
          name: ${NAME}
          securityContext:
            allowPrivilegeEscalation: false
            runAsNonRoot: true
            capabilities:
              drop:
              - ALL
            seccompProfile:
              type: RuntimeDefault
          ports:
          - containerPort: 9300
            protocol: TCP
          - containerPort: 9200
            protocol: TCP
          volumeMounts:
          - mountPath: /usr/share/elasticsearch/config/elasticsearch.yml
            subPath: elasticsearch.yml
            name: elasticsearch-config
          resources: {}
          terminationMessagePath: /dev/termination-log
          terminationMessagePolicy: File
        volumes:
        - configMap:
            defaultMode: 420
            name: ${NAME}
          name: elasticsearch-config
        serviceAccount: ${NAME}
        serviceAccountName: ${NAME}
        dnsPolicy: ClusterFirst
        restartPolicy: Always
parameters:
- name: NAME
  value: elasticsearch-server
- name: NAMESPACE
  value: openshift-logging
`)

func testdataExternalLogStoresElasticsearch6HttpNo_userDeploymentYamlBytes() ([]byte, error) {
	return _testdataExternalLogStoresElasticsearch6HttpNo_userDeploymentYaml, nil
}

func testdataExternalLogStoresElasticsearch6HttpNo_userDeploymentYaml() (*asset, error) {
	bytes, err := testdataExternalLogStoresElasticsearch6HttpNo_userDeploymentYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "testdata/external-log-stores/elasticsearch/6/http/no_user/deployment.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testdataExternalLogStoresElasticsearch6HttpUser_authConfigmapYaml = []byte(`kind: Template
apiVersion: template.openshift.io/v1
metadata:
  name: es6-cm-template
objects:
- apiVersion: v1
  data:
    add-user.sh: |+
      output = $(/usr/share/elasticsearch/bin/elasticsearch-users useradd ${USERNAME} -p ${PASSWORD} -r superuser)
      if [[ $output =~ 'already exists' ]]
      then
      return 0
      fi

    elasticsearch.yml: |
      node.name:  ${NAME}
      cluster.name: ${NAME}
      discovery.zen.minimum_master_nodes: 1
      network.host: 0.0.0.0
      http.port: 9200
      http.host: 0.0.0.0
      transport.host: 127.0.0.1
      discovery.type: single-node
      xpack.security.authc.realms:
        native:
          type: file
          order: 0
          enabled: true
          authentication.enabled: true
      xpack.security.enabled: true
      xpack.security.authc.api_key.enabled: true
      xpack.monitoring.enabled : false
      xpack.license.self_generated.type: basic
      xpack.security.http.ssl.enabled: false
      xpack.ml.enabled: ${MACHINE_LEARNING}
  kind: ConfigMap
  metadata:
    name: ${NAME}
    namespace: ${NAMESPACE}
parameters:
- name: USERNAME
  value: "fluentd"
- name: PASSWORD
  value: "redhat"
- name: NAME
  value: elasticsearch-server
- name: NAMESPACE
  value: openshift-logging
- name: MACHINE_LEARNING
  value: "true"
`)

func testdataExternalLogStoresElasticsearch6HttpUser_authConfigmapYamlBytes() ([]byte, error) {
	return _testdataExternalLogStoresElasticsearch6HttpUser_authConfigmapYaml, nil
}

func testdataExternalLogStoresElasticsearch6HttpUser_authConfigmapYaml() (*asset, error) {
	bytes, err := testdataExternalLogStoresElasticsearch6HttpUser_authConfigmapYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "testdata/external-log-stores/elasticsearch/6/http/user_auth/configmap.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testdataExternalLogStoresElasticsearch6HttpUser_authDeploymentYaml = []byte(`kind: Template
apiVersion: template.openshift.io/v1
metadata:
  name: external-log-store-template
objects:
- apiVersion: apps/v1
  kind: Deployment
  metadata:
    labels:
      app: ${NAME}
    name: ${NAME}
    namespace: ${NAMESPACE}
  spec:
    replicas: 1
    revisionHistoryLimit: 10
    selector:
      matchLabels:
        app: ${NAME}
    strategy:
      activeDeadlineSeconds: 21600
      resources: {}
      rollingParams:
        intervalSeconds: 1
        maxSurge: 25%
        maxUnavailable: 25%
        timeoutSeconds: 600
        updatePeriodSeconds: 1
      type: Recreate
    template:
      metadata:
        labels:
          app: ${NAME}
      spec:
        securityContext:
          runAsNonRoot: true
          seccompProfile:
            type: RuntimeDefault
        containers:
        - image: quay.io/openshifttest/elasticsearch@sha256:206dea14c8a2c8a4d408808a08e2b4dc932218b45aae6147ba000fa08cc7251a
          imagePullPolicy: IfNotPresent
          name: ${NAME}
          securityContext:
            allowPrivilegeEscalation: false
            runAsNonRoot: true
            capabilities:
              drop:
              - ALL
            seccompProfile:
              type: RuntimeDefault
          ports:
          - containerPort: 9300
            protocol: TCP
          - containerPort: 9200
            protocol: TCP
          readinessProbe:
            exec:
              command:
              - sh
              - /usr/share/elasticsearch/add-user.sh
            failureThreshold: 3
            initialDelaySeconds: 5
            periodSeconds: 5
            successThreshold: 1
            timeoutSeconds: 10
          resources: {}
          terminationMessagePath: /dev/termination-log
          terminationMessagePolicy: File
          volumeMounts:
          - mountPath: /usr/share/elasticsearch/config/elasticsearch.yml
            name: elasticsearch-config
            subPath: elasticsearch.yml
          - mountPath: /usr/share/elasticsearch/add-user.sh
            name: elasticsearch-config
            subPath: add-user.sh
        dnsPolicy: ClusterFirst
        restartPolicy: Always
        schedulerName: default-scheduler
        securityContext: {}
        serviceAccount: ${NAME}
        serviceAccountName: ${NAME}
        terminationGracePeriodSeconds: 30
        volumes:
        - configMap:
            defaultMode: 420
            name: ${NAME}
          name: elasticsearch-config
parameters:
- name: NAME
  value: elasticsearch-server
- name: NAMESPACE
  value: openshift-logging
`)

func testdataExternalLogStoresElasticsearch6HttpUser_authDeploymentYamlBytes() ([]byte, error) {
	return _testdataExternalLogStoresElasticsearch6HttpUser_authDeploymentYaml, nil
}

func testdataExternalLogStoresElasticsearch6HttpUser_authDeploymentYaml() (*asset, error) {
	bytes, err := testdataExternalLogStoresElasticsearch6HttpUser_authDeploymentYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "testdata/external-log-stores/elasticsearch/6/http/user_auth/deployment.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testdataExternalLogStoresElasticsearch6HttpsNo_userConfigmapYaml = []byte(`kind: Template
apiVersion: template.openshift.io/v1
metadata:
  name: external-log-store-template
objects:
- apiVersion: v1
  data:
    elasticsearch.yml: |
      node.name: ${NAME}
      cluster.name: ${NAME}
      discovery.zen.minimum_master_nodes: 1
      network.host: 0.0.0.0
      http.port: 9200
      http.host: 0.0.0.0
      transport.host: 127.0.0.1
      discovery.type: single-node
      xpack.security.enabled: true
      xpack.security.authc:
        anonymous:
          username: anonymous_user
          roles: superuser
          authz_exception: true
      xpack.security.authc.api_key.enabled: true
      xpack.monitoring.enabled : false
      xpack.license.self_generated.type: basic
      xpack.security.http.ssl.enabled: true
      xpack.security.http.ssl.key:  /usr/share/elasticsearch/config/secret/elasticsearch.key
      xpack.security.http.ssl.certificate: /usr/share/elasticsearch/config/secret/elasticsearch.crt
      xpack.security.http.ssl.certificate_authorities: [ "/usr/share/elasticsearch/config/secret/admin-ca" ]
      xpack.security.http.ssl.verification_mode: full
      xpack.security.http.ssl.client_authentication: ${CLIENT_AUTH}
      xpack.ml.enabled: ${MACHINE_LEARNING}
  kind: ConfigMap
  metadata:
    name: ${NAME}
    namespace: ${NAMESPACE}
parameters:
- name: NAME
  value: elasticsearch-server
- name: NAMESPACE
  value: openshift-logging
- name: CLIENT_AUTH
  value: none
- name: MACHINE_LEARNING
  value: "true"
`)

func testdataExternalLogStoresElasticsearch6HttpsNo_userConfigmapYamlBytes() ([]byte, error) {
	return _testdataExternalLogStoresElasticsearch6HttpsNo_userConfigmapYaml, nil
}

func testdataExternalLogStoresElasticsearch6HttpsNo_userConfigmapYaml() (*asset, error) {
	bytes, err := testdataExternalLogStoresElasticsearch6HttpsNo_userConfigmapYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "testdata/external-log-stores/elasticsearch/6/https/no_user/configmap.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testdataExternalLogStoresElasticsearch6HttpsNo_userDeploymentYaml = []byte(`kind: Template
apiVersion: template.openshift.io/v1
metadata:
  name: external-log-store-template
objects:
- apiVersion: apps/v1
  kind: Deployment
  metadata:
    labels:
      app: ${NAME}
    name: ${NAME}
    namespace: ${NAMESPACE}
  spec:
    progressDeadlineSeconds: 600
    replicas: 1
    revisionHistoryLimit: 10
    selector:
      matchLabels:
        app: ${NAME}
    strategy:
      type: Recreate
    template:
      metadata:
        creationTimestamp: null
        labels:
          app: ${NAME}
      spec:
        securityContext:
          runAsNonRoot: true
          seccompProfile:
            type: RuntimeDefault
        containers:
        - image: quay.io/openshifttest/elasticsearch@sha256:206dea14c8a2c8a4d408808a08e2b4dc932218b45aae6147ba000fa08cc7251a
          imagePullPolicy: IfNotPresent
          name: ${NAME}
          securityContext:
            allowPrivilegeEscalation: false
            runAsNonRoot: true
            capabilities:
              drop:
              - ALL
            seccompProfile:
              type: RuntimeDefault
          ports:
          - containerPort: 9300
            protocol: TCP
          - containerPort: 9200
            protocol: TCP
          resources: {}
          terminationMessagePath: /dev/termination-log
          terminationMessagePolicy: File
          volumeMounts:
          - mountPath: /usr/share/elasticsearch/config/elasticsearch.yml
            name: elasticsearch-config
            subPath: elasticsearch.yml
          - mountPath: /usr/share/elasticsearch/config/secret
            name: certificates
        dnsPolicy: ClusterFirst
        restartPolicy: Always
        schedulerName: default-scheduler
        securityContext: {}
        serviceAccount: ${NAME}
        serviceAccountName: ${NAME}
        terminationGracePeriodSeconds: 30
        volumes:
        - configMap:
            defaultMode: 420
            name: ${NAME}
          name: elasticsearch-config
        - name: certificates
          secret:
            defaultMode: 420
            secretName: ${NAME}
parameters:
- name: NAME
  value: elasticsearch-server
- name: NAMESPACE
  value: openshift-logging
`)

func testdataExternalLogStoresElasticsearch6HttpsNo_userDeploymentYamlBytes() ([]byte, error) {
	return _testdataExternalLogStoresElasticsearch6HttpsNo_userDeploymentYaml, nil
}

func testdataExternalLogStoresElasticsearch6HttpsNo_userDeploymentYaml() (*asset, error) {
	bytes, err := testdataExternalLogStoresElasticsearch6HttpsNo_userDeploymentYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "testdata/external-log-stores/elasticsearch/6/https/no_user/deployment.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testdataExternalLogStoresElasticsearch6HttpsUser_authConfigmapYaml = []byte(`kind: Template
apiVersion: template.openshift.io/v1
metadata:
  name: external-log-store-template
objects:
- apiVersion: v1
  data:
    add-user.sh: |+
      output = $(/usr/share/elasticsearch/bin/elasticsearch-users useradd ${USERNAME} -p ${PASSWORD} -r superuser)
      if [[ $output =~ 'already exists' ]]
      then
      return 0
      fi

    elasticsearch.yml: |
      node.name:  ${NAME}
      cluster.name: ${NAME}
      discovery.zen.minimum_master_nodes: 1
      network.host: 0.0.0.0
      http.port: 9200
      http.host: 0.0.0.0
      transport.host: 127.0.0.1
      discovery.type: single-node
      xpack.security.authc:
        realms:
          native:
            type: file
            order: 0
            enabled: true
            authentication.enabled: true
      xpack.security.enabled: true
      xpack.security.authc.api_key.enabled: true
      xpack.monitoring.enabled : false
      xpack.license.self_generated.type: basic
      xpack.security.http.ssl.enabled: true
      xpack.security.http.ssl.key:  /usr/share/elasticsearch/config/secret/elasticsearch.key
      xpack.security.http.ssl.certificate: /usr/share/elasticsearch/config/secret/elasticsearch.crt
      xpack.security.http.ssl.certificate_authorities: [ "/usr/share/elasticsearch/config/secret/admin-ca" ]
      xpack.security.http.ssl.verification_mode: full
      xpack.security.http.ssl.client_authentication: ${CLIENT_AUTH}
      xpack.ml.enabled: ${MACHINE_LEARNING}
  kind: ConfigMap
  metadata:
    name: ${NAME}
    namespace: ${NAMESPACE}
parameters:
- name: USERNAME
  value: "fluentd"
- name: PASSWORD
  value: "redhat"
- name: NAME
  value: elasticsearch-server
- name: NAMESPACE
  value: openshift-logging
- name: CLIENT_AUTH
  value: none
- name: MACHINE_LEARNING
  value: "true"
`)

func testdataExternalLogStoresElasticsearch6HttpsUser_authConfigmapYamlBytes() ([]byte, error) {
	return _testdataExternalLogStoresElasticsearch6HttpsUser_authConfigmapYaml, nil
}

func testdataExternalLogStoresElasticsearch6HttpsUser_authConfigmapYaml() (*asset, error) {
	bytes, err := testdataExternalLogStoresElasticsearch6HttpsUser_authConfigmapYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "testdata/external-log-stores/elasticsearch/6/https/user_auth/configmap.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testdataExternalLogStoresElasticsearch6HttpsUser_authDeploymentYaml = []byte(`kind: Template
apiVersion: template.openshift.io/v1
metadata:
  name: external-log-store-template
objects:
- apiVersion: apps/v1
  kind: Deployment
  metadata:
    labels:
      app: ${NAME}
    name: ${NAME}
    namespace: ${NAMESPACE}
  spec:
    replicas: 1
    revisionHistoryLimit: 10
    selector:
      matchLabels:
        app: ${NAME}
    strategy:
      activeDeadlineSeconds: 21600
      resources: {}
      rollingParams:
        intervalSeconds: 1
        maxSurge: 25%
        maxUnavailable: 25%
        timeoutSeconds: 600
        updatePeriodSeconds: 1
      type: Recreate
    template:
      metadata:
        labels:
          app: ${NAME}
      spec:
        securityContext:
          runAsNonRoot: true
          seccompProfile:
            type: RuntimeDefault
        containers:
        - image: quay.io/openshifttest/elasticsearch@sha256:206dea14c8a2c8a4d408808a08e2b4dc932218b45aae6147ba000fa08cc7251a
          imagePullPolicy: IfNotPresent
          name: ${NAME}
          securityContext:
            allowPrivilegeEscalation: false
            runAsNonRoot: true
            capabilities:
              drop:
              - ALL
            seccompProfile:
              type: RuntimeDefault
          ports:
          - containerPort: 9300
            protocol: TCP
          - containerPort: 9200
            protocol: TCP
          readinessProbe:
            exec:
              command:
              - sh
              - /usr/share/elasticsearch/add-user.sh
            failureThreshold: 3
            initialDelaySeconds: 5
            periodSeconds: 5
            successThreshold: 1
            timeoutSeconds: 10
          resources: {}
          terminationMessagePath: /dev/termination-log
          terminationMessagePolicy: File
          volumeMounts:
          - mountPath: /usr/share/elasticsearch/config/elasticsearch.yml
            name: elasticsearch-config
            subPath: elasticsearch.yml
          - mountPath: /usr/share/elasticsearch/add-user.sh
            name: elasticsearch-config
            subPath: add-user.sh
          - mountPath: /usr/share/elasticsearch/config/secret
            name: certificates
        dnsPolicy: ClusterFirst
        restartPolicy: Always
        schedulerName: default-scheduler
        securityContext: {}
        serviceAccount: ${NAME}
        serviceAccountName: ${NAME}
        terminationGracePeriodSeconds: 30
        volumes:
        - configMap:
            defaultMode: 420
            name: ${NAME}
          name: elasticsearch-config
        - name: certificates
          secret:
            defaultMode: 420
            secretName: ${NAME}
parameters:
- name: NAME
  value: elasticsearch-server
- name: NAMESPACE
  value: openshift-logging
`)

func testdataExternalLogStoresElasticsearch6HttpsUser_authDeploymentYamlBytes() ([]byte, error) {
	return _testdataExternalLogStoresElasticsearch6HttpsUser_authDeploymentYaml, nil
}

func testdataExternalLogStoresElasticsearch6HttpsUser_authDeploymentYaml() (*asset, error) {
	bytes, err := testdataExternalLogStoresElasticsearch6HttpsUser_authDeploymentYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "testdata/external-log-stores/elasticsearch/6/https/user_auth/deployment.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testdataExternalLogStoresElasticsearch7HttpNo_userConfigmapYaml = []byte(`kind: Template
apiVersion: template.openshift.io/v1
metadata:
  name: external-log-store-template
objects:
- apiVersion: v1
  data:
    elasticsearch.yml: |
      node.name:  ${NAME}
      cluster.name: ${NAME}
      discovery.zen.minimum_master_nodes: 1
      network.host: 0.0.0.0
      http.port: 9200
      http.host: 0.0.0.0
      transport.host: 127.0.0.1
      discovery.type: single-node
      xpack.ml.enabled: ${MACHINE_LEARNING}
      xpack.security.enabled: false
      xpack.security.authc.api_key.enabled: false
      xpack.monitoring.enabled : false
      xpack.license.self_generated.type: basic
      xpack.security.http.ssl.enabled: false
  kind: ConfigMap
  metadata:
    name: ${NAME}
    namespace: ${NAMESPACE}
parameters:
- name: NAME
  value: elasticsearch-server
- name: NAMESPACE
  value: openshift-logging
- name: MACHINE_LEARNING
  value: "true"
`)

func testdataExternalLogStoresElasticsearch7HttpNo_userConfigmapYamlBytes() ([]byte, error) {
	return _testdataExternalLogStoresElasticsearch7HttpNo_userConfigmapYaml, nil
}

func testdataExternalLogStoresElasticsearch7HttpNo_userConfigmapYaml() (*asset, error) {
	bytes, err := testdataExternalLogStoresElasticsearch7HttpNo_userConfigmapYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "testdata/external-log-stores/elasticsearch/7/http/no_user/configmap.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testdataExternalLogStoresElasticsearch7HttpNo_userDeploymentYaml = []byte(`kind: Template
apiVersion: template.openshift.io/v1
metadata:
  name: external-log-store-template
objects:
- apiVersion: apps/v1
  kind: Deployment
  metadata:
    labels:
      app: ${NAME}
    name: ${NAME}
    namespace: ${NAMESPACE}
  spec:
    replicas: 1
    revisionHistoryLimit: 10
    selector:
      matchLabels:
        app: ${NAME}
    strategy:
      activeDeadlineSeconds: 21600
      resources: {}
      rollingParams:
        intervalSeconds: 1
        maxSurge: 25%
        maxUnavailable: 25%
        timeoutSeconds: 600
        updatePeriodSeconds: 1
      type: Recreate
    template:
      metadata:
        labels:
          app: ${NAME}
      spec:
        securityContext:
          runAsNonRoot: true
          seccompProfile:
            type: RuntimeDefault
        containers:
        - image: quay.io/openshifttest/elasticsearch@sha256:420ab335838747b1d350ed39f4d88cf075479fc915cda8f91373c7e00de65887
          imagePullPolicy: IfNotPresent
          name: ${NAME}
          securityContext:
            allowPrivilegeEscalation: false
            runAsNonRoot: true
            capabilities:
              drop:
              - ALL
            seccompProfile:
              type: RuntimeDefault
          ports:
          - containerPort: 9300
            protocol: TCP
          - containerPort: 9200
            protocol: TCP
          volumeMounts:
          - mountPath: /usr/share/elasticsearch/config/elasticsearch.yml
            subPath: elasticsearch.yml
            name: elasticsearch-config
          resources:
            requests:
              cpu: 1
              memory: 2Gi
          terminationMessagePath: /dev/termination-log
          terminationMessagePolicy: File
        volumes:
        - configMap:
            defaultMode: 420
            name: ${NAME}
          name: elasticsearch-config
        serviceAccount: ${NAME}
        serviceAccountName: ${NAME}
        dnsPolicy: ClusterFirst
        restartPolicy: Always
parameters:
- name: NAME
  value: elasticsearch-server
- name: NAMESPACE
  value: openshift-logging
`)

func testdataExternalLogStoresElasticsearch7HttpNo_userDeploymentYamlBytes() ([]byte, error) {
	return _testdataExternalLogStoresElasticsearch7HttpNo_userDeploymentYaml, nil
}

func testdataExternalLogStoresElasticsearch7HttpNo_userDeploymentYaml() (*asset, error) {
	bytes, err := testdataExternalLogStoresElasticsearch7HttpNo_userDeploymentYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "testdata/external-log-stores/elasticsearch/7/http/no_user/deployment.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testdataExternalLogStoresElasticsearch7HttpUser_authConfigmapYaml = []byte(`kind: Template
apiVersion: template.openshift.io/v1
metadata:
  name: external-log-store-template
objects:
- apiVersion: v1
  data:
    add-user.sh: |+
      output = $(/usr/share/elasticsearch/bin/elasticsearch-users useradd ${USERNAME} -p ${PASSWORD} -r superuser)
      if [[ $output =~ 'already exists' ]]
      then
      return 0
      fi

    elasticsearch.yml: |
      node.name:  ${NAME}
      cluster.name: ${NAME}
      discovery.zen.minimum_master_nodes: 1
      network.host: 0.0.0.0
      http.port: 9200
      http.host: 0.0.0.0
      transport.host: 127.0.0.1
      discovery.type: single-node
      xpack.security.authc.realms:
        file:
          test:
            order: 0
            enabled: true
            authentication.enabled: true
      xpack.security.enabled: true
      xpack.ml.enabled: ${MACHINE_LEARNING}
      xpack.security.authc.api_key.enabled: true
      xpack.monitoring.enabled : false
      xpack.license.self_generated.type: basic
      xpack.security.http.ssl.enabled: false
  kind: ConfigMap
  metadata:
    name: ${NAME}
    namespace: ${NAMESPACE}
parameters:
- name: NAME
  value: elasticsearch-server
- name: NAMESPACE
  value: openshift-logging
- name: USERNAME
  value: "fluentd"
- name: PASSWORD
  value: "redhat"
- name: MACHINE_LEARNING
  value: "true"
`)

func testdataExternalLogStoresElasticsearch7HttpUser_authConfigmapYamlBytes() ([]byte, error) {
	return _testdataExternalLogStoresElasticsearch7HttpUser_authConfigmapYaml, nil
}

func testdataExternalLogStoresElasticsearch7HttpUser_authConfigmapYaml() (*asset, error) {
	bytes, err := testdataExternalLogStoresElasticsearch7HttpUser_authConfigmapYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "testdata/external-log-stores/elasticsearch/7/http/user_auth/configmap.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testdataExternalLogStoresElasticsearch7HttpUser_authDeploymentYaml = []byte(`kind: Template
apiVersion: template.openshift.io/v1
metadata:
  name: external-log-store-template
objects:
- apiVersion: apps/v1
  kind: Deployment
  metadata:
    labels:
      app: ${NAME}
    name: ${NAME}
    namespace: ${NAMESPACE}
  spec:
    replicas: 1
    revisionHistoryLimit: 10
    selector:
      matchLabels:
        app: ${NAME}
    strategy:
      activeDeadlineSeconds: 21600
      resources: {}
      rollingParams:
        intervalSeconds: 1
        maxSurge: 25%
        maxUnavailable: 25%
        timeoutSeconds: 600
        updatePeriodSeconds: 1
      type: Recreate
    template:
      metadata:
        labels:
          app: ${NAME}
      spec:
        securityContext:
          runAsNonRoot: true
          seccompProfile:
            type: RuntimeDefault
        containers:
        - image: quay.io/openshifttest/elasticsearch@sha256:420ab335838747b1d350ed39f4d88cf075479fc915cda8f91373c7e00de65887
          imagePullPolicy: IfNotPresent
          name: ${NAME}
          securityContext:
            allowPrivilegeEscalation: false
            runAsNonRoot: true
            capabilities:
              drop:
              - ALL
            seccompProfile:
              type: RuntimeDefault
          ports:
          - containerPort: 9300
            protocol: TCP
          - containerPort: 9200
            protocol: TCP
          readinessProbe:
            exec:
              command:
              - sh
              - /usr/share/elasticsearch/add-user.sh
            failureThreshold: 3
            initialDelaySeconds: 5
            periodSeconds: 5
            successThreshold: 1
            timeoutSeconds: 10
          resources:
            requests:
              cpu: 1
              memory: 2Gi
          terminationMessagePath: /dev/termination-log
          terminationMessagePolicy: File
          volumeMounts:
          - mountPath: /usr/share/elasticsearch/config/elasticsearch.yml
            name: elasticsearch-config
            subPath: elasticsearch.yml
          - mountPath: /usr/share/elasticsearch/add-user.sh
            name: elasticsearch-config
            subPath: add-user.sh
        dnsPolicy: ClusterFirst
        restartPolicy: Always
        schedulerName: default-scheduler
        securityContext: {}
        serviceAccount: ${NAME}
        serviceAccountName: ${NAME}
        terminationGracePeriodSeconds: 30
        volumes:
        - configMap:
            defaultMode: 420
            name: ${NAME}
          name: elasticsearch-config
parameters:
- name: NAME
  value: elasticsearch-server
- name: NAMESPACE
  value: openshift-logging
`)

func testdataExternalLogStoresElasticsearch7HttpUser_authDeploymentYamlBytes() ([]byte, error) {
	return _testdataExternalLogStoresElasticsearch7HttpUser_authDeploymentYaml, nil
}

func testdataExternalLogStoresElasticsearch7HttpUser_authDeploymentYaml() (*asset, error) {
	bytes, err := testdataExternalLogStoresElasticsearch7HttpUser_authDeploymentYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "testdata/external-log-stores/elasticsearch/7/http/user_auth/deployment.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testdataExternalLogStoresElasticsearch7HttpsNo_userConfigmapYaml = []byte(`kind: Template
apiVersion: template.openshift.io/v1
metadata:
  name: external-log-store-template
objects:
- apiVersion: v1
  data:
    elasticsearch.yml: |
      node.name: ${NAME}
      cluster.name: ${NAME}
      discovery.zen.minimum_master_nodes: 1
      network.host: 0.0.0.0
      http.port: 9200
      http.host: 0.0.0.0
      transport.host: 127.0.0.1
      discovery.type: single-node
      xpack.security.enabled: true
      xpack.security.authc:
        anonymous:
          username: anonymous_user
          roles: superuser
          authz_exception: true
      xpack.security.authc.api_key.enabled: true
      xpack.monitoring.enabled : false
      xpack.license.self_generated.type: basic
      xpack.security.http.ssl.enabled: true
      xpack.ml.enabled: ${MACHINE_LEARNING}
      xpack.security.http.ssl.key:  /usr/share/elasticsearch/config/secret/elasticsearch.key
      xpack.security.http.ssl.certificate: /usr/share/elasticsearch/config/secret/elasticsearch.crt
      xpack.security.http.ssl.certificate_authorities: [ "/usr/share/elasticsearch/config/secret/admin-ca" ]
      xpack.security.http.ssl.verification_mode: full
      xpack.security.http.ssl.client_authentication: ${CLIENT_AUTH}
  kind: ConfigMap
  metadata:
    name: ${NAME}
    namespace: ${NAMESPACE}
parameters:
- name: NAME
  value: elasticsearch-server
- name: NAMESPACE
  value: openshift-logging
- name: CLIENT_AUTH
  value: required
- name: MACHINE_LEARNING
  value: "true"
`)

func testdataExternalLogStoresElasticsearch7HttpsNo_userConfigmapYamlBytes() ([]byte, error) {
	return _testdataExternalLogStoresElasticsearch7HttpsNo_userConfigmapYaml, nil
}

func testdataExternalLogStoresElasticsearch7HttpsNo_userConfigmapYaml() (*asset, error) {
	bytes, err := testdataExternalLogStoresElasticsearch7HttpsNo_userConfigmapYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "testdata/external-log-stores/elasticsearch/7/https/no_user/configmap.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testdataExternalLogStoresElasticsearch7HttpsNo_userDeploymentYaml = []byte(`kind: Template
apiVersion: template.openshift.io/v1
metadata:
  name: external-log-store-template
objects:
- apiVersion: apps/v1
  kind: Deployment
  metadata:
    labels:
      app: ${NAME}
    name: ${NAME}
    namespace: ${NAMESPACE}
  spec:
    replicas: 1
    revisionHistoryLimit: 10
    selector:
      matchLabels:
        app: ${NAME}
    strategy:
      activeDeadlineSeconds: 21600
      resources: {}
      rollingParams:
        intervalSeconds: 1
        maxSurge: 25%
        maxUnavailable: 25%
        timeoutSeconds: 600
        updatePeriodSeconds: 1
      type: Recreate
    template:
      metadata:
        labels:
          app: ${NAME}
      spec:
        securityContext:
          runAsNonRoot: true
          seccompProfile:
            type: RuntimeDefault
        containers:
        - image: quay.io/openshifttest/elasticsearch@sha256:420ab335838747b1d350ed39f4d88cf075479fc915cda8f91373c7e00de65887
          imagePullPolicy: IfNotPresent
          name: ${NAME}
          securityContext:
            allowPrivilegeEscalation: false
            runAsNonRoot: true
            capabilities:
              drop:
              - ALL
            seccompProfile:
              type: RuntimeDefault
          ports:
          - containerPort: 9300
            protocol: TCP
          - containerPort: 9200
            protocol: TCP
          resources:
            requests:
              cpu: 1
              memory: 2Gi
          terminationMessagePath: /dev/termination-log
          terminationMessagePolicy: File
          volumeMounts:
          - mountPath: /usr/share/elasticsearch/config/elasticsearch.yml
            name: elasticsearch-config
            subPath: elasticsearch.yml
          - mountPath: /usr/share/elasticsearch/config/secret
            name: certificates
        dnsPolicy: ClusterFirst
        restartPolicy: Always
        schedulerName: default-scheduler
        securityContext: {}
        serviceAccount: ${NAME}
        serviceAccountName: ${NAME}
        terminationGracePeriodSeconds: 30
        volumes:
        - configMap:
            defaultMode: 420
            name: ${NAME}
          name: elasticsearch-config
        - name: certificates
          secret:
            defaultMode: 420
            secretName: ${NAME}
parameters:
- name: NAME
  value: elasticsearch-server
- name: NAMESPACE
  value: openshift-logging
`)

func testdataExternalLogStoresElasticsearch7HttpsNo_userDeploymentYamlBytes() ([]byte, error) {
	return _testdataExternalLogStoresElasticsearch7HttpsNo_userDeploymentYaml, nil
}

func testdataExternalLogStoresElasticsearch7HttpsNo_userDeploymentYaml() (*asset, error) {
	bytes, err := testdataExternalLogStoresElasticsearch7HttpsNo_userDeploymentYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "testdata/external-log-stores/elasticsearch/7/https/no_user/deployment.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testdataExternalLogStoresElasticsearch7HttpsUser_authConfigmapYaml = []byte(`kind: Template
apiVersion: template.openshift.io/v1
metadata:
  name: external-log-store-template
objects:
- apiVersion: v1
  data:
    add-user.sh: |+
      output = $(/usr/share/elasticsearch/bin/elasticsearch-users useradd ${USERNAME} -p ${PASSWORD} -r superuser)
      if [[ $output =~ 'already exists' ]]
      then
      return 0
      fi

    elasticsearch.yml: |
      node.name:  ${NAME}
      cluster.name: ${NAME}
      discovery.zen.minimum_master_nodes: 1
      network.host: 0.0.0.0
      http.port: 9200
      http.host: 0.0.0.0
      transport.host: 127.0.0.1
      discovery.type: single-node
      xpack.security.authc.realms.file:
        test:
          order: 0
          enabled: true
          authentication.enabled: true
      xpack.security.enabled: true
      xpack.security.authc.api_key.enabled: true
      xpack.monitoring.enabled : false
      xpack.license.self_generated.type: basic
      xpack.ml.enabled: ${MACHINE_LEARNING}
      xpack.security.http.ssl.enabled: true
      xpack.security.http.ssl.key:  /usr/share/elasticsearch/config/secret/elasticsearch.key
      xpack.security.http.ssl.certificate: /usr/share/elasticsearch/config/secret/elasticsearch.crt
      xpack.security.http.ssl.certificate_authorities: [ "/usr/share/elasticsearch/config/secret/admin-ca" ]
      xpack.security.http.ssl.verification_mode: full
      xpack.security.http.ssl.client_authentication: ${CLIENT_AUTH}
  kind: ConfigMap
  metadata:
    name: ${NAME}
    namespace: ${NAMESPACE}
parameters:
- name: NAME
  value: elasticsearch-server
- name: NAMESPACE
  value: openshift-logging
- name: USERNAME
  value: "fluentd"
- name: PASSWORD
  value: "redhat"
- name: CLIENT_AUTH
  value: none
- name: MACHINE_LEARNING
  value: "true"
`)

func testdataExternalLogStoresElasticsearch7HttpsUser_authConfigmapYamlBytes() ([]byte, error) {
	return _testdataExternalLogStoresElasticsearch7HttpsUser_authConfigmapYaml, nil
}

func testdataExternalLogStoresElasticsearch7HttpsUser_authConfigmapYaml() (*asset, error) {
	bytes, err := testdataExternalLogStoresElasticsearch7HttpsUser_authConfigmapYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "testdata/external-log-stores/elasticsearch/7/https/user_auth/configmap.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testdataExternalLogStoresElasticsearch7HttpsUser_authDeploymentYaml = []byte(`kind: Template
apiVersion: template.openshift.io/v1
metadata:
  name: external-log-store-template
objects:
- apiVersion: apps/v1
  kind: Deployment
  metadata:
    labels:
      app: ${NAME}
    name: ${NAME}
    namespace: ${NAMESPACE}
  spec:
    replicas: 1
    revisionHistoryLimit: 10
    selector:
      matchLabels:
        app: ${NAME}
    strategy:
      activeDeadlineSeconds: 21600
      resources: {}
      rollingParams:
        intervalSeconds: 1
        maxSurge: 25%
        maxUnavailable: 25%
        timeoutSeconds: 600
        updatePeriodSeconds: 1
      type: Recreate
    template:
      metadata:
        labels:
          app: ${NAME}
      spec:
        securityContext:
          runAsNonRoot: true
          seccompProfile:
            type: RuntimeDefault
        containers:
        - image: quay.io/openshifttest/elasticsearch@sha256:420ab335838747b1d350ed39f4d88cf075479fc915cda8f91373c7e00de65887
          imagePullPolicy: IfNotPresent
          name: ${NAME}
          securityContext:
            allowPrivilegeEscalation: false
            runAsNonRoot: true
            capabilities:
              drop:
              - ALL
            seccompProfile:
              type: RuntimeDefault
          ports:
          - containerPort: 9300
            protocol: TCP
          - containerPort: 9200
            protocol: TCP
          readinessProbe:
            exec:
              command:
              - sh
              - /usr/share/elasticsearch/add-user.sh
            failureThreshold: 3
            initialDelaySeconds: 5
            periodSeconds: 5
            successThreshold: 1
            timeoutSeconds: 10
          resources:
            requests:
              cpu: 1
              memory: 2Gi
          terminationMessagePath: /dev/termination-log
          terminationMessagePolicy: File
          volumeMounts:
          - mountPath: /usr/share/elasticsearch/config/elasticsearch.yml
            name: elasticsearch-config
            subPath: elasticsearch.yml
          - mountPath: /usr/share/elasticsearch/add-user.sh
            name: elasticsearch-config
            subPath: add-user.sh
          - mountPath: /usr/share/elasticsearch/config/secret
            name: certificates
        dnsPolicy: ClusterFirst
        restartPolicy: Always
        schedulerName: default-scheduler
        securityContext: {}
        serviceAccount: ${NAME}
        serviceAccountName: ${NAME}
        terminationGracePeriodSeconds: 30
        volumes:
        - configMap:
            defaultMode: 420
            name: ${NAME}
          name: elasticsearch-config
        - name: certificates
          secret:
            defaultMode: 420
            secretName: ${NAME}
parameters:
- name: NAME
  value: elasticsearch-server
- name: NAMESPACE
  value: openshift-logging
`)

func testdataExternalLogStoresElasticsearch7HttpsUser_authDeploymentYamlBytes() ([]byte, error) {
	return _testdataExternalLogStoresElasticsearch7HttpsUser_authDeploymentYaml, nil
}

func testdataExternalLogStoresElasticsearch7HttpsUser_authDeploymentYaml() (*asset, error) {
	bytes, err := testdataExternalLogStoresElasticsearch7HttpsUser_authDeploymentYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "testdata/external-log-stores/elasticsearch/7/https/user_auth/deployment.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testdataExternalLogStoresElasticsearch8HttpNo_userConfigmapYaml = []byte(`kind: Template
apiVersion: template.openshift.io/v1
metadata:
  name: external-log-store-template
objects:
- apiVersion: v1
  data:
    elasticsearch.yml: |
      node.name:  ${NAME}
      cluster.name: ${NAME}
      network.host: 0.0.0.0
      http.port: 9200
      http.host: 0.0.0.0
      transport.host: 127.0.0.1
      discovery.type: single-node
      xpack.security.enabled: false
      xpack.security.authc.api_key.enabled: false
      xpack.monitoring.collection.enabled: false
      xpack.license.self_generated.type: basic
      xpack.security.http.ssl.enabled: false
      xpack.ml.enabled: ${MACHINE_LEARNING}
  kind: ConfigMap
  metadata:
    name: ${NAME}
    namespace: ${NAMESPACE}
parameters:
- name: NAME
  value: elasticsearch-server
- name: NAMESPACE
  value: openshift-logging
- name: MACHINE_LEARNING
  value: "true"
`)

func testdataExternalLogStoresElasticsearch8HttpNo_userConfigmapYamlBytes() ([]byte, error) {
	return _testdataExternalLogStoresElasticsearch8HttpNo_userConfigmapYaml, nil
}

func testdataExternalLogStoresElasticsearch8HttpNo_userConfigmapYaml() (*asset, error) {
	bytes, err := testdataExternalLogStoresElasticsearch8HttpNo_userConfigmapYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "testdata/external-log-stores/elasticsearch/8/http/no_user/configmap.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testdataExternalLogStoresElasticsearch8HttpNo_userDeploymentYaml = []byte(`kind: Template
apiVersion: template.openshift.io/v1
metadata:
  name: external-log-store-template
objects:
- apiVersion: apps/v1
  kind: Deployment
  metadata:
    labels:
      app: ${NAME}
    name: ${NAME}
    namespace: ${NAMESPACE}
  spec:
    replicas: 1
    revisionHistoryLimit: 10
    selector:
      matchLabels:
        app: ${NAME}
    strategy:
      activeDeadlineSeconds: 21600
      resources: {}
      rollingParams:
        intervalSeconds: 1
        maxSurge: 25%
        maxUnavailable: 25%
        timeoutSeconds: 600
        updatePeriodSeconds: 1
      type: Recreate
    template:
      metadata:
        labels:
          app: ${NAME}
      spec:
        securityContext:
          runAsNonRoot: true
          seccompProfile:
            type: RuntimeDefault
        containers:
        - image: quay.io/openshifttest/elasticsearch@sha256:234e8ecfb6c1bdafffeb190c4a48e8e3c4b74b69d1b541b913dfbca29e952f63
          imagePullPolicy: IfNotPresent
          name: ${NAME}
          securityContext:
            allowPrivilegeEscalation: false
            runAsNonRoot: true
            capabilities:
              drop:
              - ALL
            seccompProfile:
              type: RuntimeDefault
          ports:
          - containerPort: 9300
            protocol: TCP
          - containerPort: 9200
            protocol: TCP
          volumeMounts:
          - mountPath: /usr/share/elasticsearch/config/elasticsearch.yml
            subPath: elasticsearch.yml
            name: elasticsearch-config
          resources:
            requests:
              cpu: 1
              memory: 2Gi
          terminationMessagePath: /dev/termination-log
          terminationMessagePolicy: File
        volumes:
        - configMap:
            defaultMode: 420
            name: ${NAME}
          name: elasticsearch-config
        serviceAccount: ${NAME}
        serviceAccountName: ${NAME}
        dnsPolicy: ClusterFirst
        restartPolicy: Always
parameters:
- name: NAME
  value: elasticsearch-server
- name: NAMESPACE
  value: openshift-logging
`)

func testdataExternalLogStoresElasticsearch8HttpNo_userDeploymentYamlBytes() ([]byte, error) {
	return _testdataExternalLogStoresElasticsearch8HttpNo_userDeploymentYaml, nil
}

func testdataExternalLogStoresElasticsearch8HttpNo_userDeploymentYaml() (*asset, error) {
	bytes, err := testdataExternalLogStoresElasticsearch8HttpNo_userDeploymentYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "testdata/external-log-stores/elasticsearch/8/http/no_user/deployment.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testdataExternalLogStoresElasticsearch8HttpUser_authConfigmapYaml = []byte(`kind: Template
apiVersion: template.openshift.io/v1
metadata:
  name: external-log-store-template
objects:
- apiVersion: v1
  data:
    add-user.sh: |+
      output = $(/usr/share/elasticsearch/bin/elasticsearch-users useradd ${USERNAME} -p ${PASSWORD} -r superuser)
      if [[ $output =~ 'already exists' ]]
      then
      return 0
      fi

    elasticsearch.yml: |
      node.name:  ${NAME}
      cluster.name: ${NAME}
      network.host: 0.0.0.0
      http.port: 9200
      http.host: 0.0.0.0
      transport.host: 127.0.0.1
      discovery.type: single-node
      xpack.security.authc.realms:
        file:
          test:
            order: 0
            enabled: true
            authentication.enabled: true
      xpack.security.enabled: true
      xpack.security.authc.api_key.enabled: true
      xpack.monitoring.collection.enabled: false
      xpack.license.self_generated.type: basic
      xpack.security.http.ssl.enabled: false
      xpack.ml.enabled: ${MACHINE_LEARNING}
  kind: ConfigMap
  metadata:
    name: ${NAME}
    namespace: ${NAMESPACE}
parameters:
- name: NAME
  value: elasticsearch-server
- name: NAMESPACE
  value: openshift-logging
- name: USERNAME
  value: "fluentd"
- name: PASSWORD
  value: "redhat"
- name: MACHINE_LEARNING
  value: "true"
`)

func testdataExternalLogStoresElasticsearch8HttpUser_authConfigmapYamlBytes() ([]byte, error) {
	return _testdataExternalLogStoresElasticsearch8HttpUser_authConfigmapYaml, nil
}

func testdataExternalLogStoresElasticsearch8HttpUser_authConfigmapYaml() (*asset, error) {
	bytes, err := testdataExternalLogStoresElasticsearch8HttpUser_authConfigmapYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "testdata/external-log-stores/elasticsearch/8/http/user_auth/configmap.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testdataExternalLogStoresElasticsearch8HttpUser_authDeploymentYaml = []byte(`kind: Template
apiVersion: template.openshift.io/v1
metadata:
  name: external-log-store-template
objects:
- apiVersion: apps/v1
  kind: Deployment
  metadata:
    labels:
      app: ${NAME}
    name: ${NAME}
    namespace: ${NAMESPACE}
  spec:
    replicas: 1
    revisionHistoryLimit: 10
    selector:
      matchLabels:
        app: ${NAME}
    strategy:
      activeDeadlineSeconds: 21600
      resources: {}
      rollingParams:
        intervalSeconds: 1
        maxSurge: 25%
        maxUnavailable: 25%
        timeoutSeconds: 600
        updatePeriodSeconds: 1
      type: Recreate
    template:
      metadata:
        labels:
          app: ${NAME}
      spec:
        securityContext:
          runAsNonRoot: true
          seccompProfile:
            type: RuntimeDefault
        containers:
        - image: quay.io/openshifttest/elasticsearch@sha256:234e8ecfb6c1bdafffeb190c4a48e8e3c4b74b69d1b541b913dfbca29e952f63
          imagePullPolicy: IfNotPresent
          name: ${NAME}
          securityContext:
            allowPrivilegeEscalation: false
            runAsNonRoot: true
            capabilities:
              drop:
              - ALL
            seccompProfile:
              type: RuntimeDefault
          ports:
          - containerPort: 9300
            protocol: TCP
          - containerPort: 9200
            protocol: TCP
          readinessProbe:
            exec:
              command:
              - sh
              - /usr/share/elasticsearch/add-user.sh
            failureThreshold: 3
            initialDelaySeconds: 5
            periodSeconds: 5
            successThreshold: 1
            timeoutSeconds: 10
          resources:
            requests:
              cpu: 1
              memory: 2Gi
          terminationMessagePath: /dev/termination-log
          terminationMessagePolicy: File
          volumeMounts:
          - mountPath: /usr/share/elasticsearch/config/elasticsearch.yml
            name: elasticsearch-config
            subPath: elasticsearch.yml
          - mountPath: /usr/share/elasticsearch/add-user.sh
            name: elasticsearch-config
            subPath: add-user.sh
        dnsPolicy: ClusterFirst
        restartPolicy: Always
        schedulerName: default-scheduler
        securityContext: {}
        serviceAccount: ${NAME}
        serviceAccountName: ${NAME}
        terminationGracePeriodSeconds: 30
        volumes:
        - configMap:
            defaultMode: 420
            name: ${NAME}
          name: elasticsearch-config
parameters:
- name: NAME
  value: elasticsearch-server
- name: NAMESPACE
  value: openshift-logging
`)

func testdataExternalLogStoresElasticsearch8HttpUser_authDeploymentYamlBytes() ([]byte, error) {
	return _testdataExternalLogStoresElasticsearch8HttpUser_authDeploymentYaml, nil
}

func testdataExternalLogStoresElasticsearch8HttpUser_authDeploymentYaml() (*asset, error) {
	bytes, err := testdataExternalLogStoresElasticsearch8HttpUser_authDeploymentYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "testdata/external-log-stores/elasticsearch/8/http/user_auth/deployment.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testdataExternalLogStoresElasticsearch8HttpsNo_userConfigmapYaml = []byte(`kind: Template
apiVersion: template.openshift.io/v1
metadata:
  name: external-log-store-template
objects:
- apiVersion: v1
  data:
    elasticsearch.yml: |
      node.name: ${NAME}
      cluster.name: ${NAME}
      network.host: 0.0.0.0
      http.port: 9200
      http.host: 0.0.0.0
      transport.host: 127.0.0.1
      discovery.type: single-node
      xpack.security.enabled: true
      xpack.security.authc:
        anonymous:
          username: anonymous_user
          roles: superuser
          authz_exception: true
      xpack.security.authc.api_key.enabled: true
      xpack.monitoring.collection.enabled: false
      xpack.license.self_generated.type: basic
      xpack.security.http.ssl.enabled: true
      xpack.security.http.ssl.key:  /usr/share/elasticsearch/config/secret/elasticsearch.key
      xpack.security.http.ssl.certificate: /usr/share/elasticsearch/config/secret/elasticsearch.crt
      xpack.security.http.ssl.certificate_authorities: [ "/usr/share/elasticsearch/config/secret/admin-ca" ]
      xpack.security.http.ssl.verification_mode: full
      xpack.security.http.ssl.client_authentication: ${CLIENT_AUTH}
      xpack.ml.enabled: ${MACHINE_LEARNING}
  kind: ConfigMap
  metadata:
    name: ${NAME}
    namespace: ${NAMESPACE}
parameters:
- name: NAME
  value: elasticsearch-server
- name: NAMESPACE
  value: openshift-logging
- name: CLIENT_AUTH
  value: required
- name: MACHINE_LEARNING
  value: "true"
`)

func testdataExternalLogStoresElasticsearch8HttpsNo_userConfigmapYamlBytes() ([]byte, error) {
	return _testdataExternalLogStoresElasticsearch8HttpsNo_userConfigmapYaml, nil
}

func testdataExternalLogStoresElasticsearch8HttpsNo_userConfigmapYaml() (*asset, error) {
	bytes, err := testdataExternalLogStoresElasticsearch8HttpsNo_userConfigmapYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "testdata/external-log-stores/elasticsearch/8/https/no_user/configmap.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testdataExternalLogStoresElasticsearch8HttpsNo_userDeploymentYaml = []byte(`kind: Template
apiVersion: template.openshift.io/v1
metadata:
  name: external-log-store-template
objects:
- apiVersion: apps/v1
  kind: Deployment
  metadata:
    labels:
      app: ${NAME}
    name: ${NAME}
    namespace: ${NAMESPACE}
  spec:
    replicas: 1
    revisionHistoryLimit: 10
    selector:
      matchLabels:
        app: ${NAME}
    strategy:
      activeDeadlineSeconds: 21600
      resources: {}
      rollingParams:
        intervalSeconds: 1
        maxSurge: 25%
        maxUnavailable: 25%
        timeoutSeconds: 600
        updatePeriodSeconds: 1
      type: Recreate
    template:
      metadata:
        labels:
          app: ${NAME}
      spec:
        securityContext:
          runAsNonRoot: true
          seccompProfile:
            type: RuntimeDefault
        containers:
        - image: quay.io/openshifttest/elasticsearch@sha256:234e8ecfb6c1bdafffeb190c4a48e8e3c4b74b69d1b541b913dfbca29e952f63
          imagePullPolicy: IfNotPresent
          name: ${NAME}
          securityContext:
            allowPrivilegeEscalation: false
            runAsNonRoot: true
            capabilities:
              drop:
              - ALL
            seccompProfile:
              type: RuntimeDefault
          ports:
          - containerPort: 9300
            protocol: TCP
          - containerPort: 9200
            protocol: TCP
          resources:
            requests:
              cpu: 1
              memory: 2Gi
          terminationMessagePath: /dev/termination-log
          terminationMessagePolicy: File
          volumeMounts:
          - mountPath: /usr/share/elasticsearch/config/elasticsearch.yml
            name: elasticsearch-config
            subPath: elasticsearch.yml
          - mountPath: /usr/share/elasticsearch/config/secret
            name: certificates
        dnsPolicy: ClusterFirst
        restartPolicy: Always
        schedulerName: default-scheduler
        securityContext: {}
        serviceAccount: ${NAME}
        serviceAccountName: ${NAME}
        terminationGracePeriodSeconds: 30
        volumes:
        - configMap:
            defaultMode: 420
            name: ${NAME}
          name: elasticsearch-config
        - name: certificates
          secret:
            defaultMode: 420
            secretName: ${NAME}
parameters:
- name: NAME
  value: elasticsearch-server
- name: NAMESPACE
  value: openshift-logging
`)

func testdataExternalLogStoresElasticsearch8HttpsNo_userDeploymentYamlBytes() ([]byte, error) {
	return _testdataExternalLogStoresElasticsearch8HttpsNo_userDeploymentYaml, nil
}

func testdataExternalLogStoresElasticsearch8HttpsNo_userDeploymentYaml() (*asset, error) {
	bytes, err := testdataExternalLogStoresElasticsearch8HttpsNo_userDeploymentYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "testdata/external-log-stores/elasticsearch/8/https/no_user/deployment.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testdataExternalLogStoresElasticsearch8HttpsUser_authConfigmapYaml = []byte(`kind: Template
apiVersion: template.openshift.io/v1
metadata:
  name: external-log-store-template
objects:
- apiVersion: v1
  data:
    add-user.sh: |+
      output = $(/usr/share/elasticsearch/bin/elasticsearch-users useradd ${USERNAME} -p ${PASSWORD} -r superuser)
      if [[ $output =~ 'already exists' ]]
      then
      return 0
      fi

    elasticsearch.yml: |
      node.name:  ${NAME}
      cluster.name: ${NAME}
      network.host: 0.0.0.0
      http.port: 9200
      http.host: 0.0.0.0
      transport.host: 127.0.0.1
      discovery.type: single-node
      xpack.security.authc.realms.file:
        test:
          order: 0
          enabled: true
          authentication.enabled: true
      xpack.security.enabled: true
      xpack.security.authc.api_key.enabled: true
      xpack.monitoring.collection.enabled: false
      xpack.license.self_generated.type: basic
      xpack.security.http.ssl.enabled: true
      xpack.security.http.ssl.key:  /usr/share/elasticsearch/config/secret/elasticsearch.key
      xpack.security.http.ssl.certificate: /usr/share/elasticsearch/config/secret/elasticsearch.crt
      xpack.security.http.ssl.certificate_authorities: [ "/usr/share/elasticsearch/config/secret/admin-ca" ]
      xpack.security.http.ssl.verification_mode: full
      xpack.security.http.ssl.client_authentication: ${CLIENT_AUTH}
      xpack.ml.enabled: ${MACHINE_LEARNING}
  kind: ConfigMap
  metadata:
    name: ${NAME}
    namespace: ${NAMESPACE}
parameters:
- name: NAME
  value: elasticsearch-server
- name: NAMESPACE
  value: openshift-logging
- name: USERNAME
  value: "fluentd"
- name: PASSWORD
  value: "redhat"
- name: CLIENT_AUTH
  value: none
- name: MACHINE_LEARNING
  value: "true"
`)

func testdataExternalLogStoresElasticsearch8HttpsUser_authConfigmapYamlBytes() ([]byte, error) {
	return _testdataExternalLogStoresElasticsearch8HttpsUser_authConfigmapYaml, nil
}

func testdataExternalLogStoresElasticsearch8HttpsUser_authConfigmapYaml() (*asset, error) {
	bytes, err := testdataExternalLogStoresElasticsearch8HttpsUser_authConfigmapYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "testdata/external-log-stores/elasticsearch/8/https/user_auth/configmap.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testdataExternalLogStoresElasticsearch8HttpsUser_authDeploymentYaml = []byte(`kind: Template
apiVersion: template.openshift.io/v1
metadata:
  name: external-log-store-template
objects:
- apiVersion: apps/v1
  kind: Deployment
  metadata:
    labels:
      app: ${NAME}
    name: ${NAME}
    namespace: ${NAMESPACE}
  spec:
    replicas: 1
    revisionHistoryLimit: 10
    selector:
      matchLabels:
        app: ${NAME}
    strategy:
      activeDeadlineSeconds: 21600
      resources: {}
      rollingParams:
        intervalSeconds: 1
        maxSurge: 25%
        maxUnavailable: 25%
        timeoutSeconds: 600
        updatePeriodSeconds: 1
      type: Recreate
    template:
      metadata:
        labels:
          app: ${NAME}
      spec:
        securityContext:
          runAsNonRoot: true
          seccompProfile:
            type: RuntimeDefault
        containers:
        - image: quay.io/openshifttest/elasticsearch@sha256:234e8ecfb6c1bdafffeb190c4a48e8e3c4b74b69d1b541b913dfbca29e952f63
          imagePullPolicy: IfNotPresent
          name: ${NAME}
          securityContext:
            allowPrivilegeEscalation: false
            runAsNonRoot: true
            capabilities:
              drop:
              - ALL
            seccompProfile:
              type: RuntimeDefault
          ports:
          - containerPort: 9300
            protocol: TCP
          - containerPort: 9200
            protocol: TCP
          readinessProbe:
            exec:
              command:
              - sh
              - /usr/share/elasticsearch/add-user.sh
            failureThreshold: 3
            initialDelaySeconds: 5
            periodSeconds: 5
            successThreshold: 1
            timeoutSeconds: 10
          resources:
            requests:
              cpu: 1
              memory: 2Gi
          terminationMessagePath: /dev/termination-log
          terminationMessagePolicy: File
          volumeMounts:
          - mountPath: /usr/share/elasticsearch/config/elasticsearch.yml
            name: elasticsearch-config
            subPath: elasticsearch.yml
          - mountPath: /usr/share/elasticsearch/add-user.sh
            name: elasticsearch-config
            subPath: add-user.sh
          - mountPath: /usr/share/elasticsearch/config/secret
            name: certificates
        dnsPolicy: ClusterFirst
        restartPolicy: Always
        schedulerName: default-scheduler
        securityContext: {}
        serviceAccount: ${NAME}
        serviceAccountName: ${NAME}
        terminationGracePeriodSeconds: 30
        volumes:
        - configMap:
            defaultMode: 420
            name: ${NAME}
          name: elasticsearch-config
        - name: certificates
          secret:
            defaultMode: 420
            secretName: ${NAME}
parameters:
- name: NAME
  value: elasticsearch-server
- name: NAMESPACE
  value: openshift-logging
`)

func testdataExternalLogStoresElasticsearch8HttpsUser_authDeploymentYamlBytes() ([]byte, error) {
	return _testdataExternalLogStoresElasticsearch8HttpsUser_authDeploymentYaml, nil
}

func testdataExternalLogStoresElasticsearch8HttpsUser_authDeploymentYaml() (*asset, error) {
	bytes, err := testdataExternalLogStoresElasticsearch8HttpsUser_authDeploymentYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "testdata/external-log-stores/elasticsearch/8/https/user_auth/deployment.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testdataExternalLogStoresFluentdInsecureConfigmapYaml = []byte(`kind: Template
apiVersion: template.openshift.io/v1
metadata:
  name: external-log-store-template
objects:
- apiVersion: v1
  data:
    fluent.conf: |
      <source>
        @type forward
        port  24224
      </source>

      <match kubernetes.var.log.pods.openshift-*_** kubernetes.var.log.pods.default_** kubernetes.var.log.pods.kube-*_**>
        @type file
        append true
        path /fluentd/log/infra-container.*.log
        symlink_path /fluentd/log/infra-container.log
        time_slice_format %Y%m%d
        time_slice_wait   1m
        time_format       %Y%m%dT%H%M%S%z
      </match>
      <match journal.** system.var.log**>
        @type file
        append true
        path /fluentd/log/infra.*.log
        symlink_path /fluentd/log/infra.log
        time_slice_format %Y%m%d
        time_slice_wait   1m
        time_format       %Y%m%dT%H%M%S%z
      </match>
      <match kubernetes.**>
        @type file
        append true
        path /fluentd/log/app.*.log
        symlink_path /fluentd/log/app.log
        time_slice_format %Y%m%d
        time_slice_wait   1m
        time_format       %Y%m%dT%H%M%S%z
      </match>
      <match linux-audit.log** k8s-audit.log** openshift-audit.log** ovn-audit.log**>
        @type file
        append true
        path /fluentd/log/audit.*.log
        symlink_path /fluentd/log/audit.log
        time_slice_format %Y%m%d
        time_slice_wait   1m
        time_format       %Y%m%dT%H%M%S%z
      </match>
      <match **>
        @type stdout
      </match>
  kind: ConfigMap
  metadata:
    name: ${NAME}
    namespace: ${NAMESPACE}
parameters:
- name: NAME
  value: fluentdserver
- name: NAMESPACE
  value: openshift-logging
`)

func testdataExternalLogStoresFluentdInsecureConfigmapYamlBytes() ([]byte, error) {
	return _testdataExternalLogStoresFluentdInsecureConfigmapYaml, nil
}

func testdataExternalLogStoresFluentdInsecureConfigmapYaml() (*asset, error) {
	bytes, err := testdataExternalLogStoresFluentdInsecureConfigmapYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "testdata/external-log-stores/fluentd/insecure/configmap.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testdataExternalLogStoresFluentdInsecureDeploymentYaml = []byte(`kind: Template
apiVersion: template.openshift.io/v1
metadata:
  name: external-log-store-template
objects:
- apiVersion: apps/v1
  kind: Deployment
  metadata:
    name: ${NAME}
    namespace: ${NAMESPACE}
    labels:
      provider: aosqe
      component: ${NAME}
      logging-infra: ${NAME}
  spec:
    replicas: 1
    revisionHistoryLimit: 10
    selector:
      matchLabels:
        provider: aosqe
        component: ${NAME}
        logging-infra: ${NAME}
    strategy:
      type: Recreate
    template:
      metadata:
        labels:
          logging-infra: ${NAME}
          provider: aosqe
          component: ${NAME}
      spec:
        securityContext:
          runAsNonRoot: true
          seccompProfile:
            type: RuntimeDefault
        containers:
        - name: "fluentdserver"
          image: "quay.io/openshifttest/fluentd:1.2.2"
          imagePullPolicy: "IfNotPresent"
          securityContext:
            allowPrivilegeEscalation: false
            runAsNonRoot: true
            capabilities:
              drop:
              - ALL
            seccompProfile:
              type: RuntimeDefault
          ports:
          - containerPort: 24224
            name: fluentdserver
          volumeMounts:
          - mountPath: /fluentd/etc
            name: config
            readOnly: true
        volumes:
        - configMap:
            defaultMode: 420
            name: ${NAME}
          name: config
parameters:
- name: NAME
  value: fluentdserver
- name: NAMESPACE
  value: openshift-logging
`)

func testdataExternalLogStoresFluentdInsecureDeploymentYamlBytes() ([]byte, error) {
	return _testdataExternalLogStoresFluentdInsecureDeploymentYaml, nil
}

func testdataExternalLogStoresFluentdInsecureDeploymentYaml() (*asset, error) {
	bytes, err := testdataExternalLogStoresFluentdInsecureDeploymentYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "testdata/external-log-stores/fluentd/insecure/deployment.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testdataExternalLogStoresFluentdInsecureHttpConfigmapYaml = []byte(`kind: Template
apiVersion: template.openshift.io/v1
metadata:
  name: fluentd-http-template
objects:
- apiVersion: v1
  data:
    fluent.conf: |
      <system>
        log_level info
      </system>
      <source>
        @type http
        port 24224
        bind 0.0.0.0
        body_size_limit 32m
        keepalive_timeout 10s
        add_http_headers true
        add_remote_addr true
        <parse>
          @type json
        </parse>
        @label @collector_logs
      </source>
      <source>
        @type http
        port 24224
        bind "::"
        body_size_limit 32m
        keepalive_timeout 10s
        add_http_headers true
        add_remote_addr true
        <parse>
          @type json
        </parse>
        @label @collector_logs
      </source>
      <label @collector_logs>
        <match logs.app>
          @type file
          append true
          path /fluentd/log/app.*.log
          symlink_path /fluentd/log/app.log
        </match>
        <match logs.infra>
          @type file
          append true
          path /fluentd/log/infra.*.log
          symlink_path /fluentd/log/infra.log
        </match>
        <match logs.audit>
          @type file
          append true
          path /fluentd/log/audit.*.log
          symlink_path /fluentd/log/audit.log
        </match>
      </label>
      <label @FLUENT_LOG>
        <match **>
      	  @type stdout
        </match>
      </label>
  kind: ConfigMap
  metadata:
    name: ${NAME}
    namespace: ${NAMESPACE}
parameters:
- name: NAME
  value: fluentdserver
- name: NAMESPACE
  value: openshift-logging
`)

func testdataExternalLogStoresFluentdInsecureHttpConfigmapYamlBytes() ([]byte, error) {
	return _testdataExternalLogStoresFluentdInsecureHttpConfigmapYaml, nil
}

func testdataExternalLogStoresFluentdInsecureHttpConfigmapYaml() (*asset, error) {
	bytes, err := testdataExternalLogStoresFluentdInsecureHttpConfigmapYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "testdata/external-log-stores/fluentd/insecure/http-configmap.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testdataExternalLogStoresFluentdSecureCmMtlsShareYaml = []byte(`kind: Template
apiVersion: template.openshift.io/v1
metadata:
  name: external-log-store-template
objects:
- apiVersion: v1
  data:
    fluent.conf: |
      <source>
        @type forward
        port  24224
        <transport tls>
          ca_path /etc/fluentd/secrets/ca-bundle.crt
          cert_path /etc/fluentd/secrets/tls.crt
          private_key_path /etc/fluentd/secrets/tls.key
          client_cert_auth true
        </transport>
        <security>
          shared_key ${SHARED_KEY}
          self_hostname ${NAME}
        </security>
      </source>

      <match kubernetes.var.log.pods.openshift-*_** kubernetes.var.log.pods.default_** kubernetes.var.log.pods.kube-*_**>
        @type file
        append true
        path /fluentd/log/infra-container.*.log
        symlink_path /fluentd/log/infra-container.log
        time_slice_format %Y%m%d
        time_slice_wait   1m
        time_format       %Y%m%dT%H%M%S%z
      </match>
      <match journal.** system.var.log**>
        @type file
        append true
        path /fluentd/log/infra.*.log
        symlink_path /fluentd/log/infra.log
        time_slice_format %Y%m%d
        time_slice_wait   1m
        time_format       %Y%m%dT%H%M%S%z
      </match>
      <match kubernetes.**>
        @type file
        append true
        path /fluentd/log/app.*.log
        symlink_path /fluentd/log/app.log
        time_slice_format %Y%m%d
        time_slice_wait   1m
       time_format       %Y%m%dT%H%M%S%z
      </match>
      <match linux-audit.log** k8s-audit.log** openshift-audit.log** ovn-audit.log**>
        @type file
        append true
        path /fluentd/log/audit.*.log
        symlink_path /fluentd/log/audit.log
        time_slice_format %Y%m%d
        time_slice_wait   1m
        time_format       %Y%m%dT%H%M%S%z
      </match>
      <match **>
        @type stdout
      </match>
  kind: ConfigMap
  metadata:
    name: ${NAME}
    namespace: ${NAMESPACE}
parameters:
- name: NAME
  value: fluentdserver
- name: NAMESPACE
  value: openshift-logging
- name: SHARED_KEY
`)

func testdataExternalLogStoresFluentdSecureCmMtlsShareYamlBytes() ([]byte, error) {
	return _testdataExternalLogStoresFluentdSecureCmMtlsShareYaml, nil
}

func testdataExternalLogStoresFluentdSecureCmMtlsShareYaml() (*asset, error) {
	bytes, err := testdataExternalLogStoresFluentdSecureCmMtlsShareYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "testdata/external-log-stores/fluentd/secure/cm-mtls-share.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testdataExternalLogStoresFluentdSecureCmMtlsYaml = []byte(`kind: Template
apiVersion: template.openshift.io/v1
metadata:
  name: external-log-store-template
objects:
- apiVersion: v1
  data:
    fluent.conf: |
      <source>
        @type forward
        port  24224
        <transport tls>
          ca_path /etc/fluentd/secrets/ca-bundle.crt
          cert_path /etc/fluentd/secrets/tls.crt
          private_key_path /etc/fluentd/secrets/tls.key
          client_cert_auth true
        </transport>
      </source>

      <match kubernetes.var.log.pods.openshift-*_** kubernetes.var.log.pods.default_** kubernetes.var.log.pods.kube-*_**>
        @type file
        append true
        path /fluentd/log/infra-container.*.log
        symlink_path /fluentd/log/infra-container.log
        time_slice_format %Y%m%d
        time_slice_wait   1m
        time_format       %Y%m%dT%H%M%S%z
      </match>
      <match journal.** system.var.log**>
        @type file
        append true
        path /fluentd/log/infra.*.log
        symlink_path /fluentd/log/infra.log
        time_slice_format %Y%m%d
        time_slice_wait   1m
        time_format       %Y%m%dT%H%M%S%z
      </match>
      <match kubernetes.**>
        @type file
        append true
        path /fluentd/log/app.*.log
        symlink_path /fluentd/log/app.log
        time_slice_format %Y%m%d
        time_slice_wait   1m
        time_format       %Y%m%dT%H%M%S%z
      </match>
      <match linux-audit.log** k8s-audit.log** openshift-audit.log** ovn-audit.log**>
        @type file
        append true
        path /fluentd/log/audit.*.log
        symlink_path /fluentd/log/audit.log
        time_slice_format %Y%m%d
        time_slice_wait   1m
        time_format       %Y%m%dT%H%M%S%z
      </match>
      <match **>
        @type stdout
      </match>
  kind: ConfigMap
  metadata:
    name: ${NAME}
    namespace: ${NAMESPACE}
parameters:
- name: NAME
  value: fluentdserver
- name: NAMESPACE
  value: openshift-logging
`)

func testdataExternalLogStoresFluentdSecureCmMtlsYamlBytes() ([]byte, error) {
	return _testdataExternalLogStoresFluentdSecureCmMtlsYaml, nil
}

func testdataExternalLogStoresFluentdSecureCmMtlsYaml() (*asset, error) {
	bytes, err := testdataExternalLogStoresFluentdSecureCmMtlsYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "testdata/external-log-stores/fluentd/secure/cm-mtls.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testdataExternalLogStoresFluentdSecureCmServerauthShareYaml = []byte(`kind: Template
apiVersion: template.openshift.io/v1
metadata:
  name: external-log-store-template
objects:
- apiVersion: v1
  data:
    fluent.conf: |
      <source>
        @type forward
        port  24224
        <transport tls>
          ca_cert_path /etc/fluentd/secrets/ca-bundle.crt
          ca_private_key_path /etc/fluentd/secrets/ca.key
        </transport>
        <security>
          shared_key ${SHARED_KEY}
          self_hostname ${NAME}
        </security>
      </source>

      <match kubernetes.var.log.pods.openshift-*_** kubernetes.var.log.pods.default_** kubernetes.var.log.pods.kube-*_**>
        @type file
        append true
        path /fluentd/log/infra-container.*.log
        symlink_path /fluentd/log/infra-container.log
        time_slice_format %Y%m%d
        time_slice_wait   1m
        time_format       %Y%m%dT%H%M%S%z
      </match>
      <match journal.** system.var.log**>
        @type file
        append true
        path /fluentd/log/infra.*.log
        symlink_path /fluentd/log/infra.log
        time_slice_format %Y%m%d
        time_slice_wait   1m
        time_format       %Y%m%dT%H%M%S%z
      </match>
      <match kubernetes.**>
        @type file
        append true
        path /fluentd/log/app.*.log
        symlink_path /fluentd/log/app.log
        time_slice_format %Y%m%d
        time_slice_wait   1m
        time_format       %Y%m%dT%H%M%S%z
      </match>
      <match linux-audit.log** k8s-audit.log** openshift-audit.log** ovn-audit.log**>
        @type file
        append true
        path /fluentd/log/audit.*.log
        symlink_path /fluentd/log/audit.log
        time_slice_format %Y%m%d
        time_slice_wait   1m
        time_format       %Y%m%dT%H%M%S%z
      </match>
      <match **>
        @type stdout
      </match>
  kind: ConfigMap
  metadata:
    name: ${NAME}
    namespace: ${NAMESPACE}
parameters:
- name: NAME
  value: fluentdserver
- name: NAMESPACE
  value: openshift-logging
- name: SHARED_KEY
`)

func testdataExternalLogStoresFluentdSecureCmServerauthShareYamlBytes() ([]byte, error) {
	return _testdataExternalLogStoresFluentdSecureCmServerauthShareYaml, nil
}

func testdataExternalLogStoresFluentdSecureCmServerauthShareYaml() (*asset, error) {
	bytes, err := testdataExternalLogStoresFluentdSecureCmServerauthShareYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "testdata/external-log-stores/fluentd/secure/cm-serverauth-share.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testdataExternalLogStoresFluentdSecureCmServerauthYaml = []byte(`kind: Template
apiVersion: template.openshift.io/v1
metadata:
  name: external-log-store-template
objects:
- apiVersion: v1
  data:
    fluent.conf: |
      <source>
        @type forward
        port  24224
        <transport tls>
          ca_cert_path /etc/fluentd/secrets/ca-bundle.crt
          ca_private_key_path /etc/fluentd/secrets/ca.key
        </transport>
      </source>

      <match kubernetes.var.log.pods.openshift-*_** kubernetes.var.log.pods.default_** kubernetes.var.log.pods.kube-*_**>
        @type file
        append true
        path /fluentd/log/infra-container.*.log
        symlink_path /fluentd/log/infra-container.log
        time_slice_format %Y%m%d
        time_slice_wait   1m
        time_format       %Y%m%dT%H%M%S%z
      </match>
      <match journal.** system.var.log**>
        @type file
        append true
        path /fluentd/log/infra.*.log
        symlink_path /fluentd/log/infra.log
        time_slice_format %Y%m%d
        time_slice_wait   1m
        time_format       %Y%m%dT%H%M%S%z
      </match>
      <match kubernetes.**>
        @type file
        append true
        path /fluentd/log/app.*.log
        symlink_path /fluentd/log/app.log
        time_slice_format %Y%m%d
        time_slice_wait   1m
        time_format       %Y%m%dT%H%M%S%z
      </match>
      <match linux-audit.log** k8s-audit.log** openshift-audit.log** ovn-audit.log**>
        @type file
        append true
        path /fluentd/log/audit.*.log
        symlink_path /fluentd/log/audit.log
        time_slice_format %Y%m%d
        time_slice_wait   1m
        time_format       %Y%m%dT%H%M%S%z
      </match>
      <match **>
        @type stdout
      </match>
  kind: ConfigMap
  metadata:
    name: ${NAME}
    namespace: ${NAMESPACE}
parameters:
- name: NAME
  value: fluentdserver
- name: NAMESPACE
  value: openshift-logging
`)

func testdataExternalLogStoresFluentdSecureCmServerauthYamlBytes() ([]byte, error) {
	return _testdataExternalLogStoresFluentdSecureCmServerauthYaml, nil
}

func testdataExternalLogStoresFluentdSecureCmServerauthYaml() (*asset, error) {
	bytes, err := testdataExternalLogStoresFluentdSecureCmServerauthYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "testdata/external-log-stores/fluentd/secure/cm-serverauth.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testdataExternalLogStoresFluentdSecureDeploymentYaml = []byte(`kind: Template
apiVersion: template.openshift.io/v1
metadata:
  name: external-log-store-template
objects:
- apiVersion: apps/v1
  kind: Deployment
  metadata:
    name: ${NAME}
    namespace: ${NAMESPACE}
    labels:
      provider: aosqe
      component: ${NAME}
      logging-infra: ${NAME}
  spec:
    replicas: 1
    revisionHistoryLimit: 10
    selector:
      matchLabels:
        provider: aosqe
        component: ${NAME}
        logging-infra: ${NAME}
    strategy:
      type: Recreate
    template:
      metadata:
        labels:
          logging-infra: ${NAME}
          provider: aosqe
          component: ${NAME}
      spec:
        securityContext:
          runAsNonRoot: true
          seccompProfile:
            type: RuntimeDefault
        containers:
        - name: "fluentdserver"
          image: "quay.io/openshifttest/fluentd:1.2.2"
          imagePullPolicy: "IfNotPresent"
          securityContext:
            allowPrivilegeEscalation: false
            runAsNonRoot: true
            capabilities:
              drop:
              - ALL
            seccompProfile:
              type: RuntimeDefault
          ports:
          - containerPort: 24224
            name: fluentdserver
          volumeMounts:
          - mountPath: /fluentd/etc
            name: config
            readOnly: true
          - mountPath: /etc/fluentd/secrets
            name: certs
            readOnly: true
        volumes:
        - configMap:
            defaultMode: 420
            name: ${NAME}
          name: config
        - name: certs
          secret:
            defaultMode: 420
            secretName: ${NAME}
parameters:
- name: NAME
  value: fluentdserver
- name: NAMESPACE
  value: openshift-logging
`)

func testdataExternalLogStoresFluentdSecureDeploymentYamlBytes() ([]byte, error) {
	return _testdataExternalLogStoresFluentdSecureDeploymentYaml, nil
}

func testdataExternalLogStoresFluentdSecureDeploymentYaml() (*asset, error) {
	bytes, err := testdataExternalLogStoresFluentdSecureDeploymentYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "testdata/external-log-stores/fluentd/secure/deployment.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testdataExternalLogStoresFluentdSecureHttpCmMtlsYaml = []byte(`kind: Template
apiVersion: template.openshift.io/v1
metadata:
  name: fluentd-http-template
objects:
- apiVersion: v1
  data:
    fluent.conf: |
      <system>
        log_level info
      </system>
      <source>
        @type http
        port 24224
        bind 0.0.0.0
        body_size_limit 32m
        keepalive_timeout 10s
        add_http_headers true
        add_remote_addr true
        <transport tls>
          ca_path /etc/fluentd/secrets/ca-bundle.crt
          cert_path /etc/fluentd/secrets/tls.crt
          private_key_path /etc/fluentd/secrets/tls.key
          client_cert_auth true
        </transport>
        <parse>
          @type json
        </parse>
        @label @collector_logs
      </source>
      <source>
        @type http
        port 24224
        bind "::"
        body_size_limit 32m
        keepalive_timeout 10s
        add_http_headers true
        add_remote_addr true
        <transport tls>
          ca_path /etc/fluentd/secrets/ca-bundle.crt
          cert_path /etc/fluentd/secrets/tls.crt
          private_key_path /etc/fluentd/secrets/tls.key
          client_cert_auth true
        </transport>
        <parse>
          @type json
        </parse>
        @label @collector_logs
      </source>
      <label @collector_logs>
        <match logs.app>
          @type file
          append true
          path /fluentd/log/app.*.log
          symlink_path /fluentd/log/app.log
        </match>
        <match logs.infra>
          @type file
          append true
          path /fluentd/log/infra.*.log
          symlink_path /fluentd/log/infra.log
        </match>
        <match logs.audit>
          @type file
          append true
          path /fluentd/log/audit.*.log
          symlink_path /fluentd/log/audit.log
        </match>
      </label>
      <label @FLUENT_LOG>
        <match **>
        	@type stdout
        </match>
      </label>
  kind: ConfigMap
  metadata:
    name: ${NAME}
    namespace: ${NAMESPACE}
parameters:
- name: NAME
  value: fluentdserver
- name: NAMESPACE
  value: openshift-logging
`)

func testdataExternalLogStoresFluentdSecureHttpCmMtlsYamlBytes() ([]byte, error) {
	return _testdataExternalLogStoresFluentdSecureHttpCmMtlsYaml, nil
}

func testdataExternalLogStoresFluentdSecureHttpCmMtlsYaml() (*asset, error) {
	bytes, err := testdataExternalLogStoresFluentdSecureHttpCmMtlsYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "testdata/external-log-stores/fluentd/secure/http-cm-mtls.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testdataExternalLogStoresFluentdSecureHttpCmServerauthYaml = []byte(`kind: Template
apiVersion: template.openshift.io/v1
metadata:
  name: fluentd-http-template
objects:
- apiVersion: v1
  data:
    fluent.conf: |
      <system>
        log_level info
      </system>
      <source>
        @type http
        port 24224
        bind 0.0.0.0
        body_size_limit 32m
        keepalive_timeout 10s
        add_http_headers true
        add_remote_addr true
        <transport tls>
          ca_path /etc/fluentd/secrets/ca-bundle.crt
          cert_path /etc/fluentd/secrets/tls.crt
          private_key_path /etc/fluentd/secrets/tls.key
          client_cert_auth false
        </transport>
        <parse>
          @type json
        </parse>
        @label @collector_logs
      </source>
      <source>
        @type http
        port 24224
        bind "::"
        body_size_limit 32m
        keepalive_timeout 10s
        add_http_headers true
        add_remote_addr true
        <transport tls>
          ca_path /etc/fluentd/secrets/ca-bundle.crt
          cert_path /etc/fluentd/secrets/tls.crt
          private_key_path /etc/fluentd/secrets/tls.key
          client_cert_auth false
        </transport>
        <parse>
          @type json
        </parse>
        @label @collector_logs
      </source>
      <label @collector_logs>
        <match logs.app>
          @type file
          append true
          path /fluentd/log/app.*.log
          symlink_path /fluentd/log/app.log
        </match>
        <match logs.infra>
          @type file
          append true
          path /fluentd/log/infra.*.log
          symlink_path /fluentd/log/infra.log
        </match>
        <match logs.audit>
          @type file
          append true
          path /fluentd/log/audit.*.log
          symlink_path /fluentd/log/audit.log
        </match>
      </label>
      <label @FLUENT_LOG>
        <match **>
        	@type stdout
        </match>
      </label>
  kind: ConfigMap
  metadata:
    name: ${NAME}
    namespace: ${NAMESPACE}
parameters:
- name: NAME
  value: fluentdserver
- name: NAMESPACE
  value: openshift-logging
`)

func testdataExternalLogStoresFluentdSecureHttpCmServerauthYamlBytes() ([]byte, error) {
	return _testdataExternalLogStoresFluentdSecureHttpCmServerauthYaml, nil
}

func testdataExternalLogStoresFluentdSecureHttpCmServerauthYaml() (*asset, error) {
	bytes, err := testdataExternalLogStoresFluentdSecureHttpCmServerauthYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "testdata/external-log-stores/fluentd/secure/http-cm-serverauth.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testdataExternalLogStoresKafkaAmqstreamsKafkaNoAuthClusterYaml = []byte(`kind: Template
apiVersion: template.openshift.io/v1
metadata:
  name: kafka-template
objects:
- apiVersion: kafka.strimzi.io/v1beta2
  kind: Kafka
  metadata:
    name: ${NAME}
  spec:
    kafka:
      replicas: 1
      version: ${VERSION}
      resources:
        requests:
          cpu: '64m'
          memory: 256Mi
        limits:
          memory: 4Gi
          cpu: "1"
      jvmOptions:
        '-Xms': 256m
        '-Xmx': 256m
      config:
        log.cleaner.enable: true
        log.segment.bytes: 268435456
        log.cleanup.policy: delete
        transaction.state.log.replication.factor: 1
        log.retention.bytes: 1073741824
        transaction.state.log.min.isr: 1
        log.retention.hours: 1
        auto.create.topics.enable: false
        offsets.topic.replication.factor: 1
      listeners:
      - name: plain
        port: 9092
        type: internal
        tls: false
        configuration:
          useServiceDnsDomain: true
      - name: tls
        port: 9093
        type: internal
        tls: true
        authentication:
          type: tls
      storage:
        type: ephemeral
    zookeeper:
      replicas: 1
      resources:
        limits:
          cpu: '1'
          memory: 2Gi
        requests:
          cpu: '64m'
          memory: 256Mi
      storage:
      storage:
        type: ephemeral
    entityOperator:
      topicOperator:
        reconciliationIntervalSeconds: 90
      userOperator:
        reconciliationIntervalSeconds: 120
parameters:
- name: NAME
  value: "my-cluster"
- name: VERSION
  value: "3.9.0"
`)

func testdataExternalLogStoresKafkaAmqstreamsKafkaNoAuthClusterYamlBytes() ([]byte, error) {
	return _testdataExternalLogStoresKafkaAmqstreamsKafkaNoAuthClusterYaml, nil
}

func testdataExternalLogStoresKafkaAmqstreamsKafkaNoAuthClusterYaml() (*asset, error) {
	bytes, err := testdataExternalLogStoresKafkaAmqstreamsKafkaNoAuthClusterYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "testdata/external-log-stores/kafka/amqstreams/kafka-no-auth-cluster.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testdataExternalLogStoresKafkaAmqstreamsKafkaNoAuthConsumerJobYaml = []byte(`kind: Template
apiVersion: template.openshift.io/v1
metadata:
  name: kafka-template
objects:
- apiVersion: batch/v1
  kind: Job
  metadata:
    name: ${NAME}
  spec:
    ttlSecondsAfterFinished: 60
    template:
      spec:
        containers:
        - name: kakfa-consumer
          image: "registry.redhat.io/amq7/amq-streams-kafka-25-rhel7@sha256:e719f662bd4d6b8c54b1ee2e47c51f8d75a27a238a51d9ee38007187b3a627a4"
          command: ["bin/kafka-console-consumer.sh","--bootstrap-server", "${CLUSTER_NAME}-kafka-bootstrap:9092", "--topic", "${TOPIC_NAME}", "--from-beginning"]
        restartPolicy: Never
        nodeSelector:
          kubernetes.io/arch: amd64
          kubernetes.io/os: linux
parameters:
- name: NAME
  value: "topic-logging-app-consumer"
- name: CLUSTER_NAME
  value: "my-cluster"
- name: TOPIC_NAME
  value: "topic-logging-app"
`)

func testdataExternalLogStoresKafkaAmqstreamsKafkaNoAuthConsumerJobYamlBytes() ([]byte, error) {
	return _testdataExternalLogStoresKafkaAmqstreamsKafkaNoAuthConsumerJobYaml, nil
}

func testdataExternalLogStoresKafkaAmqstreamsKafkaNoAuthConsumerJobYaml() (*asset, error) {
	bytes, err := testdataExternalLogStoresKafkaAmqstreamsKafkaNoAuthConsumerJobYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "testdata/external-log-stores/kafka/amqstreams/kafka-no-auth-consumer-job.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testdataExternalLogStoresKafkaAmqstreamsKafkaSaslClusterYaml = []byte(`kind: Template
apiVersion: template.openshift.io/v1
metadata:
  name: kafka-template
objects:
- apiVersion: kafka.strimzi.io/v1beta2
  kind: Kafka
  metadata:
    name: ${NAME}
  spec:
    entityOperator:
      topicOperator:
        reconciliationIntervalSeconds: 90
      userOperator:
        reconciliationIntervalSeconds: 120
    kafka:
      authorization:
        type: simple
      config:
        log.cleaner.enable: true
        log.segment.bytes: 268435456
        log.cleanup.policy: delete
        transaction.state.log.replication.factor: 1
        log.retention.bytes: 1073741824
        transaction.state.log.min.isr: 1
        log.retention.hours: 1
        auto.create.topics.enable: false
        offsets.topic.replication.factor: 1
      jvmOptions:
        '-Xms': 256m
        '-Xmx': 256m
      listeners:
        - authentication:
            type: scram-sha-512
          name: external
          port: 9093
          tls: true
          type: route
        - authentication:
            type: scram-sha-512
          configuration:
            useServiceDnsDomain: true
          name: plain
          port: 9092
          tls: false
          type: internal
      replicas: 1
      resources:
        limits:
          cpu: '2'
          memory: 4Gi
        requests:
          cpu: '64m'
          memory: 256Mi
      storage:
        type: ephemeral
      version: ${VERSION}
    zookeeper:
      jvmOptions:
        '-Xms': 256m
        '-Xmx': 256m
      replicas: 1
      resources:
        limits:
          cpu: '1'
          memory: 2Gi
        requests:
          cpu: '64m'
          memory: 256Mi
      storage:
        type: ephemeral
parameters:
- name: NAME
  value: "my-cluster"
- name: VERSION
  value: "3.9.0"
`)

func testdataExternalLogStoresKafkaAmqstreamsKafkaSaslClusterYamlBytes() ([]byte, error) {
	return _testdataExternalLogStoresKafkaAmqstreamsKafkaSaslClusterYaml, nil
}

func testdataExternalLogStoresKafkaAmqstreamsKafkaSaslClusterYaml() (*asset, error) {
	bytes, err := testdataExternalLogStoresKafkaAmqstreamsKafkaSaslClusterYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "testdata/external-log-stores/kafka/amqstreams/kafka-sasl-cluster.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testdataExternalLogStoresKafkaAmqstreamsKafkaSaslConsumerJobYaml = []byte(`kind: Template
apiVersion: template.openshift.io/v1
metadata:
  name: kafka-consumer-job-template
objects:
- apiVersion: batch/v1
  kind: Job
  metadata:
    name: ${NAME}
  spec:
    ttlSecondsAfterFinished: 60
    template:
      spec:
        containers:
        - name: kafka-consumer
          image: registry.redhat.io/amq-streams/kafka-35-rhel8@sha256:fc900527fa19b35ec909c2a44e9e22ff0119934dcdf6e5da3665631d724a1bf4
          command: ["bin/kafka-console-consumer.sh","--bootstrap-server=${CLUSTER_NAME}-kafka-bootstrap:9092","--topic=${TOPIC_NAME}","--consumer.config=/opt/kafka/qeclient/client.property", "--from-beginning"]
          volumeMounts:
          - mountPath: /opt/kafka/qeclient
            name: kafka-config
          - mountPath: /opt/kafka/qep12
            name: cluster-ca
        restartPolicy: Never
        volumes:
        - configMap:
            defaultMode: 420
            name: ${CLIENT_CONFIGMAP_NAME}
          name: kafka-config
        - name: cluster-ca
          secret:
            defaultMode: 288
            secretName: ${CA_SECRET_NAME}
parameters:
- name: NAME
  value: "topic-logging-consumer"
- name: CLUSTER_NAME
  value: "my-cluster"
- name: TOPIC_NAME
  value: "topic-logging"
- name: CLIENT_CONFIGMAP_NAME
  value: ""
- name: CA_SECRET_NAME
  value: ""
`)

func testdataExternalLogStoresKafkaAmqstreamsKafkaSaslConsumerJobYamlBytes() ([]byte, error) {
	return _testdataExternalLogStoresKafkaAmqstreamsKafkaSaslConsumerJobYaml, nil
}

func testdataExternalLogStoresKafkaAmqstreamsKafkaSaslConsumerJobYaml() (*asset, error) {
	bytes, err := testdataExternalLogStoresKafkaAmqstreamsKafkaSaslConsumerJobYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "testdata/external-log-stores/kafka/amqstreams/kafka-sasl-consumer-job.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testdataExternalLogStoresKafkaAmqstreamsKafkaSaslConsumersConfigYaml = []byte(`kind: Template
apiVersion: template.openshift.io/v1
metadata:
  name: kafka-client-config-template
objects:
- apiVersion: v1
  data:
    client.property: |-
      security.protocol=SASL_PLAINTEXT
      sasl.mechanism=SCRAM-SHA-512
      sasl.jaas.config=org.apache.kafka.common.security.scram.ScramLoginModule required username="${USER}" password="${PASSWORD}";
      bootstrap.servers=${KAFKA_NAME}-kafka-bootstrap:9092
      ssl.truststore.location=/opt/kafka/qep12/ca.p12
      ssl.truststore.password=${TRUSTSTORE_PASSWORD}
      ssl.truststore.type=PKCS12 
      group.id=my-group
  kind: ConfigMap
  metadata:
    name: ${NAME}
parameters:
- name: NAME
  value: "client-property"
- name: USER
  value: "my-user"
- name: PASSWORD
  value: ""
- name: TRUSTSTORE_PASSWORD
  value: ""
- name: KAFKA_NAME
  value: "my-cluster"
`)

func testdataExternalLogStoresKafkaAmqstreamsKafkaSaslConsumersConfigYamlBytes() ([]byte, error) {
	return _testdataExternalLogStoresKafkaAmqstreamsKafkaSaslConsumersConfigYaml, nil
}

func testdataExternalLogStoresKafkaAmqstreamsKafkaSaslConsumersConfigYaml() (*asset, error) {
	bytes, err := testdataExternalLogStoresKafkaAmqstreamsKafkaSaslConsumersConfigYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "testdata/external-log-stores/kafka/amqstreams/kafka-sasl-consumers-config.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testdataExternalLogStoresKafkaAmqstreamsKafkaSaslUserYaml = []byte(`kind: Template
apiVersion: template.openshift.io/v1
metadata:
  name: kafka-user-template
objects:
- apiVersion: kafka.strimzi.io/v1beta2
  kind: KafkaUser
  metadata:
    name: ${NAME}
    labels:
      strimzi.io/cluster: ${KAFKA_NAME}
  spec:
    authentication:
      type: scram-sha-512
    authorization:
      acls:
        - host: '*'
          operations:
            - Read
            - Describe
            - Write
            - Create
          resource:
            name: ${TOPIC_PREFIX}
            patternType: prefix
            type: topic
        - host: '*'
          operations:
            - Read
          resource:
            name: my-group
            patternType: literal
            type: group
      type: simple
parameters:
- name: NAME
  value: "my-user"
- name:  KAFKA_NAME
  value: "my-cluster"
- name: TOPIC_PREFIX
  value: "topic-logging"
`)

func testdataExternalLogStoresKafkaAmqstreamsKafkaSaslUserYamlBytes() ([]byte, error) {
	return _testdataExternalLogStoresKafkaAmqstreamsKafkaSaslUserYaml, nil
}

func testdataExternalLogStoresKafkaAmqstreamsKafkaSaslUserYaml() (*asset, error) {
	bytes, err := testdataExternalLogStoresKafkaAmqstreamsKafkaSaslUserYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "testdata/external-log-stores/kafka/amqstreams/kafka-sasl-user.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testdataExternalLogStoresKafkaAmqstreamsKafkaTopicYaml = []byte(`kind: Template
apiVersion: template.openshift.io/v1
metadata:
  name: kafka-template
objects:
- apiVersion: kafka.strimzi.io/v1beta1
  kind: KafkaTopic
  metadata:
    name: ${NAME}
    namespace: ${NAMESPACE}
    labels:
      strimzi.io/cluster: ${CLUSTER_NAME}
  spec:
    partitions: 1
    replicas: 1
    config:
      retention.ms: 300000
      segment.bytes: 1073741824
parameters:
- name: CLUSTER_NAME
  value: "my-cluster"
- name: NAME
  value: "logging-topic-all"
- name: NAMESPACE
  value: "amq-aosqe"
`)

func testdataExternalLogStoresKafkaAmqstreamsKafkaTopicYamlBytes() ([]byte, error) {
	return _testdataExternalLogStoresKafkaAmqstreamsKafkaTopicYaml, nil
}

func testdataExternalLogStoresKafkaAmqstreamsKafkaTopicYaml() (*asset, error) {
	bytes, err := testdataExternalLogStoresKafkaAmqstreamsKafkaTopicYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "testdata/external-log-stores/kafka/amqstreams/kafka-topic.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testdataExternalLogStoresKafkaCert_generationSh = []byte(`#!/usr/bin/env bash
set -eou pipefail

WORKING_DIR=${1:-/tmp/_working_dir}
NAMESPACE=${2:-openshift-logging}
CA_PATH=$WORKING_DIR/ca
cn_name="aosqeca"
CLIENT_CA_PATH="${WORKING_DIR}/client"
CLUSTER_CA_PATH="${WORKING_DIR}/cluster"
REGENERATE_NEEDED=0

function init_cert_files() {
if [ ! -d $CA_PATH ]; then
    mkdir $CA_PATH
fi
if [ ! -d $CLIENT_CA_PATH ]; then
    mkdir ${CLIENT_CA_PATH}
fi
if [ ! -d $CLUSTER_CA_PATH ]; then
    mkdir ${CLUSTER_CA_PATH}
fi
}

function create_root_ca() {
#fqdn=my-cluster-kafka-bootstrap.amq-aosqe.svc
#1)  generate root ca key and and self cert
openssl req -x509 -new -newkey rsa:4096 -keyout $CA_PATH/root_ca.key -nodes -days 1825 -out $CA_PATH/ca_bundle.crt -subj "/CN=aosqeroot"   -passin pass:aosqe2021 -passout pass:aosqe2021

# Create trustore jks
/usr/lib/jvm/jre-openjdk/bin/keytool -import -file $CA_PATH/ca_bundle.crt -keystore $CA_PATH/ca_bundle.jks  --srcstorepass aosqe2021 --deststorepass aosqe2021 -noprompt || exit 1
}

# Create Client CSR
function create_client_sign() {
cat <<EOF >$CLIENT_CA_PATH/client_csr.conf
[ req ]
default_bits = 4096
prompt = no
encrypt_key = yes
default_md = sha512
distinguished_name = dn
req_extensions = server_ext

[ dn ]
CN = "aosqeclient"

[ server_ext ]
basicConstraints        = CA:false
extendedKeyUsage        = serverAuth,clientAuth
subjectAltName = DNS.1:*.cluster.local,DNS.2:*.svc,DNS.3:*.pod
EOF
openssl req -new -out $CLIENT_CA_PATH/client.csr -newkey rsa:4096 -keyout $CLIENT_CA_PATH/client.key -config $CLIENT_CA_PATH/client_csr.conf -nodes


# Sign client ca by intermediate ca
cat <<EOF >$CLIENT_CA_PATH/client_sign.conf
# Simple Signing CA

# The [default] section contains global constants that can be referred to from
# the entire configuration file. It may also hold settings pertaining to more
# than one openssl command.

[ default ]
dir                     = $CLIENT_CA_PATH             # Top dir

# The remainder of the configuration file is used by the openssl ca command.
# The CA section defines the locations of CA assets, as well as the policies
# applying to the CA.

[ ca ]
default_ca              = signing_ca            # The default CA section

[ signing_ca ]
certificate             = $CA_PATH/ca_bundle.crt       # The CA cert
private_key             = $CA_PATH/root_ca.key # CA private key
new_certs_dir           = $CA_PATH/           # Certificate archive
serial                  = $CA_PATH/root_ca.serial.txt # Serial number file
crlnumber               = $CA_PATH/root_ca.crl.srl # CRL number file
database                = $CA_PATH/root_ca.db # Index file
unique_subject          = no                    # Require unique subject
default_days            = 730                   # How long to certify for
default_md              = sha512                # MD to use
policy                  = any_pol             # Default naming policy
email_in_dn             = no                    # Add email to cert DN
preserve                = no                    # Keep passed DN ordering
name_opt                = ca_default            # Subject DN display options
cert_opt                = ca_default            # Certificate display options
copy_extensions         = copy                  # Copy extensions from CSR
x509_extensions         = server_ext             # Default cert extensions
default_crl_days        = 7                     # How long before next CRL

# Naming policies control which parts of a DN end up in the certificate and
# under what circumstances certification should be denied.

[ any_pol ]
domainComponent         = optional
countryName             = optional
stateOrProvinceName     = optional
localityName            = optional
organizationName        = optional
organizationalUnitName  = optional
commonName              = optional
emailAddress            = optional

# Certificate extensions define what types of certificates the CA is able to
# create.

[ server_ext ]
basicConstraints        = CA:false
extendedKeyUsage        = serverAuth,clientAuth

[ ca_reqext ]
basicConstraints        = CA:false


# CRL extensions exist solely to point to the CA certificate that has issued
# the CRL.
EOF

touch $CA_PATH/root_ca.db
if [ ! -f $CA_PATH/root_ca.serial.txt ] ; then
    echo "01">$CA_PATH/root_ca.serial.txt
fi
openssl ca -in $CLIENT_CA_PATH/client.csr -notext -out $CLIENT_CA_PATH/client.crt -config $CLIENT_CA_PATH/client_sign.conf -batch

# Create Client keystone
openssl pkcs12 -export -in $CLIENT_CA_PATH/client.crt -inkey $CLIENT_CA_PATH/client.key -out $CLIENT_CA_PATH/client.pkcs12  -passin pass:aosqe2021 -passout pass:aosqe2021
/usr/lib/jvm/jre-openjdk/bin/keytool -importkeystore -srckeystore $CLIENT_CA_PATH/client.pkcs12 -srcstoretype PKCS12 -destkeystore $CLIENT_CA_PATH/client.jks -deststoretype JKS --srcstorepass aosqe2021 --deststorepass aosqe2021 -noprompt
}

# Create cluster csr
# https://support.dnsimple.com/articles/ssl-certificate-names/
function create_cluster_sign() {
cat <<EOF >$CLUSTER_CA_PATH/cluster_csr.conf
[ req ]
default_bits = 4096
prompt = no
encrypt_key = yes
default_md = sha512
distinguished_name = dn
req_extensions = server_ext

[ dn ]
CN = "aosqecluster"

[ server_ext ]
basicConstraints        = CA:false
extendedKeyUsage        = serverAuth,clientAuth
subjectAltName = DNS.1:kafka.${NAMESPACE}.svc.cluster.local,DNS.2:kafka.${NAMESPACE}.svc,DNS.3:kafka-0.kafka.${NAMESPACE}.svc.cluster.local,DNS.4: kafka-0.kafka.${NAMESPACE}.svc, DNS.5: kafka, DNS.6: kakfa-0
EOF
openssl req -new -out $CLUSTER_CA_PATH/cluster.csr -newkey rsa:4096 -keyout $CLUSTER_CA_PATH/cluster.key -config $CLUSTER_CA_PATH/cluster_csr.conf -nodes


cat <<EOF >$CLUSTER_CA_PATH/cluster_sign.conf
# Simple Signing CA

# The [default] section contains global constants that can be referred to from
# the entire configuration file. It may also hold settings pertaining to more
# than one openssl command.

[ default ]
dir                     = $CA_PATH             # Top dir

# The next part of the configuration file is used by the openssl req command.
# It defines the CA's key pair, its DN, and the desired extensions for the CA
# certificate.

[ req ]
default_bits            = 4096                  # RSA key size
encrypt_key             = yes                   # Protect private key
default_md              = sha512                # MD to use
utf8                    = yes                   # Input is UTF-8
string_mask             = utf8only              # Emit UTF-8 strings
prompt                  = no                    # Don't prompt for DN
distinguished_name      = ca_dn                 # DN section
req_extensions          = ca_reqext             # Desired extensions

[ ca_dn ]
commonName              = "aosqeintermediate"

[ ca_reqext ]
basicConstraints        = CA:false

# The remainder of the configuration file is used by the openssl ca command.
# The CA section defines the locations of CA assets, as well as the policies
# applying to the CA.

[ ca ]
default_ca              = signing_ca            # The default CA section

[ signing_ca ]
certificate             = $CA_PATH/ca_bundle.crt       # The CA cert
private_key             = $CA_PATH/root_ca.key # CA private key
new_certs_dir           = $CA_PATH/           # Certificate archive
serial                  = $CA_PATH/root_ca.serial.txt # Serial number file
crlnumber               = $CA_PATH/root_ca.crl.srl # CRL number file
database                = $CA_PATH/root_ca.db # Index file
unique_subject          = no                    # Require unique subject
default_days            = 730                   # How long to certify for
default_md              = sha512                # MD to use
policy                  = any_pol             # Default naming policy
email_in_dn             = no                    # Add email to cert DN
preserve                = no                    # Keep passed DN ordering
name_opt                = ca_default            # Subject DN display options
cert_opt                = ca_default            # Certificate display options
copy_extensions         = copy                  # Copy extensions from CSR
#x509_extensions         = server_ext             # Default cert extensions
default_crl_days        = 7                     # How long before next CRL

# Naming policies control which parts of a DN end up in the certificate and
# under what circumstances certification should be denied.

[ match_pol ]
domainComponent         = match                 # Must match 'simple.org'
organizationName        = match                 # Must match 'Simple Inc'
organizationalUnitName  = optional              # Included if present
commonName              = supplied              # Must be present

[ any_pol ]
domainComponent         = optional
countryName             = optional
stateOrProvinceName     = optional
localityName            = optional
organizationName        = optional
organizationalUnitName  = optional
commonName              = optional
emailAddress            = optional

# Certificate extensions define what types of certificates the CA is able to
# create.

[ server_ext ]
basicConstraints        = CA:false
extendedKeyUsage        = serverAuth,clientAuth
subjectAltName = DNS.1:kafka.${NAMESPACE}.svc.cluster.local,DNS.2:kafka.${NAMESPACE}.svc,DNS.3:kafka-0.kafka.${NAMESPACE}.svc.cluster.local,DNS.4: kafka-0.kafka.${NAMESPACE}.svc, DNS.5: kafka, DNS.6: kakfa-0
EOF

touch $CA_PATH/root_ca.db
if [ ! -f $CA_PATH/root_ca.serial.txt ] ; then
    echo "01">$CA_PATH/root_ca.serial.txt
fi
openssl ca -in $CLUSTER_CA_PATH/cluster.csr -notext -out $CLUSTER_CA_PATH/cluster.crt -config $CLUSTER_CA_PATH/cluster_sign.conf -batch

#Create keystone
openssl pkcs12 -export -in $CLUSTER_CA_PATH/cluster.crt -inkey $CLUSTER_CA_PATH/cluster.key -out $CLUSTER_CA_PATH/cluster.pkcs12  -passin pass:aosqe2021 -passout pass:aosqe2021
/usr/lib/jvm/jre-openjdk/bin/keytool -importkeystore -srckeystore $CLUSTER_CA_PATH/cluster.pkcs12 -srcstoretype PKCS12 -destkeystore $CLUSTER_CA_PATH/cluster.jks -deststoretype JKS  --srcstorepass aosqe2021 --deststorepass aosqe2021 -noprompt
}

init_cert_files
create_root_ca
create_client_sign
create_cluster_sign
`)

func testdataExternalLogStoresKafkaCert_generationShBytes() ([]byte, error) {
	return _testdataExternalLogStoresKafkaCert_generationSh, nil
}

func testdataExternalLogStoresKafkaCert_generationSh() (*asset, error) {
	bytes, err := testdataExternalLogStoresKafkaCert_generationShBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "testdata/external-log-stores/kafka/cert_generation.sh", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testdataExternalLogStoresKafkaKafkaRbacYaml = []byte(`kind: Template
apiVersion: template.openshift.io/v1
metadata:
  name: kafkaserver-template
objects:
- apiVersion: rbac.authorization.k8s.io/v1
  kind: ClusterRole
  metadata:
    name: ${NAME}
  rules:
  - apiGroups:
    - ""
    resources:
    - nodes
    verbs:
    - get
  - apiGroups:
    - ""
    resources:
    - pods
    verbs:
    - get
    - create
    - update
    - patch
    - delete
- apiVersion: rbac.authorization.k8s.io/v1
  kind: ClusterRoleBinding
  metadata:
    name: ${NAME}
  roleRef:
    apiGroup: rbac.authorization.k8s.io
    kind: ClusterRole
    name: ${NAME}
  subjects:
  - kind: ServiceAccount
    name: default
    namespace: ${NAMESPACE}
parameters:
- name: NAME
  value: "kafka-node-reader"
- name: NAMESPACE
  value: "openshift-logging"
`)

func testdataExternalLogStoresKafkaKafkaRbacYamlBytes() ([]byte, error) {
	return _testdataExternalLogStoresKafkaKafkaRbacYaml, nil
}

func testdataExternalLogStoresKafkaKafkaRbacYaml() (*asset, error) {
	bytes, err := testdataExternalLogStoresKafkaKafkaRbacYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "testdata/external-log-stores/kafka/kafka-rbac.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testdataExternalLogStoresKafkaKafkaSvcYaml = []byte(`kind: Template
apiVersion: template.openshift.io/v1
metadata:
  name: kafkaserver-template
objects:
- apiVersion: v1
  kind: Service
  metadata:
    labels:
      logging-infra: support
    name: ${NAME}
    namespace: ${NAMESPACE}
  spec:
    ports:
    - name: plaintext
      port: 9092
      protocol: TCP
      targetPort: 9092
    - name: saslplaintext
      port: 9093
      protocol: TCP
      targetPort: 9093
    - name: slasssl
      port: 9094
      protocol: TCP
      targetPort: 9093
    selector:
      component: kafka
      provider: openshift
    sessionAffinity: None
    type: ClusterIP
parameters:
- name: NAME
  value: "kafka"
- name: NAMESPACE
  value: "openshift-logging"
`)

func testdataExternalLogStoresKafkaKafkaSvcYamlBytes() ([]byte, error) {
	return _testdataExternalLogStoresKafkaKafkaSvcYaml, nil
}

func testdataExternalLogStoresKafkaKafkaSvcYaml() (*asset, error) {
	bytes, err := testdataExternalLogStoresKafkaKafkaSvcYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "testdata/external-log-stores/kafka/kafka-svc.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testdataExternalLogStoresKafkaPlaintextSslConsumerConfigmapYaml = []byte(`kind: Template
apiVersion: template.openshift.io/v1
metadata:
  name: consumer-configmap-template
objects:
- apiVersion: v1
  data:
    client.properties: |
      bootstrap.servers=kafka:9093
      #group.id=test-consumer-group
      security.protocol=SSL
      ssl.truststore.location=/etc/kafkacert/ca-bundle.jks
      ssl.truststore.password=aosqe2021
  kind: ConfigMap
  metadata:
    name: ${NAME}
    namespace: ${NAMESPACE}
parameters:
- name: NAME
  value: "kafka-client"
- name: NAMESPACE
  value: "openshift-logging"
`)

func testdataExternalLogStoresKafkaPlaintextSslConsumerConfigmapYamlBytes() ([]byte, error) {
	return _testdataExternalLogStoresKafkaPlaintextSslConsumerConfigmapYaml, nil
}

func testdataExternalLogStoresKafkaPlaintextSslConsumerConfigmapYaml() (*asset, error) {
	bytes, err := testdataExternalLogStoresKafkaPlaintextSslConsumerConfigmapYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "testdata/external-log-stores/kafka/plaintext-ssl/consumer-configmap.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testdataExternalLogStoresKafkaPlaintextSslKafkaConfigmapYaml = []byte(`kind: Template
apiVersion: template.openshift.io/v1
metadata:
  name: consumer-configmap-template
objects:
- apiVersion: v1
  data:
    init.sh: |
      #!/bin/bash
      set -e
      cp /etc/kafka-configmap/log4j.properties /etc/kafka/

      KAFKA_BROKER_ID=${HOSTNAME##*-}
      SEDS=("s/#init#broker.id=#init#/broker.id=$KAFKA_BROKER_ID/")
      LABELS="kafka-broker-id=$KAFKA_BROKER_ID"
      ANNOTATIONS=""

      hash kubectl 2>/dev/null || {
        SEDS+=("s/#init#broker.rack=#init#/#init#broker.rack=# kubectl not found in path/")
      } && {
        ZONE=$(kubectl get node "$NODE_NAME" -o=go-template='{{index .metadata.labels "failure-domain.beta.kubernetes.io/zone"}}')
        if [ "x$ZONE" == "x<no value>" ]; then
          SEDS+=("s/#init#broker.rack=#init#/#init#broker.rack=# zone label not found for node $NODE_NAME/")
        else
          SEDS+=("s/#init#broker.rack=#init#/broker.rack=$ZONE/")
          LABELS="$LABELS kafka-broker-rack=$ZONE"
        fi

        [ -z "$ADVERTISE_ADDR" ] && echo "ADVERTISE_ADDR is empty, will advertise detected DNS name"
        OUTSIDE_HOST=$(kubectl get node "$NODE_NAME" -o jsonpath='{.status.addresses[?(@.type=="InternalIP")].address}')
        OUTSIDE_PORT=$((32400 + ${KAFKA_BROKER_ID}))
        SEDS+=("s|#init#advertised.listeners=PLAINTEXT://#init#|advertised.listeners=PLAINTEXT://${ADVERTISE_ADDR}:9092,OUTSIDE://${OUTSIDE_HOST}:${OUTSIDE_PORT}|")
        ANNOTATIONS="$ANNOTATIONS kafka-listener-outside-host=$OUTSIDE_HOST kafka-listener-outside-port=$OUTSIDE_PORT"

        if [ ! -z "$LABELS" ]; then
          kubectl -n $POD_NAMESPACE label pod $POD_NAME $LABELS || echo "Failed to label $POD_NAMESPACE.$POD_NAME - RBAC issue?"
        fi
        if [ ! -z "$ANNOTATIONS" ]; then
          kubectl -n $POD_NAMESPACE annotate pod $POD_NAME $ANNOTATIONS || echo "Failed to annotate $POD_NAMESPACE.$POD_NAME - RBAC issue?"
        fi
      }
      printf '%s\n' "${SEDS[@]}" | sed -f - /etc/kafka-configmap/server.properties > /etc/kafka/server.properties.tmp
      [ $? -eq 0 ] && mv /etc/kafka/server.properties.tmp /etc/kafka/server.properties
    log4j.properties: |
      # Unspecified loggers and loggers with additivity=true output to server.log and stdout
      # Note that INFO only applies to unspecified loggers, the log level of the child logger is used otherwise
      log4j.rootLogger=INFO, stdout

      log4j.appender.stdout=org.apache.log4j.ConsoleAppender
      log4j.appender.stdout.layout=org.apache.log4j.PatternLayout
      log4j.appender.stdout.layout.ConversionPattern=[%d] %p %m (%c)%n

      log4j.appender.kafkaAppender=org.apache.log4j.DailyRollingFileAppender
      log4j.appender.kafkaAppender.DatePattern='.'yyyy-MM-dd-HH
      log4j.appender.kafkaAppender.File=${kafka.logs.dir}/server.log
      log4j.appender.kafkaAppender.layout=org.apache.log4j.PatternLayout
      log4j.appender.kafkaAppender.layout.ConversionPattern=[%d] %p %m (%c)%n

      log4j.appender.stateChangeAppender=org.apache.log4j.DailyRollingFileAppender
      log4j.appender.stateChangeAppender.DatePattern='.'yyyy-MM-dd-HH
      log4j.appender.stateChangeAppender.File=${kafka.logs.dir}/state-change.log
      log4j.appender.stateChangeAppender.layout=org.apache.log4j.PatternLayout
      log4j.appender.stateChangeAppender.layout.ConversionPattern=[%d] %p %m (%c)%n

      log4j.appender.requestAppender=org.apache.log4j.DailyRollingFileAppender
      log4j.appender.requestAppender.DatePattern='.'yyyy-MM-dd-HH
      log4j.appender.requestAppender.File=${kafka.logs.dir}/kafka-request.log
      log4j.appender.requestAppender.layout=org.apache.log4j.PatternLayout
      log4j.appender.requestAppender.layout.ConversionPattern=[%d] %p %m (%c)%n

      log4j.appender.cleanerAppender=org.apache.log4j.DailyRollingFileAppender
      log4j.appender.cleanerAppender.DatePattern='.'yyyy-MM-dd-HH
      log4j.appender.cleanerAppender.File=${kafka.logs.dir}/log-cleaner.log
      log4j.appender.cleanerAppender.layout=org.apache.log4j.PatternLayout
      log4j.appender.cleanerAppender.layout.ConversionPattern=[%d] %p %m (%c)%n

      log4j.appender.controllerAppender=org.apache.log4j.DailyRollingFileAppender
      log4j.appender.controllerAppender.DatePattern='.'yyyy-MM-dd-HH
      log4j.appender.controllerAppender.File=${kafka.logs.dir}/controller.log
      log4j.appender.controllerAppender.layout=org.apache.log4j.PatternLayout
      log4j.appender.controllerAppender.layout.ConversionPattern=[%d] %p %m (%c)%n

      log4j.appender.authorizerAppender=org.apache.log4j.DailyRollingFileAppender
      log4j.appender.authorizerAppender.DatePattern='.'yyyy-MM-dd-HH
      log4j.appender.authorizerAppender.File=${kafka.logs.dir}/kafka-authorizer.log
      log4j.appender.authorizerAppender.layout=org.apache.log4j.PatternLayout
      log4j.appender.authorizerAppender.layout.ConversionPattern=[%d] %p %m (%c)%n

      # Change the two lines below to adjust ZK client logging
      log4j.logger.org.I0Itec.zkclient.ZkClient=INFO
      log4j.logger.org.apache.zookeeper=INFO

      # Change the two lines below to adjust the general broker logging level (output to server.log and stdout)
      log4j.logger.kafka=INFO
      log4j.logger.org.apache.kafka=INFO

      # Change to DEBUG or TRACE to enable request logging
      log4j.logger.kafka.request.logger=WARN, requestAppender
      log4j.additivity.kafka.request.logger=false

      # Uncomment the lines below and change log4j.logger.kafka.network.RequestChannel$ to TRACE for additional output
      # related to the handling of requests
      #log4j.logger.kafka.network.Processor=TRACE, requestAppender
      #log4j.logger.kafka.server.KafkaApis=TRACE, requestAppender
      #log4j.additivity.kafka.server.KafkaApis=false
      log4j.logger.kafka.network.RequestChannel$=WARN, requestAppender
      log4j.additivity.kafka.network.RequestChannel$=false

      log4j.logger.kafka.controller=TRACE, controllerAppender
      log4j.additivity.kafka.controller=false

      log4j.logger.kafka.log.LogCleaner=INFO, cleanerAppender
      log4j.additivity.kafka.log.LogCleaner=false

      log4j.logger.state.change.logger=TRACE, stateChangeAppender
      log4j.additivity.state.change.logger=false

      # Change to DEBUG to enable audit log for the authorizer
      log4j.logger.kafka.authorizer.logger=WARN, authorizerAppender
      log4j.additivity.kafka.authorizer.logger=false
    server.properties: |
      #init#broker.id=#init#
      listeners=PLAINTEXT://:9092,SSL://:9093
      ssl.client.auth=requested
      ssl.keystore.location=/etc/kafkacert/cluster.jks
      ssl.keystore.password=aosqe2021
      ssl.truststore.location=/etc/kafkacert/ca_bundle.jks
      ssl.truststore.password=aosqe2021
      security.inter.broker.protocol=PLAINTEXT
      num.network.threads=3
      num.io.threads=8
      message.max.bytes=314572800
      socket.send.buffer.bytes=102400
      socket.receive.buffer.bytes=102400
      socket.request.max.bytes=104857600
      socket.request.max.bytes=314572800
      log.dirs=/tmp/kafka-logs
      num.partitions=1
      num.recovery.threads.per.data.dir=1
      offsets.topic.replication.factor=1
      transaction.state.log.replication.factor=1
      transaction.state.log.min.isr=1
      log.retention.hours=2
      log.segment.bytes=1073741824
      log.retention.check.interval.ms=300000
      zookeeper.connect=zookeeper:2181
      zookeeper.connection.timeout.ms=18000
      group.initial.rebalance.delay.ms=0
  kind: ConfigMap
  metadata:
    name: ${NAME}
    namespace: ${NAMESPACE}
parameters:
- name: NAME
  value: "kafka"
- name: NAMESPACE
  value: "openshift-logging"
`)

func testdataExternalLogStoresKafkaPlaintextSslKafkaConfigmapYamlBytes() ([]byte, error) {
	return _testdataExternalLogStoresKafkaPlaintextSslKafkaConfigmapYaml, nil
}

func testdataExternalLogStoresKafkaPlaintextSslKafkaConfigmapYaml() (*asset, error) {
	bytes, err := testdataExternalLogStoresKafkaPlaintextSslKafkaConfigmapYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "testdata/external-log-stores/kafka/plaintext-ssl/kafka-configmap.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testdataExternalLogStoresKafkaPlaintextSslKafkaConsumerDeploymentYaml = []byte(`kind: Template
apiVersion: template.openshift.io/v1
metadata:
  name: kafkaserver-template
objects:
- apiVersion: apps/v1
  kind: Deployment
  metadata:
    annotations:
      deployment.kubernetes.io/revision: "1"
    labels:
      component: kafka-consumer
      logging-infra: kafka
      provider: openshift
    name: ${NAME}
    namespace: ${NAMESPACE}
  spec:
    progressDeadlineSeconds: 600
    replicas: 1
    revisionHistoryLimit: 10
    selector:
      matchLabels:
        component: kafka-consumer
        logging-infra: kafka
        provider: openshift
    strategy:
      rollingUpdate:
        maxSurge: 25%
        maxUnavailable: 25%
      type: RollingUpdate
    template:
      metadata:
        creationTimestamp: null
        labels:
          component: kafka-consumer
          logging-infra: kafka
          provider: openshift
        name: kafka-consumer
      spec:
        securityContext:
          runAsNonRoot: true
          seccompProfile:
            type: RuntimeDefault
        containers:
        - command:
          - /bin/bash
          - -ce
          - /opt/kafka/bin/kafka-console-consumer.sh --bootstrap-server kafka:9093 --topic clo-topic --from-beginning  --consumer.config /etc/kafka-config/client.properties
          image: quay.io/openshifttest/kafka@sha256:2411662d89dd5700e1fe49aa8be1219843948da90cfe51a1c7a49bcef9d22dab
          imagePullPolicy: IfNotPresent
          name: kafka-consumer
          securityContext:
            allowPrivilegeEscalation: false
            runAsNonRoot: true
            capabilities:
              drop:
              - ALL
            seccompProfile:
              type: RuntimeDefault
          resources: {}
          terminationMessagePath: /dev/termination-log
          terminationMessagePolicy: File
          volumeMounts:
          - mountPath: /shared
            name: shared
          - mountPath: /etc/kafka-config
            name: kafka-client
          - mountPath: /etc/kafkacert
            name: kafkacert
          env:
          - name: KAFKA_OPTS
            value: -Djava.security.auth.login.config=/etc/kafka-configmap/kafka_client_jaas.conf
        dnsPolicy: ClusterFirst
        nodeSelector:
          kubernetes.io/arch: amd64
          kubernetes.io/os: linux
        restartPolicy: Always
        schedulerName: default-scheduler
        terminationGracePeriodSeconds: 30
        volumes:
        - emptyDir: {}
          name: shared
        - configMap:
            defaultMode: 420
            name: ${CM_NAME}
          name: kafka-client
        - secret:
            defaultMode: 420
            secretName: ${SECRETNAME}
          name: kafkacert
parameters:
- name: NAME
  value: "kafka-consumer-plaintext-ssl"
- name: NAMESPACE
  value: "openshift-logging"
- name: CM_NAME
  value: "kafka-client"
- name: SECRETNAME
  value: "kafka-client-cert"
`)

func testdataExternalLogStoresKafkaPlaintextSslKafkaConsumerDeploymentYamlBytes() ([]byte, error) {
	return _testdataExternalLogStoresKafkaPlaintextSslKafkaConsumerDeploymentYaml, nil
}

func testdataExternalLogStoresKafkaPlaintextSslKafkaConsumerDeploymentYaml() (*asset, error) {
	bytes, err := testdataExternalLogStoresKafkaPlaintextSslKafkaConsumerDeploymentYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "testdata/external-log-stores/kafka/plaintext-ssl/kafka-consumer-deployment.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testdataExternalLogStoresKafkaPlaintextSslKafkaStatefulsetYaml = []byte(`kind: Template
apiVersion: template.openshift.io/v1
metadata:
  name: kafkaserver-template
objects:
- apiVersion: apps/v1
  kind: StatefulSet
  metadata:
    labels:
      app: kafka
      component: kafka
      provider: openshift
    name: ${NAME}
    namespace: ${NAMESPACE}
  spec:
    podManagementPolicy: Parallel
    replicas: 1
    revisionHistoryLimit: 10
    selector:
      matchLabels:
        app: kafka
    serviceName: ${SERVICENAME}
    template:
      metadata:
        creationTimestamp: null
        labels:
          app: kafka
          component: kafka
          provider: openshift
      spec:
        securityContext:
          runAsNonRoot: true
          seccompProfile:
            type: RuntimeDefault
        containers:
        - command:
          - /opt/kafka/bin/kafka-server-start.sh
          - /etc/kafka/server.properties
          env:
          - name: CLASSPATH
            value: /opt/kafka/libs/extensions/*
          - name: KAFKA_LOG4J_OPTS
            value: -Dlog4j.configuration=file:/etc/kafka/log4j.properties
          - name: JMX_PORT
            value: "5555"
          image: quay.io/openshifttest/kafka@sha256:2411662d89dd5700e1fe49aa8be1219843948da90cfe51a1c7a49bcef9d22dab
          imagePullPolicy: IfNotPresent
          securityContext:
            allowPrivilegeEscalation: false
            runAsNonRoot: true
            capabilities:
              drop:
              - ALL
            seccompProfile:
              type: RuntimeDefault
          lifecycle:
            preStop:
              exec:
                command:
                - sh
                - -ce
                - kill -s TERM 1; while $(kill -0 1 2>/dev/null); do sleep 1; done
          name: broker
          ports:
          - containerPort: 9092
            name: inside
            protocol: TCP
          - containerPort: 9093
            name: ssl
            protocol: TCP
          - containerPort: 9094
            name: outide
            protocol: TCP
          - containerPort: 5555
            name: jmx
            protocol: TCP
          readinessProbe:
            failureThreshold: 3
            periodSeconds: 10
            successThreshold: 1
            tcpSocket:
              port: 9092
            timeoutSeconds: 1
          resources:
            limits:
              memory: 1Gi
            requests:
              cpu: 250m
              memory: 500Mi
          terminationMessagePath: /dev/termination-log
          terminationMessagePolicy: File
          volumeMounts:
          - mountPath: /etc/kafka-configmap
            name: brokerconfig
          - mountPath: /etc/kafka
            name: config
          - mountPath: /etc/kafkacert
            name: kafkacert
          - mountPath: /opt/kafka/logs
            name: brokerlogs
          - mountPath: /opt/kafka/libs/extensions
            name: extensions
          - mountPath: /var/lib/kafka/data
            name: data
        dnsPolicy: ClusterFirst
        initContainers:
        - command:
          - /bin/bash
          - /etc/kafka-configmap/init.sh
          env:
          - name: NODE_NAME
            valueFrom:
              fieldRef:
                apiVersion: v1
                fieldPath: spec.nodeName
          - name: POD_NAME
            valueFrom:
              fieldRef:
                apiVersion: v1
                fieldPath: metadata.name
          - name: POD_NAMESPACE
            valueFrom:
              fieldRef:
                apiVersion: v1
                fieldPath: metadata.namespace
          - name: ADVERTISE_ADDR
            value: kafka
          image: quay.io/openshifttest/kafka-initutils@sha256:e73ff7a44b43b85b53849c0459ba32e704540852b885a5c78af9753f86a49d68
          imagePullPolicy: IfNotPresent
          name: init-config
          securityContext:
            allowPrivilegeEscalation: false
            runAsNonRoot: true
            capabilities:
              drop:
              - ALL
            seccompProfile:
              type: RuntimeDefault
          resources: {}
          terminationMessagePath: /dev/termination-log
          terminationMessagePolicy: File
          volumeMounts:
          - mountPath: /etc/kafka-configmap
            name: brokerconfig
          - mountPath: /etc/kafka
            name: config
          - mountPath: /opt/kafka/libs/extensions
            name: extensions
        restartPolicy: Always
        nodeSelector:
          kubernetes.io/arch: amd64
          kubernetes.io/os: linux
        schedulerName: default-scheduler
        terminationGracePeriodSeconds: 30
        volumes:
        - configMap:
            defaultMode: 420
            name: ${CM_NAME}
          name: brokerconfig
        - secret:
            defaultMode: 420
            secretName: ${SECRETNAME}
          name: kafkacert
        - emptyDir: {}
          name: brokerlogs
        - emptyDir: {}
          name: config
        - emptyDir: {}
          name: extensions
        - emptyDir: {}
          name: data
    updateStrategy:
      type: RollingUpdate
parameters:
- name: NAME
  value: "kafka"
- name: NAMESPACE
  value: "openshift-logging"
- name: SERVICENAME
  value: "kafka"
- name: CM_NAME
  value: "kafka"
- name: SECRETNAME
  value: "kafka-cluster-cert"
`)

func testdataExternalLogStoresKafkaPlaintextSslKafkaStatefulsetYamlBytes() ([]byte, error) {
	return _testdataExternalLogStoresKafkaPlaintextSslKafkaStatefulsetYaml, nil
}

func testdataExternalLogStoresKafkaPlaintextSslKafkaStatefulsetYaml() (*asset, error) {
	bytes, err := testdataExternalLogStoresKafkaPlaintextSslKafkaStatefulsetYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "testdata/external-log-stores/kafka/plaintext-ssl/kafka-statefulset.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testdataExternalLogStoresKafkaSaslPlaintextConsumerConfigmapYaml = []byte(`kind: Template
apiVersion: template.openshift.io/v1
metadata:
  name: consumer-configmap-template
objects:
- apiVersion: v1
  data:
    client.properties: |
      bootstrap.servers=kafka:9092
      #group.id=test-consumer-group
      sasl.mechanism=PLAIN
      security.protocol=SASL_PLAINTEXT
      sasl.jaas.config=org.apache.kafka.common.security.plain.PlainLoginModule required \
         username="admin" \
         password="admin-secret";
    kafka_client_jaas.conf: |
      KafkaClient {
         org.apache.kafka.common.security.plain.PlainLoginModule required
         username="admin"
         password="admin-secret";
      };
    sasl-consumer.properties: |
      #export KAFKA_OPTS="-Djava.security.auth.login.config=/etc/kafka-configmap/kafka_client_jaas.conf"
      #/opt/kafka/bin/kafka-console-producer.sh --broker-list kafka:9092 --producer.config=/etc/kafka-config/sasl-producer.properties  --topic  clo-topic
      bootstrap.servers=kafka:9092
      compression.type=none
      ### SECURITY ######
      security.protocol=SASL_PLANTEXT
      sasl.mechanism=PLAIN
      sasl.jaas.config=org.apache.kafka.common.security.plain.PlainLoginModule required username="admin" password="admin-secret";
      ssl.truststore.location=/etc/kafkacert/ca-bundle.jks
  kind: ConfigMap
  metadata:
    name: ${NAME}
    namespace: ${NAMESPACE}
parameters:
- name: NAME
  value: "kafka-client"
- name: NAMESPACE
  value: "openshift-logging"
`)

func testdataExternalLogStoresKafkaSaslPlaintextConsumerConfigmapYamlBytes() ([]byte, error) {
	return _testdataExternalLogStoresKafkaSaslPlaintextConsumerConfigmapYaml, nil
}

func testdataExternalLogStoresKafkaSaslPlaintextConsumerConfigmapYaml() (*asset, error) {
	bytes, err := testdataExternalLogStoresKafkaSaslPlaintextConsumerConfigmapYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "testdata/external-log-stores/kafka/sasl-plaintext/consumer-configmap.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testdataExternalLogStoresKafkaSaslPlaintextKafkaConfigmapYaml = []byte(`kind: Template
apiVersion: template.openshift.io/v1
metadata:
  name: consumer-configmap-template
objects:
- apiVersion: v1
  data:
    init.sh: |
      #!/bin/bash
      set -e
      cp /etc/kafka-configmap/log4j.properties /etc/kafka/
      KAFKA_BROKER_ID=${HOSTNAME##*-}
      SEDS=("s/#init#broker.id=#init#/broker.id=$KAFKA_BROKER_ID/")
      LABELS="kafka-broker-id=$KAFKA_BROKER_ID"
      ANNOTATIONS=""

      hash kubectl 2>/dev/null || {
        SEDS+=("s/#init#broker.rack=#init#/#init#broker.rack=# kubectl not found in path/")
      } && {
        ZONE=$(kubectl get node "$NODE_NAME" -o=go-template='{{index .metadata.labels "failure-domain.beta.kubernetes.io/zone"}}')
        if [ "x$ZONE" == "x<no value>" ]; then
          SEDS+=("s/#init#broker.rack=#init#/#init#broker.rack=# zone label not found for node $NODE_NAME/")
        else
          SEDS+=("s/#init#broker.rack=#init#/broker.rack=$ZONE/")
          LABELS="$LABELS kafka-broker-rack=$ZONE"
        fi

        [ -z "$ADVERTISE_ADDR" ] && echo "ADVERTISE_ADDR is empty, will advertise detected DNS name"
        OUTSIDE_HOST=$(kubectl get node "$NODE_NAME" -o jsonpath='{.status.addresses[?(@.type=="InternalIP")].address}')
        OUTSIDE_PORT=$((32400 + ${KAFKA_BROKER_ID}))
        SEDS+=("s|#init#advertised.listeners=PLAINTEXT://#init#|advertised.listeners=PLAINTEXT://${ADVERTISE_ADDR}:9092,SASL_PLAINTEXT://${ADVERTISE_ADDR}:9093|")
        ANNOTATIONS="$ANNOTATIONS kafka-listener-outside-host=$OUTSIDE_HOST kafka-listener-outside-port=$OUTSIDE_PORT"

        if [ ! -z "$LABELS" ]; then
          kubectl -n $POD_NAMESPACE label pod $POD_NAME $LABELS || echo "Failed to label $POD_NAMESPACE.$POD_NAME - RBAC issue?"
        fi
        if [ ! -z "$ANNOTATIONS" ]; then
          kubectl -n $POD_NAMESPACE annotate pod $POD_NAME $ANNOTATIONS || echo "Failed to annotate $POD_NAMESPACE.$POD_NAME - RBAC issue?"
        fi
      }
      printf '%s\n' "${SEDS[@]}" | sed -f - /etc/kafka-configmap/server.properties > /etc/kafka/server.properties.tmp
      [ $? -eq 0 ] && mv /etc/kafka/server.properties.tmp /etc/kafka/server.properties
    kafka_server_jaas.conf: |
      KafkaServer {
         org.apache.kafka.common.security.plain.PlainLoginModule required
         serviceName="kafka"
         username="admin"
         password="admin-secret"
         user_admin="admin-secret"
         user_alice="alice-secret";
      };
    log4j.properties: |
      log4j.rootLogger=INFO, stdout
      log4j.appender.stdout=org.apache.log4j.ConsoleAppender
      log4j.appender.stdout.layout=org.apache.log4j.PatternLayout
      log4j.appender.stdout.layout.ConversionPattern=[%d] %p %m (%c)%n
      log4j.appender.kafkaAppender=org.apache.log4j.DailyRollingFileAppender
      log4j.appender.kafkaAppender.DatePattern='.'yyyy-MM-dd-HH
      log4j.appender.kafkaAppender.File=${kafka.logs.dir}/server.log
      log4j.appender.kafkaAppender.layout=org.apache.log4j.PatternLayout
      log4j.appender.kafkaAppender.layout.ConversionPattern=[%d] %p %m (%c)%n
      log4j.appender.stateChangeAppender=org.apache.log4j.DailyRollingFileAppender
      log4j.appender.stateChangeAppender.DatePattern='.'yyyy-MM-dd-HH
      log4j.appender.stateChangeAppender.File=${kafka.logs.dir}/state-change.log
      log4j.appender.stateChangeAppender.layout=org.apache.log4j.PatternLayout
      log4j.appender.stateChangeAppender.layout.ConversionPattern=[%d] %p %m (%c)%n
      log4j.appender.requestAppender=org.apache.log4j.DailyRollingFileAppender
      log4j.appender.requestAppender.DatePattern='.'yyyy-MM-dd-HH
      log4j.appender.requestAppender.File=${kafka.logs.dir}/kafka-request.log
      log4j.appender.requestAppender.layout=org.apache.log4j.PatternLayout
      log4j.appender.requestAppender.layout.ConversionPattern=[%d] %p %m (%c)%n
      log4j.appender.cleanerAppender=org.apache.log4j.DailyRollingFileAppender
      log4j.appender.cleanerAppender.DatePattern='.'yyyy-MM-dd-HH
      log4j.appender.cleanerAppender.File=${kafka.logs.dir}/log-cleaner.log
      log4j.appender.cleanerAppender.layout=org.apache.log4j.PatternLayout
      log4j.appender.cleanerAppender.layout.ConversionPattern=[%d] %p %m (%c)%n
      log4j.appender.controllerAppender=org.apache.log4j.DailyRollingFileAppender
      log4j.appender.controllerAppender.DatePattern='.'yyyy-MM-dd-HH
      log4j.appender.controllerAppender.File=${kafka.logs.dir}/controller.log
      log4j.appender.controllerAppender.layout=org.apache.log4j.PatternLayout
      log4j.appender.controllerAppender.layout.ConversionPattern=[%d] %p %m (%c)%n
      log4j.appender.authorizerAppender=org.apache.log4j.DailyRollingFileAppender
      log4j.appender.authorizerAppender.DatePattern='.'yyyy-MM-dd-HH
      log4j.appender.authorizerAppender.File=${kafka.logs.dir}/kafka-authorizer.log
      log4j.appender.authorizerAppender.layout=org.apache.log4j.PatternLayout
      log4j.appender.authorizerAppender.layout.ConversionPattern=[%d] %p %m (%c)%n
      log4j.logger.org.I0Itec.zkclient.ZkClient=INFO
      log4j.logger.org.apache.zookeeper=INFO
      log4j.logger.kafka=INFO
      log4j.logger.org.apache.kafka=INFO
      log4j.logger.kafka.request.logger=WARN, requestAppender
      log4j.additivity.kafka.request.logger=false
      log4j.logger.kafka.network.RequestChannel$=WARN, requestAppender
      log4j.additivity.kafka.network.RequestChannel$=false
      log4j.logger.kafka.controller=TRACE, controllerAppender
      log4j.additivity.kafka.controller=false
      log4j.logger.kafka.log.LogCleaner=INFO, cleanerAppender
      log4j.additivity.kafka.log.LogCleaner=false
      log4j.logger.state.change.logger=TRACE, stateChangeAppender
      log4j.additivity.state.change.logger=false
      log4j.logger.kafka.authorizer.logger=WARN, authorizerAppender
      log4j.additivity.kafka.authorizer.logger=false
    server.properties: |
      #https://docs.confluent.io/platform/current/kafka/authentication_sasl/authentication_sasl_plain.html
      #init#broker.id=#init#
      ssl.client.auth=none
      sasl.enabled.mechanisms=PLAIN
      sasl.mechanism.inter.broker.protocol=PLAIN
      security.inter.broker.protocol=SASL_PLAINTEXT
      listeners=SASL_PLAINTEXT://:9092
      security.protocol=SASL_PLAINTEXT
      authorizer.class.name=kafka.security.authorizer.AclAuthorizer
      super.users=User:admin
      allow.everyone.if.no.acl.found=true
      num.partitions=1
      num.network.threads=3
      num.io.threads=8
      num.recovery.threads.per.data.dir=1
      message.max.bytes=314572800
      socket.send.buffer.bytes=102400
      socket.receive.buffer.bytes=102400
      socket.request.max.bytes=104857600
      socket.request.max.bytes=314572800
      log.dirs=/tmp/kafka-logs
      offsets.topic.replication.factor=1
      transaction.state.log.replication.factor=1
      transaction.state.log.min.isr=1
      log.retention.hours=2
      log.segment.bytes=1073741824
      log.retention.check.interval.ms=300000
      zookeeper.connect=zookeeper:2181
      zookeeper.connection.timeout.ms=18000
      group.initial.rebalance.delay.ms=0
  kind: ConfigMap
  metadata:
    name: ${NAME}
    namespace: ${NAMESPACE}
parameters:
- name: NAME
  value: "kafka"
- name: NAMESPACE
  value: "openshift-logging"
`)

func testdataExternalLogStoresKafkaSaslPlaintextKafkaConfigmapYamlBytes() ([]byte, error) {
	return _testdataExternalLogStoresKafkaSaslPlaintextKafkaConfigmapYaml, nil
}

func testdataExternalLogStoresKafkaSaslPlaintextKafkaConfigmapYaml() (*asset, error) {
	bytes, err := testdataExternalLogStoresKafkaSaslPlaintextKafkaConfigmapYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "testdata/external-log-stores/kafka/sasl-plaintext/kafka-configmap.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testdataExternalLogStoresKafkaSaslPlaintextKafkaConsumerDeploymentYaml = []byte(`kind: Template
apiVersion: template.openshift.io/v1
metadata:
  name: kafkaserver-template
objects:
- apiVersion: apps/v1
  kind: Deployment
  metadata:
    annotations:
      deployment.kubernetes.io/revision: "1"
    labels:
      component: kafka-consumer
      logging-infra: kafka
      provider: openshift
    name: ${NAME}
    namespace: ${NAMESPACE}
  spec:
    progressDeadlineSeconds: 601
    replicas: 1
    revisionHistoryLimit: 10
    selector:
      matchLabels:
        component: kafka-consumer
        logging-infra: kafka
        provider: openshift
    strategy:
      rollingUpdate:
        maxSurge: 25%
        maxUnavailable: 25%
      type: RollingUpdate
    template:
      metadata:
        creationTimestamp: null
        labels:
          component: kafka-consumer
          logging-infra: kafka
          provider: openshift
        name: kafka-consumer
      spec:
        securityContext:
          runAsNonRoot: true
          seccompProfile:
            type: RuntimeDefault
        containers:
        - command:
          - /bin/bash
          - -ce
          - /opt/kafka/bin/kafka-console-consumer.sh --bootstrap-server kafka:9092 --topic clo-topic --from-beginning  --consumer.config /etc/kafka-config/client.properties
          image: quay.io/openshifttest/kafka@sha256:2411662d89dd5700e1fe49aa8be1219843948da90cfe51a1c7a49bcef9d22dab
          imagePullPolicy: IfNotPresent
          name: kafka-consumer
          securityContext:
            allowPrivilegeEscalation: false
            runAsNonRoot: true
            capabilities:
              drop:
              - ALL
            seccompProfile:
              type: RuntimeDefault
          resources: {}
          terminationMessagePath: /dev/termination-log
          terminationMessagePolicy: File
          volumeMounts:
          - mountPath: /shared
            name: shared
          - mountPath: /etc/kafka-config
            name: kafka-client
          env:
          - name: KAFKA_OPTS
            value: -Djava.security.auth.login.config=/etc/kafka-configmap/kafka_client_jaas.conf
        dnsPolicy: ClusterFirst
        nodeSelector:
          kubernetes.io/arch: amd64
          kubernetes.io/os: linux
        restartPolicy: Always
        schedulerName: default-scheduler
        terminationGracePeriodSeconds: 30
        volumes:
        - emptyDir: {}
          name: shared
        - configMap:
            defaultMode: 420
            name: ${CM_NAME}
          name: kafka-client
parameters:
- name: NAME
  value: "kafka-consumer-sasl-plaintext"
- name: NAMESPACE
  value: "openshift-logging"
- name: CM_NAME
  value: "kafka-client"
`)

func testdataExternalLogStoresKafkaSaslPlaintextKafkaConsumerDeploymentYamlBytes() ([]byte, error) {
	return _testdataExternalLogStoresKafkaSaslPlaintextKafkaConsumerDeploymentYaml, nil
}

func testdataExternalLogStoresKafkaSaslPlaintextKafkaConsumerDeploymentYaml() (*asset, error) {
	bytes, err := testdataExternalLogStoresKafkaSaslPlaintextKafkaConsumerDeploymentYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "testdata/external-log-stores/kafka/sasl-plaintext/kafka-consumer-deployment.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testdataExternalLogStoresKafkaSaslPlaintextKafkaStatefulsetYaml = []byte(`kind: Template
apiVersion: template.openshift.io/v1
metadata:
  name: kafkaserver-template
objects:
- apiVersion: apps/v1
  kind: StatefulSet
  metadata:
    labels:
      app: kafka
      component: kafka
      provider: openshift
    name: ${NAME}
    namespace: ${NAMESPACE}
  spec:
    podManagementPolicy: Parallel
    replicas: 1
    revisionHistoryLimit: 10
    selector:
      matchLabels:
        app: kafka
    serviceName: ${SERVICENAME}
    template:
      metadata:
        creationTimestamp: null
        labels:
          app: kafka
          component: kafka
          provider: openshift
      spec:
        securityContext:
          runAsNonRoot: true
          seccompProfile:
            type: RuntimeDefault
        containers:
        - command:
          - /opt/kafka/bin/kafka-server-start.sh
          - /etc/kafka/server.properties
          env:
          - name: CLASSPATH
            value: /opt/kafka/libs/extensions/*
          - name: KAFKA_LOG4J_OPTS
            value: -Dlog4j.configuration=file:/etc/kafka/log4j.properties
          - name: KAFKA_OPTS
            value: -Djava.security.auth.login.config=/etc/kafka-configmap/kafka_server_jaas.conf
          - name: JMX_PORT
            value: "5555"
          image: quay.io/openshifttest/kafka@sha256:2411662d89dd5700e1fe49aa8be1219843948da90cfe51a1c7a49bcef9d22dab
          imagePullPolicy: IfNotPresent
          securityContext:
            allowPrivilegeEscalation: false
            runAsNonRoot: true
            capabilities:
              drop:
              - ALL
            seccompProfile:
              type: RuntimeDefault
          lifecycle:
            preStop:
              exec:
                command:
                - sh
                - -ce
                - kill -s TERM 1; while $(kill -0 1 2>/dev/null); do sleep 1; done
          name: broker
          ports:
          - containerPort: 9092
            name: inside
            protocol: TCP
          - containerPort: 9093
            name: ssl
            protocol: TCP
          - containerPort: 9094
            name: outide
            protocol: TCP
          - containerPort: 5555
            name: jmx
            protocol: TCP
          readinessProbe:
            failureThreshold: 3
            periodSeconds: 10
            successThreshold: 1
            tcpSocket:
              port: 9092
            timeoutSeconds: 1
          resources:
            limits:
              memory: 1Gi
            requests:
              cpu: 250m
              memory: 500Mi
          terminationMessagePath: /dev/termination-log
          terminationMessagePolicy: File
          volumeMounts:
          - mountPath: /etc/kafka-configmap
            name: brokerconfig
          - mountPath: /etc/kafka
            name: config
          - mountPath: /etc/kafkacert
            name: kafkacert
          - mountPath: /opt/kafka/logs
            name: brokerlogs
          - mountPath: /opt/kafka/libs/extensions
            name: extensions
          - mountPath: /var/lib/kafka/data
            name: data
        dnsPolicy: ClusterFirst
        initContainers:
        - command:
          - /bin/bash
          - /etc/kafka-configmap/init.sh
          env:
          - name: NODE_NAME
            valueFrom:
              fieldRef:
                apiVersion: v1
                fieldPath: spec.nodeName
          - name: POD_NAME
            valueFrom:
              fieldRef:
                apiVersion: v1
                fieldPath: metadata.name
          - name: POD_NAMESPACE
            valueFrom:
              fieldRef:
                apiVersion: v1
                fieldPath: metadata.namespace
          - name: ADVERTISE_ADDR
            value: kafka
          image: quay.io/openshifttest/kafka-initutils@sha256:e73ff7a44b43b85b53849c0459ba32e704540852b885a5c78af9753f86a49d68
          imagePullPolicy: IfNotPresent
          name: init-config
          securityContext:
            allowPrivilegeEscalation: false
            runAsNonRoot: true
            capabilities:
              drop:
              - ALL
            seccompProfile:
              type: RuntimeDefault
          resources: {}
          terminationMessagePath: /dev/termination-log
          terminationMessagePolicy: File
          volumeMounts:
          - mountPath: /etc/kafka-configmap
            name: brokerconfig
          - mountPath: /etc/kafka
            name: config
          - mountPath: /opt/kafka/libs/extensions
            name: extensions
        restartPolicy: Always
        nodeSelector:
          kubernetes.io/arch: amd64
          kubernetes.io/os: linux
        schedulerName: default-scheduler
        terminationGracePeriodSeconds: 30
        volumes:
        - configMap:
            defaultMode: 420
            name: ${CM_NAME}
          name: brokerconfig
        - emptyDir: {}
          name: kafkacert
        - emptyDir: {}
          name: brokerlogs
        - emptyDir: {}
          name: config
        - emptyDir: {}
          name: extensions
        - emptyDir: {}
          name: data
    updateStrategy:
      type: RollingUpdate
parameters:
- name: NAME
  value: "kafka"
- name: NAMESPACE
  value: "openshift-logging"
- name: SERVICENAME
  value: "kafka"
- name: CM_NAME
  value: "kafka"
`)

func testdataExternalLogStoresKafkaSaslPlaintextKafkaStatefulsetYamlBytes() ([]byte, error) {
	return _testdataExternalLogStoresKafkaSaslPlaintextKafkaStatefulsetYaml, nil
}

func testdataExternalLogStoresKafkaSaslPlaintextKafkaStatefulsetYaml() (*asset, error) {
	bytes, err := testdataExternalLogStoresKafkaSaslPlaintextKafkaStatefulsetYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "testdata/external-log-stores/kafka/sasl-plaintext/kafka-statefulset.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testdataExternalLogStoresKafkaSaslSslConsumerConfigmapYaml = []byte(`kind: Template
apiVersion: template.openshift.io/v1
metadata:
  name: consumer-configmap-template
objects:
- apiVersion: v1
  data:
    client.properties: |
      bootstrap.servers=kafka:9093
      #group.id=test-consumer-group
      sasl.mechanism=PLAIN
      security.protocol=SASL_SSL
      sasl.jaas.config=org.apache.kafka.common.security.plain.PlainLoginModule required \
         username="admin" \
         password="admin-secret";
      ssl.truststore.location=/etc/kafkacert/ca-bundle.jks
      ssl.truststore.password=aosqe2021
    kafka_client_jaas.conf: |
      KafkaClient {
         org.apache.kafka.common.security.plain.PlainLoginModule required
         username="admin"
         password="admin-secret";
      };
    ssl-consumer.properties: |
      #export KAFKA_OPTS="-Djava.security.auth.login.config=/etc/kafka-configmap/kafka_client_jaas.conf"
      #/opt/kafka/bin/kafka-console-producer.sh --broker-list kafka:9093 --producer.config=/etc/kafka-config/ssl-producer.properties  --topic  clo-topic
      bootstrap.servers=kafka:9093
      compression.type=none
      ### SECURITY ######
      security.protocol=SASL_SSL
      sasl.mechanism=PLAIN
      sasl.jaas.config=org.apache.kafka.common.security.plain.PlainLoginModule required username="admin" password="admin-secret";
      ssl.truststore.location=/etc/kafkacert/ca-bundle.jks
  kind: ConfigMap
  metadata:
    name: ${NAME}
    namespace: ${NAMESPACE}
parameters:
- name: NAME
  value: "kafka-client"
- name: NAMESPACE
  value: "openshift-logging"
`)

func testdataExternalLogStoresKafkaSaslSslConsumerConfigmapYamlBytes() ([]byte, error) {
	return _testdataExternalLogStoresKafkaSaslSslConsumerConfigmapYaml, nil
}

func testdataExternalLogStoresKafkaSaslSslConsumerConfigmapYaml() (*asset, error) {
	bytes, err := testdataExternalLogStoresKafkaSaslSslConsumerConfigmapYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "testdata/external-log-stores/kafka/sasl-ssl/consumer-configmap.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testdataExternalLogStoresKafkaSaslSslKafkaConfigmapYaml = []byte(`kind: Template
apiVersion: template.openshift.io/v1
metadata:
  name: consumer-configmap-template
objects:
- apiVersion: v1
  data:
    init.sh: |
      #!/bin/bash
      set -e
      cp /etc/kafka-configmap/log4j.properties /etc/kafka/
      KAFKA_BROKER_ID=${HOSTNAME##*-}
      SEDS=("s/#init#broker.id=#init#/broker.id=$KAFKA_BROKER_ID/")
      LABELS="kafka-broker-id=$KAFKA_BROKER_ID"
      ANNOTATIONS=""

      hash kubectl 2>/dev/null || {
        SEDS+=("s/#init#broker.rack=#init#/#init#broker.rack=# kubectl not found in path/")
      } && {
        ZONE=$(kubectl get node "$NODE_NAME" -o=go-template='{{index .metadata.labels "failure-domain.beta.kubernetes.io/zone"}}')
        if [ "x$ZONE" == "x<no value>" ]; then
          SEDS+=("s/#init#broker.rack=#init#/#init#broker.rack=# zone label not found for node $NODE_NAME/")
        else
          SEDS+=("s/#init#broker.rack=#init#/broker.rack=$ZONE/")
          LABELS="$LABELS kafka-broker-rack=$ZONE"
        fi

        [ -z "$ADVERTISE_ADDR" ] && echo "ADVERTISE_ADDR is empty, will advertise detected DNS name"
        OUTSIDE_HOST=$(kubectl get node "$NODE_NAME" -o jsonpath='{.status.addresses[?(@.type=="InternalIP")].address}')
        OUTSIDE_PORT=$((32400 + ${KAFKA_BROKER_ID}))
        SEDS+=("s|#init#advertised.listeners=PLAINTEXT://#init#|advertised.listeners=PLAINTEXT://${ADVERTISE_ADDR}:9092,SASL_SSL://${ADVERTISE_ADDR}:9093|")
        ANNOTATIONS="$ANNOTATIONS kafka-listener-outside-host=$OUTSIDE_HOST kafka-listener-outside-port=$OUTSIDE_PORT"

        if [ ! -z "$LABELS" ]; then
          kubectl -n $POD_NAMESPACE label pod $POD_NAME $LABELS || echo "Failed to label $POD_NAMESPACE.$POD_NAME - RBAC issue?"
        fi
        if [ ! -z "$ANNOTATIONS" ]; then
          kubectl -n $POD_NAMESPACE annotate pod $POD_NAME $ANNOTATIONS || echo "Failed to annotate $POD_NAMESPACE.$POD_NAME - RBAC issue?"
        fi
      }
      printf '%s\n' "${SEDS[@]}" | sed -f - /etc/kafka-configmap/server.properties > /etc/kafka/server.properties.tmp
      [ $? -eq 0 ] && mv /etc/kafka/server.properties.tmp /etc/kafka/server.properties
    kafka_server_jaas.conf: |
      KafkaServer {
         org.apache.kafka.common.security.plain.PlainLoginModule required
         serviceName="kafka"
         username="admin"
         password="admin-secret"
         user_admin="admin-secret"
         user_alice="alice-secret";
      };
    log4j.properties: |
      log4j.rootLogger=INFO, stdout
      log4j.appender.stdout=org.apache.log4j.ConsoleAppender
      log4j.appender.stdout.layout=org.apache.log4j.PatternLayout
      log4j.appender.stdout.layout.ConversionPattern=[%d] %p %m (%c)%n
      log4j.appender.kafkaAppender=org.apache.log4j.DailyRollingFileAppender
      log4j.appender.kafkaAppender.DatePattern='.'yyyy-MM-dd-HH
      log4j.appender.kafkaAppender.File=${kafka.logs.dir}/server.log
      log4j.appender.kafkaAppender.layout=org.apache.log4j.PatternLayout
      log4j.appender.kafkaAppender.layout.ConversionPattern=[%d] %p %m (%c)%n
      log4j.appender.stateChangeAppender=org.apache.log4j.DailyRollingFileAppender
      log4j.appender.stateChangeAppender.DatePattern='.'yyyy-MM-dd-HH
      log4j.appender.stateChangeAppender.File=${kafka.logs.dir}/state-change.log
      log4j.appender.stateChangeAppender.layout=org.apache.log4j.PatternLayout
      log4j.appender.stateChangeAppender.layout.ConversionPattern=[%d] %p %m (%c)%n
      log4j.appender.requestAppender=org.apache.log4j.DailyRollingFileAppender
      log4j.appender.requestAppender.DatePattern='.'yyyy-MM-dd-HH
      log4j.appender.requestAppender.File=${kafka.logs.dir}/kafka-request.log
      log4j.appender.requestAppender.layout=org.apache.log4j.PatternLayout
      log4j.appender.requestAppender.layout.ConversionPattern=[%d] %p %m (%c)%n
      log4j.appender.cleanerAppender=org.apache.log4j.DailyRollingFileAppender
      log4j.appender.cleanerAppender.DatePattern='.'yyyy-MM-dd-HH
      log4j.appender.cleanerAppender.File=${kafka.logs.dir}/log-cleaner.log
      log4j.appender.cleanerAppender.layout=org.apache.log4j.PatternLayout
      log4j.appender.cleanerAppender.layout.ConversionPattern=[%d] %p %m (%c)%n
      log4j.appender.controllerAppender=org.apache.log4j.DailyRollingFileAppender
      log4j.appender.controllerAppender.DatePattern='.'yyyy-MM-dd-HH
      log4j.appender.controllerAppender.File=${kafka.logs.dir}/controller.log
      log4j.appender.controllerAppender.layout=org.apache.log4j.PatternLayout
      log4j.appender.controllerAppender.layout.ConversionPattern=[%d] %p %m (%c)%n
      log4j.appender.authorizerAppender=org.apache.log4j.DailyRollingFileAppender
      log4j.appender.authorizerAppender.DatePattern='.'yyyy-MM-dd-HH
      log4j.appender.authorizerAppender.File=${kafka.logs.dir}/kafka-authorizer.log
      log4j.appender.authorizerAppender.layout=org.apache.log4j.PatternLayout
      log4j.appender.authorizerAppender.layout.ConversionPattern=[%d] %p %m (%c)%n
      log4j.logger.org.I0Itec.zkclient.ZkClient=INFO
      log4j.logger.org.apache.zookeeper=INFO
      log4j.logger.kafka=INFO
      log4j.logger.org.apache.kafka=INFO
      log4j.logger.kafka.request.logger=WARN, requestAppender
      log4j.additivity.kafka.request.logger=false
      log4j.logger.kafka.network.RequestChannel$=WARN, requestAppender
      log4j.additivity.kafka.network.RequestChannel$=false
      log4j.logger.kafka.controller=TRACE, controllerAppender
      log4j.additivity.kafka.controller=false
      log4j.logger.kafka.log.LogCleaner=INFO, cleanerAppender
      log4j.additivity.kafka.log.LogCleaner=false
      log4j.logger.state.change.logger=TRACE, stateChangeAppender
      log4j.additivity.state.change.logger=false
      log4j.logger.kafka.authorizer.logger=WARN, authorizerAppender
      log4j.additivity.kafka.authorizer.logger=false
    server.properties: |
      #https://docs.confluent.io/platform/current/kafka/authentication_sasl/authentication_sasl_plain.html
      #init#broker.id=#init#
      ssl.client.auth=requested
      ssl.keystore.location=/etc/kafkacert/cluster.jks
      ssl.keystore.password=aosqe2021
      ssl.truststore.location=/etc/kafkacert/ca_bundle.jks
      ssl.truststore.password=aosqe2021
      sasl.enabled.mechanisms=PLAIN
      sasl.mechanism.inter.broker.protocol=PLAIN
      security.inter.broker.protocol=PLAINTEXT
      listeners=PLAINTEXT://:9092,SASL_SSL://:9093
      #init#advertised.listeners=PLAINTEXT://#init#
      security.protocol=SASL_SSL
      authorizer.class.name=kafka.security.authorizer.AclAuthorizer
      super.users=User:admin
      allow.everyone.if.no.acl.found=true
      num.partitions=1
      num.network.threads=3
      num.io.threads=8
      num.recovery.threads.per.data.dir=1
      message.max.bytes=314572800
      socket.send.buffer.bytes=102400
      socket.receive.buffer.bytes=102400
      socket.request.max.bytes=104857600
      socket.request.max.bytes=314572800
      log.dirs=/tmp/kafka-logs
      offsets.topic.replication.factor=1
      transaction.state.log.replication.factor=1
      transaction.state.log.min.isr=1
      log.retention.hours=2
      log.segment.bytes=1073741824
      log.retention.check.interval.ms=300000
      zookeeper.connect=zookeeper:2181
      zookeeper.connection.timeout.ms=18000
      group.initial.rebalance.delay.ms=0

  kind: ConfigMap
  metadata:
    name: ${NAME}
    namespace: ${NAMESPACE}
parameters:
- name: NAME
  value: "kafka"
- name: NAMESPACE
  value: "openshift-logging"
`)

func testdataExternalLogStoresKafkaSaslSslKafkaConfigmapYamlBytes() ([]byte, error) {
	return _testdataExternalLogStoresKafkaSaslSslKafkaConfigmapYaml, nil
}

func testdataExternalLogStoresKafkaSaslSslKafkaConfigmapYaml() (*asset, error) {
	bytes, err := testdataExternalLogStoresKafkaSaslSslKafkaConfigmapYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "testdata/external-log-stores/kafka/sasl-ssl/kafka-configmap.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testdataExternalLogStoresKafkaSaslSslKafkaConsumerDeploymentYaml = []byte(`kind: Template
apiVersion: template.openshift.io/v1
metadata:
  name: kafkaserver-template
objects:
- apiVersion: apps/v1
  kind: Deployment
  metadata:
    annotations:
      deployment.kubernetes.io/revision: "1"
    labels:
      component: kafka-consumer
      logging-infra: kafka
      provider: openshift
    name: ${NAME}
    namespace: ${NAMESPACE}
  spec:
    progressDeadlineSeconds: 600
    replicas: 1
    revisionHistoryLimit: 10
    selector:
      matchLabels:
        component: kafka-consumer
        logging-infra: kafka
        provider: openshift
    strategy:
      rollingUpdate:
        maxSurge: 25%
        maxUnavailable: 25%
      type: RollingUpdate
    template:
      metadata:
        creationTimestamp: null
        labels:
          component: kafka-consumer
          logging-infra: kafka
          provider: openshift
        name: kafka-consumer
      spec:
        securityContext:
          runAsNonRoot: true
          seccompProfile:
            type: RuntimeDefault
        containers:
        - command:
          - /bin/bash
          - -ce
          - /opt/kafka/bin/kafka-console-consumer.sh --bootstrap-server kafka:9093 --topic clo-topic --from-beginning  --consumer.config /etc/kafka-config/client.properties
          image: quay.io/openshifttest/kafka@sha256:2411662d89dd5700e1fe49aa8be1219843948da90cfe51a1c7a49bcef9d22dab
          imagePullPolicy: IfNotPresent
          name: kafka-consumer
          securityContext:
            allowPrivilegeEscalation: false
            runAsNonRoot: true
            capabilities:
              drop:
              - ALL
            seccompProfile:
              type: RuntimeDefault
          resources: {}
          terminationMessagePath: /dev/termination-log
          terminationMessagePolicy: File
          volumeMounts:
          - mountPath: /shared
            name: shared
          - mountPath: /etc/kafka-config
            name: kafka-client
          - mountPath: /etc/kafkacert
            name: kafkacert
          env:
          - name: KAFKA_OPTS
            value: -Djava.security.auth.login.config=/etc/kafka-configmap/kafka_client_jaas.conf
        dnsPolicy: ClusterFirst
        nodeSelector:
          kubernetes.io/arch: amd64
          kubernetes.io/os: linux
        restartPolicy: Always
        schedulerName: default-scheduler
        terminationGracePeriodSeconds: 30
        volumes:
        - emptyDir: {}
          name: shared
        - configMap:
            defaultMode: 420
            name: ${CM_NAME}
          name: kafka-client
        - secret:
            defaultMode: 420
            secretName: ${SECRETNAME}
          name: kafkacert
parameters:
- name: NAME
  value: "kafka-consumer-sasl-ssl"
- name: NAMESPACE
  value: "openshift-logging"
- name: CM_NAME
  value: "kafka-client"
- name: SECRETNAME
  value: "kafka-client-cert"
`)

func testdataExternalLogStoresKafkaSaslSslKafkaConsumerDeploymentYamlBytes() ([]byte, error) {
	return _testdataExternalLogStoresKafkaSaslSslKafkaConsumerDeploymentYaml, nil
}

func testdataExternalLogStoresKafkaSaslSslKafkaConsumerDeploymentYaml() (*asset, error) {
	bytes, err := testdataExternalLogStoresKafkaSaslSslKafkaConsumerDeploymentYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "testdata/external-log-stores/kafka/sasl-ssl/kafka-consumer-deployment.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testdataExternalLogStoresKafkaSaslSslKafkaStatefulsetYaml = []byte(`kind: Template
apiVersion: template.openshift.io/v1
metadata:
  name: kafkaserver-template
objects:
- apiVersion: apps/v1
  kind: StatefulSet
  metadata:
    labels:
      app: kafka
      component: kafka
      provider: openshift
    name: ${NAME}
    namespace: ${NAMESPACE}
  spec:
    podManagementPolicy: Parallel
    replicas: 1
    revisionHistoryLimit: 10
    selector:
      matchLabels:
        app: kafka
    serviceName: ${SERVICENAME}
    template:
      metadata:
        creationTimestamp: null
        labels:
          app: kafka
          component: kafka
          provider: openshift
      spec:
        securityContext:
          runAsNonRoot: true
          seccompProfile:
            type: RuntimeDefault
        containers:
        - command:
          - /opt/kafka/bin/kafka-server-start.sh
          - /etc/kafka/server.properties
          env:
          - name: CLASSPATH
            value: /opt/kafka/libs/extensions/*
          - name: KAFKA_LOG4J_OPTS
            value: -Dlog4j.configuration=file:/etc/kafka/log4j.properties
          - name: KAFKA_OPTS
            value: -Djava.security.auth.login.config=/etc/kafka-configmap/kafka_server_jaas.conf
          - name: JMX_PORT
            value: "5555"
          image: quay.io/openshifttest/kafka@sha256:2411662d89dd5700e1fe49aa8be1219843948da90cfe51a1c7a49bcef9d22dab
          imagePullPolicy: IfNotPresent
          securityContext:
            allowPrivilegeEscalation: false
            runAsNonRoot: true
            capabilities:
              drop:
              - ALL
            seccompProfile:
              type: RuntimeDefault
          lifecycle:
            preStop:
              exec:
                command:
                - sh
                - -ce
                - kill -s TERM 1; while $(kill -0 1 2>/dev/null); do sleep 1; done
          name: broker
          ports:
          - containerPort: 9092
            name: inside
            protocol: TCP
          - containerPort: 9093
            name: ssl
            protocol: TCP
          - containerPort: 9094
            name: outide
            protocol: TCP
          - containerPort: 5555
            name: jmx
            protocol: TCP
          readinessProbe:
            failureThreshold: 3
            periodSeconds: 10
            successThreshold: 1
            tcpSocket:
              port: 9092
            timeoutSeconds: 1
          resources:
            limits:
              memory: 1Gi
            requests:
              cpu: 250m
              memory: 500Mi
          terminationMessagePath: /dev/termination-log
          terminationMessagePolicy: File
          volumeMounts:
          - mountPath: /etc/kafka-configmap
            name: brokerconfig
          - mountPath: /etc/kafka
            name: config
          - mountPath: /etc/kafkacert
            name: kafkacert
          - mountPath: /opt/kafka/logs
            name: brokerlogs
          - mountPath: /opt/kafka/libs/extensions
            name: extensions
          - mountPath: /var/lib/kafka/data
            name: data
        dnsPolicy: ClusterFirst
        initContainers:
        - command:
          - /bin/bash
          - /etc/kafka-configmap/init.sh
          env:
          - name: NODE_NAME
            valueFrom:
              fieldRef:
                apiVersion: v1
                fieldPath: spec.nodeName
          - name: POD_NAME
            valueFrom:
              fieldRef:
                apiVersion: v1
                fieldPath: metadata.name
          - name: POD_NAMESPACE
            valueFrom:
              fieldRef:
                apiVersion: v1
                fieldPath: metadata.namespace
          - name: ADVERTISE_ADDR
            value: kafka
          image: quay.io/openshifttest/kafka-initutils@sha256:e73ff7a44b43b85b53849c0459ba32e704540852b885a5c78af9753f86a49d68
          imagePullPolicy: IfNotPresent
          name: init-config
          nodeSelector:
            kubernetes.io/arch: amd64
            kubernetes.io/os: linux
          securityContext:
            allowPrivilegeEscalation: false
            runAsNonRoot: true
            capabilities:
              drop:
              - ALL
            seccompProfile:
              type: RuntimeDefault
          resources: {}
          terminationMessagePath: /dev/termination-log
          terminationMessagePolicy: File
          volumeMounts:
          - mountPath: /etc/kafka-configmap
            name: brokerconfig
          - mountPath: /etc/kafka
            name: config
          - mountPath: /opt/kafka/libs/extensions
            name: extensions
        restartPolicy: Always
        schedulerName: default-scheduler
        terminationGracePeriodSeconds: 30
        volumes:
        - configMap:
            defaultMode: 420
            name: ${CM_NAME}
          name: brokerconfig
        - secret:
            defaultMode: 420
            secretName: ${SECRETNAME}
          name: kafkacert
        - emptyDir: {}
          name: brokerlogs
        - emptyDir: {}
          name: config
        - emptyDir: {}
          name: extensions
        - emptyDir: {}
          name: data
    updateStrategy:
      type: RollingUpdate
parameters:
- name: NAME
  value: "kafka"
- name: NAMESPACE
  value: "openshift-logging"
- name: SERVICENAME
  value: "kafka"
- name: CM_NAME
  value: "kafka"
- name: SECRETNAME
  value: "kafka-cluster-cert"
`)

func testdataExternalLogStoresKafkaSaslSslKafkaStatefulsetYamlBytes() ([]byte, error) {
	return _testdataExternalLogStoresKafkaSaslSslKafkaStatefulsetYaml, nil
}

func testdataExternalLogStoresKafkaSaslSslKafkaStatefulsetYaml() (*asset, error) {
	bytes, err := testdataExternalLogStoresKafkaSaslSslKafkaStatefulsetYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "testdata/external-log-stores/kafka/sasl-ssl/kafka-statefulset.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testdataExternalLogStoresKafkaZookeeperConfigmapSslYaml = []byte(`kind: Template
apiVersion: template.openshift.io/v1
metadata:
  name: zookeeper-configmap-template
objects:
- apiVersion: v1
  data:
    init.sh: |
      #!/bin/bash
      set -e
      [ -d /var/lib/zookeeper/data ] || mkdir /var/lib/zookeeper/data
      [ -z "$ID_OFFSET" ] && ID_OFFSET=1
      export ZOOKEEPER_SERVER_ID=$((${HOSTNAME##*-} + $ID_OFFSET))
      echo "${ZOOKEEPER_SERVER_ID:-1}" | tee /var/lib/zookeeper/data/myid
      cp -Lur /etc/kafka-configmap/* /etc/kafka/
    log4j.properties: |
      log4j.rootLogger=INFO, stdout
      log4j.appender.stdout=org.apache.log4j.ConsoleAppender
      log4j.appender.stdout.layout=org.apache.log4j.PatternLayout
      log4j.appender.stdout.layout.ConversionPattern=[%d] %p %m (%c)%n
      # Suppress connection log messages, three lines per livenessProbe execution
      log4j.logger.org.apache.zookeeper.server.NIOServerCnxnFactory=WARN
      log4j.logger.org.apache.zookeeper.server.NIOServerCnxn=WARN
    zookeeper.properties: |
      4lw.commands.whitelist=ruok
      tickTime=2000
      dataDir=/var/lib/zookeeper/data
      dataLogDir=/var/lib/zookeeper/log
      clientPort=2181
      authProvider.sasl=org.apache.zookeeper.server.auth.SASLAuthenticationProvider
  kind: ConfigMap
  metadata:
    name: ${NAME}
    namespace: ${NAMESPACE}
parameters:
- name: NAME
  value: "zookeeper"
- name: NAMESPACE
  value: "openshift-logging"
`)

func testdataExternalLogStoresKafkaZookeeperConfigmapSslYamlBytes() ([]byte, error) {
	return _testdataExternalLogStoresKafkaZookeeperConfigmapSslYaml, nil
}

func testdataExternalLogStoresKafkaZookeeperConfigmapSslYaml() (*asset, error) {
	bytes, err := testdataExternalLogStoresKafkaZookeeperConfigmapSslYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "testdata/external-log-stores/kafka/zookeeper/configmap-ssl.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testdataExternalLogStoresKafkaZookeeperConfigmapYaml = []byte(`kind: Template
apiVersion: template.openshift.io/v1
metadata:
  name: zookeeper-configmap-template
objects:
- apiVersion: v1
  data:
    init.sh: |
      #!/bin/bash
      set -e
      [ -d /var/lib/zookeeper/data ] || mkdir /var/lib/zookeeper/data
      [ -z "$ID_OFFSET" ] && ID_OFFSET=1
      export ZOOKEEPER_SERVER_ID=$((${HOSTNAME##*-} + $ID_OFFSET))
      echo "${ZOOKEEPER_SERVER_ID:-1}" | tee /var/lib/zookeeper/data/myid
      cp -Lur /etc/kafka-configmap/* /etc/kafka/
    log4j.properties: |
      log4j.rootLogger=INFO, stdout
      log4j.appender.stdout=org.apache.log4j.ConsoleAppender
      log4j.appender.stdout.layout=org.apache.log4j.PatternLayout
      log4j.appender.stdout.layout.ConversionPattern=[%d] %p %m (%c)%n
      # Suppress connection log messages, three lines per livenessProbe execution
      log4j.logger.org.apache.zookeeper.server.NIOServerCnxnFactory=WARN
      log4j.logger.org.apache.zookeeper.server.NIOServerCnxn=WARN
    zookeeper.properties: |
      4lw.commands.whitelist=ruok
      tickTime=2000
      dataDir=/var/lib/zookeeper/data
      dataLogDir=/var/lib/zookeeper/log
      clientPort=2181
  kind: ConfigMap
  metadata:
    name: ${NAME}
    namespace: ${NAMESPACE}
parameters:
- name: NAME
  value: "zookeeper"
- name: NAMESPACE
  value: "openshift-logging"
`)

func testdataExternalLogStoresKafkaZookeeperConfigmapYamlBytes() ([]byte, error) {
	return _testdataExternalLogStoresKafkaZookeeperConfigmapYaml, nil
}

func testdataExternalLogStoresKafkaZookeeperConfigmapYaml() (*asset, error) {
	bytes, err := testdataExternalLogStoresKafkaZookeeperConfigmapYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "testdata/external-log-stores/kafka/zookeeper/configmap.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testdataExternalLogStoresKafkaZookeeperZookeeperStatefulsetYaml = []byte(`kind: Template
apiVersion: template.openshift.io/v1
metadata:
  name: kafkaserver-template
objects:
- apiVersion: apps/v1
  kind: StatefulSet
  metadata:
    labels:
      app: zookeeper
      component: zookeeper
      provider: openshift
    name: ${NAME}
    namespace: ${NAMESPACE}
  spec:
    podManagementPolicy: Parallel
    replicas: 1
    revisionHistoryLimit: 10
    selector:
      matchLabels:
        app: zookeeper
    serviceName: ${SERVICENAME}
    template:
      metadata:
        creationTimestamp: null
        labels:
          app: zookeeper
          component: zookeeper
          provider: openshift
      spec:
        securityContext:
          runAsNonRoot: true
          seccompProfile:
            type: RuntimeDefault
        containers:
        - command:
          - /opt/kafka/bin/zookeeper-server-start.sh
          - /etc/kafka/zookeeper.properties
          env:
          - name: KAFKA_LOG4J_OPTS
            value: -Dlog4j.configuration=file:/etc/kafka/log4j.properties
          image: quay.io/openshifttest/kafka@sha256:2411662d89dd5700e1fe49aa8be1219843948da90cfe51a1c7a49bcef9d22dab
          imagePullPolicy: IfNotPresent
          securityContext:
            allowPrivilegeEscalation: false
            runAsNonRoot: true
            capabilities:
              drop:
              - ALL
            seccompProfile:
              type: RuntimeDefault
          lifecycle:
            preStop:
              exec:
                command:
                - sh
                - -ce
                - kill -s TERM 1; while $(kill -0 1 2>/dev/null); do sleep 1; done
          name: zookeeper
          ports:
          - containerPort: 2181
            name: client
            protocol: TCP
          - containerPort: 2888
            name: peer
            protocol: TCP
          - containerPort: 3888
            name: leader-election
            protocol: TCP
          resources:
            limits:
              memory: 120Mi
            requests:
              cpu: 10m
              memory: 100Mi
          terminationMessagePath: /dev/termination-log
          terminationMessagePolicy: File
          volumeMounts:
          - mountPath: /etc/kafka
            name: config
          - mountPath: /opt/kafka/logs
            name: zookeeperlogs
          - mountPath: /var/lib/zookeeper
            name: data
          - mountPath: /etc/kafka-configmap
            name: configmap
        dnsPolicy: ClusterFirst
        initContainers:
        - command:
          - /bin/bash
          - /etc/kafka-configmap/init.sh
          image: quay.io/openshifttest/kafka-initutils@sha256:e73ff7a44b43b85b53849c0459ba32e704540852b885a5c78af9753f86a49d68
          imagePullPolicy: IfNotPresent
          name: init-config
          securityContext:
            allowPrivilegeEscalation: false
            runAsNonRoot: true
            capabilities:
              drop:
              - ALL
            seccompProfile:
              type: RuntimeDefault
          resources: {}
          terminationMessagePath: /dev/termination-log
          terminationMessagePolicy: File
          volumeMounts:
          - mountPath: /etc/kafka-configmap
            name: configmap
          - mountPath: /etc/kafka
            name: config
          - mountPath: /var/lib/zookeeper
            name: data
        restartPolicy: Always
        nodeSelector:
          kubernetes.io/arch: amd64
          kubernetes.io/os: linux
        schedulerName: default-scheduler
        terminationGracePeriodSeconds: 10
        volumes:
        - configMap:
            defaultMode: 420
            name: ${CM_NAME}
          name: configmap
        - emptyDir: {}
          name: config
        - emptyDir: {}
          name: zookeeperlogs
        - emptyDir: {}
          name: data
    updateStrategy:
      type: RollingUpdate
parameters:
- name: NAME
  value: "zookeeper"
- name: NAMESPACE
  value: "openshift-logging"
- name: SERVICENAME
  value: "zookeeper"
- name: CM_NAME
  value: "zookeeper"
`)

func testdataExternalLogStoresKafkaZookeeperZookeeperStatefulsetYamlBytes() ([]byte, error) {
	return _testdataExternalLogStoresKafkaZookeeperZookeeperStatefulsetYaml, nil
}

func testdataExternalLogStoresKafkaZookeeperZookeeperStatefulsetYaml() (*asset, error) {
	bytes, err := testdataExternalLogStoresKafkaZookeeperZookeeperStatefulsetYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "testdata/external-log-stores/kafka/zookeeper/zookeeper-statefulset.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testdataExternalLogStoresKafkaZookeeperZookeeperSvcYaml = []byte(`kind: Template
apiVersion: template.openshift.io/v1
metadata:
  name: kafkaserver-template
objects:
- apiVersion: v1
  kind: Service
  metadata:
    labels:
      logging-infra: support
    name: ${NAME}
    namespace: ${NAMESPACE}
  spec:
    ports:
    - name: client
      port: 2181
      protocol: TCP
      targetPort: 2181
    - name: peer
      port: 2888
      protocol: TCP
      targetPort: 2888
    - name: leader-election
      port: 3888
      protocol: TCP
      targetPort: 3888
    selector:
      component: zookeeper
      provider: openshift
    sessionAffinity: None
    type: ClusterIP
parameters:
- name: NAME
  value: "zookeeper"
- name: NAMESPACE
  value: "openshift-logging"
`)

func testdataExternalLogStoresKafkaZookeeperZookeeperSvcYamlBytes() ([]byte, error) {
	return _testdataExternalLogStoresKafkaZookeeperZookeeperSvcYaml, nil
}

func testdataExternalLogStoresKafkaZookeeperZookeeperSvcYaml() (*asset, error) {
	bytes, err := testdataExternalLogStoresKafkaZookeeperZookeeperSvcYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "testdata/external-log-stores/kafka/zookeeper/zookeeper-svc.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testdataExternalLogStoresLokiLokiConfigmapYaml = []byte(`kind: Template
apiVersion: template.openshift.io/v1
metadata:
  name: loki-config-template
objects:
- apiVersion: v1
  kind: ConfigMap
  metadata:
    name: ${LOKICMNAME}
    namespace: ${LOKINAMESPACE}
  data:
    local-config.yaml: |
      auth_enabled: false

      server:
        http_listen_port: 3100
        grpc_listen_port: 9096
        grpc_server_max_recv_msg_size: 8388608

      ingester:
        wal:
          enabled: true
          dir: /tmp/wal
        lifecycler:
          address: 127.0.0.1
          ring:
            kvstore:
              store: inmemory
            replication_factor: 1
          final_sleep: 0s
        chunk_idle_period: 1h       # Any chunk not receiving new logs in this time will be flushed
        chunk_target_size: 8388608
        max_chunk_age: 1h           # All chunks will be flushed when they hit this age, default is 1h
        chunk_retain_period: 30s    # Must be greater than index read cache TTL if using an index cache (Default index read cache TTL is 5m)
        max_transfer_retries: 0     # Chunk transfers disabled

      schema_config:
        configs:
          - from: 2020-10-24
            store: boltdb-shipper
            object_store: filesystem
            schema: v11
            index:
              prefix: index_
              period: 24h

      storage_config:
        boltdb_shipper:
          active_index_directory: /tmp/loki/boltdb-shipper-active
          cache_location: /tmp/loki/boltdb-shipper-cache
          cache_ttl: 24h         # Can be increased for faster performance over longer query periods, uses more disk space
          shared_store: filesystem
        filesystem:
          directory: /tmp/loki/chunks

      compactor:
        working_directory: /tmp/loki/boltdb-shipper-compactor
        shared_store: filesystem

      limits_config:
        reject_old_samples: true
        reject_old_samples_max_age: 12h
        ingestion_rate_mb: 8
        ingestion_burst_size_mb: 16

      chunk_store_config:
        max_look_back_period: 0s

      table_manager:
        retention_deletes_enabled: false
        retention_period: 0s

      ruler:
        storage:
          type: local
          local:
            directory: /tmp/loki/rules
        rule_path: /tmp/loki/rules-temp
        alertmanager_url: http://localhost:9093
        ring:
          kvstore:
            store: inmemory
        enable_api: true
parameters:
- name: LOKINAMESPACE
  value: "loki-aosqe"
- name: LOKICMNAME
  value: "loki-config"
`)

func testdataExternalLogStoresLokiLokiConfigmapYamlBytes() ([]byte, error) {
	return _testdataExternalLogStoresLokiLokiConfigmapYaml, nil
}

func testdataExternalLogStoresLokiLokiConfigmapYaml() (*asset, error) {
	bytes, err := testdataExternalLogStoresLokiLokiConfigmapYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "testdata/external-log-stores/loki/loki-configmap.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testdataExternalLogStoresLokiLokiDeploymentYaml = []byte(`kind: Template
apiVersion: template.openshift.io/v1
metadata:
  name: loki-log-store-template
objects:
- apiVersion: apps/v1
  kind: Deployment
  metadata:
    name:  ${LOKISERVERNAME}
    namespace: ${LOKINAMESPACE}
    labels:
      provider: aosqe
      component: "loki"
      appname: loki-server
  spec:
    replicas: 1
    revisionHistoryLimit: 10
    selector:
      matchLabels:
        provider: aosqe
        component: "loki"
    strategy:
      type: Recreate
    template:
      metadata:
        labels:
          provider: aosqe
          component: "loki"
          appname: loki-server
      spec:
        securityContext:
          runAsNonRoot: true
          seccompProfile:
            type: RuntimeDefault
        containers:
        - name: "loki"
          image: quay.io/openshifttest/grafana-loki@sha256:bbf6dbf3264af939a541b6f3c014cba21a2bdc8f22cb7962eee7e9048b41ea5e
          imagePullPolicy: IfNotPresent
          securityContext:
            allowPrivilegeEscalation: false
            runAsNonRoot: true
            capabilities:
              drop:
              - ALL
            seccompProfile:
              type: RuntimeDefault
          ports:
          - containerPort: 3100
            name: tcp
            protocol: TCP
          volumeMounts:
          - mountPath: /etc/loki
            name: lokiconfig
            readOnly: true
        volumes:
        - configMap:
            defaultMode: 420
            name: ${LOKICMNAME}
          name: lokiconfig
parameters:
- name: LOKISERVERNAME
  value: "loki-server"
- name: LOKINAMESPACE
  value: "loki-aosqe"
- name: LOKICMNAME
  value: "loki-config"
`)

func testdataExternalLogStoresLokiLokiDeploymentYamlBytes() ([]byte, error) {
	return _testdataExternalLogStoresLokiLokiDeploymentYaml, nil
}

func testdataExternalLogStoresLokiLokiDeploymentYaml() (*asset, error) {
	bytes, err := testdataExternalLogStoresLokiLokiDeploymentYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "testdata/external-log-stores/loki/loki-deployment.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testdataExternalLogStoresOtelOtelCollectorYaml = []byte(`kind: Template
apiVersion: template.openshift.io/v1
metadata:
  name: openTelemetryCollector-template
objects:
- apiVersion: opentelemetry.io/v1beta1
  kind: OpenTelemetryCollector
  metadata:
    name: ${NAME}
    namespace: ${NAMESPACE}
  spec:
    config:
      exporters:
        debug:
          verbosity: detailed
      receivers:
        otlp:
          protocols:
            http:
              endpoint: 0.0.0.0:4318
      service:
        pipelines:
          logs:
            exporters:
            - debug
            processors: []
            receivers:
            - otlp
    managementState: managed
    mode: deployment
    replicas: 1
    upgradeStrategy: automatic
parameters:
- name: NAME
  value: "otel"
- name: NAMESPACE
  value: "openshift-opentelemetry-operator"
`)

func testdataExternalLogStoresOtelOtelCollectorYamlBytes() ([]byte, error) {
	return _testdataExternalLogStoresOtelOtelCollectorYaml, nil
}

func testdataExternalLogStoresOtelOtelCollectorYaml() (*asset, error) {
	bytes, err := testdataExternalLogStoresOtelOtelCollectorYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "testdata/external-log-stores/otel/otel-collector.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testdataExternalLogStoresRsyslogInsecureConfigmapYaml = []byte(`kind: Template
apiVersion: template.openshift.io/v1
metadata:
  name: rsyslogserver-template
objects:
- apiVersion: v1
  kind: ConfigMap
  metadata:
    name: ${NAME}
    namespace: ${NAMESPACE}
    labels:
      provider: aosqe
      component: ${NAME}
  data:
    rsyslog.conf: |+
      global(processInternalMessages="on")
      module(load="imptcp")
      module(load="imudp" TimeRequery="500")
      input(type="imptcp" port="10514")
      input(type="imudp" port="10514")
      :msg, contains, "\"log_type\":\"application\"" /var/log/clf/app-container.log
      :msg, contains, "\"log_type\":\"infrastructure\""{
        if $msg contains "\"log_source\":\"container\"" then /var/log/clf/infra-container.log
        if $msg contains "\"log_source\":\"node\"" then /var/log/clf/infra.log
      }
      :msg, contains, "\"log_type\":\"audit\"" /var/log/clf/audit.log
      :msg, contains, "\"log_source\":\"auditd\"" /var/log/clf/audit-linux.log
      :msg, contains, "\"log_source\":\"kubeAPI\"" /var/log/clf/audit-kubeAPI.log
      :msg, contains, "\"log_source\":\"openshiftAPI\"" /var/log/clf/audit-openshiftAPI.log
      :msg, contains, "\"log_source\":\"ovn\"" /var/log/clf/audit-ovn.log
      *.* /var/log/clf/other.log
parameters:
- name: NAME
  value: "rsyslogserver"
- name: NAMESPACE
  value: "openshift-logging"
`)

func testdataExternalLogStoresRsyslogInsecureConfigmapYamlBytes() ([]byte, error) {
	return _testdataExternalLogStoresRsyslogInsecureConfigmapYaml, nil
}

func testdataExternalLogStoresRsyslogInsecureConfigmapYaml() (*asset, error) {
	bytes, err := testdataExternalLogStoresRsyslogInsecureConfigmapYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "testdata/external-log-stores/rsyslog/insecure/configmap.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testdataExternalLogStoresRsyslogInsecureDeploymentYaml = []byte(`kind: Template
apiVersion: template.openshift.io/v1
metadata:
  name: rsyslogserver-template
objects:
- kind: Deployment
  apiVersion: apps/v1
  metadata:
    name: ${NAME}
    namespace: ${NAMESPACE}
    labels:
      provider: aosqe
      component: ${NAME}
  spec:
    replicas: 1
    revisionHistoryLimit: 10
    selector:
      matchLabels:
        provider: aosqe
        component: ${NAME}
    strategy:
      type: Recreate
    template:
      metadata:
        labels:
          provider: aosqe
          component: ${NAME}
      spec:
        serviceAccount: ${NAME}
        serviceAccountName: ${NAME}
        securityContext:
          runAsNonRoot: true
          seccompProfile:
            type: RuntimeDefault
        containers:
        - name: "rsyslog"
          command: ["/usr/sbin/rsyslogd", "-f", "/etc/rsyslog/conf/rsyslog.conf", "-n"]
          image: quay.io/openshifttest/rsyslogd-container@sha256:e806eb41f05d7cc6eec96bf09c7bcb692f97562d4a983cb019289bd048d9aee2
          imagePullPolicy: IfNotPresent
          securityContext:
            allowPrivilegeEscalation: false
            runAsNonRoot: true
            capabilities:
              drop:
              - ALL
            seccompProfile:
              type: RuntimeDefault
          ports:
          - containerPort: 10514
            name: rsyslog-pod-tcp
            protocol: TCP
          - containerPort: 10514
            name: rsyslog-pod-udp
            protocol: UDP
          volumeMounts:
          - mountPath: /etc/rsyslog/conf
            name: main
            readOnly: true
        volumes:
        - configMap:
            defaultMode: 420
            name: ${NAME}
          name: main
parameters:
- name: NAME
  value: "rsyslogserver"
- name: NAMESPACE
  value: "openshift-logging"
`)

func testdataExternalLogStoresRsyslogInsecureDeploymentYamlBytes() ([]byte, error) {
	return _testdataExternalLogStoresRsyslogInsecureDeploymentYaml, nil
}

func testdataExternalLogStoresRsyslogInsecureDeploymentYaml() (*asset, error) {
	bytes, err := testdataExternalLogStoresRsyslogInsecureDeploymentYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "testdata/external-log-stores/rsyslog/insecure/deployment.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testdataExternalLogStoresRsyslogInsecureSvcYaml = []byte(`kind: Template
apiVersion: template.openshift.io/v1
metadata:
  name: rsyslogserver-template
objects:
- apiVersion: v1
  kind: Service
  metadata:
    labels:
      provider: aosqe
      component: ${NAME}
    name: ${NAME}
    namespace: ${NAMESPACE}
  spec:
    ports:
    - name: rsyslogserver-tcp
      port: 514
      targetPort: 10514
      protocol: TCP
    - name: rsyslogserver-udp
      port: 514
      targetPort: 10514
      protocol: UDP
    selector:
      component: ${NAME}
      provider: aosqe
parameters:
- name: NAME
  value: "rsyslogserver"
- name: NAMESPACE
  value: "openshift-logging"
`)

func testdataExternalLogStoresRsyslogInsecureSvcYamlBytes() ([]byte, error) {
	return _testdataExternalLogStoresRsyslogInsecureSvcYaml, nil
}

func testdataExternalLogStoresRsyslogInsecureSvcYaml() (*asset, error) {
	bytes, err := testdataExternalLogStoresRsyslogInsecureSvcYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "testdata/external-log-stores/rsyslog/insecure/svc.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testdataExternalLogStoresRsyslogSecureConfigmapYaml = []byte(`kind: Template
apiVersion: template.openshift.io/v1
metadata:
  name: rsyslogserver-template
objects:
- apiVersion: v1
  kind: ConfigMap
  metadata:
    name: ${NAME}
    namespace: ${NAMESPACE}
    labels:
      provider: aosqe
  data:
    rsyslog.conf: |+
      global(
        DefaultNetstreamDriverCAFile="/opt/app-root/tls/ca_bundle.crt"
        DefaultNetstreamDriverCertFile="/opt/app-root/tls/server.crt"
        DefaultNetstreamDriverKeyFile="/opt/app-root/tls/server.key"
      )
      module( load="imtcp"
        StreamDriver.Name = "gtls"
        StreamDriver.Mode = "1"
        #https://www.rsyslog.com/doc/master/concepts/ns_ossl.html
        StreamDriver.AuthMode = "anon"
      )
      module(load="imudp" TimeRequery="500")
      input(type="imtcp" port="6514")
      input(type="imudp" port="10514")
      :msg, contains, "\"log_type\":\"application\"" /var/log/clf/app-container.log
      :msg, contains, "\"log_type\":\"infrastructure\""{
        if $msg contains "\"log_source\":\"container\"" then /var/log/clf/infra-container.log
        if $msg contains "\"log_source\":\"node\"" then /var/log/clf/infra.log
      }
      :msg, contains, "\"log_type\":\"audit\"" /var/log/clf/audit.log
      :msg, contains, "\"log_source\":\"auditd\"" /var/log/clf/audit-linux.log
      :msg, contains, "\"log_source\":\"kubeAPI\"" /var/log/clf/audit-kubeAPI.log
      :msg, contains, "\"log_source\":\"openshiftAPI\"" /var/log/clf/audit-openshiftAPI.log
      :msg, contains, "\"log_source\":\"ovn\"" /var/log/clf/audit-ovn.log
      *.* /var/log/clf/other.log
parameters:
- name: NAME
  value: "rsyslogserver"
- name: NAMESPACE
  value: "openshift-logging"
`)

func testdataExternalLogStoresRsyslogSecureConfigmapYamlBytes() ([]byte, error) {
	return _testdataExternalLogStoresRsyslogSecureConfigmapYaml, nil
}

func testdataExternalLogStoresRsyslogSecureConfigmapYaml() (*asset, error) {
	bytes, err := testdataExternalLogStoresRsyslogSecureConfigmapYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "testdata/external-log-stores/rsyslog/secure/configmap.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testdataExternalLogStoresRsyslogSecureDeploymentYaml = []byte(`kind: Template
apiVersion: template.openshift.io/v1
metadata:
  name: rsyslogserver-template
objects:
- kind: Deployment
  apiVersion: apps/v1
  metadata:
    name: ${NAME}
    namespace: ${NAMESPACE}
    labels:
      provider: aosqe
      component: ${NAME}
  spec:
    replicas: 1
    revisionHistoryLimit: 10
    selector:
      matchLabels:
        provider: aosqe
        component: ${NAME}
    strategy:
      type: Recreate
    template:
      metadata:
        labels:
          provider: aosqe
          component: ${NAME}
      spec:
        serviceAccount: ${NAME}
        serviceAccountName: ${NAME}
        securityContext:
          runAsNonRoot: true
          seccompProfile:
            type: RuntimeDefault
        containers:
        - name: "rsyslog"
          command: ["/usr/sbin/rsyslogd", "-f", "/etc/rsyslog/conf/rsyslog.conf", "-n"]
          image: quay.io/openshifttest/rsyslogd-container@sha256:e806eb41f05d7cc6eec96bf09c7bcb692f97562d4a983cb019289bd048d9aee2
          imagePullPolicy: IfNotPresent
          securityContext:
            allowPrivilegeEscalation: false
            runAsNonRoot: true
            capabilities:
              drop:
              - ALL
            seccompProfile:
              type: RuntimeDefault
          ports:
          - containerPort: 10514
            name: rsyslog-pod-tcp
            protocol: TCP
          - containerPort: 10514
            name: rsyslog-pod-udp
            protocol: UDP
          - containerPort: 6514
            name: rsyslog-pod-tls
            protocol: TCP
          volumeMounts:
          - mountPath: /etc/rsyslog/conf
            name: main
            readOnly: true
          - mountPath: /opt/app-root/tls
            name: keys
            readOnly: true
        volumes:
        - configMap:
            defaultMode: 420
            name: ${NAME}
          name: main
        - secret:
            defaultMode: 420
            secretName: ${NAME}
          name: keys
parameters:
- name: NAME
  value: "rsyslogserver"
- name: NAMESPACE
  value: "openshift-logging"
`)

func testdataExternalLogStoresRsyslogSecureDeploymentYamlBytes() ([]byte, error) {
	return _testdataExternalLogStoresRsyslogSecureDeploymentYaml, nil
}

func testdataExternalLogStoresRsyslogSecureDeploymentYaml() (*asset, error) {
	bytes, err := testdataExternalLogStoresRsyslogSecureDeploymentYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "testdata/external-log-stores/rsyslog/secure/deployment.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testdataExternalLogStoresRsyslogSecureSvcYaml = []byte(`kind: Template
apiVersion: template.openshift.io/v1
metadata:
  name: rsyslogserver-template
objects:
- apiVersion: v1
  kind: Service
  metadata:
    labels:
      provider: aosqe
      component: ${NAME}
    name: ${NAME}
    namespace: ${NAMESPACE}
  spec:
    ports:
    - name: rsyslogserver-tls
      port: 6514
      targetPort: 6514
      protocol: TCP
    - name: rsyslogserver-tcp
      port: 514
      targetPort: 10514
      protocol: TCP
    - name: rsyslogserver-udp
      port: 514
      targetPort: 10514
      protocol: UDP
    selector:
      component: ${NAME}
      provider: aosqe
parameters:
- name: NAME
  value: "rsyslogserver"
- name: NAMESPACE
  value: "openshift-logging"
`)

func testdataExternalLogStoresRsyslogSecureSvcYamlBytes() ([]byte, error) {
	return _testdataExternalLogStoresRsyslogSecureSvcYaml, nil
}

func testdataExternalLogStoresRsyslogSecureSvcYaml() (*asset, error) {
	bytes, err := testdataExternalLogStoresRsyslogSecureSvcYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "testdata/external-log-stores/rsyslog/secure/svc.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testdataExternalLogStoresSplunkRouteEdge_splunk_templateYaml = []byte(`kind: Template
apiVersion: template.openshift.io/v1
metadata:
  name: route-edge-splunk-template
objects:
- apiVersion: route.openshift.io/v1
  kind: Route
  metadata:
    name: ${NAME}
  spec:
    host: ${ROUTE_HOST}
    port:
      targetPort: ${PORT_NAME} 
    tls:
      insecureEdgeTerminationPolicy: Allow
      termination: edge
    to:
      kind: Service
      name: ${SERVICE_NAME}
    wildcardPolicy: None
parameters:
- name: NAME
  value: "splunk-default-hec"
- name: PORT_NAME
  value: "http-hec"
- name: SERVICE_NAME
  value: "splunk-default-service"
- name: ROUTE_HOST
  value: ""
`)

func testdataExternalLogStoresSplunkRouteEdge_splunk_templateYamlBytes() ([]byte, error) {
	return _testdataExternalLogStoresSplunkRouteEdge_splunk_templateYaml, nil
}

func testdataExternalLogStoresSplunkRouteEdge_splunk_templateYaml() (*asset, error) {
	bytes, err := testdataExternalLogStoresSplunkRouteEdge_splunk_templateYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "testdata/external-log-stores/splunk/route-edge_splunk_template.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testdataExternalLogStoresSplunkRoutePassthrough_splunk_templateYaml = []byte(`kind: Template
apiVersion: template.openshift.io/v1
metadata:
  name: route-splunk-passthrough-template
objects:
- apiVersion: route.openshift.io/v1
  kind: Route
  metadata:
    name: ${NAME}
  spec:
    host: ${ROUTE_HOST}
    port:
      targetPort: ${PORT_NAME}
    tls:
      termination: passthrough
    to:
      kind: Service
      name: ${SERVICE_NAME}
parameters:
- name: NAME
  value: "splunk-default-hec"
- name: PORT_NAME
  value: "http-hec"
- name: SERVICE_NAME
  value: "splunk-default-service"
- name: ROUTE_HOST
  value: ""
`)

func testdataExternalLogStoresSplunkRoutePassthrough_splunk_templateYamlBytes() ([]byte, error) {
	return _testdataExternalLogStoresSplunkRoutePassthrough_splunk_templateYaml, nil
}

func testdataExternalLogStoresSplunkRoutePassthrough_splunk_templateYaml() (*asset, error) {
	bytes, err := testdataExternalLogStoresSplunkRoutePassthrough_splunk_templateYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "testdata/external-log-stores/splunk/route-passthrough_splunk_template.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testdataExternalLogStoresSplunkSecret_splunk_templateYaml = []byte(`kind: Template
apiVersion: template.openshift.io/v1
metadata:
  name: splunk-secret-template
objects:
- apiVersion: v1
  kind: Secret
  metadata:
    name: ${NAME}
  type: Opaque
  stringData:
    default.yml: |
      splunk:
        listenOnIPv6: "yes"
        hec_token: "${HEC_TOKEN}"
        password: "${PASSWORD}"
        pass4SymmKey: "${PASSWORD}"
        idxc:
            secret: "${PASSWORD}"
        shc:
            secret: "${PASSWORD}"
        hec:
            requireClientCert: False
            ssl: False
    hec_token: ${HEC_TOKEN}
    idxc_secret: ${PASSWORD}
    pass4SymmKey: ${PASSWORD}
    password: ${PASSWORD}
    shc_secret: ${PASSWORD}
parameters:
- name: NAME
  value: "splunk-default"
- name: HEC_TOKEN
  value: "555555555-BBBB-BBBB-BBBB-555555555555"
- name: PASSWORD
  value: ""
`)

func testdataExternalLogStoresSplunkSecret_splunk_templateYamlBytes() ([]byte, error) {
	return _testdataExternalLogStoresSplunkSecret_splunk_templateYaml, nil
}

func testdataExternalLogStoresSplunkSecret_splunk_templateYaml() (*asset, error) {
	bytes, err := testdataExternalLogStoresSplunkSecret_splunk_templateYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "testdata/external-log-stores/splunk/secret_splunk_template.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testdataExternalLogStoresSplunkSecret_tls_passphrase_splunk_templateYaml = []byte(`kind: Template
apiVersion: template.openshift.io/v1
metadata:
  name: splunk-secret-template
objects:
- apiVersion: v1
  kind: Secret
  metadata:
    name: ${NAME}
  type: Opaque
  stringData:
    default.yml: |
      splunk:
        listenOnIPv6: "yes"
        hec:
            enable: true
            token: "${HEC_TOKEN}"
            requireClientCert: ${HEC_CLIENTAUTH}
            cert: "/mnt/splunk-secrets/hec.pem"
            ssl: true
        http_enableSSL: 1
        http_enableSSL_cert: "/mnt/splunk-secrets/cert.pem"
        http_enableSSL_privKey: "/mnt/splunk-secrets/key.pem"
        http_enableSSL_privKey_password: ${PASSPHASE}
        password: "${PASSWORD}"
        pass4SymmKey: "${PASSWORD}"
        idxc:
            secret: "${PASSWORD}"
        shc:
            secret: "${PASSWORD}"
    hec_token: ${HEC_TOKEN}
    idxc_secret: ${PASSWORD}
    pass4SymmKey: ${PASSWORD}
    password: ${PASSWORD}
    shc_secret: ${PASSWORD}
parameters:
- name: NAME
  value: "splunk-default"
- name: PASSWORD
  value: "password"
- name: HEC_TOKEN
  value: "555555555-BBBB-BBBB-BBBB-555555555555"
- name: PASSPHASE
  value: "password"
- name: HEC_CLIENTAUTH
  value: "False"
`)

func testdataExternalLogStoresSplunkSecret_tls_passphrase_splunk_templateYamlBytes() ([]byte, error) {
	return _testdataExternalLogStoresSplunkSecret_tls_passphrase_splunk_templateYaml, nil
}

func testdataExternalLogStoresSplunkSecret_tls_passphrase_splunk_templateYaml() (*asset, error) {
	bytes, err := testdataExternalLogStoresSplunkSecret_tls_passphrase_splunk_templateYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "testdata/external-log-stores/splunk/secret_tls_passphrase_splunk_template.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testdataExternalLogStoresSplunkSecret_tls_splunk_templateYaml = []byte(`kind: Template
apiVersion: template.openshift.io/v1
metadata:
  name: splunk-secret-template
objects:
- apiVersion: v1
  kind: Secret
  metadata:
    name: ${NAME}
  type: Opaque
  stringData:
    default.yml: |
      splunk:
        listenOnIPv6: "yes"
        hec:
            enable: true
            ssl: true
            token: "${HEC_TOKEN}"
            requireClientCert: ${HEC_CLIENTAUTH}
            cert: "/mnt/splunk-secrets/hec.pem"
        http_enableSSL: 1
        http_enableSSL_cert: "/mnt/splunk-secrets/cert.pem"
        http_enableSSL_privKey: "/mnt/splunk-secrets/key.pem"
        password: "${PASSWORD}"
        pass4SymmKey: "${PASSWORD}"
        idxc:
            secret: "${PASSWORD}"
        shc:
            secret: "${PASSWORD}"
    hec_token: "${HEC_TOKEN}"
    idxc_secret: ${PASSWORD}
    pass4SymmKey: ${PASSWORD}
    password: ${PASSWORD}
    shc_secret: ${PASSWORD}
parameters:
- name: NAME
  value: "splunk-default"
- name: PASSWORD
  value: "password"
- name: HEC_TOKEN
  value: "555555555-BBBB-BBBB-BBBB-555555555555"
- name: HEC_CLIENTAUTH
  value: "False"
`)

func testdataExternalLogStoresSplunkSecret_tls_splunk_templateYamlBytes() ([]byte, error) {
	return _testdataExternalLogStoresSplunkSecret_tls_splunk_templateYaml, nil
}

func testdataExternalLogStoresSplunkSecret_tls_splunk_templateYaml() (*asset, error) {
	bytes, err := testdataExternalLogStoresSplunkSecret_tls_splunk_templateYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "testdata/external-log-stores/splunk/secret_tls_splunk_template.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testdataExternalLogStoresSplunkStatefulset_splunk82_templateYaml = []byte(`kind: Template
apiVersion: template.openshift.io/v1
metadata:
  name: splunk-s1-standalone-template
objects:
- apiVersion: apps/v1
  kind: StatefulSet
  metadata:
    name: ${NAME}
  spec:
    podManagementPolicy: Parallel
    replicas: 1
    selector:
      matchLabels:
        app.kubernetes.io/component: splunk
        app.kubernetes.io/instance: ${NAME}
        app.kubernetes.io/name: splunk
    serviceName: ${NAME}-headless
    template:
      metadata:
        annotations:
          traffic.sidecar.istio.io/excludeOutboundPorts: 8089,8191,9997
          traffic.sidecar.istio.io/includeInboundPorts: 8000,8088
        labels:
          app.kubernetes.io/component: splunk
          app.kubernetes.io/instance: ${NAME}
          app.kubernetes.io/name: splunk
      spec:
        containers:
        - env:
          - name: DEBUG
            value: "false"
          - name: ANSIBLE_EXTRA_FLAGS
            value: "-v"
          - name: SPLUNK_DECLARATIVE_ADMIN_PASSWORD
            value: "true"
          - name: SPLUNK_DEFAULTS_URL
            value: /mnt/splunk-secrets/default.yml
          - name: SPLUNK_HOME
            value: /opt/splunk
          - name: SPLUNK_HOME_OWNERSHIP_ENFORCEMENT
            value: "false"
          - name: SPLUNK_ROLE
            value: splunk_standalone
          - name: SPLUNK_START_ARGS
            value: --accept-license
          image: quay.io/openshifttest/splunk@sha256:fbfae0b70a4884a3d23a05d3f45fa35646ea56ccd98ab73fb147b31715a41c42
          imagePullPolicy: IfNotPresent
          livenessProbe:
            exec:
              command:
              - /sbin/checkstate.sh
            failureThreshold: 3
            initialDelaySeconds: 300
            periodSeconds: 30
            successThreshold: 1
            timeoutSeconds: 30
          name: splunk
          ports:
          - containerPort: 8000
            name: http-splunkweb
            protocol: TCP
          - containerPort: 8088
            name: http-hec
            protocol: TCP
          - containerPort: 8089
            name: https-splunkd
            protocol: TCP
          - containerPort: 9997
            name: tcp-s2s
            protocol: TCP
          readinessProbe:
            exec:
              command:
              - /bin/grep
              - started
              - /opt/container_artifact/splunk-container.state
            failureThreshold: 3
            initialDelaySeconds: 10
            periodSeconds: 5
            successThreshold: 1
            timeoutSeconds: 5
          resources:
            limits:
              cpu: "4"
              memory: 8Gi
            requests:
              cpu: 100m
              memory: 512Mi
          securityContext:
            allowPrivilegeEscalation: false
            capabilities:
              drop:
              - ALL
            runAsNonRoot: true
          terminationMessagePath: /dev/termination-log
          terminationMessagePolicy: File
          volumeMounts:
          - mountPath: /mnt/splunk-secrets
            name: mnt-splunk-secrets
          - mountPath: /opt/splunk/etc
            name: pvc-etc
          - mountPath: /opt/splunk/var
            name: pvc-var
        dnsPolicy: ClusterFirst
        nodeSelector:
          kubernetes.io/arch: amd64
          kubernetes.io/os: linux
        restartPolicy: Always
        schedulerName: default-scheduler
        securityContext:
          fsGroup: 41812
          runAsNonRoot: true
          runAsUser: 41812
        terminationGracePeriodSeconds: 30
        volumes:
        - name: mnt-splunk-secrets
          secret:
            defaultMode: 420
            secretName: ${NAME}
    updateStrategy:
      type: OnDelete
    volumeClaimTemplates:
    - apiVersion: v1
      kind: PersistentVolumeClaim
      metadata:
        labels:
          app.kubernetes.io/component: splunk
          app.kubernetes.io/instance: ${NAME}
          app.kubernetes.io/name: splunk
        name: pvc-etc
      spec:
        accessModes:
        - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
        volumeMode: Filesystem
    - apiVersion: v1
      kind: PersistentVolumeClaim
      metadata:
        labels:
          app.kubernetes.io/component: splunk
          app.kubernetes.io/instance: ${NAME}
          app.kubernetes.io/name: splunk
        name: pvc-var
      spec:
        accessModes:
        - ReadWriteOnce
        resources:
          requests:
            storage: 10Gi
        volumeMode: Filesystem
- apiVersion: v1
  kind: Service
  metadata:
    labels:
      app.kubernetes.io/component: splunk
      app.kubernetes.io/instance: ${NAME}
      app.kubernetes.io/name: splunk
    name: ${NAME}-headless
  spec:
    type: ClusterIP
    clusterIP: None
    selector:
      app.kubernetes.io/component: splunk
      app.kubernetes.io/instance: ${NAME}
      app.kubernetes.io/name: splunk
    ports:
    - name: http-splunkweb
      port: 8000
      protocol: TCP
      targetPort: 8000
    - name: http-hec
      port: 8088
      protocol: TCP
      targetPort: 8088
    - name: https-splunkd
      port: 8089
      protocol: TCP
      targetPort: 8089
    - name: tcp-s2s
      port: 9997
      protocol: TCP
      targetPort: 9997
- apiVersion: v1
  kind: Service
  metadata:
    labels:
      app.kubernetes.io/component: splunk
      app.kubernetes.io/instance: ${NAME}
      app.kubernetes.io/name: splunk
    name: ${NAME}-0
  spec:
    type: ClusterIP
    selector:
      app.kubernetes.io/component: splunk
      app.kubernetes.io/instance: ${NAME}
      app.kubernetes.io/name: splunk
    internalTrafficPolicy: Cluster
    ports:
    - name: http-splunkweb
      port: 8000
      protocol: TCP
      targetPort: 8000
    - name: http-hec
      port: 8088
      protocol: TCP
      targetPort: 8088
    - name: https-splunkd
      port: 8089
      protocol: TCP
      targetPort: 8089
    - name: tcp-s2s
      port: 9997
      protocol: TCP
      targetPort: 9997
parameters:
- name: NAME
  value: "splunk-s1-standalone"
`)

func testdataExternalLogStoresSplunkStatefulset_splunk82_templateYamlBytes() ([]byte, error) {
	return _testdataExternalLogStoresSplunkStatefulset_splunk82_templateYaml, nil
}

func testdataExternalLogStoresSplunkStatefulset_splunk82_templateYaml() (*asset, error) {
	bytes, err := testdataExternalLogStoresSplunkStatefulset_splunk82_templateYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "testdata/external-log-stores/splunk/statefulset_splunk-8.2_template.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testdataExternalLogStoresSplunkStatefulset_splunk90_templateYaml = []byte(`kind: Template
apiVersion: template.openshift.io/v1
metadata:
  name: splunk-s1-standalone-template
objects:
- apiVersion: apps/v1
  kind: StatefulSet
  metadata:
    name: ${NAME}
  spec:
    podManagementPolicy: Parallel
    replicas: 1
    selector:
      matchLabels:
        app.kubernetes.io/component: splunk
        app.kubernetes.io/instance: ${NAME}
        app.kubernetes.io/name: splunk
    serviceName: ${NAME}-headless
    template:
      metadata:
        annotations:
          traffic.sidecar.istio.io/excludeOutboundPorts: 8089,8191,9997
          traffic.sidecar.istio.io/includeInboundPorts: 8000,8088
        labels:
          app.kubernetes.io/component: splunk
          app.kubernetes.io/instance: ${NAME}
          app.kubernetes.io/name: splunk
      spec:
        containers:
        - env:
          - name: DEBUG
            value: "false"
          - name: ANSIBLE_EXTRA_FLAGS
            value: "-v"
          - name: SPLUNK_DECLARATIVE_ADMIN_PASSWORD
            value: "true"
          - name: SPLUNK_DEFAULTS_URL
            value: /mnt/splunk-secrets/default.yml
          - name: SPLUNK_HOME
            value: /opt/splunk
          - name: SPLUNK_HOME_OWNERSHIP_ENFORCEMENT
            value: "false"
          - name: SPLUNK_ROLE
            value: splunk_standalone
          - name: SPLUNK_START_ARGS
            value: --accept-license
          image: quay.io/openshifttest/splunk@sha256:5762a3b61ad5090f24ad33360fc03f3ced469e16c3c75f6d8590b5ef39d95751
          imagePullPolicy: IfNotPresent
          livenessProbe:
            exec:
              command:
              - /sbin/checkstate.sh
            failureThreshold: 3
            initialDelaySeconds: 300
            periodSeconds: 30
            successThreshold: 1
            timeoutSeconds: 30
          name: splunk
          ports:
          - containerPort: 8000
            name: http-splunkweb
            protocol: TCP
          - containerPort: 8088
            name: http-hec
            protocol: TCP
          - containerPort: 8089
            name: https-splunkd
            protocol: TCP
          - containerPort: 9997
            name: tcp-s2s
            protocol: TCP
          readinessProbe:
            exec:
              command:
              - /bin/grep
              - started
              - /opt/container_artifact/splunk-container.state
            failureThreshold: 3
            initialDelaySeconds: 10
            periodSeconds: 5
            successThreshold: 1
            timeoutSeconds: 5
          resources:
            limits:
              cpu: "4"
              memory: 8Gi
            requests:
              cpu: 100m
              memory: 512Mi
          securityContext:
            allowPrivilegeEscalation: false
            capabilities:
              drop:
              - ALL
            runAsNonRoot: true
          terminationMessagePath: /dev/termination-log
          terminationMessagePolicy: File
          volumeMounts:
          - mountPath: /mnt/splunk-secrets
            name: mnt-splunk-secrets
          - mountPath: /opt/splunk/etc
            name: pvc-etc
          - mountPath: /opt/splunk/var
            name: pvc-var
        dnsPolicy: ClusterFirst
        nodeSelector:
          kubernetes.io/arch: amd64
          kubernetes.io/os: linux
        restartPolicy: Always
        schedulerName: default-scheduler
        securityContext:
          fsGroup: 41812
          runAsNonRoot: true
          runAsUser: 41812
        terminationGracePeriodSeconds: 30
        volumes:
        - name: mnt-splunk-secrets
          secret:
            defaultMode: 420
            secretName: ${NAME}
    updateStrategy:
      type: OnDelete
    volumeClaimTemplates:
    - apiVersion: v1
      kind: PersistentVolumeClaim
      metadata:
        labels:
          app.kubernetes.io/component: splunk
          app.kubernetes.io/instance: ${NAME}
          app.kubernetes.io/name: splunk
        name: pvc-etc
      spec:
        accessModes:
        - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
        volumeMode: Filesystem
    - apiVersion: v1
      kind: PersistentVolumeClaim
      metadata:
        labels:
          app.kubernetes.io/component: splunk
          app.kubernetes.io/instance: ${NAME}
          app.kubernetes.io/name: splunk
        name: pvc-var
      spec:
        accessModes:
        - ReadWriteOnce
        resources:
          requests:
            storage: 10Gi
        volumeMode: Filesystem
- apiVersion: v1
  kind: Service
  metadata:
    labels:
      app.kubernetes.io/component: splunk
      app.kubernetes.io/instance: ${NAME}
      app.kubernetes.io/name: splunk
    name: ${NAME}-headless
  spec:
    type: ClusterIP
    clusterIP: None
    selector:
      app.kubernetes.io/component: splunk
      app.kubernetes.io/instance: ${NAME}
      app.kubernetes.io/name: splunk
    ports:
    - name: http-splunkweb
      port: 8000
      protocol: TCP
      targetPort: 8000
    - name: http-hec
      port: 8088
      protocol: TCP
      targetPort: 8088
    - name: https-splunkd
      port: 8089
      protocol: TCP
      targetPort: 8089
    - name: tcp-s2s
      port: 9997
      protocol: TCP
      targetPort: 9997
- apiVersion: v1
  kind: Service
  metadata:
    labels:
      app.kubernetes.io/component: splunk
      app.kubernetes.io/instance: ${NAME}
      app.kubernetes.io/name: splunk
    name: ${NAME}-0
  spec:
    type: ClusterIP
    selector:
      app.kubernetes.io/component: splunk
      app.kubernetes.io/instance: ${NAME}
      app.kubernetes.io/name: splunk
    internalTrafficPolicy: Cluster
    ports:
    - name: http-splunkweb
      port: 8000
      protocol: TCP
      targetPort: 8000
    - name: http-hec
      port: 8088
      protocol: TCP
      targetPort: 8088
    - name: https-splunkd
      port: 8089
      protocol: TCP
      targetPort: 8089
    - name: tcp-s2s
      port: 9997
      protocol: TCP
      targetPort: 9997
parameters:
- name: NAME
  value: "splunk-s1-standalone"
`)

func testdataExternalLogStoresSplunkStatefulset_splunk90_templateYamlBytes() ([]byte, error) {
	return _testdataExternalLogStoresSplunkStatefulset_splunk90_templateYaml, nil
}

func testdataExternalLogStoresSplunkStatefulset_splunk90_templateYaml() (*asset, error) {
	bytes, err := testdataExternalLogStoresSplunkStatefulset_splunk90_templateYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "testdata/external-log-stores/splunk/statefulset_splunk-9.0_template.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testdataGeneratelog42981Yaml = []byte(`kind: Template
apiVersion: template.openshift.io/v1
metadata:
  name: ovn-audit-log-gen-template
objects:
  - kind: Namespace
    apiVersion: v1
    metadata:
      annotations:
        k8s.ovn.org/acl-logging: '{ "deny": "alert", "allow": "alert" }'
      name: ${NAMESPACE}
    spec:
      finalizers:
      - kubernetes

  - kind: Deployment
    apiVersion: apps/v1
    metadata:
      labels:
        app: ovn-app
      name: ovn-app
      namespace: ${NAMESPACE}
    spec:
      replicas: 2
      selector:
        matchLabels:
          app: ovn-app
      strategy: {}
      template:
        metadata:
          labels:
            app: ovn-app
        spec:
          containers:
          - image: quay.io/openshifttest/hello-sdn@sha256:c89445416459e7adea9a5a416b3365ed3d74f2491beb904d61dc8d1eb89a72a4
            name: hello-sdn
            resources:
              limits:
                memory: 340Mi

  - kind: Service
    apiVersion: v1
    metadata:
      labels:
        name: test-service
      name: test-service
      namespace: ${NAMESPACE}
    spec:
      ports:
      - name: http
        port: 27017
        protocol: TCP
        targetPort: 8080
      selector:
        app: ovn-app

  - kind: NetworkPolicy
    apiVersion: networking.k8s.io/v1
    metadata:
      name: default-deny
      namespace: ${NAMESPACE}
    spec:
      podSelector:

  - kind: NetworkPolicy
    apiVersion: networking.k8s.io/v1
    metadata:
      name: allow-same-namespace
      namespace: ${NAMESPACE}
    spec:
      podSelector:
      ingress:
      - from:
        - podSelector: {}

  - apiVersion: networking.k8s.io/v1
    kind: NetworkPolicy
    metadata:
      name: bad-np
      namespace: ${NAMESPACE}
    spec:
      egress:
      - {}
      podSelector:
        matchLabels:
          never-gonna: match
      policyTypes:
      - Egress

parameters:
  - name: NAMESPACE
    value: "openshift-logging"
`)

func testdataGeneratelog42981YamlBytes() ([]byte, error) {
	return _testdataGeneratelog42981Yaml, nil
}

func testdataGeneratelog42981Yaml() (*asset, error) {
	bytes, err := testdataGeneratelog42981YamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "testdata/generatelog/42981.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testdataGeneratelogContainer_json_log_templateJson = []byte(`{
  "apiVersion": "template.openshift.io/v1",
  "kind": "Template",
  "metadata": {
    "name": "centos-logtest-template"
  },
  "objects": [
    {
      "apiVersion": "v1",
      "data": {
        "ocp_logtest.cfg": "--raw --file /var/lib/svt/json.example  --text-type input --rate ${RATE}",
        "json.example": "{\"message\": \"MERGE_JSON_LOG=true\", \"level\": \"debug\",\"Layer1\": \"layer1 0\", \"layer2\": {\"name\":\"Layer2 1\", \"tips\":\"Decide by PRESERVE_JSON_LOG\"}, \"StringNumber\":\"10\", \"Number\": 10,\"foo.bar\":\"Dot Item\",\"{foobar}\":\"Brace Item\",\"[foobar]\":\"Bracket Item\", \"foo:bar\":\"Colon Item\",\"foo bar\":\"Space Item\" }"
      },
      "kind": "ConfigMap",
      "metadata": {
        "name": "${CONFIGMAP}"
      }
    },
    {
      "apiVersion": "v1",
      "kind": "ReplicationController",
      "metadata": {
        "name": "${REPLICATIONCONTROLLER}",
        "labels": "${{LABELS}}"
      },
      "spec": {
        "replicas": "${{REPLICAS}}",
        "template": {
          "metadata": {
            "generateName": "centos-logtest-",
            "annotations": {
              "containerType.logging.openshift.io/${CONTAINER}": "${CONTAINER}"
            },
            "labels": "${{LABELS}}"
          },
          "spec": {
            "containers": [
              {
                "env": [],
                "image": "quay.io/openshifttest/ocp-logtest@sha256:6e2973d7d454ce412ad90e99ce584bf221866953da42858c4629873e53778606",
                "imagePullPolicy": "IfNotPresent",
                "name": "${CONTAINER}",
                "resources": {},
                "volumeMounts": [
                  {
                    "name": "config",
                    "mountPath": "/var/lib/svt"
                  }
                ],
                "securityContext": {
                  "allowPrivilegeEscalation": false,
                  "capabilities": {
                    "drop": [
                      "ALL"
                    ]
                  }
                },
                "terminationMessagePath": "/dev/termination-log"
              }
            ],
            "securityContext": {
              "runAsNonRoot": true,
              "seccompProfile": {
                "type": "RuntimeDefault"
              }
            },
            "volumes": [
              {
                "name": "config",
                "configMap": {
                  "name": "${CONFIGMAP}"
                }
              }
            ]
          }
        }
      }
    }
  ],
  "parameters": [
    {
      "name": "REPLICAS",
      "displayName": "Replicas",
      "value": "1"
    },
    {
      "name": "LABELS",
      "displayName": "labels",
      "value": "{\"run\": \"centos-logtest\", \"test\": \"centos-logtest\"}"
    },
    {
      "name": "REPLICATIONCONTROLLER",
      "displayName": "ReplicationController",
      "value": "logging-centos-logtest"
    },
    {
      "name": "CONFIGMAP",
      "displayName": "ConfigMap",
      "value": "logtest-config"
    },
    {
      "name": "CONTAINER",
      "value": "logging-centos-logtest"
    },
    {
      "name": "RATE",
      "value": "60.0"
    }
  ]
}
`)

func testdataGeneratelogContainer_json_log_templateJsonBytes() ([]byte, error) {
	return _testdataGeneratelogContainer_json_log_templateJson, nil
}

func testdataGeneratelogContainer_json_log_templateJson() (*asset, error) {
	bytes, err := testdataGeneratelogContainer_json_log_templateJsonBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "testdata/generatelog/container_json_log_template.json", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testdataGeneratelogContainer_json_log_template_unannotedJson = []byte(`{
  "apiVersion": "template.openshift.io/v1",
  "kind": "Template",
  "metadata": {
    "name": "centos-logtest-template"
  },
  "objects": [
    {
      "apiVersion": "v1",
      "data": {
        "ocp_logtest.cfg": "--raw --file /var/lib/svt/json.example  --text-type input --rate 60.0",
        "json.example": "{\"message\": \"MERGE_JSON_LOG=true\", \"level\": \"debug\",\"Layer1\": \"layer1 0\", \"layer2\": {\"name\":\"Layer2 1\", \"tips\":\"Decide by PRESERVE_JSON_LOG\"}, \"StringNumber\":\"10\", \"Number\": 10,\"foo.bar\":\"Dot Item\",\"{foobar}\":\"Brace Item\",\"[foobar]\":\"Bracket Item\", \"foo:bar\":\"Colon Item\",\"foo bar\":\"Space Item\" }"
      },
      "kind": "ConfigMap",
      "metadata": {
        "name": "${CONFIGMAP}"
      }
    },
    {
      "apiVersion": "v1",
      "kind": "ReplicationController",
      "metadata": {
        "name": "${{REPLICATIONCONTROLLER}}",
        "labels": "${{LABELS}}"
      },
      "spec": {
        "replicas": "${{REPLICAS}}",
        "template": {
          "metadata": {
            "generateName": "centos-logtest-",
            "labels": "${{LABELS}}"
          },
          "spec": {
            "containers": [
              {
                "env": [],
                "image": "quay.io/openshifttest/ocp-logtest@sha256:6e2973d7d454ce412ad90e99ce584bf221866953da42858c4629873e53778606",
                "imagePullPolicy": "IfNotPresent",
                "name": "${CONTAINER}",
                "resources": {},
                "volumeMounts": [
                  {
                    "name": "config",
                    "mountPath": "/var/lib/svt"
                  }
                ],
                "securityContext": {
                  "allowPrivilegeEscalation": false,
                  "capabilities": {
                    "drop": [
                      "ALL"
                    ]
                  }
                },
                "terminationMessagePath": "/dev/termination-log"
              }
            ],
            "securityContext": {
              "runAsNonRoot": true,
              "seccompProfile": {
                "type": "RuntimeDefault"
              }
            },
            "volumes": [
              {
                "name": "config",
                "configMap": {
                  "name": "${{CONFIGMAP}}"
                }
              }
            ]
          }
        }
      }
    }
  ],
  "parameters": [
    {
      "name": "REPLICAS",
      "displayName": "Replicas",
      "value": "1"
    },
    {
      "name": "LABELS",
      "displayName": "labels",
      "value": "{\"run\": \"centos-logtest\", \"test\": \"centos-logtest\"}"
    },
    {
      "name": "REPLICATIONCONTROLLER",
      "displayName": "ReplicationController",
      "value": "logging-centos-logtest"
    },
    {
      "name": "CONFIGMAP",
      "displayName": "ConfigMap",
      "value": "logtest-config"
    },
    {
      "name": "CONTAINER",
      "value": "logging-centos-logtest"
    }
  ]
}
`)

func testdataGeneratelogContainer_json_log_template_unannotedJsonBytes() ([]byte, error) {
	return _testdataGeneratelogContainer_json_log_template_unannotedJson, nil
}

func testdataGeneratelogContainer_json_log_template_unannotedJson() (*asset, error) {
	bytes, err := testdataGeneratelogContainer_json_log_template_unannotedJsonBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "testdata/generatelog/container_json_log_template_unannoted.json", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testdataGeneratelogContainer_non_json_log_templateJson = []byte(`{
  "apiVersion": "template.openshift.io/v1",
  "kind": "Template",
  "metadata": {
    "name": "centos-logtest-template"
  },
  "objects": [
    {
      "apiVersion": "v1",
      "data": {
        "ocp_logtest.cfg": "--raw --file /var/lib/svt/json.example  --text-type input --rate 60.0",
        "json.example": "ㄅㄉˇˋㄓˊ˙ㄚㄞㄢㄦㄆ 中国 883.317µs ā á ǎ à ō ó ▅ ▆ ▇ █ 々"
      },
      "kind": "ConfigMap",
      "metadata": {
        "name": "${{CONFIGMAP}}"
      }
    },
    {
      "apiVersion": "v1",
      "kind": "ReplicationController",
      "metadata": {
        "name": "${{REPLICATIONCONTROLLER}}",
        "labels": {
          "run": "${{LABELS}}",
          "test": "${{LABELS}}"
        }
      },
      "spec": {
        "replicas": "${{REPLICAS}}",
        "template": {
          "metadata": {
            "generateName": "centos-logtest-",
            "labels": {
              "run": "${{LABELS}}",
              "test": "${{LABELS}}"
            }
          },
          "spec": {
            "containers": [
              {
                "env": [],
                "image": "quay.io/openshifttest/ocp-logtest@sha256:6e2973d7d454ce412ad90e99ce584bf221866953da42858c4629873e53778606",
                "imagePullPolicy": "IfNotPresent",
                "name": "logging-centos-logtest",
                "resources": {},
                "volumeMounts": [
                  {
                    "name": "config",
                    "mountPath": "/var/lib/svt"
                  }
                ],
                "securityContext": {
                  "allowPrivilegeEscalation": false,
                  "capabilities": {
                    "drop": [
                      "ALL"
                    ]
                  }
                },
                "terminationMessagePath": "/dev/termination-log"
              }
            ],
            "securityContext": {
              "runAsNonRoot": true,
              "seccompProfile": {
                "type": "RuntimeDefault"
              }
            },
            "volumes": [
              {
                "name": "config",
                "configMap": {
                  "name": "${{CONFIGMAP}}"
                }
              }
            ]
          }
        }
      }
    }
  ],
  "parameters": [
    {
      "name": "REPLICAS",
      "displayName": "Replicas",
      "value": "1"
    },
    {
      "name": "LABELS",
      "displayName": "labels",
      "value": "centos-logtest"
    },
    {
      "name": "REPLICATIONCONTROLLER",
      "displayName": "ReplicationController",
      "value": "logging-centos-logtest"
    },
    {
      "name": "CONFIGMAP",
      "displayName": "ConfigMap",
      "value": "logtest-config"
    }
  ]
}
`)

func testdataGeneratelogContainer_non_json_log_templateJsonBytes() ([]byte, error) {
	return _testdataGeneratelogContainer_non_json_log_templateJson, nil
}

func testdataGeneratelogContainer_non_json_log_templateJson() (*asset, error) {
	bytes, err := testdataGeneratelogContainer_non_json_log_templateJsonBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "testdata/generatelog/container_non_json_log_template.json", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testdataGeneratelogLoggingPerformanceAppGeneratorJson = []byte(`{
    "apiVersion": "template.openshift.io/v1",
    "kind": "Template",
    "metadata": {
      "name": "centos-logtest-template"
    },
    "objects": [
      {
        "apiVersion": "v1",
        "data": {
          "ocp_logtest.cfg": "--raw --file /var/lib/svt/json.example  --text-type input --rate ${RATE} --num-lines ${NUM_LINES}",
          "json.example": "{\"message\": \"MERGE_JSON_LOG=true\", \"level\": \"debug\",\"Layer1\": \"layer1 0\", \"layer2\": {\"name\":\"Layer2 1\", \"tips\":\"Decide by PRESERVE_JSON_LOG\"}, \"StringNumber\":\"10\", \"Number\": 10,\"foo.bar\":\"Dot Item\",\"{foobar}\":\"Brace Item\",\"[foobar]\":\"Bracket Item\", \"foo:bar\":\"Colon Item\",\"foo bar\":\"Space Item\" }"
        },
        "kind": "ConfigMap",
        "metadata": {
          "name": "${CONFIGMAP}"
        }
      },
      {
        "apiVersion": "v1",
        "kind": "ReplicationController",
        "metadata": {
          "name": "${REPLICATIONCONTROLLER}",
          "labels": "${{LABELS}}"
        },
        "spec": {
          "replicas": "${{REPLICAS}}",
          "template": {
            "metadata": {
              "generateName": "centos-logtest-",
              "annotations": {
                "containerType.logging.openshift.io/${CONTAINER}": "${CONTAINER}"
              },
              "labels": "${{LABELS}}"
            },
            "spec": {
              "nodeSelector": "${{NODE_SELECTOR}}",
              "containers": [
                {
                  "env": [],
                  "image": "quay.io/openshifttest/ocp-logtest@sha256:6e2973d7d454ce412ad90e99ce584bf221866953da42858c4629873e53778606",
                  "imagePullPolicy": "IfNotPresent",
                  "name": "${CONTAINER}",
                  "resources": {},
                  "volumeMounts": [
                    {
                      "name": "config",
                      "mountPath": "/var/lib/svt"
                    }
                  ],
                  "securityContext": {
                    "allowPrivilegeEscalation": false,
                    "capabilities": {
                      "drop": [
                        "ALL"
                      ]
                    }
                  },
                  "terminationMessagePath": "/dev/termination-log"
                }
              ],
              "securityContext": {
                "runAsNonRoot": true,
                "seccompProfile": {
                  "type": "RuntimeDefault"
                }
              },
              "volumes": [
                {
                  "name": "config",
                  "configMap": {
                    "name": "${CONFIGMAP}"
                  }
                }
              ]
            }
          }
        }
      }
    ],
    "parameters": [
      {
        "name": "REPLICAS",
        "displayName": "Replicas",
        "value": "1"
      },
      {
        "name": "LABELS",
        "displayName": "labels",
        "value": "{\"run\": \"centos-logtest\", \"test\": \"centos-logtest\"}"
      },
      {
        "name": "REPLICATIONCONTROLLER",
        "displayName": "ReplicationController",
        "value": "logging-centos-logtest"
      },
      {
        "name": "CONFIGMAP",
        "displayName": "ConfigMap",
        "value": "logtest-config"
      },
      {
        "name": "CONTAINER",
        "value": "logging-centos-logtest"
      },
      {
        "name": "RATE",
        "value": "60.0"
      },
      {
        "name": "NUM_LINES",
        "value": "1000"
      },
      {
        "name": "NODE_SELECTOR",
        "displayName": "Node Selector",
        "description": "Node selector to schedule pods on specific nodes",
        "value": "{\"node-role.kubernetes.io/worker\": \"\"}"
      }
    ]
}
`)

func testdataGeneratelogLoggingPerformanceAppGeneratorJsonBytes() ([]byte, error) {
	return _testdataGeneratelogLoggingPerformanceAppGeneratorJson, nil
}

func testdataGeneratelogLoggingPerformanceAppGeneratorJson() (*asset, error) {
	bytes, err := testdataGeneratelogLoggingPerformanceAppGeneratorJsonBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "testdata/generatelog/logging-performance-app-generator.json", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testdataGeneratelogMulti_container_json_log_templateYaml = []byte(`apiVersion: template.openshift.io/v1
kind: Template
metadata:
  name: multi-container-json-log-template
objects:
- apiVersion: v1
  data:
    ocp_logtest.cfg: "--raw --file /var/lib/svt/json.example  --text-type input --rate ${RATE}"
    json.example: "{\"message\": \"MERGE_JSON_LOG=true\", \"level\": \"debug\",\"Layer1\": \"layer1 0\", \"layer2\": {\"name\":\"Layer2 1\", \"tips\":\"Decide by PRESERVE_JSON_LOG\"}, \"StringNumber\":\"10\", \"Number\": 10,\"foo.bar\":\"Dot Item\",\"{foobar}\":\"Brace Item\",\"[foobar]\":\"Bracket Item\", \"foo:bar\":\"Colon Item\",\"foo bar\":\"Space Item\"}"
  kind: ConfigMap
  metadata:
    name: ${CMNAME}
- apiVersion: v1
  kind: ReplicationController
  metadata:
    name: ${NAME}
    labels: ${{LABELS}}
  spec:
    replicas: ${{REPLICAS}}
    template:
      metadata:
        generateName: logging-logtest-
        annotations:
          containerType.logging.openshift.io/${CONTAINER}-0: ${CONTAINER}-0
          containerType.logging.openshift.io/${CONTAINER}-1: ${CONTAINER}-1
        labels: ${{LABELS}}
      spec:
        containers:
        - env: []
          image: quay.io/openshifttest/ocp-logtest@sha256:6e2973d7d454ce412ad90e99ce584bf221866953da42858c4629873e53778606
          imagePullPolicy: IfNotPresent
          name: ${CONTAINER}-0
          resources: {}
          volumeMounts:
          - name: config
            mountPath: /var/lib/svt
          terminationMessagePath: /dev/termination-log
        - env: []
          image: quay.io/openshifttest/ocp-logtest@sha256:6e2973d7d454ce412ad90e99ce584bf221866953da42858c4629873e53778606
          imagePullPolicy: IfNotPresent
          name: ${CONTAINER}-1
          resources: {}
          volumeMounts:
          - name: config
            mountPath: /var/lib/svt
          terminationMessagePath: /dev/termination-log
        - env: []
          image: quay.io/openshifttest/ocp-logtest@sha256:6e2973d7d454ce412ad90e99ce584bf221866953da42858c4629873e53778606
          imagePullPolicy: IfNotPresent
          name: ${CONTAINER}-2
          resources: {}
          volumeMounts:
          - name: config
            mountPath: /var/lib/svt
          terminationMessagePath: /dev/termination-log
        volumes:
        - name: config
          configMap:
            name: ${CMNAME}
parameters:
- name: REPLICAS
  value: "1"
- name: LABELS
  displayName: labels
  value: "{\"run\": \"logging-logtest\", \"test\": \"logging-logtest\"}"
- name: NAME
  value: logging-logtest
- name: CMNAME
  value: multi-containers-logtest-config
- name: CONTAINER
  value: centos-logtest-container
- name: RATE
  value: "30.0"
`)

func testdataGeneratelogMulti_container_json_log_templateYamlBytes() ([]byte, error) {
	return _testdataGeneratelogMulti_container_json_log_templateYaml, nil
}

func testdataGeneratelogMulti_container_json_log_templateYaml() (*asset, error) {
	bytes, err := testdataGeneratelogMulti_container_json_log_templateYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "testdata/generatelog/multi_container_json_log_template.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testdataGeneratelogMultilineErrorLogYaml = []byte(`apiVersion: template.openshift.io/v1
kind: Template
metadata:
  name: multiline-log-template
objects:
- apiVersion: apps/v1
  kind: Deployment
  metadata:
    name: ${NAME}
    labels:
      name: multiline-log
  spec:
    progressDeadlineSeconds: 600
    replicas: 1
    revisionHistoryLimit: 10
    selector:
      matchLabels:
        name: multiline-log
    strategy:
      rollingUpdate:
        maxSurge: 25%
        maxUnavailable: 25%
      type: RollingUpdate
    template:
      metadata:
        annotations:
          capabilities: Seamless Upgrades
          containerImage: quay.io/openshifttest/multiline-log@sha256:31cabe5ffb849e79e12d7105e0d8cba68b9218d302521c8b656bad78987b0502
          support: OpenShift Logging QE
        creationTimestamp: null
        labels:
          name: multiline-log
      spec:
        containers:
        - args:
          - /run-go.sh
          command:
          - /bin/sh
          image: quay.io/openshifttest/multiline-log@sha256:31cabe5ffb849e79e12d7105e0d8cba68b9218d302521c8b656bad78987b0502
          imagePullPolicy: IfNotPresent
          name: multiline-log
          resources: {}
          terminationMessagePath: /dev/termination-log
          terminationMessagePolicy: File
          volumeMounts:
          - mountPath: /var/lib/logging/multiline-log.cfg
            subPath: multiline-log.cfg
            name: config
        dnsPolicy: ClusterFirst
        restartPolicy: Always
        schedulerName: default-scheduler
        securityContext:
          seccompProfile:
            type: RuntimeDefault
        terminationGracePeriodSeconds: 30
        volumes:
        - configMap:
            defaultMode: 420
            name: ${NAME}
          name: config
- apiVersion: v1
  data:
    multiline-log.cfg: |
      --stream ${OUT_STREAM} --rate ${RATE} --log-type ${LOG_TYPE}
  kind: ConfigMap
  metadata:
    name: ${NAME}
parameters:
- name: NAME
  value: "multiline-log"
- name: LOG_TYPE
  value: "all"
- name: RATE
  value: "30.00"
- name: OUT_STREAM
  value: "stdout"
`)

func testdataGeneratelogMultilineErrorLogYamlBytes() ([]byte, error) {
	return _testdataGeneratelogMultilineErrorLogYaml, nil
}

func testdataGeneratelogMultilineErrorLogYaml() (*asset, error) {
	bytes, err := testdataGeneratelogMultilineErrorLogYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "testdata/generatelog/multiline-error-log.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testdataLogfilemetricexporterLfmeYaml = []byte(`kind: Template
apiVersion: template.openshift.io/v1
metadata:
  name: logfilesmetricexporter-template
objects:
- apiVersion: "logging.openshift.io/v1alpha1"
  kind: LogFileMetricExporter
  metadata:
    name: ${NAME}
    namespace: ${NAMESPACE}
  spec:
    resources:
      limits:
        cpu: ${LIMIT_CPU}
        memory: ${LIMIT_MEMORY}
      requests:
        cpu: ${REQUEST_CPU}
        memory: ${REQUEST_MEMORY}
    tolerations: ${{TOLERATIONS}}
    nodeSelector: ${{NODE_SELECTOR}}
parameters:
- name: NAME
  value: "instance"
- name: NAMESPACE
  value: "openshift-logging"
- name: TOLERATIONS
  value: "[]"
- name: NODE_SELECTOR
  value: "{}"
- name: LIMIT_CPU
  value: "500m"
- name: LIMIT_MEMORY
  value: "256Mi"
- name: REQUEST_CPU
  value: "200m"
- name: REQUEST_MEMORY
  value: "128Mi"
`)

func testdataLogfilemetricexporterLfmeYamlBytes() ([]byte, error) {
	return _testdataLogfilemetricexporterLfmeYaml, nil
}

func testdataLogfilemetricexporterLfmeYaml() (*asset, error) {
	bytes, err := testdataLogfilemetricexporterLfmeYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "testdata/logfilemetricexporter/lfme.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testdataLokiLogAlertsClusterMonitoringConfigYaml = []byte(`apiVersion: template.openshift.io/v1
kind: Template
metadata:
  name: cluster-monitoring-config-temp
objects:
- apiVersion: v1
  kind: ConfigMap
  metadata:
    name: cluster-monitoring-config
    namespace: openshift-monitoring
  data:
    config.yaml: |
      enableUserWorkload: true
`)

func testdataLokiLogAlertsClusterMonitoringConfigYamlBytes() ([]byte, error) {
	return _testdataLokiLogAlertsClusterMonitoringConfigYaml, nil
}

func testdataLokiLogAlertsClusterMonitoringConfigYaml() (*asset, error) {
	bytes, err := testdataLokiLogAlertsClusterMonitoringConfigYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "testdata/loki-log-alerts/cluster-monitoring-config.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testdataLokiLogAlertsLokiAppAlertingRuleTemplateYaml = []byte(`kind: Template
apiVersion: template.openshift.io/v1
metadata:
  name: loki-app-alerting-rule-template
objects:
- apiVersion: loki.grafana.com/v1
  kind: AlertingRule
  metadata:
    labels:
      openshift.io/cluster-monitoring: 'true'
    name: ${ALERTING_RULE_NAME}
    namespace: ${NAMESPACE}
  spec:
    groups:
      - interval: 1m
        name: MyAppLogVolumeAlert
        rules:
          - alert: MyAppLogVolumeIsHigh
            annotations:
              description: My application has high amount of logs.
              summary: Your application project has high amount of logs.
            expr: ${ALERT_LOGQL_EXPR}
            for: 1m
            labels:
              severity: info
              project: ${NAMESPACE}
    tenantID: application
parameters:
- name: NAMESPACE
  value: "my-app-1"
- name: ALERTING_RULE_NAME
  value: "my-app-workload-alert"
- name: ALERT_LOGQL_EXPR
  value: ""
`)

func testdataLokiLogAlertsLokiAppAlertingRuleTemplateYamlBytes() ([]byte, error) {
	return _testdataLokiLogAlertsLokiAppAlertingRuleTemplateYaml, nil
}

func testdataLokiLogAlertsLokiAppAlertingRuleTemplateYaml() (*asset, error) {
	bytes, err := testdataLokiLogAlertsLokiAppAlertingRuleTemplateYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "testdata/loki-log-alerts/loki-app-alerting-rule-template.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testdataLokiLogAlertsLokiAppRecordingRuleTemplateYaml = []byte(`kind: Template
apiVersion: template.openshift.io/v1
metadata:
  name: loki-app-recording-rule-template
objects:
- apiVersion: loki.grafana.com/v1
  kind: RecordingRule
  metadata:
    labels:
      openshift.io/cluster-monitoring: 'true'
    name: ${RECORDING_RULE_NAME}
    namespace: ${NAMESPACE}
  spec:
    groups:
      - interval: 1m
        name: HighAppLogsToLoki1m
        rules:
          - expr: >
              count_over_time({kubernetes_namespace_name="${NAMESPACE}"}[1m]) > 10
            record: 'loki:operator:applogs:rate1m'
    tenantID: application
parameters:
- name: NAMESPACE
  value: "my-app-1"
- name: RECORDING_RULE_NAME
  value: "my-app-workload-record"
`)

func testdataLokiLogAlertsLokiAppRecordingRuleTemplateYamlBytes() ([]byte, error) {
	return _testdataLokiLogAlertsLokiAppRecordingRuleTemplateYaml, nil
}

func testdataLokiLogAlertsLokiAppRecordingRuleTemplateYaml() (*asset, error) {
	bytes, err := testdataLokiLogAlertsLokiAppRecordingRuleTemplateYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "testdata/loki-log-alerts/loki-app-recording-rule-template.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testdataLokiLogAlertsLokiInfraAlertingRuleTemplateYaml = []byte(`kind: Template
apiVersion: template.openshift.io/v1
metadata:
  name: loki-infra-alerting-rule-template
objects:
- apiVersion: loki.grafana.com/v1
  kind: AlertingRule
  metadata:
    labels:
      openshift.io/cluster-monitoring: 'true'
    name: ${ALERTING_RULE_NAME}
    namespace: ${NAMESPACE}
  spec:
    groups:
      - interval: 1m
        name: LokiOperatorLogsHigh
        rules:
          - alert: LokiOperatorLogsAreHigh
            annotations:
              description: Your Loki Operator has High amount of logs
              summary: Loki Operator Log volume is High
            expr: >
              count_over_time({kubernetes_namespace_name="${NAMESPACE}"}[1m]) > 10
            for: 1m
            labels:
              severity: info
    tenantID: infrastructure
parameters:
- name: NAMESPACE
  value: "openshift-operators-redhat"
- name: ALERTING_RULE_NAME
  value: "my-infra-workload-alert"
`)

func testdataLokiLogAlertsLokiInfraAlertingRuleTemplateYamlBytes() ([]byte, error) {
	return _testdataLokiLogAlertsLokiInfraAlertingRuleTemplateYaml, nil
}

func testdataLokiLogAlertsLokiInfraAlertingRuleTemplateYaml() (*asset, error) {
	bytes, err := testdataLokiLogAlertsLokiInfraAlertingRuleTemplateYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "testdata/loki-log-alerts/loki-infra-alerting-rule-template.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testdataLokiLogAlertsLokiInfraRecordingRuleTemplateYaml = []byte(`kind: Template
apiVersion: template.openshift.io/v1
metadata:
  name: loki-infra-recording-rule-template
objects:
- apiVersion: loki.grafana.com/v1
  kind: RecordingRule
  metadata:
    labels:
      openshift.io/cluster-monitoring: 'true'
    name: ${RECORDING_RULE_NAME}
    namespace: ${NAMESPACE}
  spec:
    groups:
      - interval: 1m
        name: LokiOperatorLogsAreHigh1m
        rules:
          - expr: >
              count_over_time({kubernetes_namespace_name="${NAMESPACE}"}[1m]) > 10
            record: 'loki:operator:infralogs:rate1m'
    tenantID: infrastructure
parameters:
- name: NAMESPACE
  value: "openshift-operators-redhat"
- name: RECORDING_RULE_NAME
  value: "my-infra-workload-record"
`)

func testdataLokiLogAlertsLokiInfraRecordingRuleTemplateYamlBytes() ([]byte, error) {
	return _testdataLokiLogAlertsLokiInfraRecordingRuleTemplateYaml, nil
}

func testdataLokiLogAlertsLokiInfraRecordingRuleTemplateYaml() (*asset, error) {
	bytes, err := testdataLokiLogAlertsLokiInfraRecordingRuleTemplateYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "testdata/loki-log-alerts/loki-infra-recording-rule-template.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testdataLokiLogAlertsUserWorkloadMonitoringConfigYaml = []byte(`apiVersion: template.openshift.io/v1
kind: Template
metadata:
  name: user-workload-monitoring-config-temp
objects:
- apiVersion: v1
  kind: ConfigMap
  metadata:
    name: user-workload-monitoring-config
    namespace: openshift-user-workload-monitoring
  data:
    config.yaml: |
      alertmanager:
        enabled: true
        enableAlertmanagerConfig: true
`)

func testdataLokiLogAlertsUserWorkloadMonitoringConfigYamlBytes() ([]byte, error) {
	return _testdataLokiLogAlertsUserWorkloadMonitoringConfigYaml, nil
}

func testdataLokiLogAlertsUserWorkloadMonitoringConfigYaml() (*asset, error) {
	bytes, err := testdataLokiLogAlertsUserWorkloadMonitoringConfigYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "testdata/loki-log-alerts/user-workload-monitoring-config.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testdataLokistackFineGrainedAccessRolesYaml = []byte(`apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: cluster-logging-application-view
rules:
- apiGroups:
  - loki.grafana.com
  resourceNames:
  - logs
  resources:
  - application
  verbs:
  - get
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: cluster-logging-infrastructure-view
rules:
- apiGroups:
  - loki.grafana.com
  resourceNames:
  - logs
  resources:
  - infrastructure
  verbs:
  - get
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: cluster-logging-audit-view
rules:
- apiGroups:
  - loki.grafana.com
  resourceNames:
  - logs
  resources:
  - audit
  verbs:
  - get
`)

func testdataLokistackFineGrainedAccessRolesYamlBytes() ([]byte, error) {
	return _testdataLokistackFineGrainedAccessRolesYaml, nil
}

func testdataLokistackFineGrainedAccessRolesYaml() (*asset, error) {
	bytes, err := testdataLokistackFineGrainedAccessRolesYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "testdata/lokistack/fine-grained-access-roles.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testdataLokistackLokistackSimpleIpv6TlsYaml = []byte(`kind: Template
apiVersion: template.openshift.io/v1
metadata:
  name: lokiStack-template
objects:
- kind: "LokiStack"
  apiVersion: "loki.grafana.com/v1"
  metadata:
    name: ${NAME}
    namespace: ${NAMESPACE}
  spec:
    networkPolicies:
      ruleSet: ${LOKISTACK_NETWORK_POLICIES_RULESET}
    limits:
      global:
        retention:
          days: 20
          streams:
          - days: 4
            priority: 1
            selector: '{kubernetes_namespace_name=~"e2e.+"}'
          - days: 1
            priority: 1
            selector: '{kubernetes_namespace_name="kube.+"}'
          - days: 15
            priority: 1
            selector: '{log_type="audit"}'
      tenants:
        application:
          retention:
            days: 1
            streams:
            - days: 4
              selector: '{kubernetes_namespace_name=~"test.+"}'
        audit:
          retention:
            days: 15
        infrastructure:
          retention:
            days: 5
            streams:
            - days: 1
              selector: '{kubernetes_namespace_name=~"openshift-cluster.+"}'
    hashRing:
      memberlist:
        enableIPv6: true
      type: memberlist
    managementState: "Managed"
    size: ${SIZE}
    storage:
      secret:
        name: ${SECRET_NAME}
        type: ${STORAGE_TYPE}
      schemas:
      - version: ${STORAGE_SCHEMA_VERSION}
        effectiveDate: ${SCHEMA_EFFECTIVE_DATE}
      tls:
        caName: ${CA_NAME}
        caKey: ${CA_KEY_NAME}
    storageClassName: ${STORAGE_CLASS}
    tenants:
      mode: "openshift-logging"
      openshift:
        adminGroups: ${{ADMIN_GROUPS}}
    rules:
      enabled: true
      selector:
        matchLabels:
          openshift.io/cluster-monitoring: "true"
      namespaceSelector:
        matchLabels:
          openshift.io/cluster-monitoring: "true"
parameters:
- name: NAME
  value: "my-loki"
- name: NAMESPACE
  value: "openshift-logging"
- name: SIZE
  value: "1x.demo"
- name: SECRET_NAME
  value: "s3-secret"
- name: STORAGE_TYPE
  value: "s3"
- name: STORAGE_CLASS
  value: "gp2"
- name: "ADMIN_GROUPS"
  value: "[]"
- name: STORAGE_SCHEMA_VERSION
  value: "v13"
- name: SCHEMA_EFFECTIVE_DATE
  value: "2023-10-15"
- name: CA_NAME
  value: ""
- name: CA_KEY_NAME
  value: "service-ca.crt"
- name: LOKISTACK_NETWORK_POLICIES_RULESET
  value: "None"
`)

func testdataLokistackLokistackSimpleIpv6TlsYamlBytes() ([]byte, error) {
	return _testdataLokistackLokistackSimpleIpv6TlsYaml, nil
}

func testdataLokistackLokistackSimpleIpv6TlsYaml() (*asset, error) {
	bytes, err := testdataLokistackLokistackSimpleIpv6TlsYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "testdata/lokistack/lokistack-simple-ipv6-tls.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testdataLokistackLokistackSimpleIpv6Yaml = []byte(`kind: Template
apiVersion: template.openshift.io/v1
metadata:
  name: lokiStack-template
objects:
- kind: "LokiStack"
  apiVersion: "loki.grafana.com/v1"
  metadata:
    name: ${NAME}
    namespace: ${NAMESPACE}
  spec:
    networkPolicies:
      ruleSet: ${LOKISTACK_NETWORK_POLICIES_RULESET}
    limits:
      global:
        retention:
          days: 20
          streams:
          - days: 4
            priority: 1
            selector: '{kubernetes_namespace_name=~"e2e.+"}'
          - days: 1
            priority: 1
            selector: '{kubernetes_namespace_name="kube.+"}'
          - days: 15
            priority: 1
            selector: '{log_type="audit"}'
      tenants:
        application:
          retention:
            days: 1
            streams:
            - days: 4
              selector: '{kubernetes_namespace_name=~"test.+"}'
        audit:
          retention:
            days: 15
        infrastructure:
          retention:
            days: 5
            streams:
            - days: 1
              selector: '{kubernetes_namespace_name=~"openshift-cluster.+"}'
    hashRing:
      memberlist:
        enableIPv6: true
      type: memberlist
    managementState: "Managed"
    size: ${SIZE}
    storage:
      secret:
        name: ${SECRET_NAME}
        type: ${STORAGE_TYPE}
      schemas:
      - version: ${STORAGE_SCHEMA_VERSION}
        effectiveDate: ${SCHEMA_EFFECTIVE_DATE}
    storageClassName: ${STORAGE_CLASS}
    tenants:
      mode: "openshift-logging"
      openshift:
        adminGroups: ${{ADMIN_GROUPS}}
    rules:
      enabled: true
      selector:
        matchLabels:
          openshift.io/cluster-monitoring: "true"
      namespaceSelector:
        matchLabels:
          openshift.io/cluster-monitoring: "true"
parameters:
- name: NAME
  value: "my-loki"
- name: NAMESPACE
  value: "openshift-logging"
- name: SIZE
  value: "1x.demo"
- name: SECRET_NAME
  value: "s3-secret"
- name: STORAGE_TYPE
  value: "s3"
- name: STORAGE_CLASS
  value: "gp2"
- name: "ADMIN_GROUPS"
  value: "[]"
- name: STORAGE_SCHEMA_VERSION
  value: "v13"
- name: SCHEMA_EFFECTIVE_DATE
  value: "2023-10-15"
- name: LOKISTACK_NETWORK_POLICIES_RULESET
  value: "None"
`)

func testdataLokistackLokistackSimpleIpv6YamlBytes() ([]byte, error) {
	return _testdataLokistackLokistackSimpleIpv6Yaml, nil
}

func testdataLokistackLokistackSimpleIpv6Yaml() (*asset, error) {
	bytes, err := testdataLokistackLokistackSimpleIpv6YamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "testdata/lokistack/lokistack-simple-ipv6.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testdataLokistackLokistackSimpleTlsYaml = []byte(`kind: Template
apiVersion: template.openshift.io/v1
metadata:
  name: lokiStack-template
objects:
- kind: "LokiStack"
  apiVersion: "loki.grafana.com/v1"
  metadata:
    name: ${NAME}
    namespace: ${NAMESPACE}
  spec:
    networkPolicies:
      ruleSet: ${LOKISTACK_NETWORK_POLICIES_RULESET}
    limits:
      global:
        retention:
          days: 20
          streams:
          - days: 4
            priority: 1
            selector: '{kubernetes_namespace_name=~"e2e.+"}'
          - days: 1
            priority: 1
            selector: '{kubernetes_namespace_name="kube.+"}'
          - days: 15
            priority: 1
            selector: '{log_type="audit"}'
      tenants:
        application:
          retention:
            days: 1
            streams:
            - days: 4
              selector: '{kubernetes_namespace_name=~"test.+"}'
        audit:
          retention:
            days: 15
        infrastructure:
          retention:
            days: 5
            streams:
            - days: 1
              selector: '{kubernetes_namespace_name=~"openshift-cluster.+"}'
    managementState: "Managed"
    size: ${SIZE}
    storage:
      secret:
        name: ${SECRET_NAME}
        type: ${STORAGE_TYPE}
      schemas:
      - version: ${STORAGE_SCHEMA_VERSION}
        effectiveDate: ${SCHEMA_EFFECTIVE_DATE}
      tls:
        caName: ${CA_NAME}
        caKey: ${CA_KEY_NAME}
    storageClassName: ${STORAGE_CLASS}
    tenants:
      mode: "openshift-logging"
      openshift:
        adminGroups: ${{ADMIN_GROUPS}}
    rules:
      enabled: true
      selector:
        matchLabels:
          openshift.io/cluster-monitoring: "true"
      namespaceSelector:
        matchLabels:
          openshift.io/cluster-monitoring: "true"
parameters:
- name: NAME
  value: "my-loki"
- name: NAMESPACE
  value: "openshift-logging"
- name: SIZE
  value: "1x.demo"
- name: SECRET_NAME
  value: "s3-secret"
- name: STORAGE_TYPE
  value: "s3"
- name: STORAGE_CLASS
  value: "gp2"
- name: "ADMIN_GROUPS"
  value: "[]"
- name: STORAGE_SCHEMA_VERSION
  value: "v13"
- name: SCHEMA_EFFECTIVE_DATE
  value: "2023-10-15"
- name: CA_NAME
  value: ""
- name: CA_KEY_NAME
  value: "service-ca.crt"
- name: LOKISTACK_NETWORK_POLICIES_RULESET
  value: "None"
`)

func testdataLokistackLokistackSimpleTlsYamlBytes() ([]byte, error) {
	return _testdataLokistackLokistackSimpleTlsYaml, nil
}

func testdataLokistackLokistackSimpleTlsYaml() (*asset, error) {
	bytes, err := testdataLokistackLokistackSimpleTlsYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "testdata/lokistack/lokistack-simple-tls.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testdataLokistackLokistackSimpleYaml = []byte(`kind: Template
apiVersion: template.openshift.io/v1
metadata:
  name: lokiStack-template
objects:
- kind: "LokiStack"
  apiVersion: "loki.grafana.com/v1"
  metadata:
    name: ${NAME}
    namespace: ${NAMESPACE}
  spec:
    networkPolicies:
      ruleSet: ${LOKISTACK_NETWORK_POLICIES_RULESET}
    limits:
      global:
        retention:
          days: 20
          streams:
          - days: 4
            priority: 1
            selector: '{kubernetes_namespace_name=~"e2e.+"}'
          - days: 1
            priority: 1
            selector: '{kubernetes_namespace_name="kube.+"}'
          - days: 15
            priority: 1
            selector: '{log_type="audit"}'
      tenants:
        application:
          retention:
            days: 1
            streams:
            - days: 4
              selector: '{kubernetes_namespace_name=~"test.+"}'
        audit:
          retention:
            days: 15
        infrastructure:
          retention:
            days: 5
            streams:
            - days: 1
              selector: '{kubernetes_namespace_name=~"openshift-cluster.+"}'
    managementState: "Managed"
    size: ${SIZE}
    storage:
      secret:
        name: ${SECRET_NAME}
        type: ${STORAGE_TYPE}
      schemas:
      - version: ${STORAGE_SCHEMA_VERSION}
        effectiveDate: ${SCHEMA_EFFECTIVE_DATE}
    storageClassName: ${STORAGE_CLASS}
    tenants:
      mode: "openshift-logging"
      openshift:
        adminGroups: ${{ADMIN_GROUPS}}
    rules:
      enabled: true
      selector:
        matchLabels:
          openshift.io/cluster-monitoring: "true"
      namespaceSelector:
        matchLabels:
          openshift.io/cluster-monitoring: "true"
parameters:
- name: NAME
  value: "my-loki"
- name: NAMESPACE
  value: "openshift-logging"
- name: SIZE
  value: "1x.demo"
- name: SECRET_NAME
  value: "s3-secret"
- name: STORAGE_TYPE
  value: "s3"
- name: STORAGE_CLASS
  value: "gp2"
- name: "ADMIN_GROUPS"
  value: "[]"
- name: STORAGE_SCHEMA_VERSION
  value: "v13"
- name: SCHEMA_EFFECTIVE_DATE
  value: "2023-10-15"
- name: LOKISTACK_NETWORK_POLICIES_RULESET
  value: "None"
`)

func testdataLokistackLokistackSimpleYamlBytes() ([]byte, error) {
	return _testdataLokistackLokistackSimpleYaml, nil
}

func testdataLokistackLokistackSimpleYaml() (*asset, error) {
	bytes, err := testdataLokistackLokistackSimpleYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "testdata/lokistack/lokistack-simple.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testdataMinioDeployYaml = []byte(`kind: Template
apiVersion: template.openshift.io/v1
metadata:
  name: minio-template
  annotations:
    description: "A MinIO service"
objects:
- kind: PersistentVolumeClaim
  apiVersion: v1
  metadata:
    name: minio-pv-claim
    namespace: ${NAMESPACE}
  spec:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 10Gi
- kind: Deployment
  apiVersion: apps/v1
  metadata:
    name: ${NAME}
    namespace: ${NAMESPACE}
  spec:
    selector:
      matchLabels:
        app: ${NAME}
    strategy:
      type: Recreate
    template:
      metadata:
        labels:
          app: ${NAME}
      spec:
        volumes:
        - name: data
          persistentVolumeClaim:
            claimName: minio-pv-claim
        containers:
        - name: minio
          volumeMounts:
          - name: data
            mountPath: "/data"
          image: ${IMAGE}
          args:
          - server
          - /data
          - --console-address
          - ":9001"
          env:
          - name: MINIO_ROOT_USER
            valueFrom:
              secretKeyRef:
                name: ${SECRET_NAME}
                key: access_key_id
          - name: MINIO_ROOT_PASSWORD
            valueFrom:
              secretKeyRef:
                name: ${SECRET_NAME}
                key: secret_access_key
          - name: MINIO_DOMAIN
            value: ${MINIO_DOMAIN}
          ports:
          - containerPort: 9000
          readinessProbe:
            httpGet:
              path: /minio/health/ready
              port: 9000
            initialDelaySeconds: 120
            periodSeconds: 20
          livenessProbe:
            httpGet:
              path: /minio/health/live
              port: 9000
            initialDelaySeconds: 120
            periodSeconds: 20
- kind: Service
  apiVersion: v1
  metadata:
    name: ${NAME}
    namespace: ${NAMESPACE}
  spec:
    ports:
      - port: 9000
        targetPort: 9000
        protocol: TCP
    selector:
      app: ${NAME}
- kind: Service
  apiVersion: v1
  metadata:
    name: minio-service-console
    namespace: ${NAMESPACE}
  spec:
    ports:
      - port: 9001
        targetPort: 9001
        protocol: TCP
    selector:
      app: ${NAME}
- kind: Route
  apiVersion: route.openshift.io/v1
  metadata:
    labels:
      app: ${NAME}
    name: ${NAME}
    namespace: ${NAMESPACE}
  spec:
    host: ${MINIO_DOMAIN}
    port:
      targetPort: 9000
    to:
      kind: Service
      name: ${NAME}
- kind: Route
  apiVersion: route.openshift.io/v1
  metadata:
    labels:
      app: ${NAME}
    name: minio-console
    namespace: ${NAMESPACE}
  spec:
    port:
      targetPort: 9001
    to:
      kind: Service
      name: minio-service-console
parameters:
  - name: IMAGE
    displayName: " The MinIO image"
    value: "quay.io/openshifttest/minio:latest"
  - name: NAMESPACE
    displayName: Namespace
    value: "minio-aosqe"
  - name: NAME
    value: "minio"
  - name: SECRET_NAME
    value: "minio-creds"
  - name: MINIO_DOMAIN
    value: ""
`)

func testdataMinioDeployYamlBytes() ([]byte, error) {
	return _testdataMinioDeployYaml, nil
}

func testdataMinioDeployYaml() (*asset, error) {
	bytes, err := testdataMinioDeployYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "testdata/minIO/deploy.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testdataObservabilityOpenshiftIo_clusterlogforwarder48593Yaml = []byte(`kind: Template
apiVersion: template.openshift.io/v1
metadata:
  name: clusterlogforwarder-template
objects:
- apiVersion: "observability.openshift.io/v1"
  kind: ClusterLogForwarder
  metadata:
    name: ${NAME}
    namespace: ${NAMESPACE}
  spec:
    filters:
    - name: app-logs
      type: openshiftLabels
      openshiftLabels:
        logging: app-logs
    - name: infra-logs
      type: openshiftLabels
      openshiftLabels:
        logging: infra-logs
    - name: audit-logs
      type: openshiftLabels
      openshiftLabels:
        logging: audit-logs
    outputs:
    - name: es-created-by-user
      type: elasticsearch
      elasticsearch:
        url: ${ES_URL}
        version: ${{ES_VERSION}}
        index: ${INDEX}
    pipelines:
    - name: forward-app-logs
      inputRefs:
      - application
      filterRefs:
      - app-logs
      outputRefs:
      - es-created-by-user
    - name: forward-infra-logs
      inputRefs:
      - infrastructure
      filterRefs:
      - infra-logs
      outputRefs:
      - es-created-by-user
    - name: forward-audit-logs
      inputRefs:
      - audit
      filterRefs:
      - audit-logs
      outputRefs:
      - es-created-by-user
    serviceAccount:
      name: ${SERVICE_ACCOUNT_NAME}
parameters:
- name: NAME
  value: "instance"
- name: NAMESPACE
  value: "openshift-logging"
- name: ES_URL
  value: "http://elasticsearch-server.es-aosqe-qa.svc:9200"
- name: SERVICE_ACCOUNT_NAME
  value: ""
- name: ES_VERSION
  value: "6"
- name: INDEX
  value: "{.log_type||\"none-typed-logs\"}"
`)

func testdataObservabilityOpenshiftIo_clusterlogforwarder48593YamlBytes() ([]byte, error) {
	return _testdataObservabilityOpenshiftIo_clusterlogforwarder48593Yaml, nil
}

func testdataObservabilityOpenshiftIo_clusterlogforwarder48593Yaml() (*asset, error) {
	bytes, err := testdataObservabilityOpenshiftIo_clusterlogforwarder48593YamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "testdata/observability.openshift.io_clusterlogforwarder/48593.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testdataObservabilityOpenshiftIo_clusterlogforwarder67421Yaml = []byte(`kind: Template
apiVersion: template.openshift.io/v1
metadata:
  name: clusterlogforwarder-template
objects:
- apiVersion: "observability.openshift.io/v1"
  kind: ClusterLogForwarder
  metadata:
    name: ${NAME}
    namespace: ${NAMESPACE}
  spec:
    filters:
    - kubeAPIAudit:
        omitStages:
        - RequestReceived
        rules:
        - level: RequestResponse
          resources:
          - group: ""
            resources:
            - pods
      name: my-policy-0
      type: kubeAPIAudit
    - kubeAPIAudit:
        rules:
        - level: Request
          resources:
          - group: ""
            resources:
            - pods/status
            - pods/binding
      name: my-policy-1
      type: kubeAPIAudit
    managementState: ${MANAGEMENT_STATE}
    outputs:
    - name: lokistack
      type: lokiStack
      lokiStack:
        authentication:
          token:
            from: serviceAccount
        target:
          name: ${LOKISTACK_NAME}
          namespace: ${LOKISTACK_NAMESPACE}
      tls:
        ca:
          key: ca-bundle.crt
          secretName: ${SECRET_NAME}
    - name: es-created-by-user
      type: elasticsearch
      elasticsearch:
        url: ${ES_URL}
        index: ${INDEX}
        version: ${{ES_VERSION}}
    pipelines:
    - name: forward-to-external-es
      inputRefs:
      - audit
      outputRefs:
      - es-created-by-user
      filterRefs:
      - my-policy-0
    - name: forward-to-lokistack
      inputRefs:
      - audit
      outputRefs:
      - lokistack
      filterRefs:
      - my-policy-1
    serviceAccount:
      name: ${SERVICE_ACCOUNT_NAME}
parameters:
- name: NAME
  value: "instance"
- name: NAMESPACE
  value: "openshift-logging"
- name: ES_URL
  value: "http://elasticsearch-server.es-aosqe-qa.svc:9200"
- name: ES_VERSION
  value: "6"
- name: SERVICE_ACCOUNT_NAME
  value: ""
- name: LOKISTACK_NAME
  value: "logging-loki"
- name: LOKISTACK_NAMESPACE
  value: "openshift-logging"
- name: MANAGEMENT_STATE
  value: "Managed"
- name: INDEX
  value: "{.log_type||\"none-typed-logs\"}"
- name: SECRET_NAME
  value: ""
`)

func testdataObservabilityOpenshiftIo_clusterlogforwarder67421YamlBytes() ([]byte, error) {
	return _testdataObservabilityOpenshiftIo_clusterlogforwarder67421Yaml, nil
}

func testdataObservabilityOpenshiftIo_clusterlogforwarder67421Yaml() (*asset, error) {
	bytes, err := testdataObservabilityOpenshiftIo_clusterlogforwarder67421YamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "testdata/observability.openshift.io_clusterlogforwarder/67421.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testdataObservabilityOpenshiftIo_clusterlogforwarder68318Yaml = []byte(`kind: Template
apiVersion: template.openshift.io/v1
metadata:
  name: clusterlogforwarder-template
objects:
- apiVersion: "observability.openshift.io/v1"
  kind: ClusterLogForwarder
  metadata:
    name: ${NAME}
    namespace: ${NAMESPACE}
  spec:
    filters:
    - kubeAPIAudit:
        omitStages:
        - RequestReceived
        rules:
        - level: RequestResponse
          resources:
          - group: ""
            resources:
            - pods
      name: my-policy-0
      type: kubeAPIAudit
    - kubeAPIAudit:
        omitResponseCodes: []
        rules:
        - level: Request
          resources:
          - group: ""
            resources:
            - pods
            - secrets
      name: my-policy-1
      type: kubeAPIAudit
    managementState: ${MANAGEMENT_STATE}
    outputs:
    - name: lokistack
      type: lokiStack
      lokiStack:
        authentication:
          token:
            from: serviceAccount
        target:
          name: ${LOKISTACK_NAME}
          namespace: ${LOKISTACK_NAMESPACE}
      tls:
        ca:
          key: ca-bundle.crt
          secretName: ${SECRET_NAME}
    pipelines:
    - name: forward-to-default
      inputRefs:
      - audit
      outputRefs:
      - lokistack
      filterRefs:
      - my-policy-0
      - my-policy-1
    serviceAccount:
      name: ${SERVICE_ACCOUNT_NAME}
parameters:
- name: NAME
  value: "instance"
- name: NAMESPACE
  value: "openshift-logging"
- name: SERVICE_ACCOUNT_NAME
  value: ""
- name: LOKISTACK_NAME
  value: "logging-loki"
- name: LOKISTACK_NAMESPACE
  value: "openshift-logging"
- name: MANAGEMENT_STATE
  value: "Managed"
- name: SECRET_NAME
  value: ""
`)

func testdataObservabilityOpenshiftIo_clusterlogforwarder68318YamlBytes() ([]byte, error) {
	return _testdataObservabilityOpenshiftIo_clusterlogforwarder68318Yaml, nil
}

func testdataObservabilityOpenshiftIo_clusterlogforwarder68318Yaml() (*asset, error) {
	bytes, err := testdataObservabilityOpenshiftIo_clusterlogforwarder68318YamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "testdata/observability.openshift.io_clusterlogforwarder/68318.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testdataObservabilityOpenshiftIo_clusterlogforwarder71049Yaml = []byte(`kind: Template
apiVersion: template.openshift.io/v1
metadata:
  name: clusterlogforwarder-template
objects:
- apiVersion: observability.openshift.io/v1
  kind: ClusterLogForwarder
  metadata:
    name: ${NAME}
    namespace: ${NAMESPACE}
  spec:
    collector: ${{COLLECTOR}}
    managementState: ${MANAGEMENT_STATE}
    inputs:
    - name: syslog
      receiver:
        type: syslog
        port: 6514
      type: receiver
    outputs:
    - name: lokistack
      type: lokiStack
      lokiStack:
        authentication:
          token:
            from: serviceAccount
        target:
          name: ${LOKISTACK_NAME}
          namespace: ${LOKISTACK_NAMESPACE}
      tls:
        ca:
          key: ca-bundle.crt
          secretName: ${SECRET_NAME}
    pipelines:
    - name: forward-to-lokistack
      inputRefs:
      - syslog
      outputRefs:
      - lokistack
    serviceAccount:
      name: ${SERVICE_ACCOUNT_NAME}
parameters:
- name: NAME
  value: "instance"
- name: NAMESPACE
  value: "openshift-logging"
- name: SERVICE_ACCOUNT_NAME
  value: ""
- name: LOKISTACK_NAME
  value: "logging-loki"
- name: LOKISTACK_NAMESPACE
  value: "openshift-logging"
- name: MANAGEMENT_STATE
  value: "Managed"
- name: SECRET_NAME
  value: ""
- name: COLLECTOR
  value: "{}"
`)

func testdataObservabilityOpenshiftIo_clusterlogforwarder71049YamlBytes() ([]byte, error) {
	return _testdataObservabilityOpenshiftIo_clusterlogforwarder71049Yaml, nil
}

func testdataObservabilityOpenshiftIo_clusterlogforwarder71049Yaml() (*asset, error) {
	bytes, err := testdataObservabilityOpenshiftIo_clusterlogforwarder71049YamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "testdata/observability.openshift.io_clusterlogforwarder/71049.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testdataObservabilityOpenshiftIo_clusterlogforwarder71749Yaml = []byte(`kind: Template
apiVersion: template.openshift.io/v1
metadata:
  name: clusterlogforwarder-template
objects:
- apiVersion: "observability.openshift.io/v1"
  kind: ClusterLogForwarder
  metadata:
    name: ${NAME}
    namespace: ${NAMESPACE}
  spec:
    filters:
    - name: drop-logs-1
      type: drop
      drop:
      - test:
        - field: .log_type
          matches: "application"
        - field: .kubernetes.pod_name
          notMatches: "logging-centos-logtest.+"
      - test:
        - field: .message
          matches: (?i)\berror\b
        - field: .level
          matches: error
      - test:
        - field: .kubernetes.labels."test.logging.io/logging.qe-test-label"
          matches: .+
    - name: drop-logs-2
      type: drop
      drop:
      - test:
        - field: .kubernetes.namespace_name
          matches: "openshift*"
    - name: drop-logs-3
      type: drop
      drop:
      - test:
        - field: .log_type
          matches: "infrastructure"
        - field: .log_source
          matches: "container"
        - field: .kubernetes.namespace_name
          notMatches: "openshift-cluster*"
    managementState: ${MANAGEMENT_STATE}
    outputs:
    - name: lokistack
      type: lokiStack
      lokiStack:
        authentication:
          token:
            from: serviceAccount
        target:
          name: ${LOKISTACK_NAME}
          namespace: ${LOKISTACK_NAMESPACE}
      tls:
        ca:
          key: ca-bundle.crt
          secretName: ${SECRET_NAME}
    pipelines:
    - name: forward-to-lokistack
      inputRefs: ${{INPUT_REFS}}
      outputRefs:
      - lokistack
      filterRefs:
      - drop-logs-3
      - drop-logs-2
      - drop-logs-1
    serviceAccount:
      name: ${SERVICE_ACCOUNT_NAME}
parameters:
- name: NAME
  value: "instance"
- name: NAMESPACE
  value: "openshift-logging"
- name: INPUT_REFS
  value: "[\"infrastructure\", \"audit\", \"application\"]"
- name: SERVICE_ACCOUNT_NAME
  value: ""
- name: LOKISTACK_NAME
  value: "logging-loki"
- name: LOKISTACK_NAMESPACE
  value: "openshift-logging"
- name: MANAGEMENT_STATE
  value: "Managed"
- name: SECRET_NAME
  value: ""
`)

func testdataObservabilityOpenshiftIo_clusterlogforwarder71749YamlBytes() ([]byte, error) {
	return _testdataObservabilityOpenshiftIo_clusterlogforwarder71749Yaml, nil
}

func testdataObservabilityOpenshiftIo_clusterlogforwarder71749Yaml() (*asset, error) {
	bytes, err := testdataObservabilityOpenshiftIo_clusterlogforwarder71749YamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "testdata/observability.openshift.io_clusterlogforwarder/71749.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testdataObservabilityOpenshiftIo_clusterlogforwarderAffinity81397Yaml = []byte(`kind: Template
apiversion: template.openshift.io/v1
metadata:
  name: clusterlogforwarder-template
objects:
- apiVersion: observability.openshift.io/v1
  kind: ClusterLogForwarder
  metadata:
    name: ${NAME}
    namespace: ${NAMESPACE}
    annotations:
      logging.openshift.io/dev-preview-enable-collector-as-deployment: "${DEPLOYMENT}"
  spec:
    collector:
      affinity:
        nodeAffinity:
          preferredDuringSchedulingIgnoredDuringExecution:
            - preference:
                matchExpressions:
                  - key: label-1
                    operator: Exists
              weight: 1
            - preference:
                matchExpressions:
                  - key: label-2
                    operator: In
                    values:
                      - key-2
              weight: 50
          requiredDuringSchedulingIgnoredDuringExecution:
            nodeSelectorTerms:
              - matchExpressions:
                  - key: kubernetes.io/os
                    operator: In
                    values:
                      - linux
              - matchExpressions:
                  - key: node-role.kubernetes.io/worker
                    operator: Exists
        podAffinity:
          preferredDuringSchedulingIgnoredDuringExecution:
            - podAffinityTerm:
                labelSelector:
                  matchExpressions:
                    - key: qe-test
                      operator: In
                      values:
                        - value1
                topologyKey: kubernetes.io/hostname
              weight: 50
          requiredDuringSchedulingIgnoredDuringExecution:
            - labelSelector:
                matchExpressions:
                  - key: run
                    operator: In
                    values:
                      - centos-logtest
              namespaceSelector: {}
              topologyKey: kubernetes.io/hostname
        podAntiAffinity:
          preferredDuringSchedulingIgnoredDuringExecution:
            - podAffinityTerm:
                labelSelector:
                  matchExpressions:
                    - key: security
                      operator: In
                      values:
                        - S2
                topologyKey: topology.kubernetes.io/zone
              weight: 100
    inputs:
    - name: collector-receiver
      receiver:
        http:
          format: kubeAPIAudit
        port: 8443
        type: http
      type: receiver
    outputs:
    - name: rsyslog
      type: syslog
      syslog:
        rfc: ${RFC}
        url: ${URL}
    pipelines:
    - inputRefs:
      - collector-receiver
      name: forward-to-syslog
      outputRefs:
        - rsyslog
    serviceAccount:
      name: ${SERVICE_ACCOUNT_NAME}
parameters:
- name: NAME
  value: "logcollector"
- name: NAMESPACE
  value: "openshift-logging"
- name: URL
  value: "udp://rsyslogserver.openshift-logging.svc:514"
- name: SERVICE_ACCOUNT_NAME
  value: ""
- name: RFC
  value: RFC5424
- name: DEPLOYMENT
  value: "true"
`)

func testdataObservabilityOpenshiftIo_clusterlogforwarderAffinity81397YamlBytes() ([]byte, error) {
	return _testdataObservabilityOpenshiftIo_clusterlogforwarderAffinity81397Yaml, nil
}

func testdataObservabilityOpenshiftIo_clusterlogforwarderAffinity81397Yaml() (*asset, error) {
	bytes, err := testdataObservabilityOpenshiftIo_clusterlogforwarderAffinity81397YamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "testdata/observability.openshift.io_clusterlogforwarder/affinity-81397.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testdataObservabilityOpenshiftIo_clusterlogforwarderAffinity81398Yaml = []byte(`kind: Template
apiversion: template.openshift.io/v1
metadata:
  name: clusterlogforwarder-template
objects:
- apiVersion: observability.openshift.io/v1
  kind: ClusterLogForwarder
  metadata:
    name: ${NAME}
    namespace: ${NAMESPACE}
  spec:
    collector:
      affinity:
        nodeAffinity:
          preferredDuringSchedulingIgnoredDuringExecution:
            - preference:
                matchExpressions:
                  - key: label-1
                    operator: Exists
              weight: 1
            - preference:
                matchExpressions:
                  - key: label-2
                    operator: In
                    values:
                      - key-2
              weight: 50
          requiredDuringSchedulingIgnoredDuringExecution:
            nodeSelectorTerms:
              - matchExpressions:
                  - key: kubernetes.io/os
                    operator: In
                    values:
                      - linux
              - matchExpressions:
                  - key: node-role.kubernetes.io/worker
                    operator: Exists
        podAffinity:
          preferredDuringSchedulingIgnoredDuringExecution:
            - podAffinityTerm:
                labelSelector:
                  matchExpressions:
                    - key: qe-test
                      operator: In
                      values:
                        - value1
                topologyKey: kubernetes.io/hostname
              weight: 50
          requiredDuringSchedulingIgnoredDuringExecution:
            - labelSelector:
                matchExpressions:
                  - key: run
                    operator: In
                    values:
                      - centos-logtest
              namespaceSelector: {}
              topologyKey: kubernetes.io/hostname
        podAntiAffinity:
          preferredDuringSchedulingIgnoredDuringExecution:
            - podAffinityTerm:
                labelSelector:
                  matchExpressions:
                    - key: security
                      operator: In
                      values:
                        - S2
                topologyKey: topology.kubernetes.io/zone
              weight: 100
    inputs:
    - application:
        includes:
        - namespace: ${NAMESPACE_PATTERN}
      name: app-input-namespace
      type: application
    outputs:
    - name: rsyslog
      type: syslog
      syslog:
        rfc: ${RFC}
        url: ${URL}
    pipelines:
    - inputRefs: ${{INPUT_REFS}}
      name: forward-to-syslog
      outputRefs:
        - rsyslog
    serviceAccount:
      name: ${SERVICE_ACCOUNT_NAME}
parameters:
- name: NAME
  value: "instance"
- name: NAMESPACE
  value: "openshift-logging"
- name: URL
  value: "udp://rsyslogserver.openshift-logging.svc:514"
- name: SERVICE_ACCOUNT_NAME
  value: ""
- name: NAMESPACE_PATTERN
  value: ""
- name: INPUT_REFS
  value: "[\"infrastructure\", \"audit\", \"app-input-namespace\"]"
- name: RFC
  value: RFC5424
`)

func testdataObservabilityOpenshiftIo_clusterlogforwarderAffinity81398YamlBytes() ([]byte, error) {
	return _testdataObservabilityOpenshiftIo_clusterlogforwarderAffinity81398Yaml, nil
}

func testdataObservabilityOpenshiftIo_clusterlogforwarderAffinity81398Yaml() (*asset, error) {
	bytes, err := testdataObservabilityOpenshiftIo_clusterlogforwarderAffinity81398YamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "testdata/observability.openshift.io_clusterlogforwarder/affinity-81398.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testdataObservabilityOpenshiftIo_clusterlogforwarderAuditPolicyYaml = []byte(`kind: Template
apiversion: template.openshift.io/v1
metadata:
  name: clusterlogforwarder-template
objects:
- apiVersion: observability.openshift.io/v1
  kind: ClusterLogForwarder
  metadata:
    name: ${NAME}
    namespace: ${NAMESPACE}
  spec:
    filters:
    - name: my-policy
      type: kubeAPIAudit
      kubeAPIAudit:
        omitStages:
        - "RequestReceived"
        rules:
        - level: RequestResponse
          resources:
          - group: ""
            resources: ["pods"]
        - level: Request
          resources:
          - group: ""
            resources: ["pods/binding", "pods/status"]
        - level: None
          resources:
          - group: ""
            resources: ["configmaps"]
            resourceNames: ["merged-trusted-image-registry-ca"]
        - level: Request
          resources:
          - group: ""
            resources: ["configmaps"]
          namespaces: ["openshift-multus"]
        - level: RequestResponse
          resources:
          - group: ""
            resources: ["secrets", "configmaps"]
        - level: None
          users: ["system:serviceaccount:openshift-monitoring:prometheus-k8s"]
          verbs: ["watch"]
          resources:
          - group: ""
            resources: ["endpoints", "services", "pods"]
        - level: None
          userGroups: ["system:authenticated"]
          nonResourceURLs:
          - "/openapi*"
          - "/metrics"
        - level: Request
          resources:
          - group: ""
          - group: "operators.coreos.com"
          - group: "rbac.authorization.k8s.io"
        - level: Metadata
    managementState: ${MANAGEMENT_STATE}
    outputs:
    - name: lokistack
      type: lokiStack
      lokiStack:
        authentication:
          token:
            from: serviceAccount
        target:
          name: ${LOKISTACK_NAME}
          namespace: ${LOKISTACK_NAMESPACE}
      tls:
        ca:
          key: ca-bundle.crt
          secretName: ${SECRET_NAME}
    pipelines:
    - name: enable-audit-polciy
      filterRefs: ${{FILTER_REFS}}
      inputRefs: ${{INPUT_REFS}}
      outputRefs:
      - lokistack
    serviceAccount:
      name: ${SERVICE_ACCOUNT_NAME}
parameters:
- name: NAME
  value: "instance"
- name: NAMESPACE
  value: "openshift-logging"
- name: INPUT_REFS
  value: "[\"infrastructure\", \"audit\", \"application\"]"
- name: SERVICE_ACCOUNT_NAME
  value: ""
- name: LOKISTACK_NAME
  value: "logging-loki"
- name: LOKISTACK_NAMESPACE
  value: "openshift-logging"
- name: MANAGEMENT_STATE
  value: "Managed"
- name: FILTER_REFS
  value: "[\"my-policy\"]"
- name: SECRET_NAME
  value: ""
`)

func testdataObservabilityOpenshiftIo_clusterlogforwarderAuditPolicyYamlBytes() ([]byte, error) {
	return _testdataObservabilityOpenshiftIo_clusterlogforwarderAuditPolicyYaml, nil
}

func testdataObservabilityOpenshiftIo_clusterlogforwarderAuditPolicyYaml() (*asset, error) {
	bytes, err := testdataObservabilityOpenshiftIo_clusterlogforwarderAuditPolicyYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "testdata/observability.openshift.io_clusterlogforwarder/audit-policy.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testdataObservabilityOpenshiftIo_clusterlogforwarderAzuremonitorMinOptsYaml = []byte(`kind: Template
apiVersion: template.openshift.io/v1
metadata:
  name: clusterlogforwarder-template
objects:
- apiVersion: "observability.openshift.io/v1"
  kind: ClusterLogForwarder
  metadata:
    name: ${NAME}
    namespace: ${NAMESPACE}
  spec:
    collector: ${{COLLECTOR}}
    outputs:
    - name: azure-app
      type: azureMonitor
      azureMonitor:
        customerId: ${CUSTOMER_ID}
        logType: ${PREFIX_OR_NAME}app_log
        authentication:
          sharedKey:
            key: shared_key
            secretName: ${SECRET_NAME}
    - name: azure-infra
      type: azureMonitor
      azureMonitor:
        customerId: ${CUSTOMER_ID}
        logType: ${PREFIX_OR_NAME}infra_log
        authentication:
          sharedKey:
            key: shared_key
            secretName: ${SECRET_NAME}
    - name: azure-audit
      type: azureMonitor
      azureMonitor:
        customerId: ${CUSTOMER_ID}
        logType: ${PREFIX_OR_NAME}audit_log
        authentication:
          sharedKey:
            key: shared_key
            secretName: ${SECRET_NAME}
    pipelines:
    - name: pipe1
      inputRefs:
      - application
      outputRefs:
      - azure-app
    - name: pipe2
      inputRefs:
      - infrastructure
      outputRefs:
      - azure-infra
    - name: pipe3
      inputRefs:
      - audit
      outputRefs:
      - azure-audit
    serviceAccount:
      name: ${SERVICE_ACCOUNT_NAME}
parameters:
- name: NAME
  value: "instance"
- name: NAMESPACE
  value: "openshift-logging"
- name: SECRET_NAME
  value: ""
- name: PREFIX_OR_NAME
  value: ""
- name: CUSTOMER_ID
  value: ""
- name: SERVICE_ACCOUNT_NAME
  value: ""
- name: COLLECTOR
  value: "{}"
`)

func testdataObservabilityOpenshiftIo_clusterlogforwarderAzuremonitorMinOptsYamlBytes() ([]byte, error) {
	return _testdataObservabilityOpenshiftIo_clusterlogforwarderAzuremonitorMinOptsYaml, nil
}

func testdataObservabilityOpenshiftIo_clusterlogforwarderAzuremonitorMinOptsYaml() (*asset, error) {
	bytes, err := testdataObservabilityOpenshiftIo_clusterlogforwarderAzuremonitorMinOptsYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "testdata/observability.openshift.io_clusterlogforwarder/azureMonitor-min-opts.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testdataObservabilityOpenshiftIo_clusterlogforwarderAzuremonitorYaml = []byte(`kind: Template
apiVersion: template.openshift.io/v1
metadata:
  name: clusterlogforwarder-template
objects:
- apiVersion: "observability.openshift.io/v1"
  kind: ClusterLogForwarder
  metadata:
    name: ${NAME}
    namespace: ${NAMESPACE}
  spec:
    collector: ${{COLLECTOR}}
    managementState: Managed
    outputs:
    - name: azure-app
      type: azureMonitor
      azureMonitor:
        authentication:
          sharedKey:
            key: shared_key
            secretName: ${SECRET_NAME}
        customerId: ${CUSTOMER_ID}
        logType: ${PREFIX_OR_NAME}app_log
        azureResourceId: ${RESOURCE_ID}
        host: ${AZURE_HOST}
    - name: azure-infra
      type: azureMonitor
      azureMonitor:
        authentication:
          sharedKey:
            key: shared_key
            secretName: ${SECRET_NAME}
        customerId: ${CUSTOMER_ID}
        logType: ${PREFIX_OR_NAME}infra_log
        azureResourceId: ${RESOURCE_ID}
        host: ${AZURE_HOST}
    - name: azure-audit
      type: azureMonitor
      azureMonitor:
        authentication:
          sharedKey:
            key: shared_key
            secretName: ${SECRET_NAME}
        customerId: ${CUSTOMER_ID}
        logType: ${PREFIX_OR_NAME}audit_log
        azureResourceId: ${RESOURCE_ID}
        host: ${AZURE_HOST}
    pipelines:
    - name: pipe1
      inputRefs:
      - application
      outputRefs:
      - azure-app
    - name: pipe2
      inputRefs:
      - infrastructure
      outputRefs:
      - azure-infra
    - name: pipe3
      inputRefs:
      - audit
      outputRefs:
      - azure-audit
    serviceAccount:
      name: ${SERVICE_ACCOUNT_NAME}
parameters:
- name: NAME
  value: "instance"
- name: NAMESPACE
  value: "openshift-logging"
- name: SECRET_NAME
  value: ""
- name: PREFIX_OR_NAME
  value: ""
- name: CUSTOMER_ID
  value: ""
- name: RESOURCE_ID
  value: ""
- name: AZURE_HOST
  value: ""
- name: SERVICE_ACCOUNT_NAME
  value: ""
- name: COLLECTOR
  value: "{}"
`)

func testdataObservabilityOpenshiftIo_clusterlogforwarderAzuremonitorYamlBytes() ([]byte, error) {
	return _testdataObservabilityOpenshiftIo_clusterlogforwarderAzuremonitorYaml, nil
}

func testdataObservabilityOpenshiftIo_clusterlogforwarderAzuremonitorYaml() (*asset, error) {
	bytes, err := testdataObservabilityOpenshiftIo_clusterlogforwarderAzuremonitorYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "testdata/observability.openshift.io_clusterlogforwarder/azureMonitor.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testdataObservabilityOpenshiftIo_clusterlogforwarderClfExternalLokiWithSecretTenantkeyYaml = []byte(`kind: Template
apiVersion: template.openshift.io/v1
metadata:
  name: clusterlogforwarder-template
objects:
- apiVersion: observability.openshift.io/v1
  kind: ClusterLogForwarder
  metadata:
    name: ${NAME}
    namespace: ${NAMESPACE}
  spec:
    collector: ${{COLLECTOR}}
    outputs:
    - name: loki-server
      type: loki
      loki:
        authentication:
          password:
            key: password
            secretName: ${SECRET_NAME}
          username:
            key: username
            secretName: ${SECRET_NAME}
        url: ${LOKI_URL}
        tenantKey: ${TENANTKEY}
    pipelines:
      - name: to-loki
        inputRefs:
        - application
        outputRefs:
        - loki-server
    serviceAccount:
      name: ${SERVICE_ACCOUNT_NAME}
parameters:
- name: NAME
  value: "instance"
- name: NAMESPACE
  value: "openshift-logging"
- name: LOKI_URL
  required: true
- name: SECRET_NAME
  value: "loki-client"
- name: TENANTKEY
  value: "{.log_type||\"none\"}"
  required: true
- name: SERVICE_ACCOUNT_NAME
  value: ""
- name: COLLECTOR
  value: "{}"
`)

func testdataObservabilityOpenshiftIo_clusterlogforwarderClfExternalLokiWithSecretTenantkeyYamlBytes() ([]byte, error) {
	return _testdataObservabilityOpenshiftIo_clusterlogforwarderClfExternalLokiWithSecretTenantkeyYaml, nil
}

func testdataObservabilityOpenshiftIo_clusterlogforwarderClfExternalLokiWithSecretTenantkeyYaml() (*asset, error) {
	bytes, err := testdataObservabilityOpenshiftIo_clusterlogforwarderClfExternalLokiWithSecretTenantkeyYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "testdata/observability.openshift.io_clusterlogforwarder/clf-external-loki-with-secret-tenantKey.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testdataObservabilityOpenshiftIo_clusterlogforwarderClfExternalLokiWithSecretYaml = []byte(`kind: Template
apiVersion: template.openshift.io/v1
metadata:
  name: clusterlogforwarder-template
objects:
- apiVersion: observability.openshift.io/v1
  kind: ClusterLogForwarder
  metadata:
    name: ${NAME}
    namespace: ${NAMESPACE}
  spec:
    collector: ${{COLLECTOR}}
    outputs:
    - name: ${OUTPUTNAME}
      type: loki
      loki:
        authentication:
          password:
            key: password
            secretName: ${SECRET_NAME}
          username:
            key: username
            secretName: ${SECRET_NAME}
        url: ${LOKI_URL}
        tuning: ${{TUNING}}
    pipelines:
    - name: to-loki
      inputRefs: ${{INPUTREFS}}
      outputRefs: ${{OUTPUTREFS}}
    serviceAccount:
      name: ${SERVICE_ACCOUNT_NAME}
parameters:
- name: NAME
  value: "instance"
- name: NAMESPACE
  value: "openshift-logging"
- name: LOKI_URL
  required: true
- name: SECRET_NAME
  value: "loki-client"
- name: INPUTREFS
  value: "[\"infrastructure\", \"audit\", \"application\"]"
- name: OUTPUTREFS
  value: "[\"loki-server\"]"
- name: OUTPUTNAME
  value: "loki-server"
- name: SERVICE_ACCOUNT_NAME
  value: ""
- name: TUNING
  value: "{}"
- name: COLLECTOR
  value: "{}"
`)

func testdataObservabilityOpenshiftIo_clusterlogforwarderClfExternalLokiWithSecretYamlBytes() ([]byte, error) {
	return _testdataObservabilityOpenshiftIo_clusterlogforwarderClfExternalLokiWithSecretYaml, nil
}

func testdataObservabilityOpenshiftIo_clusterlogforwarderClfExternalLokiWithSecretYaml() (*asset, error) {
	bytes, err := testdataObservabilityOpenshiftIo_clusterlogforwarderClfExternalLokiWithSecretYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "testdata/observability.openshift.io_clusterlogforwarder/clf-external-loki-with-secret.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testdataObservabilityOpenshiftIo_clusterlogforwarderClfKafkaMultiBrokersYaml = []byte(`kind: Template
apiVersion: template.openshift.io/v1
metadata:
  name: forward-to-kafka-template
objects:
- apiVersion: observability.openshift.io/v1
  kind: ClusterLogForwarder
  metadata:
    name: ${NAME}
    namespace: ${NAMESPACE}
  spec:
    collector: ${{COLLECTOR}}
    managementState: Managed
    inputs:
    - name: selected-app
      application:
        includes:
          - namespace: ${NAMESPACE_PATTERN}
      type: application
    outputs:
    - kafka:
        brokers: ${{BROKERS}}
        topic: ${TOPIC}
        tuning: ${{TUNING}}
      name: kafka-brokers
      type: kafka
    pipelines:
    - inputRefs:
      - selected-app
      - audit
      name: pipe1
      outputRefs:
      - kafka-brokers
    serviceAccount:
      name: ${SERVICE_ACCOUNT_NAME}
parameters:
- name: NAME
  value: "instance"
- name: BROKERS
  value: "[\"tls://my-cluster-kafka-bootstrap.amq-aosqe1.svc:9092\", \"tls://my-cluster-kafka-bootstrap.amq-aosqe2.svc:9092\"]"
- name: TOPIC
  value: "{.log_type||\"none-typed-logs\"}"
- name: SERVICE_ACCOUNT_NAME
  value: ""
- name: TUNING
  value: "{}"
- name: NAMESPACE_PATTERN
  value: "e2e*"
- name: NAMESPACE
  value: "openshift-logging"
- name: COLLECTOR
  value: "{}"
`)

func testdataObservabilityOpenshiftIo_clusterlogforwarderClfKafkaMultiBrokersYamlBytes() ([]byte, error) {
	return _testdataObservabilityOpenshiftIo_clusterlogforwarderClfKafkaMultiBrokersYaml, nil
}

func testdataObservabilityOpenshiftIo_clusterlogforwarderClfKafkaMultiBrokersYaml() (*asset, error) {
	bytes, err := testdataObservabilityOpenshiftIo_clusterlogforwarderClfKafkaMultiBrokersYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "testdata/observability.openshift.io_clusterlogforwarder/clf-kafka-multi-brokers.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testdataObservabilityOpenshiftIo_clusterlogforwarderClfKafkaMultiTopicsYaml = []byte(`kind: Template
apiVersion: template.openshift.io/v1
metadata:
  name: forward-to-kafka-template
objects:
- apiVersion: observability.openshift.io/v1
  kind: ClusterLogForwarder
  metadata:
    name: ${NAME}
    namespace: ${NAMESPACE}
  spec:
    collector: ${{COLLECTOR}}
    inputs:
    - application:
        includes:
        - namespace: ${NAMESPACE_PATTERN}
      name: app-input-namespace
      type: application
    outputs:
    - name: kafka-infra
      type: kafka
      kafka:
        brokers:
        - tcp://${BOOTSTRAP_SVC}
        topic: ${INFRA_TOPIC}
        authentication:
          sasl:
            mechanism: "SCRAM-SHA-512"
            password:
              key: password
              secretName: ${SECRET_NAME}
            username:
              key: username
              secretName: ${SECRET_NAME}
    - name: kafka-app
      type: kafka
      kafka:
        brokers:
        - tcp://${BOOTSTRAP_SVC}
        topic: ${APP_TOPIC}
        authentication:
          sasl:
            mechanism: "SCRAM-SHA-512"
            password:
              key: password
              secretName: ${SECRET_NAME}
            username:
              key: username
              secretName: ${SECRET_NAME}
    - name: kafka-audit
      type: kafka
      kafka:
        brokers:
        - tcp://${BOOTSTRAP_SVC}
        topic: ${AUDIT_TOPIC}
        authentication:
          sasl:
            mechanism: "SCRAM-SHA-512"
            password:
              key: password
              secretName: ${SECRET_NAME}
            username:
              key: username
              secretName: ${SECRET_NAME}
    pipelines:
    - inputRefs:
      - infrastructure
      name: test-infra
      outputRefs:
      - kafka-infra
    - inputRefs:
      - app-input-namespace
      name: test-app
      outputRefs:
      - kafka-app
    - inputRefs:
      - audit
      name: test-audit
      outputRefs:
      - kafka-audit
    serviceAccount:
      name: ${SERVICE_ACCOUNT_NAME}
parameters:
- name: NAME
  value: "instance"
- name: BOOTSTRAP_SVC
  value: ""
- name: NAMESPACE
  value: "openshift-logging"
- name: NAMESPACE_PATTERN
  value: "e2e*"
- name: SERVICE_ACCOUNT_NAME
  value: ""
- name: SECRET_NAME
  value: ""
- name: APP_TOPIC
  value: "{.log_type||\"none-typed-logs\"}"
- name: INFRA_TOPIC
  value: "{.log_type||\"none-typed-logs\"}"
- name: AUDIT_TOPIC
  value: "{.log_type||\"none-typed-logs\"}"
- name: COLLECTOR
  value: "{}"
`)

func testdataObservabilityOpenshiftIo_clusterlogforwarderClfKafkaMultiTopicsYamlBytes() ([]byte, error) {
	return _testdataObservabilityOpenshiftIo_clusterlogforwarderClfKafkaMultiTopicsYaml, nil
}

func testdataObservabilityOpenshiftIo_clusterlogforwarderClfKafkaMultiTopicsYaml() (*asset, error) {
	bytes, err := testdataObservabilityOpenshiftIo_clusterlogforwarderClfKafkaMultiTopicsYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "testdata/observability.openshift.io_clusterlogforwarder/clf-kafka-multi-topics.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testdataObservabilityOpenshiftIo_clusterlogforwarderClfKafkaNoAuthYaml = []byte(`kind: Template
apiVersion: template.openshift.io/v1
metadata:
  name: clusterlogforwarder-template
objects:
- apiVersion: observability.openshift.io/v1
  kind: ClusterLogForwarder
  metadata:
    name: ${NAME}
    namespace: ${NAMESPACE}
  spec:
    collector: ${{COLLECTOR}}
    outputs:
    - kafka:
        authentication:
          sasl:
            mechanism: ${SASL_MECHANISM}
        url: ${URL}
      name: kafka-app
      tls:
        ca:
          key: ${TLS_CA_KEY}
          secretName: ${SECRET_NAME}
        certificate:
          key: ${TLS_CERTIFICATE_KEY}
          secretName: ${SECRET_NAME}
        key:
          key: ${TLS_KEY}
          secretName: ${SECRET_NAME}
      type: kafka
    pipelines:
    - name: test-app
      inputRefs:
      - application
      outputRefs:
      - kafka-app
    serviceAccount:
      name: ${SERVICE_ACCOUNT_NAME}
parameters:
- name: NAME
  value: "instance"
- name: NAMESPACE
  value: "openshift-logging"
- name: URL
  value: "tls://kafka.openshift-logging.svc.cluster.local:9093/clo-topic"
- name: SECRET_NAME
  value: "kafka-vector"
- name: SERVICE_ACCOUNT_NAME
  value: ""
- name: SASL_MECHANISM
  value: "PLAIN"
- name: TLS_CA_KEY
  value: "ca-bundle.crt"
- name: TLS_CERTIFICATE_KEY
  value: "tls.crt"
- name: TLS_KEY
  value: "tls.key"
- name: COLLECTOR
  value: "{}"
`)

func testdataObservabilityOpenshiftIo_clusterlogforwarderClfKafkaNoAuthYamlBytes() ([]byte, error) {
	return _testdataObservabilityOpenshiftIo_clusterlogforwarderClfKafkaNoAuthYaml, nil
}

func testdataObservabilityOpenshiftIo_clusterlogforwarderClfKafkaNoAuthYaml() (*asset, error) {
	bytes, err := testdataObservabilityOpenshiftIo_clusterlogforwarderClfKafkaNoAuthYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "testdata/observability.openshift.io_clusterlogforwarder/clf-kafka-no-auth.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testdataObservabilityOpenshiftIo_clusterlogforwarderClfKafkaSaslSslYaml = []byte(`kind: Template
apiVersion: template.openshift.io/v1
metadata:
  name: clusterlogforwarder-template
objects:
- apiVersion: observability.openshift.io/v1
  kind: ClusterLogForwarder
  metadata:
    name: ${NAME}
    namespace: ${NAMESPACE}
  spec:
    collector: ${{COLLECTOR}}
    inputs:
    - name: selected-app
      application:
        includes:
          - namespace: ${NAMESPACE_PATTERN}
      type: application
    outputs:
    - name: amq-instance
      type: kafka
      kafka:
        authentication:
          sasl:
            mechanism: "SCRAM-SHA-512"
            password:
              key: password
              secretName: ${SECRET_NAME}
            username:
              key: username
              secretName: ${SECRET_NAME}
        url: ${URL}
      tls:
        ca:
          key: ca-bundle.crt
          secretName: ${SECRET_NAME}
      type: kafka
    pipelines:
    - name: pipe1
      inputRefs:
      - selected-app
      outputRefs:
      - amq-instance
    serviceAccount:
      name: ${SERVICE_ACCOUNT_NAME}
parameters:
- name: NAME
  value: "instance"
- name: URL
  value: "tls://kafka.openshift-logging.svc.cluster.local:9093/clo-topic"
- name: SECRET_NAME
  value: "vector-kafka"
- name: SERVICE_ACCOUNT_NAME
  value: ""
- name: NAMESPACE_PATTERN
  value: "e2e*"
- name: NAMESPACE
  value: "openshift-logging"
- name: COLLECTOR
  value: "{}"
`)

func testdataObservabilityOpenshiftIo_clusterlogforwarderClfKafkaSaslSslYamlBytes() ([]byte, error) {
	return _testdataObservabilityOpenshiftIo_clusterlogforwarderClfKafkaSaslSslYaml, nil
}

func testdataObservabilityOpenshiftIo_clusterlogforwarderClfKafkaSaslSslYaml() (*asset, error) {
	bytes, err := testdataObservabilityOpenshiftIo_clusterlogforwarderClfKafkaSaslSslYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "testdata/observability.openshift.io_clusterlogforwarder/clf-kafka-sasl-ssl.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testdataObservabilityOpenshiftIo_clusterlogforwarderClfKafkaWithAuthYaml = []byte(`kind: Template
apiVersion: template.openshift.io/v1
metadata:
  name: clusterlogforwarder-template
objects:
- apiVersion: observability.openshift.io/v1
  kind: ClusterLogForwarder
  metadata:
    name: ${NAME}
    namespace: ${NAMESPACE}
  spec:
    collector: ${{COLLECTOR}}
    outputs:
    - kafka:
        authentication:
          sasl:
            mechanism: ${SASL_MECHANISM}
            password:
              key: password
              secretName: ${SECRET_NAME}
            username:
              key: username
              secretName: ${SECRET_NAME}
        url: ${URL}
        tuning:
          compression: ${COMPRESSION}
          deliveryMode: ${DELIVERY}
          maxWrite: ${MAX_WRITE}
      name: kafka-app
      tls:
        ca:
          key: ${TLS_CA_KEY}
          secretName: ${TLS_SECRET_NAME}
        certificate:
          key: ${TLS_CERTIFICATE_KEY}
          secretName: ${TLS_SECRET_NAME}
        key:
          key: ${TLS_KEY}
          secretName: ${TLS_SECRET_NAME}
      type: kafka
    pipelines:
    - name: test-app
      inputRefs:
      - application
      outputRefs:
      - kafka-app
    serviceAccount:
      name: ${SERVICE_ACCOUNT_NAME}
parameters:
- name: NAME
  value: "instance"
- name: NAMESPACE
  value: "openshift-logging"
- name: URL
  value: "tls://kafka.openshift-logging.svc.cluster.local:9093/clo-topic"
- name: SECRET_NAME
  value: "vector-kafka"
- name: SERVICE_ACCOUNT_NAME
  value: ""
- name: SASL_MECHANISM
  value: "PLAIN"
- name: TLS_CA_KEY
  value: "ca-bundle.crt"
- name: TLS_CERTIFICATE_KEY
  value: "tls.crt"
- name: TLS_KEY
  value: "tls.key"
- name: TLS_SECRET_NAME
  value: ""
- name: COMPRESSION
  value: "zstd"
- name: MAX_WRITE
  value: "10M"
- name: DELIVERY
  value: "AtLeastOnce"
- name: COLLECTOR
  value: "{}"
`)

func testdataObservabilityOpenshiftIo_clusterlogforwarderClfKafkaWithAuthYamlBytes() ([]byte, error) {
	return _testdataObservabilityOpenshiftIo_clusterlogforwarderClfKafkaWithAuthYaml, nil
}

func testdataObservabilityOpenshiftIo_clusterlogforwarderClfKafkaWithAuthYaml() (*asset, error) {
	bytes, err := testdataObservabilityOpenshiftIo_clusterlogforwarderClfKafkaWithAuthYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "testdata/observability.openshift.io_clusterlogforwarder/clf-kafka-with-auth.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testdataObservabilityOpenshiftIo_clusterlogforwarderCloudwatchAccesskeyYaml = []byte(`kind: Template
apiversion: template.openshift.io/v1
metadata:
  name: clusterlogforwarder-template
objects:
- apiVersion: "observability.openshift.io/v1"
  kind: ClusterLogForwarder
  metadata:
    name: ${NAME}
    namespace: ${NAMESPACE}
  spec:
    collector: ${{COLLECTOR}}
    managementState: Managed
    outputs:
    - name: cloudwatch
      type: cloudwatch
      cloudwatch:
        authentication:
          awsAccessKey:
            keyId:
              key: aws_access_key_id
              secretName: ${SECRET_NAME}
            keySecret:
              key: aws_secret_access_key
              secretName: ${SECRET_NAME}
          type: awsAccessKey
        groupName: ${GROUP_NAME}
        region: ${REGION}
        url: "https://logs.${REGION}.amazonaws.com"
        tuning: ${{TUNING}}
    pipelines:
    - name: to-cloudwatch
      inputRefs: ${{INPUT_REFS}}
      outputRefs:
      - cloudwatch
    serviceAccount:
      name: ${SERVICE_ACCOUNT_NAME}
parameters:
- name: NAME
  value: "instance"
- name: NAMESPACE
  value: "openshift-logging"
- name: SECRET_NAME
  value: "cw-secret"
- name: REGION
  value: "us-east-2"
- name: GROUP_NAME
  value: "{.log_type||\"none-typed-logs\"}"
- name: INPUT_REFS
  value: "[\"infrastructure\", \"audit\", \"application\"]"
- name: SERVICE_ACCOUNT_NAME
  value: ""
- name: TUNING
  value: "{}"
- name: COLLECTOR
  value: "{}"
`)

func testdataObservabilityOpenshiftIo_clusterlogforwarderCloudwatchAccesskeyYamlBytes() ([]byte, error) {
	return _testdataObservabilityOpenshiftIo_clusterlogforwarderCloudwatchAccesskeyYaml, nil
}

func testdataObservabilityOpenshiftIo_clusterlogforwarderCloudwatchAccesskeyYaml() (*asset, error) {
	bytes, err := testdataObservabilityOpenshiftIo_clusterlogforwarderCloudwatchAccesskeyYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "testdata/observability.openshift.io_clusterlogforwarder/cloudwatch-accessKey.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testdataObservabilityOpenshiftIo_clusterlogforwarderCloudwatchIamroleYaml = []byte(`kind: Template
apiversion: template.openshift.io/v1
metadata:
  name: clusterlogforwarder-template
objects:
- apiVersion: "observability.openshift.io/v1"
  kind: ClusterLogForwarder
  metadata:
    name: ${NAME}
    namespace: ${NAMESPACE}
  spec:
    collector: ${{COLLECTOR}}
    managementState: Managed
    outputs:
    - name: cloudwatch
      type: cloudwatch
      cloudwatch:
        authentication:
          iamRole:
            roleARN:
              key: role_arn
              secretName: ${SECRET_NAME}
            token:
              from: serviceAccount
          type: iamRole
        groupName: ${GROUP_NAME}
        region: ${REGION}
        tuning: ${{TUNING}}
    pipelines:
    - name: to-cloudwatch
      inputRefs: ${{INPUT_REFS}}
      outputRefs:
      - cloudwatch
    serviceAccount:
      name: ${SERVICE_ACCOUNT_NAME}
parameters:
- name: NAME
  value: "instance"
- name: NAMESPACE
  value: "openshift-logging"
- name: SECRET_NAME
  value: "cw-secret"
- name: REGION
  value: "us-east-2"
- name: GROUP_NAME
  value: "{.log_type||\"none-typed-logs\"}"
- name: INPUT_REFS
  value: "[\"infrastructure\", \"audit\", \"application\"]"
- name: SERVICE_ACCOUNT_NAME
  value: ""
- name: TUNING
  value: "{}"
- name: COLLECTOR
  value: "{}"
`)

func testdataObservabilityOpenshiftIo_clusterlogforwarderCloudwatchIamroleYamlBytes() ([]byte, error) {
	return _testdataObservabilityOpenshiftIo_clusterlogforwarderCloudwatchIamroleYaml, nil
}

func testdataObservabilityOpenshiftIo_clusterlogforwarderCloudwatchIamroleYaml() (*asset, error) {
	bytes, err := testdataObservabilityOpenshiftIo_clusterlogforwarderCloudwatchIamroleYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "testdata/observability.openshift.io_clusterlogforwarder/cloudwatch-iamRole.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testdataObservabilityOpenshiftIo_clusterlogforwarderCloudwatchMultipleIamroleYaml = []byte(`kind: Template
apiversion: template.openshift.io/v1
metadata:
  name: clusterlogforwarder-template
objects:
- apiVersion: "observability.openshift.io/v1"
  kind: ClusterLogForwarder
  metadata:
    name: ${NAME}
    namespace: ${NAMESPACE}
  spec:
    collector: ${{COLLECTOR}}
    managementState: Managed
    outputs:
    - name: cloudwatch-application
      type: cloudwatch
      cloudwatch:
        authentication:
          iamRole:
            roleARN:
              key: role_arn
              secretName: ${SECRET_NAME_1}
            token:
              from: serviceAccount
          type: iamRole
        groupName: ${GROUP_NAME_1}
        region: ${REGION_1}
    - cloudwatch:
        authentication:
          iamRole:
            roleARN:
              key: role_arn
              secretName: ${SECRET_NAME_2}
            token:
              from: secret
              secret:
                key: token
                name: ${SECRET_NAME_2}
          type: iamRole
        groupName: ${GROUP_NAME_2}
        region: ${REGION_2}
      name: cloudwatch-infra-audit
      type: cloudwatch
    pipelines:
    - name: cloudwatch-1
      inputRefs:
      - application
      outputRefs:
      - cloudwatch-application
    - name: cloudwatch-2
      inputRefs:
      - infrastructure
      - audit
      outputRefs:
      - cloudwatch-infra-audit
    serviceAccount:
      name: ${SERVICE_ACCOUNT_NAME}
parameters:
- name: NAME
  value: "instance"
- name: NAMESPACE
  value: "openshift-logging"
- name: SECRET_NAME_1
- name: SECRET_NAME_2
- name: REGION_1
- name: REGION_2
- name: GROUP_NAME_1
- name: GROUP_NAME_2
- name: SERVICE_ACCOUNT_NAME
- name: COLLECTOR
  value: "{}"
`)

func testdataObservabilityOpenshiftIo_clusterlogforwarderCloudwatchMultipleIamroleYamlBytes() ([]byte, error) {
	return _testdataObservabilityOpenshiftIo_clusterlogforwarderCloudwatchMultipleIamroleYaml, nil
}

func testdataObservabilityOpenshiftIo_clusterlogforwarderCloudwatchMultipleIamroleYaml() (*asset, error) {
	bytes, err := testdataObservabilityOpenshiftIo_clusterlogforwarderCloudwatchMultipleIamroleYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "testdata/observability.openshift.io_clusterlogforwarder/cloudwatch-multiple-iamRole.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testdataObservabilityOpenshiftIo_clusterlogforwarderElasticsearchHttpsYaml = []byte(`kind: Template
apiVersion: template.openshift.io/v1
metadata:
  name: clusterlogforwarder-template
objects:
- apiVersion: "observability.openshift.io/v1"
  kind: ClusterLogForwarder
  metadata:
    name: ${NAME}
    namespace: ${NAMESPACE}
  spec:
    collector: ${{COLLECTOR}}
    managementState: ${MANAGEMENT_STATE}
    outputs:
    - name: es-created-by-user
      type: elasticsearch
      elasticsearch:
        index: ${INDEX}
        tuning: ${{TUNING}}
        url: ${ES_URL}
        version: ${{ES_VERSION}}
      tls:
        ca:
          key: ca-bundle.crt
          secretName: ${SECRET_NAME}
        insecureSkipVerify: ${{INSECURE_SKIP_VERIFY}}
        securityProfile: ${{SECURITY_PROFILE}}
    pipelines:
    - name: forward-to-external-es
      inputRefs: ${{INPUT_REFS}}
      outputRefs:
      - es-created-by-user
    serviceAccount:
      name: ${SERVICE_ACCOUNT_NAME}
parameters:
- name: NAME
  value: "instance"
- name: NAMESPACE
  value: "openshift-logging"
- name: ES_URL
  value: "https://elasticsearch-server.es-aosqe-qa.svc:9200"
- name: SECRET_NAME
  value: "pipelinesecret"
- name: ES_VERSION
  value: "6"
- name: INPUT_REFS
  value: "[\"infrastructure\", \"audit\", \"application\"]"
- name: SERVICE_ACCOUNT_NAME
  value: ""
- name: MANAGEMENT_STATE
  value: "Managed"
- name: TUNING
  value: "{}"
- name: INDEX
  value: "{.log_type||\"none-typed-logs\"}"
- name: INSECURE_SKIP_VERIFY
  value: "false"
- name: SECURITY_PROFILE
  value: "{}"
- name: COLLECTOR
  value: "{}"
`)

func testdataObservabilityOpenshiftIo_clusterlogforwarderElasticsearchHttpsYamlBytes() ([]byte, error) {
	return _testdataObservabilityOpenshiftIo_clusterlogforwarderElasticsearchHttpsYaml, nil
}

func testdataObservabilityOpenshiftIo_clusterlogforwarderElasticsearchHttpsYaml() (*asset, error) {
	bytes, err := testdataObservabilityOpenshiftIo_clusterlogforwarderElasticsearchHttpsYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "testdata/observability.openshift.io_clusterlogforwarder/elasticsearch-https.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testdataObservabilityOpenshiftIo_clusterlogforwarderElasticsearchMtlsYaml = []byte(`kind: Template
apiVersion: template.openshift.io/v1
metadata:
  name: clusterlogforwarder-template
objects:
- apiVersion: "observability.openshift.io/v1"
  kind: ClusterLogForwarder
  metadata:
    name: ${NAME}
    namespace: ${NAMESPACE}
  spec:
    collector: ${{COLLECTOR}}
    managementState: ${MANAGEMENT_STATE}
    outputs:
    - name: es-created-by-user
      type: elasticsearch
      elasticsearch:
        index: ${INDEX}
        tuning: ${{TUNING}}
        url: ${ES_URL}
        version: ${{ES_VERSION}}
      tls:
        ca:
          configMapName:
          key: ca-bundle.crt
          secretName: ${SECRET_NAME} #configmap or secret
        certificate:
          configMapName:
          key: tls.crt
          secretName: ${SECRET_NAME} #configmap or secret
        insecureSkipVerify: ${{INSECURE_SKIP_VERIFY}}
        key:
          key: tls.key
          secretName: ${SECRET_NAME}
    pipelines:
    - name: forward-to-external-es
      inputRefs: ${{INPUT_REFS}}
      outputRefs:
      - es-created-by-user
    serviceAccount:
      name: ${SERVICE_ACCOUNT_NAME}
parameters:
- name: NAME
  value: "instance"
- name: NAMESPACE
  value: "openshift-logging"
- name: ES_URL
  value: "https://elasticsearch-server.es-aosqe-qa.svc:9200"
- name: SECRET_NAME
  value: "pipelinesecret"
- name: ES_VERSION
  value: "6"
- name: INPUT_REFS
  value: "[\"infrastructure\", \"audit\", \"application\"]"
- name: SERVICE_ACCOUNT_NAME
  value: ""
- name: COLLECTOR
  value: "{}"
- name: MANAGEMENT_STATE
  value: "Managed"
- name: TUNING
  value: "{}"
- name: INDEX
  value: "{.log_type||\"none-typed-logs\"}"
- name: INSECURE_SKIP_VERIFY
  value: "false"
`)

func testdataObservabilityOpenshiftIo_clusterlogforwarderElasticsearchMtlsYamlBytes() ([]byte, error) {
	return _testdataObservabilityOpenshiftIo_clusterlogforwarderElasticsearchMtlsYaml, nil
}

func testdataObservabilityOpenshiftIo_clusterlogforwarderElasticsearchMtlsYaml() (*asset, error) {
	bytes, err := testdataObservabilityOpenshiftIo_clusterlogforwarderElasticsearchMtlsYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "testdata/observability.openshift.io_clusterlogforwarder/elasticsearch-mtls.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testdataObservabilityOpenshiftIo_clusterlogforwarderElasticsearchUserauthHttpsYaml = []byte(`kind: Template
apiVersion: template.openshift.io/v1
metadata:
  name: clusterlogforwarder-template
objects:
- apiVersion: "observability.openshift.io/v1"
  kind: ClusterLogForwarder
  metadata:
    name: ${NAME}
    namespace: ${NAMESPACE}
  spec:
    collector: ${{COLLECTOR}}
    managementState: ${MANAGEMENT_STATE}
    outputs:
    - name: es-created-by-user
      type: elasticsearch
      elasticsearch:
        authentication:
          password:
            key: password
            secretName: ${SECRET_NAME}
          username:
            key: username
            secretName: ${SECRET_NAME}
        index: ${INDEX}
        tuning: ${{TUNING}}
        url: ${ES_URL}
        version: ${{ES_VERSION}}
      tls:
        ca:
          configMapName:
          key: ca-bundle.crt
          secretName: ${SECRET_NAME}
        insecureSkipVerify: ${{INSECURE_SKIP_VERIFY}}
    pipelines:
    - name: forward-to-external-es
      inputRefs: ${{INPUT_REFS}}
      outputRefs:
      - es-created-by-user
    serviceAccount:
      name: ${SERVICE_ACCOUNT_NAME}
parameters:
- name: NAME
  value: "instance"
- name: NAMESPACE
  value: "openshift-logging"
- name: ES_URL
  value: "https://elasticsearch-server.es-aosqe-qa.svc:9200"
- name: SECRET_NAME
  value: "pipelinesecret"
- name: ES_VERSION
  value: "6"
- name: INPUT_REFS
  value: "[\"infrastructure\", \"audit\", \"application\"]"
- name: SERVICE_ACCOUNT_NAME
  value: ""
- name: COLLECTOR
  value: "{}"
- name: MANAGEMENT_STATE
  value: "Managed"
- name: TUNING
  value: "{}"
- name: INDEX
  value: "{.log_type||\"none-typed-logs\"}"
- name: INSECURE_SKIP_VERIFY
  value: "false"
`)

func testdataObservabilityOpenshiftIo_clusterlogforwarderElasticsearchUserauthHttpsYamlBytes() ([]byte, error) {
	return _testdataObservabilityOpenshiftIo_clusterlogforwarderElasticsearchUserauthHttpsYaml, nil
}

func testdataObservabilityOpenshiftIo_clusterlogforwarderElasticsearchUserauthHttpsYaml() (*asset, error) {
	bytes, err := testdataObservabilityOpenshiftIo_clusterlogforwarderElasticsearchUserauthHttpsYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "testdata/observability.openshift.io_clusterlogforwarder/elasticsearch-userauth-https.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testdataObservabilityOpenshiftIo_clusterlogforwarderElasticsearchUserauthMtlsYaml = []byte(`kind: Template
apiVersion: template.openshift.io/v1
metadata:
  name: clusterlogforwarder-template
objects:
- apiVersion: "observability.openshift.io/v1"
  kind: ClusterLogForwarder
  metadata:
    name: ${NAME}
    namespace: ${NAMESPACE}
  spec:
    collector: ${{COLLECTOR}}
    managementState: ${MANAGEMENT_STATE}
    outputs:
    - name: es-created-by-user
      type: elasticsearch
      elasticsearch:
        authentication:
          password:
            key: password
            secretName: ${SECRET_NAME}
          username:
            key: username
            secretName: ${SECRET_NAME}
        index: ${INDEX}
        tuning: ${{TUNING}}
        url: ${ES_URL}
        version: ${{ES_VERSION}}
      tls:
        ca:
          configMapName:
          key: ca-bundle.crt
          secretName: ${SECRET_NAME}
        certificate:
          configMapName:
          key: tls.crt
          secretName: ${SECRET_NAME}
        insecureSkipVerify: ${{INSECURE_SKIP_VERIFY}}
        key:
          key: tls.key
          secretName: ${SECRET_NAME}
    pipelines:
    - name: forward-to-external-es
      inputRefs: ${{INPUT_REFS}}
      outputRefs:
      - es-created-by-user
    serviceAccount:
      name: ${SERVICE_ACCOUNT_NAME}
parameters:
- name: NAME
  value: "instance"
- name: NAMESPACE
  value: "openshift-logging"
- name: ES_URL
  value: "https://elasticsearch-server.es-aosqe-qa.svc:9200"
- name: SECRET_NAME
  value: "pipelinesecret"
- name: ES_VERSION
  value: "6"
- name: INPUT_REFS
  value: "[\"infrastructure\", \"audit\", \"application\"]"
- name: SERVICE_ACCOUNT_NAME
  value: ""
- name: INSECURE_SKIP_VERIFY
  value: "false"
- name: COLLECTOR
  value: "{}"
- name: MANAGEMENT_STATE
  value: "Managed"
- name: TUNING
  value: "{}"
- name: INDEX
  value: "{.log_type||\"none-typed-logs\"}"
`)

func testdataObservabilityOpenshiftIo_clusterlogforwarderElasticsearchUserauthMtlsYamlBytes() ([]byte, error) {
	return _testdataObservabilityOpenshiftIo_clusterlogforwarderElasticsearchUserauthMtlsYaml, nil
}

func testdataObservabilityOpenshiftIo_clusterlogforwarderElasticsearchUserauthMtlsYaml() (*asset, error) {
	bytes, err := testdataObservabilityOpenshiftIo_clusterlogforwarderElasticsearchUserauthMtlsYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "testdata/observability.openshift.io_clusterlogforwarder/elasticsearch-userauth-mtls.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testdataObservabilityOpenshiftIo_clusterlogforwarderElasticsearchUserauthYaml = []byte(`kind: Template
apiVersion: template.openshift.io/v1
metadata:
  name: clusterlogforwarder-template
objects:
- apiVersion: "observability.openshift.io/v1"
  kind: ClusterLogForwarder
  metadata:
    name: ${NAME}
    namespace: ${NAMESPACE}
  spec:
    collector: ${{COLLECTOR}}
    managementState: ${MANAGEMENT_STATE}
    outputs:
    - name: es-created-by-user
      type: elasticsearch
      elasticsearch:
        authentication:
          password:
            key: password
            secretName: ${SECRET_NAME}
          username:
            key: username
            secretName: ${SECRET_NAME}
        index: ${INDEX}
        tuning: ${{TUNING}}
        url: ${ES_URL}
        version: ${{ES_VERSION}}
    pipelines:
    - name: forward-to-external-es
      inputRefs: ${{INPUT_REFS}}
      outputRefs:
      - es-created-by-user
    serviceAccount:
      name: ${SERVICE_ACCOUNT_NAME}
parameters:
- name: NAME
  value: "instance"
- name: NAMESPACE
  value: "openshift-logging"
- name: ES_URL
  value: "http://elasticsearch-server.es-aosqe-qa.svc:9200"
- name: SECRET_NAME
  value: "pipelinesecret"
- name: ES_VERSION
  value: "6"
- name: INPUT_REFS
  value: "[\"infrastructure\", \"audit\", \"application\"]"
- name: SERVICE_ACCOUNT_NAME
  value: ""
- name: COLLECTOR
  value: "{}"
- name: MANAGEMENT_STATE
  value: "Managed"
- name: TUNING
  value: "{}"
- name: INDEX
  value: "{.log_type||\"none-typed-logs\"}"
`)

func testdataObservabilityOpenshiftIo_clusterlogforwarderElasticsearchUserauthYamlBytes() ([]byte, error) {
	return _testdataObservabilityOpenshiftIo_clusterlogforwarderElasticsearchUserauthYaml, nil
}

func testdataObservabilityOpenshiftIo_clusterlogforwarderElasticsearchUserauthYaml() (*asset, error) {
	bytes, err := testdataObservabilityOpenshiftIo_clusterlogforwarderElasticsearchUserauthYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "testdata/observability.openshift.io_clusterlogforwarder/elasticsearch-userauth.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testdataObservabilityOpenshiftIo_clusterlogforwarderElasticsearchYaml = []byte(`kind: Template
apiVersion: template.openshift.io/v1
metadata:
  name: clusterlogforwarder-template
objects:
- apiVersion: "observability.openshift.io/v1"
  kind: ClusterLogForwarder
  metadata:
    name: ${NAME}
    namespace: ${NAMESPACE}
  spec:
    collector: ${{COLLECTOR}}
    managementState: ${MANAGEMENT_STATE}
    outputs:
    - name: es-created-by-user
      type: elasticsearch
      elasticsearch:
        index: ${INDEX}
        tuning: ${{TUNING}}
        url: ${ES_URL}
        version: ${{ES_VERSION}}
    pipelines:
    - name: forward-to-external-es
      inputRefs: ${{INPUT_REFS}}
      outputRefs:
      - es-created-by-user
    serviceAccount:
      name: ${SERVICE_ACCOUNT_NAME}
parameters:
- name: NAME
  value: "instance"
- name: NAMESPACE
  value: "openshift-logging"
- name: ES_URL
  value: "http://elasticsearch-server.es-aosqe-qa.svc:9200"
- name: ES_VERSION
  value: "6"
- name: INPUT_REFS
  value: "[\"infrastructure\", \"audit\", \"application\"]"
- name: SERVICE_ACCOUNT_NAME
  value: ""
- name: COLLECTOR
  value: "{}"
- name: MANAGEMENT_STATE
  value: "Managed"
- name: TUNING
  value: "{}"
- name: INDEX
  value: "{.log_type||\"none-typed-logs\"}"
`)

func testdataObservabilityOpenshiftIo_clusterlogforwarderElasticsearchYamlBytes() ([]byte, error) {
	return _testdataObservabilityOpenshiftIo_clusterlogforwarderElasticsearchYaml, nil
}

func testdataObservabilityOpenshiftIo_clusterlogforwarderElasticsearchYaml() (*asset, error) {
	bytes, err := testdataObservabilityOpenshiftIo_clusterlogforwarderElasticsearchYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "testdata/observability.openshift.io_clusterlogforwarder/elasticsearch.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testdataObservabilityOpenshiftIo_clusterlogforwarderGoogleCloudLoggingMultiLogidsYaml = []byte(`kind: Template
apiVersion: template.openshift.io/v1
metadata:
  name: clusterlogforwarder-template
objects:
- apiVersion: "observability.openshift.io/v1"
  kind: ClusterLogForwarder
  metadata:
    name: ${NAME}
    namespace: ${NAMESPACE}
  spec:
    collector: ${{COLLECTOR}}
    managementState: Managed
    outputs:
    - name: gcp-app
      type: googleCloudLogging
      googleCloudLogging:
        authentication:
          credentials:
            key: ${CRED_KEY}
            secretName: ${SECRET_NAME}
        id:
          type: ${ID_TYPE}
          value: ${ID_VALUE}
        logId: ${LOG_ID}-application
    - name: gcp-infra
      type: googleCloudLogging
      googleCloudLogging:
        authentication:
          credentials:
            key: ${CRED_KEY}
            secretName: ${SECRET_NAME}
        id:
          type: ${ID_TYPE}
          value: ${ID_VALUE}
        logId: ${LOG_ID}-infrastructure
    - name: gcp-audit
      type: googleCloudLogging
      googleCloudLogging:
        authentication:
          credentials:
            key: ${CRED_KEY}
            secretName: ${SECRET_NAME}
        id:
          type: ${ID_TYPE}
          value: ${ID_VALUE}
        logId: ${LOG_ID}-audit
    pipelines:
    - name: gcp-app-pipeline
      inputRefs:
        - application
      outputRefs:
        - gcp-app
    - name: gcp-infra-pipeline
      inputRefs:
        - infrastructure
      outputRefs:
        - gcp-infra
    - name: gcp-audit-pipeline
      inputRefs:
        - audit
      outputRefs:
        - gcp-audit
    serviceAccount:
      name: ${SERVICE_ACCOUNT_NAME}
parameters:
- name: NAME
  value: "instance"
- name: NAMESPACE
  value: "openshift-logging"
- name: CRED_KEY
  value: "google-application-credentials.json"
- name: SECRET_NAME
  value: "gcp-secret"
- name: LOG_ID
  value: "{.log_type||\"none-typed-logs\"}"
- name: ID_TYPE
  value: "project"
- name: ID_VALUE
  value: "openshift-qe"
- name: SERVICE_ACCOUNT_NAME
  value: ""
- name: COLLECTOR
  value: "{}"
`)

func testdataObservabilityOpenshiftIo_clusterlogforwarderGoogleCloudLoggingMultiLogidsYamlBytes() ([]byte, error) {
	return _testdataObservabilityOpenshiftIo_clusterlogforwarderGoogleCloudLoggingMultiLogidsYaml, nil
}

func testdataObservabilityOpenshiftIo_clusterlogforwarderGoogleCloudLoggingMultiLogidsYaml() (*asset, error) {
	bytes, err := testdataObservabilityOpenshiftIo_clusterlogforwarderGoogleCloudLoggingMultiLogidsYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "testdata/observability.openshift.io_clusterlogforwarder/google-cloud-logging-multi-logids.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testdataObservabilityOpenshiftIo_clusterlogforwarderGooglecloudloggingYaml = []byte(`kind: Template
apiVersion: template.openshift.io/v1
metadata:
  name: clusterlogforwarder-template
objects:
- apiVersion: "observability.openshift.io/v1"
  kind: ClusterLogForwarder
  metadata:
    name: ${NAME}
    namespace: ${NAMESPACE}
  spec:
    collector: ${{COLLECTOR}}
    managementState: Managed
    outputs:
    - name: gcp-logging
      type: googleCloudLogging
      googleCloudLogging:
        authentication:
          credentials:
            key: ${CRED_KEY}
            secretName: ${SECRET_NAME}
        id:
          type: ${ID_TYPE}
          value: ${ID_VALUE}
        logId : ${LOG_ID}
        tuning:
          deliveryMode: ${DELIVERY}
          maxWrite: ${MAX_WRITE}
          minRetryDuration: ${{MIN_RETRY_DURATION}}
          maxRetryDuration: ${{MAX_RETRY_DURATION}}
    pipelines:
    - name: test-google-cloud-logging
      inputRefs: ${{INPUT_REFS}}
      outputRefs:
      - gcp-logging
    serviceAccount:
      name: ${SERVICE_ACCOUNT_NAME}
parameters:
- name: NAME
  value: "instance"
- name: NAMESPACE
  value: "openshift-logging"
- name: CRED_KEY
  value: "google-application-credentials.json"
- name: SECRET_NAME
  value: "gcp-secret"
- name: LOG_ID
  value: "{.log_type||\"none-typed-logs\"}"
- name: ID_TYPE
  value: "project"
- name: ID_VALUE
  value: "openshift-qe"
- name: INPUT_REFS
  value: "[\"infrastructure\", \"audit\", \"application\"]"
- name: SERVICE_ACCOUNT_NAME
  value: ""
- name: DELIVERY
  value: "AtLeastOnce"
- name: MAX_WRITE
  value: "10M"
- name: MIN_RETRY_DURATION
  value: "10"
- name: MAX_RETRY_DURATION
  value: "20"
- name: COLLECTOR
  value: "{}"
`)

func testdataObservabilityOpenshiftIo_clusterlogforwarderGooglecloudloggingYamlBytes() ([]byte, error) {
	return _testdataObservabilityOpenshiftIo_clusterlogforwarderGooglecloudloggingYaml, nil
}

func testdataObservabilityOpenshiftIo_clusterlogforwarderGooglecloudloggingYaml() (*asset, error) {
	bytes, err := testdataObservabilityOpenshiftIo_clusterlogforwarderGooglecloudloggingYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "testdata/observability.openshift.io_clusterlogforwarder/googleCloudLogging.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testdataObservabilityOpenshiftIo_clusterlogforwarderHttpOutput85490Yaml = []byte(`kind: Template
apiVersion: template.openshift.io/v1
metadata:
  name: clusterlogforwarder-template
objects:
- apiVersion: observability.openshift.io/v1
  kind: ClusterLogForwarder
  metadata:
    name: ${NAME}
    namespace: ${NAMESPACE}
  spec:
    collector: ${{COLLECTOR}}
    outputs:
    - name: httpout-app
      type: http
      http:
        headers:
          h1: v1
          h2: v2
        method: POST
        url: ${URL_APP}
      tls:
        ca:
          key: ${TLS_CA_KEY}
          secretName: ${SECRET_APP}
    - name: httpout-infra
      type: http
      http:
        headers:
          h1: v1
          h2: v2
        method: POST
        url: ${URL_INFRA}
      tls:
        ca:
          key: ${TLS_CA_KEY}
          secretName: ${SECRET_INFRA}
    - name: httpout-audit
      type: http
      http:
        headers:
          h1: v1
          h2: v2
        method: POST
        url: ${URL_AUDIT}
      tls:
        ca:
          key: ${TLS_CA_KEY}
          secretName: ${SECRET_AUDIT}
    pipelines:
    - inputRefs:
      - application
      name: app-logs
      outputRefs:
      - httpout-app
    - inputRefs:
      - infrastructure
      name: infra-logs
      outputRefs:
      - httpout-infra
    - inputRefs:
      - audit
      name: audit-logs
      outputRefs:
      - httpout-audit
    serviceAccount:
      name: ${SERVICE_ACCOUNT_NAME}
parameters:
- name: NAME
  value: instance
- name: NAMESPACE
  value: openshift-logging
- name: URL_APP
  value: https://fluentdserver.openshift-logging.svc:24224
- name: URL_AUDIT
  value: https://fluentdserver.openshift-logging.svc:24224
- name: URL_INFRA
  value: https://fluentdserver.openshift-logging.svc:24224
- name: SECRET_APP
- name: SECRET_AUDIT
- name: SECRET_INFRA
- name: SERVICE_ACCOUNT_NAME
  value: ""
- name: COLLECTOR
  value: "{}"
- name: TLS_CA_KEY
  value: "ca-bundle.crt"
`)

func testdataObservabilityOpenshiftIo_clusterlogforwarderHttpOutput85490YamlBytes() ([]byte, error) {
	return _testdataObservabilityOpenshiftIo_clusterlogforwarderHttpOutput85490Yaml, nil
}

func testdataObservabilityOpenshiftIo_clusterlogforwarderHttpOutput85490Yaml() (*asset, error) {
	bytes, err := testdataObservabilityOpenshiftIo_clusterlogforwarderHttpOutput85490YamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "testdata/observability.openshift.io_clusterlogforwarder/http-output-85490.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testdataObservabilityOpenshiftIo_clusterlogforwarderHttpOutputYaml = []byte(`kind: Template
apiVersion: template.openshift.io/v1
metadata:
  name: clusterlogforwarder-template
objects:
- apiVersion: observability.openshift.io/v1
  kind: ClusterLogForwarder
  metadata:
    name: ${NAME}
    namespace: ${NAMESPACE}
  spec:
    collector: ${{COLLECTOR}}
    managementState: Managed
    outputs:
    - http:
        headers:
          h1: v1
          h2: v2
        method: POST
        timeout:
        tuning: ${{TUNING}}
        url: ${URL}/logs/app
      name: httpout-app
      type: http
    - http:
        headers:
          h1: v1
          h2: v2
        method: POST
        timeout:
        tuning: ${{TUNING}}
        url: ${URL}/logs/infra
      name: httpout-infra
      type: http
    - http:
        headers:
          h1: v1
          h2: v2
        method: POST
        timeout:
        tuning: ${{TUNING}}
        url: ${URL}/logs/audit
      name: httpout-audit
      type: http
    pipelines:
    - inputRefs:
      - application
      name: app-logs
      outputRefs:
      - httpout-app
    - inputRefs:
      - infrastructure
      name: infra-logs
      outputRefs:
      - httpout-infra
    - inputRefs:
      - audit
      name: audit-logs
      outputRefs:
      - httpout-audit
    serviceAccount:
      name: ${SERVICE_ACCOUNT_NAME}
parameters:
- name: NAME
  value: instance
- name: NAMESPACE
  value: openshift-logging
- name: URL
  value: http://fluentdserver.openshift-logging.svc:24224
- name: SERVICE_ACCOUNT_NAME
  value: ""
- name: TUNING
  value: "{}"
- name: COLLECTOR
  value: "{}"
`)

func testdataObservabilityOpenshiftIo_clusterlogforwarderHttpOutputYamlBytes() ([]byte, error) {
	return _testdataObservabilityOpenshiftIo_clusterlogforwarderHttpOutputYaml, nil
}

func testdataObservabilityOpenshiftIo_clusterlogforwarderHttpOutputYaml() (*asset, error) {
	bytes, err := testdataObservabilityOpenshiftIo_clusterlogforwarderHttpOutputYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "testdata/observability.openshift.io_clusterlogforwarder/http-output.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testdataObservabilityOpenshiftIo_clusterlogforwarderHttps61567Yaml = []byte(`kind: Template
apiVersion: template.openshift.io/v1
metadata:
  name: clusterlogforwarder-template
objects:
- apiVersion: observability.openshift.io/v1
  kind: ClusterLogForwarder
  metadata:
    name: ${NAME}
    namespace: ${NAMESPACE}
  spec:
    outputs:
    - name: httpout-app
      http:
        headers:
          h1: v1
          h2: v2
        method: POST
        url: ${URL}/logs/app
      tls:
        ca:
          key: ca-bundle.crt
          secretName: ${SECRET_NAME}
      type: http
    pipelines:
    - inputRefs:
      - application
      name: app-logs
      outputRefs:
      - httpout-app
    serviceAccount:
      name: ${SERVICE_ACCOUNT_NAME}
parameters:
- name: NAME
  value: instance
- name: NAMESPACE
  value: openshift-logging
- name: URL
  value: https://fluentdserver.openshift-logging.svc:24224
- name: SECRET_NAME
  value: to-fluentdserver
- name: SERVICE_ACCOUNT_NAME
  value: ""
`)

func testdataObservabilityOpenshiftIo_clusterlogforwarderHttps61567YamlBytes() ([]byte, error) {
	return _testdataObservabilityOpenshiftIo_clusterlogforwarderHttps61567Yaml, nil
}

func testdataObservabilityOpenshiftIo_clusterlogforwarderHttps61567Yaml() (*asset, error) {
	bytes, err := testdataObservabilityOpenshiftIo_clusterlogforwarderHttps61567YamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "testdata/observability.openshift.io_clusterlogforwarder/https-61567.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testdataObservabilityOpenshiftIo_clusterlogforwarderHttpsOutputCaYaml = []byte(`kind: Template
apiVersion: template.openshift.io/v1
metadata:
  name: clusterlogforwarder-template
objects:
- apiVersion: observability.openshift.io/v1
  kind: ClusterLogForwarder
  metadata:
    name: ${NAME}
    namespace: ${NAMESPACE}
  spec:
    collector: ${{COLLECTOR}}
    outputs:
    - name: httpout-app
      type: http
      http:
        headers:
          h1: v1
          h2: v2
        method: POST
        url: ${URL}/logs/app
      tls:
        ca:
          key: ca-bundle.crt
          secretName: ${SECRET_NAME}
    - name: httpout-infra
      type: http
      http:
        headers:
          h1: v1
          h2: v2
        method: POST
        url: ${URL}/logs/infra
      tls:
        ca:
          key: ca-bundle.crt
          secretName: ${SECRET_NAME}
    - name: httpout-audit
      type: http
      http:
        headers:
          h1: v1
          h2: v2
        method: POST
        url: ${URL}/logs/audit
      tls:
        ca:
          key: ca-bundle.crt
          secretName: ${SECRET_NAME}
    pipelines:
    - inputRefs:
      - application
      name: app-logs
      outputRefs:
      - httpout-app
    - inputRefs:
      - infrastructure
      name: infra-logs
      outputRefs:
      - httpout-infra
    - inputRefs:
      - audit
      name: audit-logs
      outputRefs:
      - httpout-audit
    serviceAccount:
      name: ${SERVICE_ACCOUNT_NAME}
parameters:
- name: NAME
  value: instance
- name: NAMESPACE
  value: openshift-logging
- name: URL
  value: https://fluentdserver.openshift-logging.svc:24224
- name: SECRET_NAME
  value: to-fluentdserver
- name: SERVICE_ACCOUNT_NAME
  value: ""
- name: COLLECTOR
  value: "{}"
`)

func testdataObservabilityOpenshiftIo_clusterlogforwarderHttpsOutputCaYamlBytes() ([]byte, error) {
	return _testdataObservabilityOpenshiftIo_clusterlogforwarderHttpsOutputCaYaml, nil
}

func testdataObservabilityOpenshiftIo_clusterlogforwarderHttpsOutputCaYaml() (*asset, error) {
	bytes, err := testdataObservabilityOpenshiftIo_clusterlogforwarderHttpsOutputCaYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "testdata/observability.openshift.io_clusterlogforwarder/https-output-ca.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testdataObservabilityOpenshiftIo_clusterlogforwarderHttpsOutputMtlsYaml = []byte(`kind: Template
apiVersion: template.openshift.io/v1
metadata:
  name: clusterlogforwarder-template
objects:
- apiVersion: observability.openshift.io/v1
  kind: ClusterLogForwarder
  metadata:
    name: ${NAME}
    namespace: ${NAMESPACE}
  spec:
    collector: ${{COLLECTOR}}
    outputs:
    - name: httpout-app
      type: http
      http:
        headers:
          h1: v1
          h2: v2
        method: POST
        url: ${URL}/logs/app
      tls:
        ca:
          key: ca-bundle.crt
          secretName: ${SECRET_NAME}
        certificate:
          key: tls.crt
          secretName: ${SECRET_NAME}
        key:
          key: tls.key
          secretName: ${SECRET_NAME}
        keyPassphrase:
          key: passphrase
          secretName: ${SECRET_NAME}
    - name: httpout-infra
      type: http
      http:
        headers:
          h1: v1
          h2: v2
        method: POST
        url: ${URL}/logs/infra
      tls:
        ca:
          key: ca-bundle.crt
          secretName: ${SECRET_NAME}
        certificate:
          key: tls.crt
          secretName: ${SECRET_NAME}
        key:
          key: tls.key
          secretName: ${SECRET_NAME}
        keyPassphrase:
          key: passphrase
          secretName: ${SECRET_NAME}
    - name: httpout-audit
      type: http
      http:
        headers:
          h1: v1
          h2: v2
        method: POST
        url: ${URL}/logs/audit
      tls:
        ca:
          key: ca-bundle.crt
          secretName: ${SECRET_NAME}
        certificate:
          key: tls.crt
          secretName: ${SECRET_NAME}
        key:
          key: tls.key
          secretName: ${SECRET_NAME}
        keyPassphrase:
          key: passphrase
          secretName: ${SECRET_NAME}
    pipelines:
    - inputRefs:
      - application
      name: app-logs
      outputRefs:
      - httpout-app
    - inputRefs:
      - infrastructure
      name: infra-logs
      outputRefs:
      - httpout-infra
    - inputRefs:
      - audit
      name: audit-logs
      outputRefs:
      - httpout-audit
    serviceAccount:
      name: ${SERVICE_ACCOUNT_NAME}
parameters:
- name: NAME
  value: instance
- name: NAMESPACE
  value: openshift-logging
- name: URL
  value: https://fluentdserver.openshift-logging.svc:24224
- name: SECRET_NAME
  value: to-fluentdserver
- name: SERVICE_ACCOUNT_NAME
  value: ""
- name: COLLECTOR
  value: "{}"
`)

func testdataObservabilityOpenshiftIo_clusterlogforwarderHttpsOutputMtlsYamlBytes() ([]byte, error) {
	return _testdataObservabilityOpenshiftIo_clusterlogforwarderHttpsOutputMtlsYaml, nil
}

func testdataObservabilityOpenshiftIo_clusterlogforwarderHttpsOutputMtlsYaml() (*asset, error) {
	bytes, err := testdataObservabilityOpenshiftIo_clusterlogforwarderHttpsOutputMtlsYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "testdata/observability.openshift.io_clusterlogforwarder/https-output-mtls.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testdataObservabilityOpenshiftIo_clusterlogforwarderHttpsOutputSkipverifyYaml = []byte(`kind: Template
apiVersion: template.openshift.io/v1
metadata:
  name: clusterlogforwarder-template
objects:
- apiVersion: observability.openshift.io/v1
  kind: ClusterLogForwarder
  metadata:
    name: ${NAME}
    namespace: ${NAMESPACE}
  spec:
    collector: ${{COLLECTOR}}
    outputs:
    - name: httpout-app
      type: http
      http:
        headers:
          h1: v1
          h2: v2
        method: POST
        url: ${URL}/logs/app
      tls:
        insecureSkipVerify: true
    - name: httpout-infra
      type: http
      http:
        headers:
          h1: v1
          h2: v2
        method: POST
        url: ${URL}/logs/infra
      tls:
        insecureSkipVerify: true
    - name: httpout-audit
      type: http
      http:
        headers:
          h1: v1
          h2: v2
        method: POST
        url: ${URL}/logs/audit
      tls:
        insecureSkipVerify: true
    pipelines:
    - inputRefs:
      - application
      name: app-logs
      outputRefs:
      - httpout-app
    - inputRefs:
      - infrastructure
      name: infra-logs
      outputRefs:
      - httpout-infra
    - inputRefs:
      - audit
      name: audit-logs
      outputRefs:
      - httpout-audit
    serviceAccount:
      name: ${SERVICE_ACCOUNT_NAME}
parameters:
- name: NAME
  value: instance
- name: NAMESPACE
  value: openshift-logging
- name: URL
  value: https://fluentdserver.openshift-logging.svc:24224
- name: SECRET_NAME
  value: to-fluentdserver
- name: SERVICE_ACCOUNT_NAME
  value: ""
- name: COLLECTOR
  value: "{}"
`)

func testdataObservabilityOpenshiftIo_clusterlogforwarderHttpsOutputSkipverifyYamlBytes() ([]byte, error) {
	return _testdataObservabilityOpenshiftIo_clusterlogforwarderHttpsOutputSkipverifyYaml, nil
}

func testdataObservabilityOpenshiftIo_clusterlogforwarderHttpsOutputSkipverifyYaml() (*asset, error) {
	bytes, err := testdataObservabilityOpenshiftIo_clusterlogforwarderHttpsOutputSkipverifyYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "testdata/observability.openshift.io_clusterlogforwarder/https-output-skipverify.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testdataObservabilityOpenshiftIo_clusterlogforwarderHttpserverToHttpoutputYaml = []byte(`kind: Template
apiVersion: template.openshift.io/v1
metadata:
  name: clusterlogforwarder-template
objects:
- apiVersion: observability.openshift.io/v1
  kind: ClusterLogForwarder
  metadata:
    name: ${NAME}
    namespace: ${NAMESPACE}
  spec:
    collector: ${{COLLECTOR}}
    inputs:
    - name: ${HTTPSERVER_NAME}
      receiver:
        type: http
        http:
          format: kubeAPIAudit
        port: 8443
      type: receiver
    outputs:
    - name: httpout-audit
      http:
        headers:
          h1: v1
          h2: v2
        method: POST
        url: ${URL}/logs/audit
      type: http
      tls:
        ca:
          key: ca-bundle.crt
          secretName: ${SECRET_NAME}
        certificate:
          key: tls.crt
          secretName: ${SECRET_NAME}
        key:
          key: tls.key
          secretName: ${SECRET_NAME}
        keyPassphrase:
          key: passphrase
          secretName: ${SECRET_NAME}
    pipelines:
    - inputRefs:
      - ${HTTPSERVER_NAME}
      name: audit-logs
      outputRefs:
      - httpout-audit
    serviceAccount:
      name: ${SERVICE_ACCOUNT_NAME}
parameters:
- name: NAME
  value: instance
- name: NAMESPACE
  value: openshift-logging
- name: URL
  value: http://fluentdserver.openshift-logging.svc:24224
- name: SERVICE_ACCOUNT_NAME
  value: ""
- name: HTTPSERVER_NAME
  value: "httpserver"
- name: SECRET_NAME
  value: ""
- name: COLLECTOR
  value: "{}"
`)

func testdataObservabilityOpenshiftIo_clusterlogforwarderHttpserverToHttpoutputYamlBytes() ([]byte, error) {
	return _testdataObservabilityOpenshiftIo_clusterlogforwarderHttpserverToHttpoutputYaml, nil
}

func testdataObservabilityOpenshiftIo_clusterlogforwarderHttpserverToHttpoutputYaml() (*asset, error) {
	bytes, err := testdataObservabilityOpenshiftIo_clusterlogforwarderHttpserverToHttpoutputYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "testdata/observability.openshift.io_clusterlogforwarder/httpserver-to-httpoutput.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testdataObservabilityOpenshiftIo_clusterlogforwarderHttpserverToSplunkYaml = []byte(`kind: Template
apiVersion: template.openshift.io/v1
metadata:
  name: splunk-clf-template
objects:
- apiVersion: observability.openshift.io/v1
  kind: ClusterLogForwarder
  metadata:
    name: ${NAME}
    namespace: ${NAMESPACE}
  spec:
    collector: ${{COLLECTOR}}
    managementState: Managed
    inputs:
    - name: httpserver1
      type: receiver
      receiver:
        http:
          format: kubeAPIAudit
        port: 8081
        type: http
    - name: httpserver2
      type: receiver
      receiver:
        http:
          format: kubeAPIAudit
        port: 8082
        type: http
      type: receiver
    - name: httpserver3
      type: receiver
      receiver:
        http:
          format: kubeAPIAudit
        port: 8083
        type: http
    outputs:
    - name: splunk-aosqe
      type: splunk
      splunk:
        authentication:
          token:
            key: hecToken
            secretName: ${SECRET_NAME}
        index: ${INDEX}
        tuning: ${{TUNING}}
        url: ${URL}
    pipelines:
    - name: forward-log-splunk
      inputRefs:
      - httpserver1
      - httpserver2
      - httpserver3
      outputRefs:
      - splunk-aosqe
    serviceAccount:
      name: ${SERVICE_ACCOUNT_NAME}
parameters:
- name: NAME
  value: "instance"
- name: SECRET_NAME
  value: "to-splunk-secret"
- name: NAMESPACE
  value: "openshift-logging"
- name: URL
  value: "https://splunk-default-service.splunk-aosqe.svc:8088"
  required: true
- name: SERVICE_ACCOUNT_NAME
  value: ""
- name: INDEX
  value: "main"
- name: TUNING
  value: "{}"
- name: COLLECTOR
  value: "{}"
`)

func testdataObservabilityOpenshiftIo_clusterlogforwarderHttpserverToSplunkYamlBytes() ([]byte, error) {
	return _testdataObservabilityOpenshiftIo_clusterlogforwarderHttpserverToSplunkYaml, nil
}

func testdataObservabilityOpenshiftIo_clusterlogforwarderHttpserverToSplunkYaml() (*asset, error) {
	bytes, err := testdataObservabilityOpenshiftIo_clusterlogforwarderHttpserverToSplunkYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "testdata/observability.openshift.io_clusterlogforwarder/httpserver-to-splunk.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testdataObservabilityOpenshiftIo_clusterlogforwarderLokiYaml = []byte(`kind: Template
apiVersion: template.openshift.io/v1
metadata:
  name: clusterlogforwarder-template
objects:
- apiVersion: observability.openshift.io/v1
  kind: ClusterLogForwarder
  metadata:
    name: ${NAME}
    namespace: ${NAMESPACE}
  spec:
    collector: ${{COLLECTOR}}
    managementState: Managed
    outputs:
    - name: loki-server
      type: loki
      loki:
        labelKeys: ${{LABEL_KEYS}}
        tenantKey: ${TENANT_KEY}
        tuning: ${{TUNING}}
        url: ${URL}
    pipelines:
    - name: forward-to-loki
      inputRefs: ${{INPUT_REFS}}
      outputRefs:
      - loki-server
    serviceAccount:
      name: ${SERVICE_ACCOUNT_NAME}
parameters:
- name: NAME
  value: "instance"
- name: NAMESPACE
  value: "openshift-logging"
- name: LABEL_KEYS
  value: "[]"
- name: URL
  value: ""
- name: SERVICE_ACCOUNT_NAME
  value: ""
- name: INPUT_REFS
  value: "[\"infrastructure\", \"audit\", \"application\"]"
- name: TENANT_KEY
  value: "{.log_type||\"none-typed-logs\"}"
- name: TUNING
  value: "{}"
- name: COLLECTOR
  value: "{}"
`)

func testdataObservabilityOpenshiftIo_clusterlogforwarderLokiYamlBytes() ([]byte, error) {
	return _testdataObservabilityOpenshiftIo_clusterlogforwarderLokiYaml, nil
}

func testdataObservabilityOpenshiftIo_clusterlogforwarderLokiYaml() (*asset, error) {
	bytes, err := testdataObservabilityOpenshiftIo_clusterlogforwarderLokiYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "testdata/observability.openshift.io_clusterlogforwarder/loki.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testdataObservabilityOpenshiftIo_clusterlogforwarderLokistackWithLabelkeysYaml = []byte(`kind: Template
apiVersion: template.openshift.io/v1
metadata:
  name: clusterlogforwarder-template
objects:
- apiVersion: observability.openshift.io/v1
  kind: ClusterLogForwarder
  metadata:
    name: ${NAME}
    namespace: ${NAMESPACE}
  spec:
    collector: ${{COLLECTOR}}
    managementState: ${MANAGEMENT_STATE}
    outputs:
    - name: lokistack
      type: lokiStack
      lokiStack:
        authentication:
          token:
            from: serviceAccount
        labelKeys:
          application:
            ignoreGlobal: ${{IGNORE_GLOBAL_APP}}
            labelKeys: ${{APP_LABELKEYS}}
          audit:
            ignoreGlobal: ${{IGNORE_GLOBAL_AUDIT}}
            labelKeys: ${{AUDIT_LABELKEYS}}
          infrastructure:
            ignoreGlobal: ${{IGNORE_GLOBAL_INFRA}}
            labelKeys: ${{INFRA_LABELKEYS}}
          global: ${{GLOBAL_LABELKEYS}}
        target:
          name: ${LOKISTACK_NAME}
          namespace: ${LOKISTACK_NAMESPACE}
        tuning: ${{TUNING}}
      tls:
        ca:
          configMapName: openshift-service-ca.crt
          key: service-ca.crt
    pipelines:
    - name: forward-to-lokistack
      inputRefs: ${{INPUT_REFS}}
      outputRefs:
      - lokistack
    serviceAccount:
      name: ${SERVICE_ACCOUNT_NAME}
parameters:
- name: NAME
  value: "instance"
- name: NAMESPACE
  value: "openshift-logging"
- name: INPUT_REFS
  value: "[\"infrastructure\", \"audit\", \"application\"]"
- name: SERVICE_ACCOUNT_NAME
  value: ""
- name: LOKISTACK_NAME
  value: "logging-loki"
- name: LOKISTACK_NAMESPACE
  value: "openshift-logging"
- name: MANAGEMENT_STATE
  value: "Managed"
- name: TUNING
  value: "{}"
- name: IGNORE_GLOBAL_APP
  value: "false"
- name: APP_LABELKEYS
  value: "[]"
- name: IGNORE_GLOBAL_AUDIT
  value: "false"
- name: AUDIT_LABELKEYS
  value: "[]"
- name: IGNORE_GLOBAL_INFRA
  value: "false"
- name: INFRA_LABELKEYS
  value: "[]"
- name: GLOBAL_LABELKEYS
  value: "[\"log_type\",\"kubernetes.container_name\",\"kubernetes.namespace_name\",\"kubernetes.pod_name\"]"
- name: COLLECTOR
  value: "{}"
`)

func testdataObservabilityOpenshiftIo_clusterlogforwarderLokistackWithLabelkeysYamlBytes() ([]byte, error) {
	return _testdataObservabilityOpenshiftIo_clusterlogforwarderLokistackWithLabelkeysYaml, nil
}

func testdataObservabilityOpenshiftIo_clusterlogforwarderLokistackWithLabelkeysYaml() (*asset, error) {
	bytes, err := testdataObservabilityOpenshiftIo_clusterlogforwarderLokistackWithLabelkeysYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "testdata/observability.openshift.io_clusterlogforwarder/lokistack-with-labelkeys.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testdataObservabilityOpenshiftIo_clusterlogforwarderLokistackYaml = []byte(`kind: Template
apiVersion: template.openshift.io/v1
metadata:
  name: clusterlogforwarder-template
objects:
- apiVersion: observability.openshift.io/v1
  kind: ClusterLogForwarder
  metadata:
    name: ${NAME}
    namespace: ${NAMESPACE}
    annotations:
      observability.openshift.io/tech-preview-otlp-output: "enabled"
  spec:
    collector: ${{COLLECTOR}}
    managementState: ${MANAGEMENT_STATE}
    outputs:
    - name: lokistack
      type: lokiStack
      lokiStack:
        authentication:
          token:
            from: serviceAccount
        dataModel: ${DATAMODEL}
        target:
          name: ${LOKISTACK_NAME}
          namespace: ${LOKISTACK_NAMESPACE}
        tuning: ${{TUNING}}
      tls:
        ca:
          key: ca-bundle.crt
          secretName: ${SECRET_NAME}
    pipelines:
    - name: forward-to-lokistack
      inputRefs: ${{INPUT_REFS}}
      outputRefs:
      - lokistack
    serviceAccount:
      name: ${SERVICE_ACCOUNT_NAME}
parameters:
- name: NAME
  value: "instance"
- name: NAMESPACE
  value: "openshift-logging"
- name: INPUT_REFS
  value: "[\"infrastructure\", \"audit\", \"application\"]"
- name: SERVICE_ACCOUNT_NAME
  value: ""
- name: LOKISTACK_NAME
  value: "logging-loki"
- name: LOKISTACK_NAMESPACE
  value: "openshift-logging"
- name: MANAGEMENT_STATE
  value: "Managed"
- name: TUNING
  value: "{}"
- name: SECRET_NAME
  value: ""
- name: DATAMODEL
  value: "Viaq"
- name: COLLECTOR
  value: "{}"
`)

func testdataObservabilityOpenshiftIo_clusterlogforwarderLokistackYamlBytes() ([]byte, error) {
	return _testdataObservabilityOpenshiftIo_clusterlogforwarderLokistackYaml, nil
}

func testdataObservabilityOpenshiftIo_clusterlogforwarderLokistackYaml() (*asset, error) {
	bytes, err := testdataObservabilityOpenshiftIo_clusterlogforwarderLokistackYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "testdata/observability.openshift.io_clusterlogforwarder/lokistack.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testdataObservabilityOpenshiftIo_clusterlogforwarderLokistack_gateway_https_secretYaml = []byte(`kind: Template
apiVersion: template.openshift.io/v1
metadata:
  name: clusterlogforwarder-template
objects:
- apiVersion: observability.openshift.io/v1
  kind: ClusterLogForwarder
  metadata:
    name: ${NAME}
    namespace: ${NAMESPACE}
  spec:
    collector: ${{COLLECTOR}}
    outputs:
    - name: loki-app
      type: loki
      loki:
        authentication:
          token:
            from: serviceAccount
        url: https://${GATEWAY_SVC}/api/logs/v1/application
      tls:
        ca:
          key: ca-bundle.crt
          secretName: ${SECRET_NAME}
    - name: loki-infra
      type: loki
      loki:
        authentication:
          token:
            from: serviceAccount
        url: https://${GATEWAY_SVC}/api/logs/v1/infrastructure
      tls:
        ca:
          key: ca-bundle.crt
          secretName: ${SECRET_NAME}
    - name: loki-audit
      type: loki
      loki:
        authentication:
          token:
            from: serviceAccount
        url: https://${GATEWAY_SVC}/api/logs/v1/audit
      tls:
        ca:
          key: ca-bundle.crt
          secretName: ${SECRET_NAME}
    pipelines:
    - name: infro-to-loki
      inputRefs:
      - infrastructure
      outputRefs:
      - loki-infra
    - name: app-to-loki
      inputRefs:
      - application
      outputRefs:
      - loki-app
    - name: audit-to-loki
      inputRefs:
      - audit
      outputRefs:
      - loki-audit
    serviceAccount:
      name: ${SERVICE_ACCOUNT_NAME}
parameters:
- name: NAME
  value: "instance"
- name: NAMESPACE
  value: "openshift-logging"
- name: GATEWAY_SVC
- name: SECRET_NAME
  value: "lokistack-gateway-bearer-token"
- name: SERVICE_ACCOUNT_NAME
  value: ""
- name: COLLECTOR
  value: "{}"
`)

func testdataObservabilityOpenshiftIo_clusterlogforwarderLokistack_gateway_https_secretYamlBytes() ([]byte, error) {
	return _testdataObservabilityOpenshiftIo_clusterlogforwarderLokistack_gateway_https_secretYaml, nil
}

func testdataObservabilityOpenshiftIo_clusterlogforwarderLokistack_gateway_https_secretYaml() (*asset, error) {
	bytes, err := testdataObservabilityOpenshiftIo_clusterlogforwarderLokistack_gateway_https_secretYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "testdata/observability.openshift.io_clusterlogforwarder/lokistack_gateway_https_secret.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testdataObservabilityOpenshiftIo_clusterlogforwarderOtlpLokistackYaml = []byte(`kind: Template
apiVersion: template.openshift.io/v1
metadata:
  name: clusterlogforwarder-template
objects:
- apiVersion: observability.openshift.io/v1
  kind: ClusterLogForwarder
  metadata:
    annotations:
      observability.openshift.io/tech-preview-otlp-output: "enabled"
    name: ${NAME}
    namespace: ${NAMESPACE}
  spec:
    collector: ${{COLLECTOR}}
    managementState: ${MANAGEMENT_STATE}
    filters:
      - name: detect-multiline-exception
        type: detectMultilineException
    outputs:
    - name: apps
      type: otlp
      otlp:
        authentication:
          token:
            from: serviceAccount
        url: ${URL}/api/logs/v1/application/otlp/v1/logs
      tls:
        ca:
          key: ca-bundle.crt
          secretName: ${SECRET_NAME}
    - name: audit
      type: otlp
      otlp:
        authentication:
          token:
            from: serviceAccount
        url: ${URL}/api/logs/v1/audit/otlp/v1/logs
      tls:
        ca:
          key: ca-bundle.crt
          secretName: ${SECRET_NAME}
    - name: infra
      type: otlp
      otlp:
        authentication:
          token:
            from: serviceAccount
        url: ${URL}/api/logs/v1/infrastructure/otlp/v1/logs
      tls:
        ca:
          key: service-ca.crt
          configMapName: ${CM_NAME}
    pipelines:
    - inputRefs:
      - application
      name: apps
      filterRefs:
      - detect-multiline-exception
      outputRefs:
      - apps
    - inputRefs:
      - audit
      name: audit
      outputRefs:
      - audit
    - inputRefs:
      - infrastructure
      filterRefs:
      - detect-multiline-exception
      name: infra
      outputRefs:
      - infra
    serviceAccount:
      name: ${SERVICE_ACCOUNT_NAME}
parameters:
- name: NAME
  value: "instance"
- name: NAMESPACE
  value: "openshift-logging"
- name: SERVICE_ACCOUNT_NAME
  value: ""
- name: MANAGEMENT_STATE
  value: "Managed"
- name: SECRET_NAME
  value: ""
- name: CM_NAME
  value: "openshift-service-ca.crt"
- name: URL
  value: "https://logging-loki-gateway-http.openshift-logging.svc.cluster.local:8080"
- name: COLLECTOR
  value: "{}"
`)

func testdataObservabilityOpenshiftIo_clusterlogforwarderOtlpLokistackYamlBytes() ([]byte, error) {
	return _testdataObservabilityOpenshiftIo_clusterlogforwarderOtlpLokistackYaml, nil
}

func testdataObservabilityOpenshiftIo_clusterlogforwarderOtlpLokistackYaml() (*asset, error) {
	bytes, err := testdataObservabilityOpenshiftIo_clusterlogforwarderOtlpLokistackYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "testdata/observability.openshift.io_clusterlogforwarder/otlp-lokistack.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testdataObservabilityOpenshiftIo_clusterlogforwarderOtlpYaml = []byte(`kind: Template
apiVersion: template.openshift.io/v1
metadata:
  name: clusterlogforwarder-template
objects:
- apiVersion: observability.openshift.io/v1
  kind: ClusterLogForwarder
  metadata:
    annotations:
      observability.openshift.io/tech-preview-otlp-output: "enabled"
    name: ${NAME}
    namespace: ${NAMESPACE}
  spec:
    collector: ${{COLLECTOR}}
    managementState: ${MANAGEMENT_STATE}
    outputs:
    - name: otlp
      type: otlp
      otlp:
        tuning:
          compression: gzip
          deliveryMode: AtLeastOnce
          maxRetryDuration: 20
          maxWrite: 10M
          minRetryDuration: 5
        url: ${URL}/v1/logs
    pipelines:
    - inputRefs: ${{INPUT_REFS}}
      name: otlp-logs
      outputRefs:
      - otlp
    serviceAccount:
      name: ${SERVICE_ACCOUNT_NAME}
parameters:
- name: NAME
  value: "instance"
- name: NAMESPACE
  value: "openshift-logging"
- name: SERVICE_ACCOUNT_NAME
  value: ""
- name: MANAGEMENT_STATE
  value: "Managed"
- name: URL
  value: "http://otel-collector.openshift-opentelemetry-operator.svc:4318"
- name: COLLECTOR
  value: "{}"
- name: INPUT_REFS
  value: "[\"infrastructure\", \"audit\", \"application\"]"
`)

func testdataObservabilityOpenshiftIo_clusterlogforwarderOtlpYamlBytes() ([]byte, error) {
	return _testdataObservabilityOpenshiftIo_clusterlogforwarderOtlpYaml, nil
}

func testdataObservabilityOpenshiftIo_clusterlogforwarderOtlpYaml() (*asset, error) {
	bytes, err := testdataObservabilityOpenshiftIo_clusterlogforwarderOtlpYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "testdata/observability.openshift.io_clusterlogforwarder/otlp.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testdataObservabilityOpenshiftIo_clusterlogforwarderRsyslogMtlsYaml = []byte(`kind: Template
apiVersion: template.openshift.io/v1
metadata:
  name: clusterlogforwarder-template
objects:
- apiVersion: observability.openshift.io/v1
  kind: ClusterLogForwarder
  metadata:
    name: ${NAME}
    namespace: ${NAMESPACE}
  spec:
    collector: ${{COLLECTOR}}
    outputs:
    - name: external-syslog
      type: syslog
      syslog:
        url: ${URL}
        rfc: ${RFC}
      tls:
        ca:
          key: ca-bundle.crt
          secretName: ${SECRET_NAME}
        key:
          key: tls.key
          secretName: ${SECRET_NAME}
        certificate:
          key: tls.crt
          secretName: ${SECRET_NAME}
        keyPassphrase:
          key: passphrase
          secretName: ${SECRET_NAME}
    pipelines:
    - name: forward-to-external-syslog
      inputRefs: ${{INPUTREFS}}
      outputRefs:
      - external-syslog
    serviceAccount:
      name: ${SERVICE_ACCOUNT_NAME}
parameters:
- name: NAME
  value: "instance"
- name: NAMESPACE
  value: "openshift-logging"
- name: SECRET_NAME
  value: pipelinesecret
- name: URL
  value: tls://rsyslogserver.openshift-logging.svc:6514
- name: INPUTREFS
  value: "[\"infrastructure\", \"audit\", \"application\"]"
- name: SERVICE_ACCOUNT_NAME
  value: ""
- name: RFC
  value: RFC5424
- name: COLLECTOR
  value: "{}"
`)

func testdataObservabilityOpenshiftIo_clusterlogforwarderRsyslogMtlsYamlBytes() ([]byte, error) {
	return _testdataObservabilityOpenshiftIo_clusterlogforwarderRsyslogMtlsYaml, nil
}

func testdataObservabilityOpenshiftIo_clusterlogforwarderRsyslogMtlsYaml() (*asset, error) {
	bytes, err := testdataObservabilityOpenshiftIo_clusterlogforwarderRsyslogMtlsYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "testdata/observability.openshift.io_clusterlogforwarder/rsyslog-mtls.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testdataObservabilityOpenshiftIo_clusterlogforwarderRsyslogServerauthYaml = []byte(`kind: Template
apiVersion: template.openshift.io/v1
metadata:
  name: clusterlogforwarder-template
objects:
- apiVersion: observability.openshift.io/v1
  kind: ClusterLogForwarder
  metadata:
    name: ${NAME}
    namespace: ${NAMESPACE}
  spec:
    collector: ${{COLLECTOR}}
    inputs:
    - application:
        includes:
        - namespace: ${NAMESPACE_PATTERN}
      name: app-input-namespace
      type: application
    outputs:
    - name: external-syslog
      type: syslog
      syslog:
        facility: local0
        rfc: ${RFC}
        severity: informational
        url: ${URL}
      tls:
        ca:
          key: ca-bundle.crt
          secretName: ${SECRET_NAME}
    pipelines:
    - name: forward-to-external-syslog
      inputRefs: ${{INPUTREFS}}
      outputRefs:
        - external-syslog
    serviceAccount:
      name: ${SERVICE_ACCOUNT_NAME}
parameters:
- name: NAME
  value: "instance"
- name: NAMESPACE
  value: "openshift-logging"
- name: NAMESPACE_PATTERN
  value: "e2e*"
- name: SECRET_NAME
  value: pipelinesecret
- name: URL
  value: tls://rsyslogserver.openshift-logging.svc:6514
- name: RFC
  value: RFC5424
- name: INPUTREFS
  value: "[\"infrastructure\", \"audit\", \"app-input-namespace\"]"
- name: SERVICE_ACCOUNT_NAME
  value: ""
- name: COLLECTOR
  value: "{}"
`)

func testdataObservabilityOpenshiftIo_clusterlogforwarderRsyslogServerauthYamlBytes() ([]byte, error) {
	return _testdataObservabilityOpenshiftIo_clusterlogforwarderRsyslogServerauthYaml, nil
}

func testdataObservabilityOpenshiftIo_clusterlogforwarderRsyslogServerauthYaml() (*asset, error) {
	bytes, err := testdataObservabilityOpenshiftIo_clusterlogforwarderRsyslogServerauthYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "testdata/observability.openshift.io_clusterlogforwarder/rsyslog-serverAuth.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testdataObservabilityOpenshiftIo_clusterlogforwarderSplunkMtlsPassphraseYaml = []byte(`kind: Template
apiVersion: template.openshift.io/v1
metadata:
  name: splunk-clf-template
objects:
- apiVersion: observability.openshift.io/v1
  kind: ClusterLogForwarder
  metadata:
    name: ${NAME}
    namespace: ${NAMESPACE}
  spec:
    collector: ${{COLLECTOR}}
    managementState: Managed
    outputs:
    - name: splunk-aosqe
      type: splunk
      splunk:
        authentication:
          token:
            key: hecToken
            secretName: ${SECRET_NAME}
        index: ${INDEX}
        tuning: ${{TUNING}}
        url: ${URL}
      tls:
        ca:
          key: ca-bundle.crt
          secretName: ${SECRET_NAME}
        key:
          key: tls.key
          secretName: ${SECRET_NAME}
        certificate:
          key: tls.crt
          secretName: ${SECRET_NAME}
        keyPassphrase:
          key: passphrase
          secretName: ${SECRET_NAME}
    pipelines:
    - name: forward-log-splunk
      inputRefs: ${{INPUT_REFS}}
      outputRefs:
      - splunk-aosqe
    serviceAccount:
      name: ${SERVICE_ACCOUNT_NAME}
parameters:
- name: NAME
  value: "instance"
- name: SECRET_NAME
  value: "to-splunk-secret"
- name: NAMESPACE
  value: "openshift-logging"
- name: URL
  value: "https://splunk-default-service.splunk-aosqe.svc:8088"
  required: true
- name: INPUT_REFS
  value: "[\"infrastructure\", \"audit\", \"application\"]"
- name: SERVICE_ACCOUNT_NAME
  value: ""
- name: INDEX
  value: "main"
- name: TUNING
  value: "{}"
- name: COLLECTOR
  value: "{}"
`)

func testdataObservabilityOpenshiftIo_clusterlogforwarderSplunkMtlsPassphraseYamlBytes() ([]byte, error) {
	return _testdataObservabilityOpenshiftIo_clusterlogforwarderSplunkMtlsPassphraseYaml, nil
}

func testdataObservabilityOpenshiftIo_clusterlogforwarderSplunkMtlsPassphraseYaml() (*asset, error) {
	bytes, err := testdataObservabilityOpenshiftIo_clusterlogforwarderSplunkMtlsPassphraseYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "testdata/observability.openshift.io_clusterlogforwarder/splunk-mtls-passphrase.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testdataObservabilityOpenshiftIo_clusterlogforwarderSplunkMtlsYaml = []byte(`kind: Template
apiVersion: template.openshift.io/v1
metadata:
  name: splunk-clf-template
objects:
- apiVersion: observability.openshift.io/v1
  kind: ClusterLogForwarder
  metadata:
    name: ${NAME}
    namespace: ${NAMESPACE}
  spec:
    collector: ${{COLLECTOR}}
    managementState: Managed
    outputs:
    - name: splunk-aosqe
      type: splunk
      splunk:
        authentication:
          token:
            key: hecToken
            secretName: ${SECRET_NAME}
        index: ${INDEX}
        tuning: ${{TUNING}}
        url: ${URL}
      tls:
        insecureSkipVerify: ${{SKIPVERIFY}}
        ca:
          key: ca-bundle.crt
          secretName: ${SECRET_NAME}
        key:
          key: tls.key
          secretName: ${SECRET_NAME}
        certificate:
          key: tls.crt
          secretName: ${SECRET_NAME}
    pipelines:
    - name: forward-log-splunk
      inputRefs: ${{INPUT_REFS}}
      outputRefs:
      - splunk-aosqe
    serviceAccount:
      name: ${SERVICE_ACCOUNT_NAME}
parameters:
- name: NAME
  value: "instance"
- name: SECRET_NAME
  value: "to-splunk-secret"
- name: NAMESPACE
  value: "openshift-logging"
- name: URL
  value: "https://splunk-default-service.splunk-aosqe.svc:8088"
  required: true
- name: INPUT_REFS
  value: "[\"infrastructure\", \"audit\", \"application\"]"
- name: SERVICE_ACCOUNT_NAME
  value: ""
- name: INDEX
  value: "main"
- name: TUNING
  value: "{}"
- name: SKIPVERIFY
  value: "false"
- name: COLLECTOR
  value: "{}"
`)

func testdataObservabilityOpenshiftIo_clusterlogforwarderSplunkMtlsYamlBytes() ([]byte, error) {
	return _testdataObservabilityOpenshiftIo_clusterlogforwarderSplunkMtlsYaml, nil
}

func testdataObservabilityOpenshiftIo_clusterlogforwarderSplunkMtlsYaml() (*asset, error) {
	bytes, err := testdataObservabilityOpenshiftIo_clusterlogforwarderSplunkMtlsYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "testdata/observability.openshift.io_clusterlogforwarder/splunk-mtls.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testdataObservabilityOpenshiftIo_clusterlogforwarderSplunkServeronlyYaml = []byte(`kind: Template
apiVersion: template.openshift.io/v1
metadata:
  name: splunk-clf-template
objects:
- apiVersion: observability.openshift.io/v1
  kind: ClusterLogForwarder
  metadata:
    name: ${NAME}
    namespace: ${NAMESPACE}
  spec:
    collector: ${{COLLECTOR}}
    managementState: Managed
    outputs:
    - name: splunk-aosqe
      type: splunk
      splunk:
        authentication:
          token:
            key: hecToken
            secretName: ${SECRET_NAME}
        index: ${INDEX}
        tuning: ${{TUNING}}
        url: ${URL}
      tls:
        insecureSkipVerify: ${{SKIP_VERIFY}}
        ca:
          key: ca-bundle.crt
          secretName: ${SECRET_NAME}
    pipelines:
    - name: forward-log-splunk
      inputRefs: ${{INPUT_REFS}}
      outputRefs:
      - splunk-aosqe
    serviceAccount:
      name: ${SERVICE_ACCOUNT_NAME}
parameters:
- name: NAME
  value: "instance"
- name: SECRET_NAME
  value: "to-splunk-secret"
- name: NAMESPACE
  value: "openshift-logging"
- name: URL
  value: "https://splunk-default-service.splunk-aosqe.svc:8088"
  required: true
- name: INPUT_REFS
  value: "[\"infrastructure\", \"audit\", \"application\"]"
- name: SERVICE_ACCOUNT_NAME
  value: ""
- name: SKIP_VERIFY
  value: "false"
- name: INDEX
  value: "main"
- name: TUNING
  value: "{}"
- name: COLLECTOR
  value: "{}"
`)

func testdataObservabilityOpenshiftIo_clusterlogforwarderSplunkServeronlyYamlBytes() ([]byte, error) {
	return _testdataObservabilityOpenshiftIo_clusterlogforwarderSplunkServeronlyYaml, nil
}

func testdataObservabilityOpenshiftIo_clusterlogforwarderSplunkServeronlyYaml() (*asset, error) {
	bytes, err := testdataObservabilityOpenshiftIo_clusterlogforwarderSplunkServeronlyYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "testdata/observability.openshift.io_clusterlogforwarder/splunk-serveronly.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testdataObservabilityOpenshiftIo_clusterlogforwarderSplunkYaml = []byte(`kind: Template
apiVersion: template.openshift.io/v1
metadata:
  name: splunk-clf-template
objects:
- apiVersion: observability.openshift.io/v1
  kind: ClusterLogForwarder
  metadata:
    name: ${NAME}
    namespace: ${NAMESPACE}
  spec:
    collector: ${{COLLECTOR}}
    managementState: Managed
    outputs:
    - name: splunk-aosqe
      type: splunk
      splunk:
        authentication:
          token:
            key: hecToken
            secretName: ${SECRET_NAME}
        index: ${INDEX}
        tuning: ${{TUNING}}
        url: ${URL}
    pipelines:
    - name: forward-log-splunk
      inputRefs: ${{INPUT_REFS}}
      outputRefs:
      - splunk-aosqe
    serviceAccount:
      name: ${SERVICE_ACCOUNT_NAME}
parameters:
- name: NAME
  value: "instance"
- name: SECRET_NAME
  value: "to-splunk-secret"
- name: NAMESPACE
  value: "openshift-logging"
- name: URL
  value: "https://splunk-default-service.splunk-aosqe.svc:8088"
  required: true
- name: INPUT_REFS
  value: "[\"infrastructure\", \"audit\", \"application\"]"
- name: SERVICE_ACCOUNT_NAME
  value: ""
- name: INDEX
  value: "main"
- name: TUNING
  value: "{}"
- name: COLLECTOR
  value: "{}"
`)

func testdataObservabilityOpenshiftIo_clusterlogforwarderSplunkYamlBytes() ([]byte, error) {
	return _testdataObservabilityOpenshiftIo_clusterlogforwarderSplunkYaml, nil
}

func testdataObservabilityOpenshiftIo_clusterlogforwarderSplunkYaml() (*asset, error) {
	bytes, err := testdataObservabilityOpenshiftIo_clusterlogforwarderSplunkYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "testdata/observability.openshift.io_clusterlogforwarder/splunk.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testdataObservabilityOpenshiftIo_clusterlogforwarderSyslog75317Yaml = []byte(`kind: Template
apiversion: template.openshift.io/v1
metadata:
  name: clusterlogforwarder-template
objects:
- apiVersion: observability.openshift.io/v1
  kind: ClusterLogForwarder
  metadata:
    name: ${NAME}
    namespace: ${NAMESPACE}
    annotations:
      observability.openshift.io/log-level: ${LOG_LEVEL}
  spec:
    inputs:
    - application:
        includes:
        - namespace: ${NAMESPACE_PATTERN}
      name: app-input-namespace
      type: application
    outputs:
    - name: syslog-app
      type: syslog
      syslog:
        appName: '{.kubernetes.container_name||.log_source||"container"}'
        facility: '1'
        procId: '{.kubernetes.pod_name||.log_source||"container"}'
        msgId: '{.openshift.sequence||"unknown"}'
        rfc: ${RFC}
        severity: informational
        url: ${URL}
    - name: syslog-infra
      type: syslog
      syslog:
        appName: '{.kubernetes.container_name||.log_source||"unknown"}'
        facility: 'local0'
        procId: '{.kubernetes.pod_name||.log_source||"unknown"}'
        msgId: '{._SYSTEMD_INVOCATION_ID||"unknown"}'
        rfc: ${RFC}
        severity: Notice
        url: ${URL}
    - name: syslog-audit
      type: syslog
      syslog:
        appName: '{.log_source||"audit"}'
        facility: '13'
        procId: '{.log_source||"audit"}'
        msgId: '{.openshift.sequence||"unknown"}'
        rfc: ${RFC}
        severity: Alert
        url: ${URL}
    pipelines:
    - inputRefs:
        - app-input-namespace
      name: pipe1
      outputRefs:
        - syslog-app
    - inputRefs:
        - infrastructure
      name: pipe2
      outputRefs:
        - syslog-infra
    - inputRefs:
        - audit
      name: pipe3
      outputRefs:
        - syslog-audit
    serviceAccount:
      name: ${SERVICE_ACCOUNT_NAME}
parameters:
- name: NAME
  value: "instance"
- name: NAMESPACE
  value: "openshift-logging"
- name: NAMESPACE_PATTERN
  value: "e2e*"
- name: URL
  value: "udp://rsyslogserver.openshift-logging.svc:514"
- name: SERVICE_ACCOUNT_NAME
  value: ""
- name: RFC
  value: RFC5424
- name: LOG_LEVEL
  value: "off"
`)

func testdataObservabilityOpenshiftIo_clusterlogforwarderSyslog75317YamlBytes() ([]byte, error) {
	return _testdataObservabilityOpenshiftIo_clusterlogforwarderSyslog75317Yaml, nil
}

func testdataObservabilityOpenshiftIo_clusterlogforwarderSyslog75317Yaml() (*asset, error) {
	bytes, err := testdataObservabilityOpenshiftIo_clusterlogforwarderSyslog75317YamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "testdata/observability.openshift.io_clusterlogforwarder/syslog-75317.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testdataObservabilityOpenshiftIo_clusterlogforwarderSyslog75431Yaml = []byte(`kind: Template
apiversion: template.openshift.io/v1
metadata:
  name: clusterlogforwarder-template
objects:
- apiVersion: observability.openshift.io/v1
  kind: ClusterLogForwarder
  metadata:
    name: ${NAME}
    namespace: ${NAMESPACE}
    annotations:
      observability.openshift.io/log-level: ${LOG_LEVEL}
  spec:
    inputs:
    - application:
        includes:
        - namespace: ${NAMESPACE_PATTERN}
      name: app-input-namespace
      type: application
    outputs:
    - name: syslog-out
      type: syslog
      syslog:
        enrichment: KubernetesMinimal
        rfc: ${RFC}
        url: ${URL}
    pipelines:
    - inputRefs:
        - app-input-namespace
      name: pipe1
      outputRefs:
        - syslog-out
    - inputRefs:
        - infrastructure
      name: pipe2
      outputRefs:
        - syslog-out
    - inputRefs:
        - audit
      name: pipe3
      outputRefs:
        - syslog-out
    serviceAccount:
      name: ${SERVICE_ACCOUNT_NAME}
parameters:
- name: NAME
  value: "instance"
- name: NAMESPACE
  value: "openshift-logging"
- name: NAMESPACE_PATTERN
  value: "e2e*"
- name: URL
  value: "udp://rsyslogserver.openshift-logging.svc:514"
- name: SERVICE_ACCOUNT_NAME
  value: ""
- name: RFC
  value: RFC5424
- name: LOG_LEVEL
  value: "off"
`)

func testdataObservabilityOpenshiftIo_clusterlogforwarderSyslog75431YamlBytes() ([]byte, error) {
	return _testdataObservabilityOpenshiftIo_clusterlogforwarderSyslog75431Yaml, nil
}

func testdataObservabilityOpenshiftIo_clusterlogforwarderSyslog75431Yaml() (*asset, error) {
	bytes, err := testdataObservabilityOpenshiftIo_clusterlogforwarderSyslog75431YamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "testdata/observability.openshift.io_clusterlogforwarder/syslog-75431.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testdataObservabilityOpenshiftIo_clusterlogforwarderSyslog81512Yaml = []byte(`kind: Template
apiversion: template.openshift.io/v1
metadata:
  name: clusterlogforwarder-template
objects:
- apiVersion: observability.openshift.io/v1
  kind: ClusterLogForwarder
  metadata:
    name: ${NAME}
    namespace: ${NAMESPACE}
    annotations:
      observability.openshift.io/log-level: ${LOG_LEVEL}
  spec:
    inputs:
    - application:
        includes:
        - namespace: ${NAMESPACE_PATTERN}
      name: app-input-namespace
      type: application
    outputs:
    - name: syslog-server
      type: syslog
      syslog:
        payloadKey: '{${PAYLOAD_KEY}}'
        rfc: ${RFC}
        url: ${URL}
      tls:
        ca:
          key: ca-bundle.crt
          secretName: ${SECRET_NAME}
    pipelines:
    - inputRefs:
        - app-input-namespace
      name: pipe1
      outputRefs:
        - syslog-server
    serviceAccount:
      name: ${SERVICE_ACCOUNT_NAME}
parameters:
- name: NAME
  value: "logcollector"
- name: NAMESPACE
  value: "openshift-logging"
- name: NAMESPACE_PATTERN
  value: "e2e*"
- name: URL
  value: "tls://rsyslogserver.openshift-logging.svc:6514"
- name: SERVICE_ACCOUNT_NAME
  value: ""
- name: SECRET_NAME
  value: ""
- name: RFC
  value: RFC5424
- name: PAYLOAD_KEY
  value: ".message"
- name: LOG_LEVEL
  value: "off"
`)

func testdataObservabilityOpenshiftIo_clusterlogforwarderSyslog81512YamlBytes() ([]byte, error) {
	return _testdataObservabilityOpenshiftIo_clusterlogforwarderSyslog81512Yaml, nil
}

func testdataObservabilityOpenshiftIo_clusterlogforwarderSyslog81512Yaml() (*asset, error) {
	bytes, err := testdataObservabilityOpenshiftIo_clusterlogforwarderSyslog81512YamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "testdata/observability.openshift.io_clusterlogforwarder/syslog-81512.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testdataObservabilityOpenshiftIo_clusterlogforwarderSyslogSelectedNsYaml = []byte(`kind: Template
apiversion: template.openshift.io/v1
metadata:
  name: clusterlogforwarder-template
objects:
- apiVersion: observability.openshift.io/v1
  kind: ClusterLogForwarder
  metadata:
    name: ${NAME}
    namespace: ${NAMESPACE}
  spec:
    inputs:
    - application:
        includes:
        - namespace: ${NAMESPACE_PATTERN}
      name: app-input-namespace
      type: application
    outputs:
    - name: rsyslog
      type: syslog
      syslog:
        rfc: ${RFC}
        url: ${URL}
    pipelines:
    - inputRefs: ${{INPUT_REFS}}
      name: forward-to-syslog
      outputRefs:
        - rsyslog
    serviceAccount:
      name: ${SERVICE_ACCOUNT_NAME}
parameters:
- name: NAME
  value: "instance"
- name: NAMESPACE
  value: "openshift-logging"
- name: URL
  value: "udp://rsyslogserver.openshift-logging.svc:514"
- name: SERVICE_ACCOUNT_NAME
  value: ""
- name: NAMESPACE_PATTERN
  value: ""
- name: INPUT_REFS
  value: "[\"infrastructure\", \"audit\", \"app-input-namespace\"]"
- name: RFC
  value: RFC5424
`)

func testdataObservabilityOpenshiftIo_clusterlogforwarderSyslogSelectedNsYamlBytes() ([]byte, error) {
	return _testdataObservabilityOpenshiftIo_clusterlogforwarderSyslogSelectedNsYaml, nil
}

func testdataObservabilityOpenshiftIo_clusterlogforwarderSyslogSelectedNsYaml() (*asset, error) {
	bytes, err := testdataObservabilityOpenshiftIo_clusterlogforwarderSyslogSelectedNsYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "testdata/observability.openshift.io_clusterlogforwarder/syslog-selected-ns.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testdataObservabilityOpenshiftIo_clusterlogforwarderSyslogYaml = []byte(`kind: Template
apiversion: template.openshift.io/v1
metadata:
  name: clusterlogforwarder-template
objects:
- apiVersion: observability.openshift.io/v1
  kind: ClusterLogForwarder
  metadata:
    name: ${NAME}
    namespace: ${NAMESPACE}
  spec:
    collector: ${{COLLECTOR}}
    outputs:
    - name: rsyslog
      type: syslog
      syslog:
        facility: ${FACILITY}
        rfc: ${RFC}
        severity: ${SEVERITY}
        url: ${URL}
    pipelines:
    - inputRefs: ${{INPUT_REFS}}
      name: forward-to-syslog
      outputRefs:
        - rsyslog
    serviceAccount:
      name: ${SERVICE_ACCOUNT_NAME}
parameters:
- name: NAME
  value: "instance"
- name: NAMESPACE
  value: "openshift-logging"
- name: FACILITY
  value: "local0"
- name: SEVERITY
  value: "informational"
- name: URL
  value: "udp://rsyslogserver.openshift-logging.svc:514"
- name: SERVICE_ACCOUNT_NAME
  value: ""
- name: INPUT_REFS
  value: "[\"infrastructure\", \"audit\", \"application\"]"
- name: RFC
  value: RFC5424
- name: COLLECTOR
  value: "{}"
`)

func testdataObservabilityOpenshiftIo_clusterlogforwarderSyslogYamlBytes() ([]byte, error) {
	return _testdataObservabilityOpenshiftIo_clusterlogforwarderSyslogYaml, nil
}

func testdataObservabilityOpenshiftIo_clusterlogforwarderSyslogYaml() (*asset, error) {
	bytes, err := testdataObservabilityOpenshiftIo_clusterlogforwarderSyslogYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "testdata/observability.openshift.io_clusterlogforwarder/syslog.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testdataOdfObjectbucketclaimYaml = []byte(`kind: Template
apiVersion: template.openshift.io/v1
metadata:
  name: objectBucketClaim-template
objects:
- apiVersion: objectbucket.io/v1alpha1
  kind: ObjectBucketClaim
  metadata:
    name: ${NAME}
    namespace: ${NAMESPACE}
  spec:
    additionalConfig:
      bucketclass: ${BUCKETCLASS}
    generateBucketName: ${NAME}
    bucketName: ${NAME}
    storageClassName: ${STORAGE_CLASS_NAME}
parameters:
- name: NAME
  value: logging-loki
- name: NAMESPACE
  value: openshift-storage
- name: BUCKETCLASS
  value: noobaa-default-bucket-class
- name: STORAGE_CLASS_NAME
  value: openshift-storage.noobaa.io
`)

func testdataOdfObjectbucketclaimYamlBytes() ([]byte, error) {
	return _testdataOdfObjectbucketclaimYaml, nil
}

func testdataOdfObjectbucketclaimYaml() (*asset, error) {
	bytes, err := testdataOdfObjectbucketclaimYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "testdata/odf/objectBucketClaim.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testdataPrometheusK8sRbacYaml = []byte(`apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: prometheus-k8s
rules:
  - apiGroups:
      - ""
    resources:
      - services
      - endpoints
      - pods
    verbs:
      - get
      - list
      - watch
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: prometheus-k8s
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: prometheus-k8s
subjects:
  - kind: ServiceAccount
    name: prometheus-k8s
    namespace: openshift-monitoring
`)

func testdataPrometheusK8sRbacYamlBytes() ([]byte, error) {
	return _testdataPrometheusK8sRbacYaml, nil
}

func testdataPrometheusK8sRbacYaml() (*asset, error) {
	bytes, err := testdataPrometheusK8sRbacYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "testdata/prometheus-k8s-rbac.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testdataRapidastCustomscanPolicy = []byte(`<?xml version="1.0" encoding="UTF-8" standalone="no"?>
<configuration>
    <policy>helm-custom-scan</policy>
    <scanner>
        <level>MEDIUM</level>
        <strength>MEDIUM</strength>
    </scanner>
    <plugins>
        <p6>
            <enabled>false</enabled>
            <level>OFF</level>
        </p6>
        <p7>
            <enabled>true</enabled>
            <level>DEFAULT</level>
        </p7>
        <p10045>
            <enabled>false</enabled>
            <level>OFF</level>
        </p10045>
        <p20019>
            <enabled>true</enabled>
            <level>DEFAULT</level>
        </p20019>
        <p40009>
            <enabled>true</enabled>
            <level>DEFAULT</level>
        </p40009>
        <p40012>
            <enabled>true</enabled>
            <level>DEFAULT</level>
        </p40012>
        <p40014>
            <enabled>false</enabled>
            <level>OFF</level>
        </p40014>
        <p40018>
            <enabled>true</enabled>
            <level>DEFAULT</level>
        </p40018>
        <p90019>
            <enabled>true</enabled>
            <level>DEFAULT</level>
        </p90019>
        <p90020>
            <enabled>true</enabled>
            <level>DEFAULT</level>
        </p90020>
        <p0>
            <enabled>false</enabled>
            <level>OFF</level>
        </p0>
        <p30001>
            <enabled>false</enabled>
            <level>OFF</level>
        </p30001>
        <p30002>
            <enabled>false</enabled>
            <level>OFF</level>
        </p30002>
        <p40003>
            <enabled>true</enabled>
            <level>DEFAULT</level>
        </p40003>
        <p40008>
            <enabled>true</enabled>
            <level>DEFAULT</level>
        </p40008>
        <p40028>
            <enabled>false</enabled>
            <level>OFF</level>
        </p40028>
        <p40032>
            <enabled>true</enabled>
        </p40032>
        <p40016>
            <enabled>false</enabled>
            <level>OFF</level>
        </p40016>
        <p40017>
            <enabled>false</enabled>
            <level>OFF</level>
        </p40017>
        <p50000>
            <enabled>true</enabled>
            <level>DEFAULT</level>
        </p50000>
        <p41>
            <enabled>false</enabled>
            <level>OFF</level>
        </p41>
        <p43>
            <enabled>false</enabled>
            <level>OFF</level>
        </p43>
        <p10048>
            <enabled>false</enabled>
            <level>OFF</level>
        </p10048>
        <p10107>
            <enabled>false</enabled>
            <level>OFF</level>
        </p10107>
        <p20012>
            <enabled>false</enabled>
            <level>OFF</level>
        </p20012>
        <p20015>
            <enabled>false</enabled>
            <level>OFF</level>
        </p20015>
        <p20016>
            <enabled>false</enabled>
            <level>OFF</level>
        </p20016>
        <p20017>
            <enabled>false</enabled>
            <level>OFF</level>
        </p20017>
        <p20018>
            <enabled>false</enabled>
            <level>OFF</level>
        </p20018>
        <p40013>
            <enabled>false</enabled>
            <level>OFF</level>
        </p40013>
        <p40019>
            <enabled>false</enabled>
            <level>OFF</level>
        </p40019>
        <p40020>
            <enabled>false</enabled>
            <level>OFF</level>
        </p40020>
        <p40021>
            <enabled>false</enabled>
            <level>OFF</level>
        </p40021>
        <p40022>
            <enabled>false</enabled>
            <level>OFF</level>
        </p40022>
        <p40024>
            <enabled>false</enabled>
            <level>OFF</level>
        </p40024>
        <p40026>
            <enabled>false</enabled>
            <level>OFF</level>
        </p40026>
        <p40027>
            <enabled>false</enabled>
            <level>OFF</level>
        </p40027>
        <p90021>
            <enabled>true</enabled>
            <level>DEFAULT</level>
        </p90021>
        <p90023>
            <enabled>true</enabled>
            <level>DEFAULT</level>
        </p90023>
        <p90024>
            <enabled>false</enabled>
            <level>OFF</level>
        </p90024>
        <p90025>
            <enabled>false</enabled>
            <level>OFF</level>
        </p90025>
        <p90034>
            <enabled>true</enabled>
            <level>DEFAULT</level>
        </p90034>
        <p42>
            <enabled>false</enabled>
            <level>OFF</level>
        </p42>
        <p10051>
            <enabled>false</enabled>
            <level>OFF</level>
        </p10051>
        <p10053>
            <enabled>false</enabled>
            <level>OFF</level>
        </p10053>
        <p10095>
            <enabled>false</enabled>
            <level>OFF</level>
        </p10095>
        <p10106>
            <enabled>false</enabled>
            <level>OFF</level>
        </p10106>
        <p30003>
            <enabled>false</enabled>
            <level>OFF</level>
        </p30003>
        <p40025>
            <enabled>false</enabled>
            <level>OFF</level>
        </p40025>
        <p40029>
            <enabled>false</enabled>
            <level>OFF</level>
        </p40029>
        <p40034>
            <enabled>true</enabled>
        </p40034>
        <p40035>
            <enabled>true</enabled>
            <level>DEFAULT</level>
        </p40035>
        <p90017>
            <enabled>true</enabled>
        </p90017>
        <p90028>
            <enabled>false</enabled>
            <level>OFF</level>
        </p90028>
        <p10047>
            <enabled>false</enabled>
            <level>OFF</level>
        </p10047>
        <p10058>
            <enabled>true</enabled>
            <level>DEFAULT</level>
        </p10058>
        <p10104>
            <enabled>false</enabled>
            <level>OFF</level>
        </p10104>
        <p20014>
            <enabled>false</enabled>
            <level>OFF</level>
        </p20014>
        <p40023>
            <enabled>false</enabled>
            <level>OFF</level>
        </p40023>
        <p90027>
            <enabled>false</enabled>
            <level>OFF</level>
        </p90027>
        <p40015>
            <enabled>false</enabled>
            <level>OFF</level>
        </p40015>
        <p40033>
            <enabled>true</enabled>
            <level>DEFAULT</level>
        </p40033>
        <p60100>
            <enabled>false</enabled>
            <level>OFF</level>
        </p60100>
        <p60101>
            <enabled>false</enabled>
            <level>OFF</level>
        </p60101>
        <p90026>
            <enabled>false</enabled>
            <level>OFF</level>
        </p90026>
        <p90029>
            <enabled>false</enabled>
            <level>OFF</level>
        </p90029>
        <p40038>
            <enabled>false</enabled>
            <level>OFF</level>
        </p40038>
        <p40039>
            <enabled>false</enabled>
            <level>OFF</level>
        </p40039>
        <p40040>
            <enabled>false</enabled>
            <level>OFF</level>
        </p40040>
    </plugins>
</configuration>
`)

func testdataRapidastCustomscanPolicyBytes() ([]byte, error) {
	return _testdataRapidastCustomscanPolicy, nil
}

func testdataRapidastCustomscanPolicy() (*asset, error) {
	bytes, err := testdataRapidastCustomscanPolicyBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "testdata/rapidast/customscan.policy", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testdataRapidastData_rapidastconfig_logging_v1Yaml = []byte(`config:
    configVersion: 4
application:
  shortName: "ocptest"
  url: "https://kubernetes.default.svc"
general:
    authentication:
        type: "http_header"
        parameters:
            name: "Authorization"
            value: "Bearer sha256~xxxxxxxx"
    container:
        type: "none"
scanners:
    zap:
        apiScan:
            apis:
                apiUrl: "https://kubernetes.default.svc/openapi/v3/apis/logging.openshift.io/v1"
        passiveScan:
            disabledRules: "2,10015,10027,10096,10024,10054"
        activeScan:
            policy: "Operator-scan"
        miscOptions:
          enableUI: False
          updateAddons: False
`)

func testdataRapidastData_rapidastconfig_logging_v1YamlBytes() ([]byte, error) {
	return _testdataRapidastData_rapidastconfig_logging_v1Yaml, nil
}

func testdataRapidastData_rapidastconfig_logging_v1Yaml() (*asset, error) {
	bytes, err := testdataRapidastData_rapidastconfig_logging_v1YamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "testdata/rapidast/data_rapidastconfig_logging_v1.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testdataRapidastData_rapidastconfig_logging_v1alpha1Yaml = []byte(`config:
    configVersion: 4
application:
  shortName: "ocptest"
  url: "https://kubernetes.default.svc"
general:
    authentication:
        type: "http_header"
        parameters:
            name: "Authorization"
            value: "Bearer sha256~xxxxxxxx"
    container:
        type: "none"
scanners:
    zap:
        apiScan:
            apis:
                apiUrl: "https://kubernetes.default.svc/openapi/v3/apis/logging.openshift.io/v1alpha1"
        passiveScan:
            disabledRules: "2,10015,10027,10096,10024,10054"
        activeScan:
            policy: "Operator-scan"
        miscOptions:
          enableUI: False
          updateAddons: False
`)

func testdataRapidastData_rapidastconfig_logging_v1alpha1YamlBytes() ([]byte, error) {
	return _testdataRapidastData_rapidastconfig_logging_v1alpha1Yaml, nil
}

func testdataRapidastData_rapidastconfig_logging_v1alpha1Yaml() (*asset, error) {
	bytes, err := testdataRapidastData_rapidastconfig_logging_v1alpha1YamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "testdata/rapidast/data_rapidastconfig_logging_v1alpha1.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testdataRapidastData_rapidastconfig_loki_v1Yaml = []byte(`config:
    configVersion: 4
application:
  shortName: "ocptest"
  url: "https://kubernetes.default.svc"
general:
    authentication:
        type: "http_header"
        parameters:
            name: "Authorization"
            value: "Bearer sha256~xxxxxxxx"
    container:
        type: "none"
scanners:
    zap:
        apiScan:
            apis:
                apiUrl: "https://kubernetes.default.svc/openapi/v3/apis/loki.grafana.com/v1"
        results: "*stdout"   
        passiveScan:
            disabledRules: "2,10015,10027,10096,10024,10054"
        activeScan:
            policy: "Operator-scan"
        miscOptions:
          enableUI: False
          updateAddons: False
`)

func testdataRapidastData_rapidastconfig_loki_v1YamlBytes() ([]byte, error) {
	return _testdataRapidastData_rapidastconfig_loki_v1Yaml, nil
}

func testdataRapidastData_rapidastconfig_loki_v1Yaml() (*asset, error) {
	bytes, err := testdataRapidastData_rapidastconfig_loki_v1YamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "testdata/rapidast/data_rapidastconfig_loki_v1.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testdataRapidastData_rapidastconfig_observability_v1Yaml = []byte(`config:
    configVersion: 4
application:
  shortName: "ocptest"
  url: "https://kubernetes.default.svc"
general:
    authentication:
        type: "http_header"
        parameters:
            name: "Authorization"
            value: "Bearer sha256~xxxxxxxx"
    container:
        # currently supported: ` + "`" + `podman` + "`" + ` and ` + "`" + `none` + "`" + `
        type: "none"
scanners:
    zap:
        apiScan:
            apis:
                apiUrl: "https://kubernetes.default.svc/openapi/v3/apis/observability.openshift.io/v1"
        passiveScan:
            disabledRules: "2,10015,10027,10096,10024,10054"
        activeScan:
            policy: "Operator-scan"
        miscOptions:
          enableUI: False
          updateAddons: False
`)

func testdataRapidastData_rapidastconfig_observability_v1YamlBytes() ([]byte, error) {
	return _testdataRapidastData_rapidastconfig_observability_v1Yaml, nil
}

func testdataRapidastData_rapidastconfig_observability_v1Yaml() (*asset, error) {
	bytes, err := testdataRapidastData_rapidastconfig_observability_v1YamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "testdata/rapidast/data_rapidastconfig_observability_v1.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testdataRapidastJob_rapidastYaml = []byte(`kind: Template
apiVersion: template.openshift.io/v1
metadata:
  name: rapidast-job-template
objects:
- apiVersion: batch/v1
  kind: Job
  metadata:
    name: ${NAME}
  spec:
    backoffLimit: 0
    completionMode: NonIndexed
    completions: 1
    parallelism: 1
    selector:
      matchLabels:
        job-name: ${NAME}
    suspend: false
    template:
      metadata:
        labels:
          job-name: ${NAME}
        name: rapidast-job
      spec:
        containers:
        - command: ["/bin/sh"]
          args:
          - "-c"
          - | 
            export HOME=/home/rapidast 
            mkdir -p $HOME/.ZAP/policies 
            cp /opt/rapidast/config/customscan.policy $HOME/.ZAP/policies/custom-scan.policy 
            rapidast.py --config /opt/rapidast/config/rapidastconfig.yaml
            echo "--------------- show rapidash result -----------------"  
            find $HOME/results/ocptest -name zap-report.json -exec cat {} \;
            echo "--------------- rapidash result end -----------------"  
          image: quay.io/redhatproductsecurity/rapidast:latest
          workingDir: "/home/rapidast"
          imagePullPolicy: Always
          securityContext:
            allowPrivilegeEscalation: false
            capabilities:
              drop:
              - ALL
            runAsNonRoot: true
          name: rapidast
          resources: {}
          terminationMessagePath: /dev/termination-log
          terminationMessagePolicy: File
          volumeMounts:
          - mountPath: /opt/rapidast/config
            name: config-volume
          - mountPath: /home/rapidast
            name: work-volume
        dnsPolicy: ClusterFirst
        restartPolicy: Never
        schedulerName: default-scheduler
        securityContext:
          seccompProfile:
            type: RuntimeDefault
        terminationGracePeriodSeconds: 30
        nodeSelector:
          kubernetes.io/os: linux
          kubernetes.io/arch: amd64
        volumes:
        - configMap:
            defaultMode: 420
            name: ${NAME}
          name: config-volume
        - name: work-volume
          emptyDir:
            sizeLimit: 10Mi
parameters:
- name: NAME
  value: rapidast-job
`)

func testdataRapidastJob_rapidastYamlBytes() ([]byte, error) {
	return _testdataRapidastJob_rapidastYaml, nil
}

func testdataRapidastJob_rapidastYaml() (*asset, error) {
	bytes, err := testdataRapidastJob_rapidastYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "testdata/rapidast/job_rapidast.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testdataSubscriptionAllnamespaceOgYaml = []byte(`kind: Template
apiVersion: template.openshift.io/v1
metadata:
  name: elasticsearch-operator-og-template
objects:
- apiVersion: operators.coreos.com/v1
  kind: OperatorGroup
  metadata:
    name: ${OG_NAME}
    namespace: ${NAMESPACE}
  spec: {}
parameters:
  - name: OG_NAME
    value: "openshift-operators-redhat"
  - name: NAMESPACE
    value: "openshift-operators-redhat"
`)

func testdataSubscriptionAllnamespaceOgYamlBytes() ([]byte, error) {
	return _testdataSubscriptionAllnamespaceOgYaml, nil
}

func testdataSubscriptionAllnamespaceOgYaml() (*asset, error) {
	bytes, err := testdataSubscriptionAllnamespaceOgYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "testdata/subscription/allnamespace-og.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testdataSubscriptionCatsrcYaml = []byte(`kind: Template
apiVersion: template.openshift.io/v1
metadata:
  name: catsrc-template
objects:
- apiVersion: operators.coreos.com/v1alpha1
  kind: CatalogSource
  metadata:
    name: ${NAME}
    namespace: ${NAMESPACE}
  spec:
    sourceType: grpc
    image: ${IMAGE}
parameters:
- name: NAME
- name: NAMESPACE
  value: openshift-marketplace
- name: IMAGE
  value: quay.io/openshift-qe-optional-operators/ocp4-index:latest
`)

func testdataSubscriptionCatsrcYamlBytes() ([]byte, error) {
	return _testdataSubscriptionCatsrcYaml, nil
}

func testdataSubscriptionCatsrcYaml() (*asset, error) {
	bytes, err := testdataSubscriptionCatsrcYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "testdata/subscription/catsrc.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testdataSubscriptionNamespaceYaml = []byte(`kind: Template
apiVersion: template.openshift.io/v1
metadata:
  name: namespace-template
objects:
- kind: Namespace
  apiVersion: v1
  metadata:
    name: ${NAMESPACE_NAME}
    annotations:
      openshift.io/node-selector: ""
    labels:
      openshift.io/cluster-monitoring: "true"
parameters:
- name: NAMESPACE_NAME
`)

func testdataSubscriptionNamespaceYamlBytes() ([]byte, error) {
	return _testdataSubscriptionNamespaceYaml, nil
}

func testdataSubscriptionNamespaceYaml() (*asset, error) {
	bytes, err := testdataSubscriptionNamespaceYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "testdata/subscription/namespace.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testdataSubscriptionSinglenamespaceOgYaml = []byte(`kind: Template
apiVersion: template.openshift.io/v1
metadata:
  name: clusterlogging-og-template
objects:
- apiVersion: operators.coreos.com/v1
  kind: OperatorGroup
  metadata:
    name: ${OG_NAME}
    namespace: ${NAMESPACE}
  spec:
    targetNamespaces:
    - ${NAMESPACE}
parameters:
  - name: OG_NAME
    value: "openshift-logging"
  - name: NAMESPACE
    value: "openshift-logging"
`)

func testdataSubscriptionSinglenamespaceOgYamlBytes() ([]byte, error) {
	return _testdataSubscriptionSinglenamespaceOgYaml, nil
}

func testdataSubscriptionSinglenamespaceOgYaml() (*asset, error) {
	bytes, err := testdataSubscriptionSinglenamespaceOgYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "testdata/subscription/singlenamespace-og.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testdataSubscriptionSubTemplateYaml = []byte(`kind: Template
apiVersion: template.openshift.io/v1
metadata:
  name: subscription-template
objects:
- apiVersion: operators.coreos.com/v1alpha1
  kind: Subscription
  metadata:
    name: ${PACKAGE_NAME}
    namespace: ${NAMESPACE}
  spec:
    channel: ${CHANNEL}
    installPlanApproval: ${INSTALL_PLAN_APPROVAL}
    name: ${PACKAGE_NAME}
    source: ${SOURCE}
    sourceNamespace: ${SOURCE_NAMESPACE}
parameters:
  - name: PACKAGE_NAME
  - name: NAMESPACE
  - name: CHANNEL
  - name: SOURCE
    value: "qe-app-registry"
  - name: SOURCE_NAMESPACE
    value: "openshift-marketplace"
  - name: INSTALL_PLAN_APPROVAL
    value: Automatic
`)

func testdataSubscriptionSubTemplateYamlBytes() ([]byte, error) {
	return _testdataSubscriptionSubTemplateYaml, nil
}

func testdataSubscriptionSubTemplateYaml() (*asset, error) {
	bytes, err := testdataSubscriptionSubTemplateYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "testdata/subscription/sub-template.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

// Asset loads and returns the asset for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func Asset(name string) ([]byte, error) {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[cannonicalName]; ok {
		a, err := f()
		if err != nil {
			return nil, fmt.Errorf("Asset %s can't read by error: %v", name, err)
		}
		return a.bytes, nil
	}
	return nil, fmt.Errorf("Asset %s not found", name)
}

// MustAsset is like Asset but panics when Asset would return an error.
// It simplifies safe initialization of global variables.
func MustAsset(name string) []byte {
	a, err := Asset(name)
	if err != nil {
		panic("asset: Asset(" + name + "): " + err.Error())
	}

	return a
}

// AssetInfo loads and returns the asset info for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func AssetInfo(name string) (os.FileInfo, error) {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[cannonicalName]; ok {
		a, err := f()
		if err != nil {
			return nil, fmt.Errorf("AssetInfo %s can't read by error: %v", name, err)
		}
		return a.info, nil
	}
	return nil, fmt.Errorf("AssetInfo %s not found", name)
}

// AssetNames returns the names of the assets.
func AssetNames() []string {
	names := make([]string, 0, len(_bindata))
	for name := range _bindata {
		names = append(names, name)
	}
	return names
}

// _bindata is a table, holding each asset generator, mapped to its name.
var _bindata = map[string]func() (*asset, error){
	"testdata/UIPlugin/UIPlugin.yaml":                                                                      testdataUipluginUipluginYaml,
	"testdata/eventrouter/eventrouter.yaml":                                                                testdataEventrouterEventrouterYaml,
	"testdata/external-log-stores/cert_generation.sh":                                                      testdataExternalLogStoresCert_generationSh,
	"testdata/external-log-stores/elasticsearch/6/http/no_user/configmap.yaml":                             testdataExternalLogStoresElasticsearch6HttpNo_userConfigmapYaml,
	"testdata/external-log-stores/elasticsearch/6/http/no_user/deployment.yaml":                            testdataExternalLogStoresElasticsearch6HttpNo_userDeploymentYaml,
	"testdata/external-log-stores/elasticsearch/6/http/user_auth/configmap.yaml":                           testdataExternalLogStoresElasticsearch6HttpUser_authConfigmapYaml,
	"testdata/external-log-stores/elasticsearch/6/http/user_auth/deployment.yaml":                          testdataExternalLogStoresElasticsearch6HttpUser_authDeploymentYaml,
	"testdata/external-log-stores/elasticsearch/6/https/no_user/configmap.yaml":                            testdataExternalLogStoresElasticsearch6HttpsNo_userConfigmapYaml,
	"testdata/external-log-stores/elasticsearch/6/https/no_user/deployment.yaml":                           testdataExternalLogStoresElasticsearch6HttpsNo_userDeploymentYaml,
	"testdata/external-log-stores/elasticsearch/6/https/user_auth/configmap.yaml":                          testdataExternalLogStoresElasticsearch6HttpsUser_authConfigmapYaml,
	"testdata/external-log-stores/elasticsearch/6/https/user_auth/deployment.yaml":                         testdataExternalLogStoresElasticsearch6HttpsUser_authDeploymentYaml,
	"testdata/external-log-stores/elasticsearch/7/http/no_user/configmap.yaml":                             testdataExternalLogStoresElasticsearch7HttpNo_userConfigmapYaml,
	"testdata/external-log-stores/elasticsearch/7/http/no_user/deployment.yaml":                            testdataExternalLogStoresElasticsearch7HttpNo_userDeploymentYaml,
	"testdata/external-log-stores/elasticsearch/7/http/user_auth/configmap.yaml":                           testdataExternalLogStoresElasticsearch7HttpUser_authConfigmapYaml,
	"testdata/external-log-stores/elasticsearch/7/http/user_auth/deployment.yaml":                          testdataExternalLogStoresElasticsearch7HttpUser_authDeploymentYaml,
	"testdata/external-log-stores/elasticsearch/7/https/no_user/configmap.yaml":                            testdataExternalLogStoresElasticsearch7HttpsNo_userConfigmapYaml,
	"testdata/external-log-stores/elasticsearch/7/https/no_user/deployment.yaml":                           testdataExternalLogStoresElasticsearch7HttpsNo_userDeploymentYaml,
	"testdata/external-log-stores/elasticsearch/7/https/user_auth/configmap.yaml":                          testdataExternalLogStoresElasticsearch7HttpsUser_authConfigmapYaml,
	"testdata/external-log-stores/elasticsearch/7/https/user_auth/deployment.yaml":                         testdataExternalLogStoresElasticsearch7HttpsUser_authDeploymentYaml,
	"testdata/external-log-stores/elasticsearch/8/http/no_user/configmap.yaml":                             testdataExternalLogStoresElasticsearch8HttpNo_userConfigmapYaml,
	"testdata/external-log-stores/elasticsearch/8/http/no_user/deployment.yaml":                            testdataExternalLogStoresElasticsearch8HttpNo_userDeploymentYaml,
	"testdata/external-log-stores/elasticsearch/8/http/user_auth/configmap.yaml":                           testdataExternalLogStoresElasticsearch8HttpUser_authConfigmapYaml,
	"testdata/external-log-stores/elasticsearch/8/http/user_auth/deployment.yaml":                          testdataExternalLogStoresElasticsearch8HttpUser_authDeploymentYaml,
	"testdata/external-log-stores/elasticsearch/8/https/no_user/configmap.yaml":                            testdataExternalLogStoresElasticsearch8HttpsNo_userConfigmapYaml,
	"testdata/external-log-stores/elasticsearch/8/https/no_user/deployment.yaml":                           testdataExternalLogStoresElasticsearch8HttpsNo_userDeploymentYaml,
	"testdata/external-log-stores/elasticsearch/8/https/user_auth/configmap.yaml":                          testdataExternalLogStoresElasticsearch8HttpsUser_authConfigmapYaml,
	"testdata/external-log-stores/elasticsearch/8/https/user_auth/deployment.yaml":                         testdataExternalLogStoresElasticsearch8HttpsUser_authDeploymentYaml,
	"testdata/external-log-stores/fluentd/insecure/configmap.yaml":                                         testdataExternalLogStoresFluentdInsecureConfigmapYaml,
	"testdata/external-log-stores/fluentd/insecure/deployment.yaml":                                        testdataExternalLogStoresFluentdInsecureDeploymentYaml,
	"testdata/external-log-stores/fluentd/insecure/http-configmap.yaml":                                    testdataExternalLogStoresFluentdInsecureHttpConfigmapYaml,
	"testdata/external-log-stores/fluentd/secure/cm-mtls-share.yaml":                                       testdataExternalLogStoresFluentdSecureCmMtlsShareYaml,
	"testdata/external-log-stores/fluentd/secure/cm-mtls.yaml":                                             testdataExternalLogStoresFluentdSecureCmMtlsYaml,
	"testdata/external-log-stores/fluentd/secure/cm-serverauth-share.yaml":                                 testdataExternalLogStoresFluentdSecureCmServerauthShareYaml,
	"testdata/external-log-stores/fluentd/secure/cm-serverauth.yaml":                                       testdataExternalLogStoresFluentdSecureCmServerauthYaml,
	"testdata/external-log-stores/fluentd/secure/deployment.yaml":                                          testdataExternalLogStoresFluentdSecureDeploymentYaml,
	"testdata/external-log-stores/fluentd/secure/http-cm-mtls.yaml":                                        testdataExternalLogStoresFluentdSecureHttpCmMtlsYaml,
	"testdata/external-log-stores/fluentd/secure/http-cm-serverauth.yaml":                                  testdataExternalLogStoresFluentdSecureHttpCmServerauthYaml,
	"testdata/external-log-stores/kafka/amqstreams/kafka-no-auth-cluster.yaml":                             testdataExternalLogStoresKafkaAmqstreamsKafkaNoAuthClusterYaml,
	"testdata/external-log-stores/kafka/amqstreams/kafka-no-auth-consumer-job.yaml":                        testdataExternalLogStoresKafkaAmqstreamsKafkaNoAuthConsumerJobYaml,
	"testdata/external-log-stores/kafka/amqstreams/kafka-sasl-cluster.yaml":                                testdataExternalLogStoresKafkaAmqstreamsKafkaSaslClusterYaml,
	"testdata/external-log-stores/kafka/amqstreams/kafka-sasl-consumer-job.yaml":                           testdataExternalLogStoresKafkaAmqstreamsKafkaSaslConsumerJobYaml,
	"testdata/external-log-stores/kafka/amqstreams/kafka-sasl-consumers-config.yaml":                       testdataExternalLogStoresKafkaAmqstreamsKafkaSaslConsumersConfigYaml,
	"testdata/external-log-stores/kafka/amqstreams/kafka-sasl-user.yaml":                                   testdataExternalLogStoresKafkaAmqstreamsKafkaSaslUserYaml,
	"testdata/external-log-stores/kafka/amqstreams/kafka-topic.yaml":                                       testdataExternalLogStoresKafkaAmqstreamsKafkaTopicYaml,
	"testdata/external-log-stores/kafka/cert_generation.sh":                                                testdataExternalLogStoresKafkaCert_generationSh,
	"testdata/external-log-stores/kafka/kafka-rbac.yaml":                                                   testdataExternalLogStoresKafkaKafkaRbacYaml,
	"testdata/external-log-stores/kafka/kafka-svc.yaml":                                                    testdataExternalLogStoresKafkaKafkaSvcYaml,
	"testdata/external-log-stores/kafka/plaintext-ssl/consumer-configmap.yaml":                             testdataExternalLogStoresKafkaPlaintextSslConsumerConfigmapYaml,
	"testdata/external-log-stores/kafka/plaintext-ssl/kafka-configmap.yaml":                                testdataExternalLogStoresKafkaPlaintextSslKafkaConfigmapYaml,
	"testdata/external-log-stores/kafka/plaintext-ssl/kafka-consumer-deployment.yaml":                      testdataExternalLogStoresKafkaPlaintextSslKafkaConsumerDeploymentYaml,
	"testdata/external-log-stores/kafka/plaintext-ssl/kafka-statefulset.yaml":                              testdataExternalLogStoresKafkaPlaintextSslKafkaStatefulsetYaml,
	"testdata/external-log-stores/kafka/sasl-plaintext/consumer-configmap.yaml":                            testdataExternalLogStoresKafkaSaslPlaintextConsumerConfigmapYaml,
	"testdata/external-log-stores/kafka/sasl-plaintext/kafka-configmap.yaml":                               testdataExternalLogStoresKafkaSaslPlaintextKafkaConfigmapYaml,
	"testdata/external-log-stores/kafka/sasl-plaintext/kafka-consumer-deployment.yaml":                     testdataExternalLogStoresKafkaSaslPlaintextKafkaConsumerDeploymentYaml,
	"testdata/external-log-stores/kafka/sasl-plaintext/kafka-statefulset.yaml":                             testdataExternalLogStoresKafkaSaslPlaintextKafkaStatefulsetYaml,
	"testdata/external-log-stores/kafka/sasl-ssl/consumer-configmap.yaml":                                  testdataExternalLogStoresKafkaSaslSslConsumerConfigmapYaml,
	"testdata/external-log-stores/kafka/sasl-ssl/kafka-configmap.yaml":                                     testdataExternalLogStoresKafkaSaslSslKafkaConfigmapYaml,
	"testdata/external-log-stores/kafka/sasl-ssl/kafka-consumer-deployment.yaml":                           testdataExternalLogStoresKafkaSaslSslKafkaConsumerDeploymentYaml,
	"testdata/external-log-stores/kafka/sasl-ssl/kafka-statefulset.yaml":                                   testdataExternalLogStoresKafkaSaslSslKafkaStatefulsetYaml,
	"testdata/external-log-stores/kafka/zookeeper/configmap-ssl.yaml":                                      testdataExternalLogStoresKafkaZookeeperConfigmapSslYaml,
	"testdata/external-log-stores/kafka/zookeeper/configmap.yaml":                                          testdataExternalLogStoresKafkaZookeeperConfigmapYaml,
	"testdata/external-log-stores/kafka/zookeeper/zookeeper-statefulset.yaml":                              testdataExternalLogStoresKafkaZookeeperZookeeperStatefulsetYaml,
	"testdata/external-log-stores/kafka/zookeeper/zookeeper-svc.yaml":                                      testdataExternalLogStoresKafkaZookeeperZookeeperSvcYaml,
	"testdata/external-log-stores/loki/loki-configmap.yaml":                                                testdataExternalLogStoresLokiLokiConfigmapYaml,
	"testdata/external-log-stores/loki/loki-deployment.yaml":                                               testdataExternalLogStoresLokiLokiDeploymentYaml,
	"testdata/external-log-stores/otel/otel-collector.yaml":                                                testdataExternalLogStoresOtelOtelCollectorYaml,
	"testdata/external-log-stores/rsyslog/insecure/configmap.yaml":                                         testdataExternalLogStoresRsyslogInsecureConfigmapYaml,
	"testdata/external-log-stores/rsyslog/insecure/deployment.yaml":                                        testdataExternalLogStoresRsyslogInsecureDeploymentYaml,
	"testdata/external-log-stores/rsyslog/insecure/svc.yaml":                                               testdataExternalLogStoresRsyslogInsecureSvcYaml,
	"testdata/external-log-stores/rsyslog/secure/configmap.yaml":                                           testdataExternalLogStoresRsyslogSecureConfigmapYaml,
	"testdata/external-log-stores/rsyslog/secure/deployment.yaml":                                          testdataExternalLogStoresRsyslogSecureDeploymentYaml,
	"testdata/external-log-stores/rsyslog/secure/svc.yaml":                                                 testdataExternalLogStoresRsyslogSecureSvcYaml,
	"testdata/external-log-stores/splunk/route-edge_splunk_template.yaml":                                  testdataExternalLogStoresSplunkRouteEdge_splunk_templateYaml,
	"testdata/external-log-stores/splunk/route-passthrough_splunk_template.yaml":                           testdataExternalLogStoresSplunkRoutePassthrough_splunk_templateYaml,
	"testdata/external-log-stores/splunk/secret_splunk_template.yaml":                                      testdataExternalLogStoresSplunkSecret_splunk_templateYaml,
	"testdata/external-log-stores/splunk/secret_tls_passphrase_splunk_template.yaml":                       testdataExternalLogStoresSplunkSecret_tls_passphrase_splunk_templateYaml,
	"testdata/external-log-stores/splunk/secret_tls_splunk_template.yaml":                                  testdataExternalLogStoresSplunkSecret_tls_splunk_templateYaml,
	"testdata/external-log-stores/splunk/statefulset_splunk-8.2_template.yaml":                             testdataExternalLogStoresSplunkStatefulset_splunk82_templateYaml,
	"testdata/external-log-stores/splunk/statefulset_splunk-9.0_template.yaml":                             testdataExternalLogStoresSplunkStatefulset_splunk90_templateYaml,
	"testdata/generatelog/42981.yaml":                                                                      testdataGeneratelog42981Yaml,
	"testdata/generatelog/container_json_log_template.json":                                                testdataGeneratelogContainer_json_log_templateJson,
	"testdata/generatelog/container_json_log_template_unannoted.json":                                      testdataGeneratelogContainer_json_log_template_unannotedJson,
	"testdata/generatelog/container_non_json_log_template.json":                                            testdataGeneratelogContainer_non_json_log_templateJson,
	"testdata/generatelog/logging-performance-app-generator.json":                                          testdataGeneratelogLoggingPerformanceAppGeneratorJson,
	"testdata/generatelog/multi_container_json_log_template.yaml":                                          testdataGeneratelogMulti_container_json_log_templateYaml,
	"testdata/generatelog/multiline-error-log.yaml":                                                        testdataGeneratelogMultilineErrorLogYaml,
	"testdata/logfilemetricexporter/lfme.yaml":                                                             testdataLogfilemetricexporterLfmeYaml,
	"testdata/loki-log-alerts/cluster-monitoring-config.yaml":                                              testdataLokiLogAlertsClusterMonitoringConfigYaml,
	"testdata/loki-log-alerts/loki-app-alerting-rule-template.yaml":                                        testdataLokiLogAlertsLokiAppAlertingRuleTemplateYaml,
	"testdata/loki-log-alerts/loki-app-recording-rule-template.yaml":                                       testdataLokiLogAlertsLokiAppRecordingRuleTemplateYaml,
	"testdata/loki-log-alerts/loki-infra-alerting-rule-template.yaml":                                      testdataLokiLogAlertsLokiInfraAlertingRuleTemplateYaml,
	"testdata/loki-log-alerts/loki-infra-recording-rule-template.yaml":                                     testdataLokiLogAlertsLokiInfraRecordingRuleTemplateYaml,
	"testdata/loki-log-alerts/user-workload-monitoring-config.yaml":                                        testdataLokiLogAlertsUserWorkloadMonitoringConfigYaml,
	"testdata/lokistack/fine-grained-access-roles.yaml":                                                    testdataLokistackFineGrainedAccessRolesYaml,
	"testdata/lokistack/lokistack-simple-ipv6-tls.yaml":                                                    testdataLokistackLokistackSimpleIpv6TlsYaml,
	"testdata/lokistack/lokistack-simple-ipv6.yaml":                                                        testdataLokistackLokistackSimpleIpv6Yaml,
	"testdata/lokistack/lokistack-simple-tls.yaml":                                                         testdataLokistackLokistackSimpleTlsYaml,
	"testdata/lokistack/lokistack-simple.yaml":                                                             testdataLokistackLokistackSimpleYaml,
	"testdata/minIO/deploy.yaml":                                                                           testdataMinioDeployYaml,
	"testdata/observability.openshift.io_clusterlogforwarder/48593.yaml":                                   testdataObservabilityOpenshiftIo_clusterlogforwarder48593Yaml,
	"testdata/observability.openshift.io_clusterlogforwarder/67421.yaml":                                   testdataObservabilityOpenshiftIo_clusterlogforwarder67421Yaml,
	"testdata/observability.openshift.io_clusterlogforwarder/68318.yaml":                                   testdataObservabilityOpenshiftIo_clusterlogforwarder68318Yaml,
	"testdata/observability.openshift.io_clusterlogforwarder/71049.yaml":                                   testdataObservabilityOpenshiftIo_clusterlogforwarder71049Yaml,
	"testdata/observability.openshift.io_clusterlogforwarder/71749.yaml":                                   testdataObservabilityOpenshiftIo_clusterlogforwarder71749Yaml,
	"testdata/observability.openshift.io_clusterlogforwarder/affinity-81397.yaml":                          testdataObservabilityOpenshiftIo_clusterlogforwarderAffinity81397Yaml,
	"testdata/observability.openshift.io_clusterlogforwarder/affinity-81398.yaml":                          testdataObservabilityOpenshiftIo_clusterlogforwarderAffinity81398Yaml,
	"testdata/observability.openshift.io_clusterlogforwarder/audit-policy.yaml":                            testdataObservabilityOpenshiftIo_clusterlogforwarderAuditPolicyYaml,
	"testdata/observability.openshift.io_clusterlogforwarder/azureMonitor-min-opts.yaml":                   testdataObservabilityOpenshiftIo_clusterlogforwarderAzuremonitorMinOptsYaml,
	"testdata/observability.openshift.io_clusterlogforwarder/azureMonitor.yaml":                            testdataObservabilityOpenshiftIo_clusterlogforwarderAzuremonitorYaml,
	"testdata/observability.openshift.io_clusterlogforwarder/clf-external-loki-with-secret-tenantKey.yaml": testdataObservabilityOpenshiftIo_clusterlogforwarderClfExternalLokiWithSecretTenantkeyYaml,
	"testdata/observability.openshift.io_clusterlogforwarder/clf-external-loki-with-secret.yaml":           testdataObservabilityOpenshiftIo_clusterlogforwarderClfExternalLokiWithSecretYaml,
	"testdata/observability.openshift.io_clusterlogforwarder/clf-kafka-multi-brokers.yaml":                 testdataObservabilityOpenshiftIo_clusterlogforwarderClfKafkaMultiBrokersYaml,
	"testdata/observability.openshift.io_clusterlogforwarder/clf-kafka-multi-topics.yaml":                  testdataObservabilityOpenshiftIo_clusterlogforwarderClfKafkaMultiTopicsYaml,
	"testdata/observability.openshift.io_clusterlogforwarder/clf-kafka-no-auth.yaml":                       testdataObservabilityOpenshiftIo_clusterlogforwarderClfKafkaNoAuthYaml,
	"testdata/observability.openshift.io_clusterlogforwarder/clf-kafka-sasl-ssl.yaml":                      testdataObservabilityOpenshiftIo_clusterlogforwarderClfKafkaSaslSslYaml,
	"testdata/observability.openshift.io_clusterlogforwarder/clf-kafka-with-auth.yaml":                     testdataObservabilityOpenshiftIo_clusterlogforwarderClfKafkaWithAuthYaml,
	"testdata/observability.openshift.io_clusterlogforwarder/cloudwatch-accessKey.yaml":                    testdataObservabilityOpenshiftIo_clusterlogforwarderCloudwatchAccesskeyYaml,
	"testdata/observability.openshift.io_clusterlogforwarder/cloudwatch-iamRole.yaml":                      testdataObservabilityOpenshiftIo_clusterlogforwarderCloudwatchIamroleYaml,
	"testdata/observability.openshift.io_clusterlogforwarder/cloudwatch-multiple-iamRole.yaml":             testdataObservabilityOpenshiftIo_clusterlogforwarderCloudwatchMultipleIamroleYaml,
	"testdata/observability.openshift.io_clusterlogforwarder/elasticsearch-https.yaml":                     testdataObservabilityOpenshiftIo_clusterlogforwarderElasticsearchHttpsYaml,
	"testdata/observability.openshift.io_clusterlogforwarder/elasticsearch-mtls.yaml":                      testdataObservabilityOpenshiftIo_clusterlogforwarderElasticsearchMtlsYaml,
	"testdata/observability.openshift.io_clusterlogforwarder/elasticsearch-userauth-https.yaml":            testdataObservabilityOpenshiftIo_clusterlogforwarderElasticsearchUserauthHttpsYaml,
	"testdata/observability.openshift.io_clusterlogforwarder/elasticsearch-userauth-mtls.yaml":             testdataObservabilityOpenshiftIo_clusterlogforwarderElasticsearchUserauthMtlsYaml,
	"testdata/observability.openshift.io_clusterlogforwarder/elasticsearch-userauth.yaml":                  testdataObservabilityOpenshiftIo_clusterlogforwarderElasticsearchUserauthYaml,
	"testdata/observability.openshift.io_clusterlogforwarder/elasticsearch.yaml":                           testdataObservabilityOpenshiftIo_clusterlogforwarderElasticsearchYaml,
	"testdata/observability.openshift.io_clusterlogforwarder/google-cloud-logging-multi-logids.yaml":       testdataObservabilityOpenshiftIo_clusterlogforwarderGoogleCloudLoggingMultiLogidsYaml,
	"testdata/observability.openshift.io_clusterlogforwarder/googleCloudLogging.yaml":                      testdataObservabilityOpenshiftIo_clusterlogforwarderGooglecloudloggingYaml,
	"testdata/observability.openshift.io_clusterlogforwarder/http-output-85490.yaml":                       testdataObservabilityOpenshiftIo_clusterlogforwarderHttpOutput85490Yaml,
	"testdata/observability.openshift.io_clusterlogforwarder/http-output.yaml":                             testdataObservabilityOpenshiftIo_clusterlogforwarderHttpOutputYaml,
	"testdata/observability.openshift.io_clusterlogforwarder/https-61567.yaml":                             testdataObservabilityOpenshiftIo_clusterlogforwarderHttps61567Yaml,
	"testdata/observability.openshift.io_clusterlogforwarder/https-output-ca.yaml":                         testdataObservabilityOpenshiftIo_clusterlogforwarderHttpsOutputCaYaml,
	"testdata/observability.openshift.io_clusterlogforwarder/https-output-mtls.yaml":                       testdataObservabilityOpenshiftIo_clusterlogforwarderHttpsOutputMtlsYaml,
	"testdata/observability.openshift.io_clusterlogforwarder/https-output-skipverify.yaml":                 testdataObservabilityOpenshiftIo_clusterlogforwarderHttpsOutputSkipverifyYaml,
	"testdata/observability.openshift.io_clusterlogforwarder/httpserver-to-httpoutput.yaml":                testdataObservabilityOpenshiftIo_clusterlogforwarderHttpserverToHttpoutputYaml,
	"testdata/observability.openshift.io_clusterlogforwarder/httpserver-to-splunk.yaml":                    testdataObservabilityOpenshiftIo_clusterlogforwarderHttpserverToSplunkYaml,
	"testdata/observability.openshift.io_clusterlogforwarder/loki.yaml":                                    testdataObservabilityOpenshiftIo_clusterlogforwarderLokiYaml,
	"testdata/observability.openshift.io_clusterlogforwarder/lokistack-with-labelkeys.yaml":                testdataObservabilityOpenshiftIo_clusterlogforwarderLokistackWithLabelkeysYaml,
	"testdata/observability.openshift.io_clusterlogforwarder/lokistack.yaml":                               testdataObservabilityOpenshiftIo_clusterlogforwarderLokistackYaml,
	"testdata/observability.openshift.io_clusterlogforwarder/lokistack_gateway_https_secret.yaml":          testdataObservabilityOpenshiftIo_clusterlogforwarderLokistack_gateway_https_secretYaml,
	"testdata/observability.openshift.io_clusterlogforwarder/otlp-lokistack.yaml":                          testdataObservabilityOpenshiftIo_clusterlogforwarderOtlpLokistackYaml,
	"testdata/observability.openshift.io_clusterlogforwarder/otlp.yaml":                                    testdataObservabilityOpenshiftIo_clusterlogforwarderOtlpYaml,
	"testdata/observability.openshift.io_clusterlogforwarder/rsyslog-mtls.yaml":                            testdataObservabilityOpenshiftIo_clusterlogforwarderRsyslogMtlsYaml,
	"testdata/observability.openshift.io_clusterlogforwarder/rsyslog-serverAuth.yaml":                      testdataObservabilityOpenshiftIo_clusterlogforwarderRsyslogServerauthYaml,
	"testdata/observability.openshift.io_clusterlogforwarder/splunk-mtls-passphrase.yaml":                  testdataObservabilityOpenshiftIo_clusterlogforwarderSplunkMtlsPassphraseYaml,
	"testdata/observability.openshift.io_clusterlogforwarder/splunk-mtls.yaml":                             testdataObservabilityOpenshiftIo_clusterlogforwarderSplunkMtlsYaml,
	"testdata/observability.openshift.io_clusterlogforwarder/splunk-serveronly.yaml":                       testdataObservabilityOpenshiftIo_clusterlogforwarderSplunkServeronlyYaml,
	"testdata/observability.openshift.io_clusterlogforwarder/splunk.yaml":                                  testdataObservabilityOpenshiftIo_clusterlogforwarderSplunkYaml,
	"testdata/observability.openshift.io_clusterlogforwarder/syslog-75317.yaml":                            testdataObservabilityOpenshiftIo_clusterlogforwarderSyslog75317Yaml,
	"testdata/observability.openshift.io_clusterlogforwarder/syslog-75431.yaml":                            testdataObservabilityOpenshiftIo_clusterlogforwarderSyslog75431Yaml,
	"testdata/observability.openshift.io_clusterlogforwarder/syslog-81512.yaml":                            testdataObservabilityOpenshiftIo_clusterlogforwarderSyslog81512Yaml,
	"testdata/observability.openshift.io_clusterlogforwarder/syslog-selected-ns.yaml":                      testdataObservabilityOpenshiftIo_clusterlogforwarderSyslogSelectedNsYaml,
	"testdata/observability.openshift.io_clusterlogforwarder/syslog.yaml":                                  testdataObservabilityOpenshiftIo_clusterlogforwarderSyslogYaml,
	"testdata/odf/objectBucketClaim.yaml":                                                                  testdataOdfObjectbucketclaimYaml,
	"testdata/prometheus-k8s-rbac.yaml":                                                                    testdataPrometheusK8sRbacYaml,
	"testdata/rapidast/customscan.policy":                                                                  testdataRapidastCustomscanPolicy,
	"testdata/rapidast/data_rapidastconfig_logging_v1.yaml":                                                testdataRapidastData_rapidastconfig_logging_v1Yaml,
	"testdata/rapidast/data_rapidastconfig_logging_v1alpha1.yaml":                                          testdataRapidastData_rapidastconfig_logging_v1alpha1Yaml,
	"testdata/rapidast/data_rapidastconfig_loki_v1.yaml":                                                   testdataRapidastData_rapidastconfig_loki_v1Yaml,
	"testdata/rapidast/data_rapidastconfig_observability_v1.yaml":                                          testdataRapidastData_rapidastconfig_observability_v1Yaml,
	"testdata/rapidast/job_rapidast.yaml":                                                                  testdataRapidastJob_rapidastYaml,
	"testdata/subscription/allnamespace-og.yaml":                                                           testdataSubscriptionAllnamespaceOgYaml,
	"testdata/subscription/catsrc.yaml":                                                                    testdataSubscriptionCatsrcYaml,
	"testdata/subscription/namespace.yaml":                                                                 testdataSubscriptionNamespaceYaml,
	"testdata/subscription/singlenamespace-og.yaml":                                                        testdataSubscriptionSinglenamespaceOgYaml,
	"testdata/subscription/sub-template.yaml":                                                              testdataSubscriptionSubTemplateYaml,
}

// AssetDir returns the file names below a certain
// directory embedded in the file by go-bindata.
// For example if you run go-bindata on data/... and data contains the
// following hierarchy:
//
//	data/
//	  foo.txt
//	  img/
//	    a.png
//	    b.png
//
// then AssetDir("data") would return []string{"foo.txt", "img"}
// AssetDir("data/img") would return []string{"a.png", "b.png"}
// AssetDir("foo.txt") and AssetDir("notexist") would return an error
// AssetDir("") will return []string{"data"}.
func AssetDir(name string) ([]string, error) {
	node := _bintree
	if len(name) != 0 {
		cannonicalName := strings.Replace(name, "\\", "/", -1)
		pathList := strings.Split(cannonicalName, "/")
		for _, p := range pathList {
			node = node.Children[p]
			if node == nil {
				return nil, fmt.Errorf("Asset %s not found", name)
			}
		}
	}
	if node.Func != nil {
		return nil, fmt.Errorf("Asset %s not found", name)
	}
	rv := make([]string, 0, len(node.Children))
	for childName := range node.Children {
		rv = append(rv, childName)
	}
	return rv, nil
}

type bintree struct {
	Func     func() (*asset, error)
	Children map[string]*bintree
}

var _bintree = &bintree{nil, map[string]*bintree{
	"testdata": {nil, map[string]*bintree{
		"UIPlugin": {nil, map[string]*bintree{
			"UIPlugin.yaml": {testdataUipluginUipluginYaml, map[string]*bintree{}},
		}},
		"eventrouter": {nil, map[string]*bintree{
			"eventrouter.yaml": {testdataEventrouterEventrouterYaml, map[string]*bintree{}},
		}},
		"external-log-stores": {nil, map[string]*bintree{
			"cert_generation.sh": {testdataExternalLogStoresCert_generationSh, map[string]*bintree{}},
			"elasticsearch": {nil, map[string]*bintree{
				"6": {nil, map[string]*bintree{
					"http": {nil, map[string]*bintree{
						"no_user": {nil, map[string]*bintree{
							"configmap.yaml":  {testdataExternalLogStoresElasticsearch6HttpNo_userConfigmapYaml, map[string]*bintree{}},
							"deployment.yaml": {testdataExternalLogStoresElasticsearch6HttpNo_userDeploymentYaml, map[string]*bintree{}},
						}},
						"user_auth": {nil, map[string]*bintree{
							"configmap.yaml":  {testdataExternalLogStoresElasticsearch6HttpUser_authConfigmapYaml, map[string]*bintree{}},
							"deployment.yaml": {testdataExternalLogStoresElasticsearch6HttpUser_authDeploymentYaml, map[string]*bintree{}},
						}},
					}},
					"https": {nil, map[string]*bintree{
						"no_user": {nil, map[string]*bintree{
							"configmap.yaml":  {testdataExternalLogStoresElasticsearch6HttpsNo_userConfigmapYaml, map[string]*bintree{}},
							"deployment.yaml": {testdataExternalLogStoresElasticsearch6HttpsNo_userDeploymentYaml, map[string]*bintree{}},
						}},
						"user_auth": {nil, map[string]*bintree{
							"configmap.yaml":  {testdataExternalLogStoresElasticsearch6HttpsUser_authConfigmapYaml, map[string]*bintree{}},
							"deployment.yaml": {testdataExternalLogStoresElasticsearch6HttpsUser_authDeploymentYaml, map[string]*bintree{}},
						}},
					}},
				}},
				"7": {nil, map[string]*bintree{
					"http": {nil, map[string]*bintree{
						"no_user": {nil, map[string]*bintree{
							"configmap.yaml":  {testdataExternalLogStoresElasticsearch7HttpNo_userConfigmapYaml, map[string]*bintree{}},
							"deployment.yaml": {testdataExternalLogStoresElasticsearch7HttpNo_userDeploymentYaml, map[string]*bintree{}},
						}},
						"user_auth": {nil, map[string]*bintree{
							"configmap.yaml":  {testdataExternalLogStoresElasticsearch7HttpUser_authConfigmapYaml, map[string]*bintree{}},
							"deployment.yaml": {testdataExternalLogStoresElasticsearch7HttpUser_authDeploymentYaml, map[string]*bintree{}},
						}},
					}},
					"https": {nil, map[string]*bintree{
						"no_user": {nil, map[string]*bintree{
							"configmap.yaml":  {testdataExternalLogStoresElasticsearch7HttpsNo_userConfigmapYaml, map[string]*bintree{}},
							"deployment.yaml": {testdataExternalLogStoresElasticsearch7HttpsNo_userDeploymentYaml, map[string]*bintree{}},
						}},
						"user_auth": {nil, map[string]*bintree{
							"configmap.yaml":  {testdataExternalLogStoresElasticsearch7HttpsUser_authConfigmapYaml, map[string]*bintree{}},
							"deployment.yaml": {testdataExternalLogStoresElasticsearch7HttpsUser_authDeploymentYaml, map[string]*bintree{}},
						}},
					}},
				}},
				"8": {nil, map[string]*bintree{
					"http": {nil, map[string]*bintree{
						"no_user": {nil, map[string]*bintree{
							"configmap.yaml":  {testdataExternalLogStoresElasticsearch8HttpNo_userConfigmapYaml, map[string]*bintree{}},
							"deployment.yaml": {testdataExternalLogStoresElasticsearch8HttpNo_userDeploymentYaml, map[string]*bintree{}},
						}},
						"user_auth": {nil, map[string]*bintree{
							"configmap.yaml":  {testdataExternalLogStoresElasticsearch8HttpUser_authConfigmapYaml, map[string]*bintree{}},
							"deployment.yaml": {testdataExternalLogStoresElasticsearch8HttpUser_authDeploymentYaml, map[string]*bintree{}},
						}},
					}},
					"https": {nil, map[string]*bintree{
						"no_user": {nil, map[string]*bintree{
							"configmap.yaml":  {testdataExternalLogStoresElasticsearch8HttpsNo_userConfigmapYaml, map[string]*bintree{}},
							"deployment.yaml": {testdataExternalLogStoresElasticsearch8HttpsNo_userDeploymentYaml, map[string]*bintree{}},
						}},
						"user_auth": {nil, map[string]*bintree{
							"configmap.yaml":  {testdataExternalLogStoresElasticsearch8HttpsUser_authConfigmapYaml, map[string]*bintree{}},
							"deployment.yaml": {testdataExternalLogStoresElasticsearch8HttpsUser_authDeploymentYaml, map[string]*bintree{}},
						}},
					}},
				}},
			}},
			"fluentd": {nil, map[string]*bintree{
				"insecure": {nil, map[string]*bintree{
					"configmap.yaml":      {testdataExternalLogStoresFluentdInsecureConfigmapYaml, map[string]*bintree{}},
					"deployment.yaml":     {testdataExternalLogStoresFluentdInsecureDeploymentYaml, map[string]*bintree{}},
					"http-configmap.yaml": {testdataExternalLogStoresFluentdInsecureHttpConfigmapYaml, map[string]*bintree{}},
				}},
				"secure": {nil, map[string]*bintree{
					"cm-mtls-share.yaml":       {testdataExternalLogStoresFluentdSecureCmMtlsShareYaml, map[string]*bintree{}},
					"cm-mtls.yaml":             {testdataExternalLogStoresFluentdSecureCmMtlsYaml, map[string]*bintree{}},
					"cm-serverauth-share.yaml": {testdataExternalLogStoresFluentdSecureCmServerauthShareYaml, map[string]*bintree{}},
					"cm-serverauth.yaml":       {testdataExternalLogStoresFluentdSecureCmServerauthYaml, map[string]*bintree{}},
					"deployment.yaml":          {testdataExternalLogStoresFluentdSecureDeploymentYaml, map[string]*bintree{}},
					"http-cm-mtls.yaml":        {testdataExternalLogStoresFluentdSecureHttpCmMtlsYaml, map[string]*bintree{}},
					"http-cm-serverauth.yaml":  {testdataExternalLogStoresFluentdSecureHttpCmServerauthYaml, map[string]*bintree{}},
				}},
			}},
			"kafka": {nil, map[string]*bintree{
				"amqstreams": {nil, map[string]*bintree{
					"kafka-no-auth-cluster.yaml":       {testdataExternalLogStoresKafkaAmqstreamsKafkaNoAuthClusterYaml, map[string]*bintree{}},
					"kafka-no-auth-consumer-job.yaml":  {testdataExternalLogStoresKafkaAmqstreamsKafkaNoAuthConsumerJobYaml, map[string]*bintree{}},
					"kafka-sasl-cluster.yaml":          {testdataExternalLogStoresKafkaAmqstreamsKafkaSaslClusterYaml, map[string]*bintree{}},
					"kafka-sasl-consumer-job.yaml":     {testdataExternalLogStoresKafkaAmqstreamsKafkaSaslConsumerJobYaml, map[string]*bintree{}},
					"kafka-sasl-consumers-config.yaml": {testdataExternalLogStoresKafkaAmqstreamsKafkaSaslConsumersConfigYaml, map[string]*bintree{}},
					"kafka-sasl-user.yaml":             {testdataExternalLogStoresKafkaAmqstreamsKafkaSaslUserYaml, map[string]*bintree{}},
					"kafka-topic.yaml":                 {testdataExternalLogStoresKafkaAmqstreamsKafkaTopicYaml, map[string]*bintree{}},
				}},
				"cert_generation.sh": {testdataExternalLogStoresKafkaCert_generationSh, map[string]*bintree{}},
				"kafka-rbac.yaml":    {testdataExternalLogStoresKafkaKafkaRbacYaml, map[string]*bintree{}},
				"kafka-svc.yaml":     {testdataExternalLogStoresKafkaKafkaSvcYaml, map[string]*bintree{}},
				"plaintext-ssl": {nil, map[string]*bintree{
					"consumer-configmap.yaml":        {testdataExternalLogStoresKafkaPlaintextSslConsumerConfigmapYaml, map[string]*bintree{}},
					"kafka-configmap.yaml":           {testdataExternalLogStoresKafkaPlaintextSslKafkaConfigmapYaml, map[string]*bintree{}},
					"kafka-consumer-deployment.yaml": {testdataExternalLogStoresKafkaPlaintextSslKafkaConsumerDeploymentYaml, map[string]*bintree{}},
					"kafka-statefulset.yaml":         {testdataExternalLogStoresKafkaPlaintextSslKafkaStatefulsetYaml, map[string]*bintree{}},
				}},
				"sasl-plaintext": {nil, map[string]*bintree{
					"consumer-configmap.yaml":        {testdataExternalLogStoresKafkaSaslPlaintextConsumerConfigmapYaml, map[string]*bintree{}},
					"kafka-configmap.yaml":           {testdataExternalLogStoresKafkaSaslPlaintextKafkaConfigmapYaml, map[string]*bintree{}},
					"kafka-consumer-deployment.yaml": {testdataExternalLogStoresKafkaSaslPlaintextKafkaConsumerDeploymentYaml, map[string]*bintree{}},
					"kafka-statefulset.yaml":         {testdataExternalLogStoresKafkaSaslPlaintextKafkaStatefulsetYaml, map[string]*bintree{}},
				}},
				"sasl-ssl": {nil, map[string]*bintree{
					"consumer-configmap.yaml":        {testdataExternalLogStoresKafkaSaslSslConsumerConfigmapYaml, map[string]*bintree{}},
					"kafka-configmap.yaml":           {testdataExternalLogStoresKafkaSaslSslKafkaConfigmapYaml, map[string]*bintree{}},
					"kafka-consumer-deployment.yaml": {testdataExternalLogStoresKafkaSaslSslKafkaConsumerDeploymentYaml, map[string]*bintree{}},
					"kafka-statefulset.yaml":         {testdataExternalLogStoresKafkaSaslSslKafkaStatefulsetYaml, map[string]*bintree{}},
				}},
				"zookeeper": {nil, map[string]*bintree{
					"configmap-ssl.yaml":         {testdataExternalLogStoresKafkaZookeeperConfigmapSslYaml, map[string]*bintree{}},
					"configmap.yaml":             {testdataExternalLogStoresKafkaZookeeperConfigmapYaml, map[string]*bintree{}},
					"zookeeper-statefulset.yaml": {testdataExternalLogStoresKafkaZookeeperZookeeperStatefulsetYaml, map[string]*bintree{}},
					"zookeeper-svc.yaml":         {testdataExternalLogStoresKafkaZookeeperZookeeperSvcYaml, map[string]*bintree{}},
				}},
			}},
			"loki": {nil, map[string]*bintree{
				"loki-configmap.yaml":  {testdataExternalLogStoresLokiLokiConfigmapYaml, map[string]*bintree{}},
				"loki-deployment.yaml": {testdataExternalLogStoresLokiLokiDeploymentYaml, map[string]*bintree{}},
			}},
			"otel": {nil, map[string]*bintree{
				"otel-collector.yaml": {testdataExternalLogStoresOtelOtelCollectorYaml, map[string]*bintree{}},
			}},
			"rsyslog": {nil, map[string]*bintree{
				"insecure": {nil, map[string]*bintree{
					"configmap.yaml":  {testdataExternalLogStoresRsyslogInsecureConfigmapYaml, map[string]*bintree{}},
					"deployment.yaml": {testdataExternalLogStoresRsyslogInsecureDeploymentYaml, map[string]*bintree{}},
					"svc.yaml":        {testdataExternalLogStoresRsyslogInsecureSvcYaml, map[string]*bintree{}},
				}},
				"secure": {nil, map[string]*bintree{
					"configmap.yaml":  {testdataExternalLogStoresRsyslogSecureConfigmapYaml, map[string]*bintree{}},
					"deployment.yaml": {testdataExternalLogStoresRsyslogSecureDeploymentYaml, map[string]*bintree{}},
					"svc.yaml":        {testdataExternalLogStoresRsyslogSecureSvcYaml, map[string]*bintree{}},
				}},
			}},
			"splunk": {nil, map[string]*bintree{
				"route-edge_splunk_template.yaml":            {testdataExternalLogStoresSplunkRouteEdge_splunk_templateYaml, map[string]*bintree{}},
				"route-passthrough_splunk_template.yaml":     {testdataExternalLogStoresSplunkRoutePassthrough_splunk_templateYaml, map[string]*bintree{}},
				"secret_splunk_template.yaml":                {testdataExternalLogStoresSplunkSecret_splunk_templateYaml, map[string]*bintree{}},
				"secret_tls_passphrase_splunk_template.yaml": {testdataExternalLogStoresSplunkSecret_tls_passphrase_splunk_templateYaml, map[string]*bintree{}},
				"secret_tls_splunk_template.yaml":            {testdataExternalLogStoresSplunkSecret_tls_splunk_templateYaml, map[string]*bintree{}},
				"statefulset_splunk-8.2_template.yaml":       {testdataExternalLogStoresSplunkStatefulset_splunk82_templateYaml, map[string]*bintree{}},
				"statefulset_splunk-9.0_template.yaml":       {testdataExternalLogStoresSplunkStatefulset_splunk90_templateYaml, map[string]*bintree{}},
			}},
		}},
		"generatelog": {nil, map[string]*bintree{
			"42981.yaml":                                 {testdataGeneratelog42981Yaml, map[string]*bintree{}},
			"container_json_log_template.json":           {testdataGeneratelogContainer_json_log_templateJson, map[string]*bintree{}},
			"container_json_log_template_unannoted.json": {testdataGeneratelogContainer_json_log_template_unannotedJson, map[string]*bintree{}},
			"container_non_json_log_template.json":       {testdataGeneratelogContainer_non_json_log_templateJson, map[string]*bintree{}},
			"logging-performance-app-generator.json":     {testdataGeneratelogLoggingPerformanceAppGeneratorJson, map[string]*bintree{}},
			"multi_container_json_log_template.yaml":     {testdataGeneratelogMulti_container_json_log_templateYaml, map[string]*bintree{}},
			"multiline-error-log.yaml":                   {testdataGeneratelogMultilineErrorLogYaml, map[string]*bintree{}},
		}},
		"logfilemetricexporter": {nil, map[string]*bintree{
			"lfme.yaml": {testdataLogfilemetricexporterLfmeYaml, map[string]*bintree{}},
		}},
		"loki-log-alerts": {nil, map[string]*bintree{
			"cluster-monitoring-config.yaml":          {testdataLokiLogAlertsClusterMonitoringConfigYaml, map[string]*bintree{}},
			"loki-app-alerting-rule-template.yaml":    {testdataLokiLogAlertsLokiAppAlertingRuleTemplateYaml, map[string]*bintree{}},
			"loki-app-recording-rule-template.yaml":   {testdataLokiLogAlertsLokiAppRecordingRuleTemplateYaml, map[string]*bintree{}},
			"loki-infra-alerting-rule-template.yaml":  {testdataLokiLogAlertsLokiInfraAlertingRuleTemplateYaml, map[string]*bintree{}},
			"loki-infra-recording-rule-template.yaml": {testdataLokiLogAlertsLokiInfraRecordingRuleTemplateYaml, map[string]*bintree{}},
			"user-workload-monitoring-config.yaml":    {testdataLokiLogAlertsUserWorkloadMonitoringConfigYaml, map[string]*bintree{}},
		}},
		"lokistack": {nil, map[string]*bintree{
			"fine-grained-access-roles.yaml": {testdataLokistackFineGrainedAccessRolesYaml, map[string]*bintree{}},
			"lokistack-simple-ipv6-tls.yaml": {testdataLokistackLokistackSimpleIpv6TlsYaml, map[string]*bintree{}},
			"lokistack-simple-ipv6.yaml":     {testdataLokistackLokistackSimpleIpv6Yaml, map[string]*bintree{}},
			"lokistack-simple-tls.yaml":      {testdataLokistackLokistackSimpleTlsYaml, map[string]*bintree{}},
			"lokistack-simple.yaml":          {testdataLokistackLokistackSimpleYaml, map[string]*bintree{}},
		}},
		"minIO": {nil, map[string]*bintree{
			"deploy.yaml": {testdataMinioDeployYaml, map[string]*bintree{}},
		}},
		"observability.openshift.io_clusterlogforwarder": {nil, map[string]*bintree{
			"48593.yaml":                 {testdataObservabilityOpenshiftIo_clusterlogforwarder48593Yaml, map[string]*bintree{}},
			"67421.yaml":                 {testdataObservabilityOpenshiftIo_clusterlogforwarder67421Yaml, map[string]*bintree{}},
			"68318.yaml":                 {testdataObservabilityOpenshiftIo_clusterlogforwarder68318Yaml, map[string]*bintree{}},
			"71049.yaml":                 {testdataObservabilityOpenshiftIo_clusterlogforwarder71049Yaml, map[string]*bintree{}},
			"71749.yaml":                 {testdataObservabilityOpenshiftIo_clusterlogforwarder71749Yaml, map[string]*bintree{}},
			"affinity-81397.yaml":        {testdataObservabilityOpenshiftIo_clusterlogforwarderAffinity81397Yaml, map[string]*bintree{}},
			"affinity-81398.yaml":        {testdataObservabilityOpenshiftIo_clusterlogforwarderAffinity81398Yaml, map[string]*bintree{}},
			"audit-policy.yaml":          {testdataObservabilityOpenshiftIo_clusterlogforwarderAuditPolicyYaml, map[string]*bintree{}},
			"azureMonitor-min-opts.yaml": {testdataObservabilityOpenshiftIo_clusterlogforwarderAzuremonitorMinOptsYaml, map[string]*bintree{}},
			"azureMonitor.yaml":          {testdataObservabilityOpenshiftIo_clusterlogforwarderAzuremonitorYaml, map[string]*bintree{}},
			"clf-external-loki-with-secret-tenantKey.yaml": {testdataObservabilityOpenshiftIo_clusterlogforwarderClfExternalLokiWithSecretTenantkeyYaml, map[string]*bintree{}},
			"clf-external-loki-with-secret.yaml":           {testdataObservabilityOpenshiftIo_clusterlogforwarderClfExternalLokiWithSecretYaml, map[string]*bintree{}},
			"clf-kafka-multi-brokers.yaml":                 {testdataObservabilityOpenshiftIo_clusterlogforwarderClfKafkaMultiBrokersYaml, map[string]*bintree{}},
			"clf-kafka-multi-topics.yaml":                  {testdataObservabilityOpenshiftIo_clusterlogforwarderClfKafkaMultiTopicsYaml, map[string]*bintree{}},
			"clf-kafka-no-auth.yaml":                       {testdataObservabilityOpenshiftIo_clusterlogforwarderClfKafkaNoAuthYaml, map[string]*bintree{}},
			"clf-kafka-sasl-ssl.yaml":                      {testdataObservabilityOpenshiftIo_clusterlogforwarderClfKafkaSaslSslYaml, map[string]*bintree{}},
			"clf-kafka-with-auth.yaml":                     {testdataObservabilityOpenshiftIo_clusterlogforwarderClfKafkaWithAuthYaml, map[string]*bintree{}},
			"cloudwatch-accessKey.yaml":                    {testdataObservabilityOpenshiftIo_clusterlogforwarderCloudwatchAccesskeyYaml, map[string]*bintree{}},
			"cloudwatch-iamRole.yaml":                      {testdataObservabilityOpenshiftIo_clusterlogforwarderCloudwatchIamroleYaml, map[string]*bintree{}},
			"cloudwatch-multiple-iamRole.yaml":             {testdataObservabilityOpenshiftIo_clusterlogforwarderCloudwatchMultipleIamroleYaml, map[string]*bintree{}},
			"elasticsearch-https.yaml":                     {testdataObservabilityOpenshiftIo_clusterlogforwarderElasticsearchHttpsYaml, map[string]*bintree{}},
			"elasticsearch-mtls.yaml":                      {testdataObservabilityOpenshiftIo_clusterlogforwarderElasticsearchMtlsYaml, map[string]*bintree{}},
			"elasticsearch-userauth-https.yaml":            {testdataObservabilityOpenshiftIo_clusterlogforwarderElasticsearchUserauthHttpsYaml, map[string]*bintree{}},
			"elasticsearch-userauth-mtls.yaml":             {testdataObservabilityOpenshiftIo_clusterlogforwarderElasticsearchUserauthMtlsYaml, map[string]*bintree{}},
			"elasticsearch-userauth.yaml":                  {testdataObservabilityOpenshiftIo_clusterlogforwarderElasticsearchUserauthYaml, map[string]*bintree{}},
			"elasticsearch.yaml":                           {testdataObservabilityOpenshiftIo_clusterlogforwarderElasticsearchYaml, map[string]*bintree{}},
			"google-cloud-logging-multi-logids.yaml":       {testdataObservabilityOpenshiftIo_clusterlogforwarderGoogleCloudLoggingMultiLogidsYaml, map[string]*bintree{}},
			"googleCloudLogging.yaml":                      {testdataObservabilityOpenshiftIo_clusterlogforwarderGooglecloudloggingYaml, map[string]*bintree{}},
			"http-output-85490.yaml":                       {testdataObservabilityOpenshiftIo_clusterlogforwarderHttpOutput85490Yaml, map[string]*bintree{}},
			"http-output.yaml":                             {testdataObservabilityOpenshiftIo_clusterlogforwarderHttpOutputYaml, map[string]*bintree{}},
			"https-61567.yaml":                             {testdataObservabilityOpenshiftIo_clusterlogforwarderHttps61567Yaml, map[string]*bintree{}},
			"https-output-ca.yaml":                         {testdataObservabilityOpenshiftIo_clusterlogforwarderHttpsOutputCaYaml, map[string]*bintree{}},
			"https-output-mtls.yaml":                       {testdataObservabilityOpenshiftIo_clusterlogforwarderHttpsOutputMtlsYaml, map[string]*bintree{}},
			"https-output-skipverify.yaml":                 {testdataObservabilityOpenshiftIo_clusterlogforwarderHttpsOutputSkipverifyYaml, map[string]*bintree{}},
			"httpserver-to-httpoutput.yaml":                {testdataObservabilityOpenshiftIo_clusterlogforwarderHttpserverToHttpoutputYaml, map[string]*bintree{}},
			"httpserver-to-splunk.yaml":                    {testdataObservabilityOpenshiftIo_clusterlogforwarderHttpserverToSplunkYaml, map[string]*bintree{}},
			"loki.yaml":                                    {testdataObservabilityOpenshiftIo_clusterlogforwarderLokiYaml, map[string]*bintree{}},
			"lokistack-with-labelkeys.yaml":                {testdataObservabilityOpenshiftIo_clusterlogforwarderLokistackWithLabelkeysYaml, map[string]*bintree{}},
			"lokistack.yaml":                               {testdataObservabilityOpenshiftIo_clusterlogforwarderLokistackYaml, map[string]*bintree{}},
			"lokistack_gateway_https_secret.yaml":          {testdataObservabilityOpenshiftIo_clusterlogforwarderLokistack_gateway_https_secretYaml, map[string]*bintree{}},
			"otlp-lokistack.yaml":                          {testdataObservabilityOpenshiftIo_clusterlogforwarderOtlpLokistackYaml, map[string]*bintree{}},
			"otlp.yaml":                                    {testdataObservabilityOpenshiftIo_clusterlogforwarderOtlpYaml, map[string]*bintree{}},
			"rsyslog-mtls.yaml":                            {testdataObservabilityOpenshiftIo_clusterlogforwarderRsyslogMtlsYaml, map[string]*bintree{}},
			"rsyslog-serverAuth.yaml":                      {testdataObservabilityOpenshiftIo_clusterlogforwarderRsyslogServerauthYaml, map[string]*bintree{}},
			"splunk-mtls-passphrase.yaml":                  {testdataObservabilityOpenshiftIo_clusterlogforwarderSplunkMtlsPassphraseYaml, map[string]*bintree{}},
			"splunk-mtls.yaml":                             {testdataObservabilityOpenshiftIo_clusterlogforwarderSplunkMtlsYaml, map[string]*bintree{}},
			"splunk-serveronly.yaml":                       {testdataObservabilityOpenshiftIo_clusterlogforwarderSplunkServeronlyYaml, map[string]*bintree{}},
			"splunk.yaml":                                  {testdataObservabilityOpenshiftIo_clusterlogforwarderSplunkYaml, map[string]*bintree{}},
			"syslog-75317.yaml":                            {testdataObservabilityOpenshiftIo_clusterlogforwarderSyslog75317Yaml, map[string]*bintree{}},
			"syslog-75431.yaml":                            {testdataObservabilityOpenshiftIo_clusterlogforwarderSyslog75431Yaml, map[string]*bintree{}},
			"syslog-81512.yaml":                            {testdataObservabilityOpenshiftIo_clusterlogforwarderSyslog81512Yaml, map[string]*bintree{}},
			"syslog-selected-ns.yaml":                      {testdataObservabilityOpenshiftIo_clusterlogforwarderSyslogSelectedNsYaml, map[string]*bintree{}},
			"syslog.yaml":                                  {testdataObservabilityOpenshiftIo_clusterlogforwarderSyslogYaml, map[string]*bintree{}},
		}},
		"odf": {nil, map[string]*bintree{
			"objectBucketClaim.yaml": {testdataOdfObjectbucketclaimYaml, map[string]*bintree{}},
		}},
		"prometheus-k8s-rbac.yaml": {testdataPrometheusK8sRbacYaml, map[string]*bintree{}},
		"rapidast": {nil, map[string]*bintree{
			"customscan.policy":                         {testdataRapidastCustomscanPolicy, map[string]*bintree{}},
			"data_rapidastconfig_logging_v1.yaml":       {testdataRapidastData_rapidastconfig_logging_v1Yaml, map[string]*bintree{}},
			"data_rapidastconfig_logging_v1alpha1.yaml": {testdataRapidastData_rapidastconfig_logging_v1alpha1Yaml, map[string]*bintree{}},
			"data_rapidastconfig_loki_v1.yaml":          {testdataRapidastData_rapidastconfig_loki_v1Yaml, map[string]*bintree{}},
			"data_rapidastconfig_observability_v1.yaml": {testdataRapidastData_rapidastconfig_observability_v1Yaml, map[string]*bintree{}},
			"job_rapidast.yaml":                         {testdataRapidastJob_rapidastYaml, map[string]*bintree{}},
		}},
		"subscription": {nil, map[string]*bintree{
			"allnamespace-og.yaml":    {testdataSubscriptionAllnamespaceOgYaml, map[string]*bintree{}},
			"catsrc.yaml":             {testdataSubscriptionCatsrcYaml, map[string]*bintree{}},
			"namespace.yaml":          {testdataSubscriptionNamespaceYaml, map[string]*bintree{}},
			"singlenamespace-og.yaml": {testdataSubscriptionSinglenamespaceOgYaml, map[string]*bintree{}},
			"sub-template.yaml":       {testdataSubscriptionSubTemplateYaml, map[string]*bintree{}},
		}},
	}},
}}

// RestoreAsset restores an asset under the given directory
func RestoreAsset(dir, name string) error {
	data, err := Asset(name)
	if err != nil {
		return err
	}
	info, err := AssetInfo(name)
	if err != nil {
		return err
	}
	err = os.MkdirAll(_filePath(dir, filepath.Dir(name)), os.FileMode(0755))
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(_filePath(dir, name), data, info.Mode())
	if err != nil {
		return err
	}
	err = os.Chtimes(_filePath(dir, name), info.ModTime(), info.ModTime())
	if err != nil {
		return err
	}
	return nil
}

// RestoreAssets restores an asset under the given directory recursively
func RestoreAssets(dir, name string) error {
	children, err := AssetDir(name)
	// File
	if err != nil {
		return RestoreAsset(dir, name)
	}
	// Dir
	for _, child := range children {
		err = RestoreAssets(dir, filepath.Join(name, child))
		if err != nil {
			return err
		}
	}
	return nil
}

func _filePath(dir, name string) string {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	return filepath.Join(append([]string{dir}, strings.Split(cannonicalName, "/")...)...)
}
