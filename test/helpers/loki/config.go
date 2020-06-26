package loki

const (
	lokiYaml = `
auth_enabled: false
ingester:
  chunk_idle_period: 3m
  chunk_block_size: 262144
  chunk_retain_period: 1m
  max_transfer_retries: 0
  lifecycler:
    ring:
      kvstore:
        store: inmemory
      replication_factor: 1
limits_config:
  enforce_metric_name: false
  ingestion_rate_mb: 32
  ingestion_burst_size_mb: 32

  reject_old_samples: false
  reject_old_samples_max_age: 336h
schema_config:
  configs:
  - from: 2018-04-15
    store: boltdb
    object_store: filesystem
    schema: v9
    index:
      prefix: index_
      period: 336h
server:
  http_listen_port: 3100
  grpc_server_max_recv_msg_size: 10485760
  grpc_server_max_send_msg_size: 10485760
storage_config:
  boltdb:
    directory: /data/loki/index
  filesystem:
    directory: /data/loki/chunks
chunk_store_config:
  max_look_back_period: 0
table_manager:
  retention_deletes_enabled: false
  retention_period: 0
    `

	lokiUtil = `
#!/bin/sh

HOSTNAME=$1
INDEX_NAME=$2

curl -G -s "http://${HOSTNAME}/loki/api/v1/query" --data-urlencode "query={index_name=\"${INDEX_NAME}\"}" --data-urlencode 'limit=5'
    `
)
