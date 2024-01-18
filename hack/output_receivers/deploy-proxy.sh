#!/bin/bash

# This script deploys a forward proxy container that can be used
# for manually testing log forwarding with a proxy

set -euo pipefail

NAMESPACE=$1

cat <<EOF | oc create -n $NAMESPACE -f -
apiVersion: v1
kind: ConfigMap
metadata:
  name: httpd
data:
  httpd.conf: |
    Listen 8080
    LoadModule mpm_event_module modules/mod_mpm_event.so
    LoadModule authn_file_module modules/mod_authn_file.so
    LoadModule authn_core_module modules/mod_authn_core.so
    LoadModule authz_host_module modules/mod_authz_host.so
    LoadModule authz_groupfile_module modules/mod_authz_groupfile.so
    LoadModule authz_user_module modules/mod_authz_user.so
    LoadModule authz_core_module modules/mod_authz_core.so
    LoadModule reqtimeout_module modules/mod_reqtimeout.so
    LoadModule filter_module modules/mod_filter.so
    LoadModule log_config_module modules/mod_log_config.so
    LoadModule env_module modules/mod_env.so
    LoadModule headers_module modules/mod_headers.so
    LoadModule setenvif_module modules/mod_setenvif.so
    LoadModule version_module modules/mod_version.so
    LoadModule unixd_module modules/mod_unixd.so
    LoadModule status_module modules/mod_status.so
    LoadModule autoindex_module modules/mod_autoindex.so
    LoadModule proxy_module modules/mod_proxy.so
    LoadModule proxy_connect_module modules/mod_proxy_connect.so
    LoadModule proxy_ftp_module modules/mod_proxy_ftp.so
    LoadModule proxy_http_module modules/mod_proxy_http.so
    LoadModule proxy_fcgi_module modules/mod_proxy_fcgi.so
    LoadModule proxy_scgi_module modules/mod_proxy_scgi.so
    LoadModule proxy_uwsgi_module modules/mod_proxy_uwsgi.so
    LoadModule proxy_fdpass_module modules/mod_proxy_fdpass.so
    LoadModule proxy_wstunnel_module modules/mod_proxy_wstunnel.so
    LoadModule proxy_ajp_module modules/mod_proxy_ajp.so
    LoadModule proxy_balancer_module modules/mod_proxy_balancer.so
    LoadModule proxy_express_module modules/mod_proxy_express.so
    LoadModule proxy_hcheck_module modules/mod_proxy_hcheck.so
    LoadModule slotmem_shm_module modules/mod_slotmem_shm.so
    LoadModule slotmem_plain_module modules/mod_slotmem_plain.so
    LoadModule watchdog_module modules/mod_watchdog.so
    ProxyRequests On
    <Proxy *>
    </Proxy>
    <Directory />
        AllowOverride none
        Require all denied
    </Directory>
    ErrorLog /proc/self/fd/2
    LogLevel warn
    <IfModule log_config_module>
        LogFormat "%h %l %u %t \"%r\" %>s %b \"%{Referer}i\" \"%{User-Agent}i\"" combined
        LogFormat "%h %l %u %t \"%r\" %>s %b" common
        CustomLog /proc/self/fd/1 common
    </IfModule>
EOF

cat <<EOF | oc create -n $NAMESPACE -f -
apiVersion: apps/v1
kind: Deployment
metadata:
  name: mock-proxy
  labels:
    app: mock-proxy
    component: test
spec:
  replicas: 1
  selector:
    matchLabels:
      app: proxy
  template:
    metadata:
      labels:
        app: proxy
        component: test
    spec:
      volumes:
      - configMap:
          defaultMode: 420
          name: httpd
        name: config
      - emptyDir:
          defaultMode: 420
        name: logs
      containers:
      - name: proxy
        image: httpd:2.4
        ports:
        - containerPort: 8080
        volumeMounts:
        - mountPath: /usr/local/apache2/logs
          name: logs
        - mountPath: /usr/local/apache2/conf
          name: config
EOF

cat <<EOF | oc create -n $NAMESPACE -f -
apiVersion: v1
kind: Service
metadata:
  labels:
    app: proxy
    component: test
  name: mock-proxy
spec:
  ports:
  - port: 8080
    protocol: TCP
  selector:
    app: proxy
    component: test
EOF

