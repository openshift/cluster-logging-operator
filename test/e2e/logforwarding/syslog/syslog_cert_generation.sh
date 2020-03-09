#!/bin/bash

set -e

WORKING_DIR=$1
NAMESPACE=$2
LOG_STORE=$3

function generate_signing_ca() {
  openssl req -x509 \
              -new \
              -newkey rsa:4096 \
              -keyout ${WORKING_DIR}/ca-syslog.key \
              -nodes \
              -days 1825 \
              -out ${WORKING_DIR}/ca-syslog.crt \
              -subj "/CN=openshift-cluster-logging-signer"
}

function init_cert_files() {

  if [ ! -f ${WORKING_DIR}/ca.db ]; then
    touch ${WORKING_DIR}/ca.db
  fi

  if [ ! -f ${WORKING_DIR}/ca.serial.txt ]; then
    echo 00 > ${WORKING_DIR}/ca.serial.txt
  fi
}

function generate_cert_config() {
  local subject=$1
  cat <<EOF > "${WORKING_DIR}/syslog-server.conf"
[ req ]
default_bits = 4096
prompt = no
encrypt_key = yes
default_md = sha512
distinguished_name = dn
[ dn ]
CN = ${subject}
OU = OpenShift
O = Logging
EOF
}

function generate_request() {
  openssl req -new                                        \
          -out ${WORKING_DIR}/syslog-server.csr           \
          -newkey rsa:4096                                \
          -keyout ${WORKING_DIR}/syslog-server.key        \
          -config ${WORKING_DIR}/syslog-server.conf       \
          -days 712                                       \
          -nodes
}

function sign_cert() {
  openssl ca \
          -in ${WORKING_DIR}/syslog-server.csr  \
          -notext                               \
          -out ${WORKING_DIR}/syslog-server.crt \
          -config ${WORKING_DIR}/signing.conf   \
          -extensions v3_req                    \
          -batch                                \
          -extensions server_ext
}

function generate_certs() {
  local subject=$1

  generate_cert_config $subject
  generate_request
  sign_cert
}

function create_signing_conf() {
  cat <<EOF > "${WORKING_DIR}/signing.conf"
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
certificate             = \$dir/ca-syslog.crt       # The CA cert
private_key             = \$dir/ca-syslog.key # CA private key
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


generate_signing_ca
init_cert_files
create_signing_conf

generate_certs ${LOG_STORE}.${NAMESPACE}.svc
