package util

import (
	"encoding/xml"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// SearchResult example
/*
{
  "took" : 75,
  "timed_out" : false,
  "_shards" : {
    "total" : 14,
    "successful" : 14,
    "skipped" : 0,
    "failed" : 0
  },
  "hits" : {
    "total" : 63767,
    "max_score" : 1.0,
    "hits" : [
      {
        "_index" : "app-centos-logtest-000001",
        "_type" : "_doc",
        "_id" : "ODlhMmYzZDgtMTc4NC00M2Q0LWIyMGQtMThmMGY3NTNlNWYw",
        "_score" : 1.0,
        "_source" : {
          "kubernetes" : {
            "container_image_id" : "quay.io/openshifttest/ocp-logtest@sha256:f23bea6f669d125f2f317e3097a0a4da48e8792746db32838725b45efa6c64a4",
            "container_name" : "centos-logtest",
            "namespace_id" : "c74f42bb-3407-418a-b483-d5f33e08f6a5",
            "flat_labels" : [
              "run=centos-logtest",
              "test=centos-logtest"
            ],
            "host" : "ip-10-0-174-131.us-east-2.compute.internal",
            "master_url" : "https://kubernetes.default.svc",
            "pod_id" : "242e7eb4-47ca-4708-9993-db0390d18268",
            "namespace_labels" : {
              "kubernetes_io/metadata_name" : "e2e-test--lg56q"
            },
            "container_image" : "quay.io/openshifttest/ocp-logtest@sha256:f23bea6f669d125f2f317e3097a0a4da48e8792746db32838725b45efa6c64a4",
            "namespace_name" : "e2e-test--lg56q",
            "pod_name" : "centos-logtest-vnwjn"
          },
          "viaq_msg_id" : "ODlhMmYzZDgtMTc4NC00M2Q0LWIyMGQtMThmMGY3NTNlNWYw",
          "level" : "unknown",
          "message" : "{\"message\": \"MERGE_JSON_LOG=true\", \"level\": \"debug\",\"Layer1\": \"layer1 0\", \"layer2\": {\"name\":\"Layer2 1\", \"tips\":\"Decide by PRESERVE_JSON_LOG\"}, \"StringNumber\":\"10\", \"Number\": 10,\"foo.bar\":\"Dot Item\",\"{foobar}\":\"Brace Item\",\"[foobar]\":\"Bracket Item\", \"foo:bar\":\"Colon Item\",\"foo bar\":\"Space Item\" }",
          "docker" : {
            "container_id" : "b3b84d9f11cefa8abf335e8257e394414133b853dc65c8bc1d50120fc3f86da5"
          },
          "hostname" : "ip-10-0-174-131.us-east-2.compute.internal",
          "@timestamp" : "2021-07-09T01:57:44.400169+00:00",
          "pipeline_metadata" : {
            "collector" : {
              "received_at" : "2021-07-09T01:57:44.688935+00:00",
              "name" : "fluentd",
              "inputname" : "fluent-plugin-systemd",
              "version" : "1.7.4 1.6.0",
              "ipaddr4" : "10.0.174.131"
            }
          },
          "structured" : {
            "foo:bar" : "Colon Item",
            "foo.bar" : "Dot Item",
            "Number" : 10,
            "level" : "debug",
            "{foobar}" : "Brace Item",
            "foo bar" : "Space Item",
            "StringNumber" : "10",
            "layer2" : {
              "name" : "Layer2 1",
              "tips" : "Decide by PRESERVE_JSON_LOG"
            },
            "message" : "MERGE_JSON_LOG=true",
            "Layer1" : "layer1 0",
            "[foobar]" : "Bracket Item"
          }
        }
      }
    ]
  }
}
*/
type SearchResult struct {
	Took     int64 `json:"took"`
	TimedOut bool  `json:"timed_out"`
	Shards   struct {
		Total      int64 `json:"total"`
		Successful int64 `json:"successful"`
		Skipped    int64 `json:"skipped"`
		Failed     int64 `json:"failed"`
	} `json:"_shards"`
	Hits struct {
		Total    int64   `json:"total"`
		MaxScore float32 `json:"max_score"`
		DataHits []struct {
			Index  string    `json:"_index"`
			Type   string    `json:"_type"`
			ID     string    `json:"_id"`
			Score  float32   `json:"_score"`
			Source LogEntity `json:"_source"`
		} `json:"hits"`
	} `json:"hits"`
	Aggregations struct {
		LoggingAggregations struct {
			DocCount          int64 `json:"doc_count,omitempty"`
			InnerAggregations struct {
				DocCountErrorUpperBound int64 `json:"doc_count_error_upper_bound,omitempty"`
				SumOtherDocCount        int64 `json:"sum_other_doc_count,omitempty"`
				Buckets                 []struct {
					Key      string `json:"key,omitempty"`
					DocCount int64  `json:"doc_count,omitempty"`
				} `json:"buckets,omitempty"`
			} `json:"inner_aggregations,omitempty"`
		} `json:"logging_aggregations,omitempty"`
	} `json:"aggregations,omitempty"`
}

/*
The aggregation query string must be set as:
{
    "aggs" : {
        "logging_aggregations": {
            "filter": {
                "exists": {
                    "field":"kubernetes"
                }
            },
            "aggs" : {
                "inner_aggregations": {
                    "terms" : {
                        "field" : "hostname"
                    }
                }
            }
        }
    }
}
AggregationResult example
{
	"aggregations": {
		"logging_aggregations": {
		    "doc_count": 13089,
		    "inner_aggregations": {
			    "doc_count_error_upper_bound": 0,
			    "sum_other_doc_count": 0,
			    "buckets": [
			        {
				        "key": "ip-10-0-202-141",
				    	"doc_count": 3250
			  		},
			  		{
						"key": "ip-10-0-147-235",
						"doc_count": 3064
			  		},
			  		{
						"key": "ip-10-0-210-50",
						"doc_count": 2515
			  		},
			  		{
						"key": "ip-10-0-167-109",
						"doc_count": 1832
			  		},
			  		{
						"key": "ip-10-0-186-71",
						"doc_count": 1321
			 		},
			  		{
						"key": "ip-10-0-143-89",
						"doc_count": 1107
			 		 }
				]
		  	}
		}
	}
}
*/

// LogEntity the entity of log data
type LogEntity struct {
	Kubernetes struct {
		Annotations       map[string]string `json:"annotations,omitempty"`
		ContainerID       string            `json:"container_id,omitempty"`
		ContainerImage    string            `json:"container_image"`
		ContainerImageID  string            `json:"container_image_id,omitempty"`
		ContainerIOStream string            `json:"container_iostream,omitempty"`
		ContainerName     string            `json:"container_name"`
		FlatLabels        []string          `json:"flat_labels"`
		Host              string            `json:"host"`
		Lables            map[string]string `json:"labels,omitempty"`
		MasterURL         string            `json:"master_url,omitempty"`
		NamespaceID       string            `json:"namespace_id"`
		NamespaceLabels   map[string]string `json:"namespace_labels,omitempty"`
		NamespaceName     string            `json:"namespace_name"`
		PodID             string            `json:"pod_id"`
		PodIP             string            `json:"pod_ip,omitempty"`
		PodName           string            `json:"pod_name"`
		PodOwner          string            `json:"pod_owner"`
	} `json:"kubernetes,omitempty"`
	Systemd struct {
		SystemdT struct {
			SystemdInvocationID string `json:"SYSTEMD_INVOCATION_ID"`
			BootID              string `json:"BOOT_ID"`
			GID                 string `json:"GID"`
			CmdLine             string `json:"CMDLINE"`
			PID                 string `json:"PID"`
			SystemSlice         string `json:"SYSTEMD_SLICE"`
			SelinuxContext      string `json:"SELINUX_CONTEXT"`
			UID                 string `json:"UID"`
			StreamID            string `json:"STREAM_ID"`
			Transport           string `json:"TRANSPORT"`
			Comm                string `json:"COMM"`
			EXE                 string
			SystemdUnit         string `json:"SYSTEMD_UNIT"`
			CapEffective        string `json:"CAP_EFFECTIVE"`
			MachineID           string `json:"MACHINE_ID"`
			SystemdCgroup       string `json:"SYSTEMD_CGROUP"`
		} `json:"t"`
		SystemdU struct {
			SyslogIdntifier string `json:"SYSLOG_IDENTIFIER"`
			SyslogFacility  string `json:"SYSLOG_FACILITY"`
		} `json:"u"`
	} `json:"systemd,omitempty"`
	ViaqMsgID string `json:"viaq_msg_id,omitempty"`
	Level     string `json:"level"`
	LogSource string `json:"log_source"`
	LogType   string `json:"log_type,omitempty"`
	Message   string `json:"message"`
	Docker    struct {
		ContainerID string `json:"container_id"`
	} `json:"docker,omitempty"`
	HostName  string `json:"hostname"`
	TimeStamp string `json:"@timestamp"`
	File      string `json:"file,omitempty"`
	OpenShift struct {
		ClusterID string            `json:"cluster_id,omitempty"`
		Sequence  int64             `json:"sequence"`
		Labels    map[string]string `json:"labels,omitempty"`
	} `json:"openshift,omitempty"`
	PipelineMetadata struct {
		Collector struct {
			ReceivedAt string `json:"received_at"`
			Name       string `json:"name"`
			InputName  string `json:"inputname"`
			Version    string `json:"version"`
			IPaddr4    string `json:"ipaddr4"`
		} `json:"collector"`
	} `json:"pipeline_metadata,omitempty"`
	Structured struct {
		Level        string `json:"level,omitempty"`
		StringNumber string `json:"StringNumber,omitempty"`
		Message      string `json:"message,omitempty"`
		Number       int    `json:"Number,omitempty"`
		Layer1       string `json:"Layer1,omitempty"`
		FooColonBar  string `json:"foo:bar,omitempty"`
		FooDotBar    string `json:"foo.bar,omitempty"`
		BraceItem    string `json:"{foobar},omitempty"`
		BracketItem  string `json:"[foobar],omitempty"`
		Layer2       struct {
			Name string `json:"name,omitempty"`
			Tips string `json:"tips,omitempty"`
		} `json:"layer2,omitempty"`
	} `json:"structured,omitempty"`
}

// CountResult example
/*
{
  "count" : 453558,
  "_shards" : {
    "total" : 39,
    "successful" : 39,
    "skipped" : 0,
    "failed" : 0
  }
}
*/
type CountResult struct {
	Count  int64 `json:"count"`
	Shards struct {
		Total      int64 `json:"total"`
		Successful int64 `json:"successful"`
		Skipped    int64 `json:"skipped"`
		Failed     int64 `json:"failed"`
	} `json:"_shards"`
}

// ESIndex example
/*
  {
    "health": "green",
    "status": "open",
    "index": "infra-000015",
    "uuid": "uHqlf91RQAqit072gI9LaA",
    "pri": "3",
    "rep": "1",
    "docs.count": "37323",
    "docs.deleted": "0",
    "store.size": "58.8mb",
    "pri.store.size": "29.3mb"
  }
*/
type ESIndex struct {
	Health       string `json:"health"`
	Status       string `json:"status"`
	Index        string `json:"index"`
	UUID         string `json:"uuid"`
	PrimaryCount string `json:"pri"`
	ReplicaCount string `json:"rep"`
	DocsCount    string `json:"docs.count"`
	DocsDeleted  string `json:"docs.deleted"`
	StoreSize    string `json:"store.size"`
	PriStoreSize string `json:"pri.store.size"`
}

// PackageManifest gets the status filed of a packagemanifest
type PackageManifest struct {
	metav1.ObjectMeta `json:"metadata"`
	Status            struct {
		CatalogSource          string `json:"catalogSource"`
		CatalogSourceNamespace string `json:"catalogSourceNamespace"`
		Channels               []struct {
			CurrentCSV string `json:"currentCSV"`
			Name       string `json:"name"`
		} `json:"channels"`
		DefaultChannel string `json:"defaultChannel"`
	} `json:"status"`
}

// OperatorHub gets the status field of an operatorhub object
type OperatorHub struct {
	Status struct {
		Sources []struct {
			Disabled bool   `json:"disabled"`
			Name     string `json:"name"`
			Status   string `json:"status"`
		} `json:"sources"`
	} `json:"status"`
}

//LokiLogQuery result example
/*
{
	"status": "success",
	"data": {
		"resultType": "streams",
		"result": [{
			"stream": {
				"kubernetes_container_name": "centos-logtest",
				"kubernetes_host": "ip-10-0-161-168.us-east-2.compute.internal",
				"kubernetes_namespace_name": "test",
				"kubernetes_pod_name": "centos-logtest-qt6pz",
				"log_type": "application",
				"tag": "kubernetes.var.log.containers.centos-logtest-qt6pz_test_centos-logtest-da3cf8c0493625dc4f42c8592954aad95f3f4ce2a2098ab97ab6a4ad58276617.log",
				"fluentd_thread": "flush_thread_0"
			},
			"values": [
				["1637005525936482085", "{\"docker\":{\"container_id\":\"da3cf8c0493625dc4f42c8592954aad95f3f4ce2a2098ab97ab6a4ad58276617\"},\"kubernetes\":{\"container_name\":\"centos-logtest\",\"namespace_name\":\"test\",\"pod_name\":\"centos-logtest-qt6pz\",\"container_image\":\"quay.io/openshifttest/ocp-logtest@sha256:f23bea6f669d125f2f317e3097a0a4da48e8792746db32838725b45efa6c64a4\",\"container_image_id\":\"quay.io/openshifttest/ocp-logtest@sha256:f23bea6f669d125f2f317e3097a0a4da48e8792746db32838725b45efa6c64a4\",\"pod_id\":\"d77cae4f-2b8a-4c30-8142-417756aa3daf\",\"pod_ip\":\"10.129.2.66\",\"host\":\"ip-10-0-161-168.us-east-2.compute.internal\",\"labels\":{\"run\":\"centos-logtest\",\"test\":\"centos-logtest\"},\"master_url\":\"https://kubernetes.default.svc\",\"namespace_id\":\"18a06953-fdca-4760-b265-a4ef9b98128e\",\"namespace_labels\":{\"kubernetes_io/metadata_name\":\"test\"}},\"message\":\"{\\\"message\\\": \\\"MERGE_JSON_LOG=true\\\", \\\"level\\\": \\\"debug\\\",\\\"Layer1\\\": \\\"layer1 0\\\", \\\"layer2\\\": {\\\"name\\\":\\\"Layer2 1\\\", \\\"tips\\\":\\\"Decide by PRESERVE_JSON_LOG\\\"}, \\\"StringNumber\\\":\\\"10\\\", \\\"Number\\\": 10,\\\"foo.bar\\\":\\\"Dot Item\\\",\\\"{foobar}\\\":\\\"Brace Item\\\",\\\"[foobar]\\\":\\\"Bracket Item\\\", \\\"foo:bar\\\":\\\"Colon Item\\\",\\\"foo bar\\\":\\\"Space Item\\\" }\",\"level\":\"unknown\",\"hostname\":\"ip-10-0-161-168.us-east-2.compute.internal\",\"pipeline_metadata\":{\"collector\":{\"ipaddr4\":\"10.0.161.168\",\"inputname\":\"fluent-plugin-systemd\",\"name\":\"fluentd\",\"received_at\":\"2021-11-15T19:45:26.753126+00:00\",\"version\":\"1.7.4 1.6.0\"}},\"@timestamp\":\"2021-11-15T19:45:25.936482+00:00\",\"viaq_index_name\":\"app-write\",\"viaq_msg_id\":\"NmM5YWIyOGMtM2M4MS00MTFkLWJjNjEtZGIxZDE4MWViNzk0\",\"log_type\":\"application\"}"]
			]
		}, {
			"stream": {
				"kubernetes_host": "ip-10-0-161-168.us-east-2.compute.internal",
				"kubernetes_namespace_name": "test",
				"kubernetes_pod_name": "centos-logtest-qt6pz",
				"log_type": "application",
				"tag": "kubernetes.var.log.containers.centos-logtest-qt6pz_test_centos-logtest-da3cf8c0493625dc4f42c8592954aad95f3f4ce2a2098ab97ab6a4ad58276617.log",
				"fluentd_thread": "flush_thread_1",
				"kubernetes_container_name": "centos-logtest"
			},
			"values": [
				["1637005500907904677", "{\"docker\":{\"container_id\":\"da3cf8c0493625dc4f42c8592954aad95f3f4ce2a2098ab97ab6a4ad58276617\"},\"kubernetes\":{\"container_name\":\"centos-logtest\",\"namespace_name\":\"test\",\"pod_name\":\"centos-logtest-qt6pz\",\"container_image\":\"quay.io/openshifttest/ocp-logtest@sha256:f23bea6f669d125f2f317e3097a0a4da48e8792746db32838725b45efa6c64a4\",\"container_image_id\":\"quay.io/openshifttest/ocp-logtest@sha256:f23bea6f669d125f2f317e3097a0a4da48e8792746db32838725b45efa6c64a4\",\"pod_id\":\"d77cae4f-2b8a-4c30-8142-417756aa3daf\",\"pod_ip\":\"10.129.2.66\",\"host\":\"ip-10-0-161-168.us-east-2.compute.internal\",\"labels\":{\"run\":\"centos-logtest\",\"test\":\"centos-logtest\"},\"master_url\":\"https://kubernetes.default.svc\",\"namespace_id\":\"18a06953-fdca-4760-b265-a4ef9b98128e\",\"namespace_labels\":{\"kubernetes_io/metadata_name\":\"test\"}},\"message\":\"{\\\"message\\\": \\\"MERGE_JSON_LOG=true\\\", \\\"level\\\": \\\"debug\\\",\\\"Layer1\\\": \\\"layer1 0\\\", \\\"layer2\\\": {\\\"name\\\":\\\"Layer2 1\\\", \\\"tips\\\":\\\"Decide by PRESERVE_JSON_LOG\\\"}, \\\"StringNumber\\\":\\\"10\\\", \\\"Number\\\": 10,\\\"foo.bar\\\":\\\"Dot Item\\\",\\\"{foobar}\\\":\\\"Brace Item\\\",\\\"[foobar]\\\":\\\"Bracket Item\\\", \\\"foo:bar\\\":\\\"Colon Item\\\",\\\"foo bar\\\":\\\"Space Item\\\" }\",\"level\":\"unknown\",\"hostname\":\"ip-10-0-161-168.us-east-2.compute.internal\",\"pipeline_metadata\":{\"collector\":{\"ipaddr4\":\"10.0.161.168\",\"inputname\":\"fluent-plugin-systemd\",\"name\":\"fluentd\",\"received_at\":\"2021-11-15T19:45:01.753261+00:00\",\"version\":\"1.7.4 1.6.0\"}},\"@timestamp\":\"2021-11-15T19:45:00.907904+00:00\",\"viaq_index_name\":\"app-write\",\"viaq_msg_id\":\"Yzc1MmJkZDQtNzk4NS00NzA5LWFkN2ItNTlmZmE3NTgxZmUy\",\"log_type\":\"application\"}"]
			]
		}],
		"stats": {
			"summary": {
				"bytesProcessedPerSecond": 48439028,
				"linesProcessedPerSecond": 39619,
				"totalBytesProcessed": 306872,
				"totalLinesProcessed": 251,
				"execTime": 0.006335222
			},
			"store": {
				"totalChunksRef": 0,
				"totalChunksDownloaded": 0,
				"chunksDownloadTime": 0,
				"headChunkBytes": 0,
				"headChunkLines": 0,
				"decompressedBytes": 0,
				"decompressedLines": 0,
				"compressedBytes": 0,
				"totalDuplicates": 0
			},
			"ingester": {
				"totalReached": 1,
				"totalChunksMatched": 2,
				"totalBatches": 1,
				"totalLinesSent": 28,
				"headChunkBytes": 41106,
				"headChunkLines": 85,
				"decompressedBytes": 265766,
				"decompressedLines": 166,
				"compressedBytes": 13457,
				"totalDuplicates": 0
			}
		}
	}
}
*/

// OTEL data module
/*
{
  "status": "success",
  "data": {
    "resultType": "streams",
    "result": [
      {
        "stream": {
          "detected_level": "debug",
          "k8s_container_name": "logging-centos-logtest",
          "k8s_namespace_name": "e2e-test-vector-otlp-pzpdm",
          "k8s_node_name": "ip-10-0-70-97.us-east-2.compute.internal",
          "k8s_pod_label_run": "centos-logtest",
          "k8s_pod_label_test": "centos-logtest",
          "k8s_pod_name": "logging-centos-logtest-9bbn2",
          "k8s_pod_uid": "de4948fe-07bb-42c2-986b-227a24f38a8b",
          "kubernetes_container_name": "logging-centos-logtest",
          "kubernetes_host": "ip-10-0-70-97.us-east-2.compute.internal",
          "kubernetes_namespace_name": "e2e-test-vector-otlp-pzpdm",
          "kubernetes_pod_name": "logging-centos-logtest-9bbn2",
          "log_iostream": "stdout",
          "log_source": "container",
          "log_type": "application",
          "observed_timestamp": "1730165206237598776",
          "openshift_cluster_id": "de026959-72d3-4924-ada8-d6f935c0cdf7",
          "openshift_cluster_uid": "de026959-72d3-4924-ada8-d6f935c0cdf7",
          "openshift_log_source": "container",
          "openshift_log_type": "application",
          "severity_text": "default"
        },
        "values": [
          [
            "1730165205734904307",
            "{\"message\": \"MERGE_JSON_LOG=true\", \"level\": \"debug\",\"Layer1\": \"layer1 0\", \"layer2\": {\"name\":\"Layer2 1\", \"tips\":\"Decide by PRESERVE_JSON_LOG\"}, \"StringNumber\":\"10\", \"Number\": 10,\"foo.bar\":\"Dot Item\",\"{foobar}\":\"Brace Item\",\"[foobar]\":\"Bracket Item\", \"foo:bar\":\"Colon Item\",\"foo bar\":\"Space Item\" }"
          ]
        ]
      }
	]
  }
}
*/
type lokiQueryResponse struct {
	Status string `json:"status"`
	Data   struct {
		ResultType string `json:"resultType"`
		Result     []struct {
			Stream *struct {
				DetectedLevel           string `json:"detected_level,omitempty"`
				K8sContainerName        string `json:"k8s_container_name,omitempty"`
				K8sNamespaceName        string `json:"k8s_namespace_name,omitempty"`
				K8sNodeName             string `json:"k8s_node_name,omitempty"`
				K8sPodName              string `json:"k8s_pod_name,omitempty"`
				K8sPodUID               string `json:"k8s_pod_uid,omitempty"`
				LogType                 string `json:"log_type,omitempty"`
				Tag                     string `json:"tag,omitempty"`
				FluentdThread           string `json:"fluentd_thread,omitempty"`
				KubernetesContainerName string `json:"kubernetes_container_name,omitempty"`
				KubernetesHost          string `json:"kubernetes_host,omitempty"`
				KubernetesNamespaceName string `json:"kubernetes_namespace_name,omitempty"`
				KubernetesPodName       string `json:"kubernetes_pod_name,omitempty"`
				LogIOStream             string `json:"log_iostream,omitempty"`
				LogSource               string `json:"log_source,omitempty"`
				ObservedTimestamp       string `json:"observed_timestamp,omitempty"`
				OpenshiftClusterID      string `json:"openshift_cluster_id,omitempty"`
				OpenshiftClusterUID     string `json:"openshift_cluster_uid,omitempty"`
				OpenshiftLogSource      string `json:"openshift_log_source,omitempty"`
				OpenshiftLogType        string `json:"openshift_log_type,omitempty"`
				SeverityText            string `json:"severity_text,omitempty"`
			} `json:"stream,omitempty"`
			Metric *struct {
				LogType                 string `json:"log_type,omitempty"`
				KubernetesContainerName string `json:"kubernetes_container_name,omitempty"`
				KubernetesHost          string `json:"kubernetes_host,omitempty"`
				KubernetesNamespaceName string `json:"kubernetes_namespace_name,omitempty"`
				KubernetesPodName       string `json:"kubernetes_pod_name,omitempty"`
			} `json:"metric,omitempty"`
			Values []interface{} `json:"values,omitempty"`
			Value  interface{}   `json:"value,omitempty"`
		} `json:"result"`
		Stats struct {
			Summary struct {
				BytesProcessedPerSecond int     `json:"bytesProcessedPerSecond"`
				LinesProcessedPerSecond int     `json:"linesProcessedPerSecond"`
				TotalBytesProcessed     int     `json:"totalBytesProcessed"`
				TotalLinesProcessed     int     `json:"totalLinesProcessed"`
				ExecTime                float32 `json:"execTime"`
			} `json:"summary"`
			Store struct {
				TotalChunksRef        int `json:"totalChunksRef"`
				TotalChunksDownloaded int `json:"totalChunksDownloaded"`
				ChunksDownloadTime    int `json:"chunksDownloadTime"`
				HeadChunkBytes        int `json:"headChunkBytes"`
				HeadChunkLines        int `json:"headChunkLines"`
				DecompressedBytes     int `json:"decompressedBytes"`
				DecompressedLines     int `json:"decompressedLines"`
				CompressedBytes       int `json:"compressedBytes"`
				TotalDuplicates       int `json:"totalDuplicates"`
			} `json:"store"`
			Ingester struct {
				TotalReached       int `json:"totalReached"`
				TotalChunksMatched int `json:"totalChunksMatched"`
				TotalBatches       int `json:"totalBatches"`
				TotalLinesSent     int `json:"totalLinesSent"`
				HeadChunkBytes     int `json:"headChunkBytes"`
				HeadChunkLines     int `json:"headChunkLines"`
				DecompressedBytes  int `json:"decompressedBytes"`
				DecompressedLines  int `json:"decompressedLines"`
				CompressedBytes    int `json:"compressedBytes"`
				TotalDuplicates    int `json:"totalDuplicates"`
			} `json:"ingester"`
		} `json:"stats"`
	} `json:"data"`
}

//labelResponse result example
/*
 {
	"status": "success",
	"data": ["__name__", "fluentd_thread", "kubernetes_container_name", "kubernetes_host", "kubernetes_namespace_name", "kubernetes_pod_name", "log_type", "tag"]
}
*/
type labelResponse struct {
	SearchStatus string   `json:"status"`
	Data         []string `json:"data"`
}

// prometheusQueryResult the response of querying prometheus APIs
type prometheusQueryResult struct {
	Data struct {
		Result     []metric `json:"result"`
		ResultType string   `json:"resultType"`
		Alerts     []alert  `json:"alerts,omitempty"`
	} `json:"data"`
	Status string `json:"status"`
}

// metric the prometheus metric
type metric struct {
	Metric struct {
		Name              string `json:"__name__"`
		Cluster           string `json:"cluster,omitempty"`
		Container         string `json:"container,omitempty"`
		ContainerName     string `json:"containername,omitempty"`
		Endpoint          string `json:"endpoint,omitempty"`
		Instance          string `json:"instance,omitempty"`
		Job               string `json:"job,omitempty"`
		Namespace         string `json:"namespace,omitempty"`
		Path              string `json:"path,omitempty"`
		Pod               string `json:"pod,omitempty"`
		PodName           string `json:"podname,omitempty"`
		Service           string `json:"service,omitempty"`
		ExportedNamespace string `json:"exported_namespace,omitempty"`
		State             string `json:"state,omitempty"`
	} `json:"metric"`
	Value []interface{} `json:"value"`
}

// alert the pending/firing alert
type alert struct {
	Labels struct {
		AlertName string `json:"alertname,omitempty"`
		Condition string `json:"condition,omitempty"`
		Endpoint  string `json:"endpoint,omitempty"`
		Namespace string `json:"namespace,omitempty"`
		Pod       string `json:"pod,omitempty"`
		Instance  string `json:"instance,omitempty"`
		Severity  string `json:"severity,omitempty"`
	} `json:"labels,omitempty"`
	Annotations struct {
		Message    string `json:"message,omitempty"`
		RunBookURL string `json:"runbook_url,omitempty"`
		Summary    string `json:"summary,omitempty"`
	} `json:"annotations,omitempty"`
	State    string `json:"state,omitempty"`
	ActiveAt string `json:"activeAt,omitempty"`
	Value    string `json:"value,omitempty"`
}

// The splunkPod search which deploy on same Openshift Server
type splunkPodServer struct {
	name          string // The splunk name, default: splunk-s1-standalone
	namespace     string // The namespace where splunk is deployed in, default: splunk-aosqe
	authType      string // http(insecure http),tls_mutual,tls_serveronly. Note: when authType==http, you can still access splunk via https://${splunk_route}
	version       string // The splunk version: 8.2 or 9.0, default: 9.0
	hecToken      string // hec_token
	adminUser     string // admin user
	adminPassword string // admin password
	serviceName   string // http service name
	serviceURL    string // http service URL
	hecRoute      string // hec route
	webRoute      string // web route
	splunkdRoute  string // splunkd route
	caFile        string // The ca File
	keyFile       string // The Key File
	certFile      string // The cert File
	passphrase    string // The passphase
}

// The secret used in CLF to splunk server
type toSplunkSecret struct {
	name       string // The secret name
	namespace  string // The namespace where secret will be created
	hecToken   string // The Splunk hec_token
	caFile     string // The collector ca_file
	keyFile    string // The collector Key File
	certFile   string // The collector cert File
	passphrase string // The passphase for the collect key
}

// The splunk response  for a search request.  It includes batch id which can be used to fetch log records
type splunkSearchResp struct {
	XMLName xml.Name `xml:"response"`
	Sid     string   `xml:"sid"`
}

// The log Record in splunk server which is sent out by collector
type splunkLogRecord struct {
	Bkt           string   `json:"_bkt"`
	Cd            string   `json:"_cd"`
	IndexTime     string   `json:"_indextime"`
	Raw           string   `json:"_raw"`
	Serial        string   `json:"_serial"`
	Si            []string `json:"_si"`
	TagSourceType string   `json:"_sourcetype"`
	SubSecond     string   `json:"_subsecond"`
	Time          string   `json:"_time"`
	Host          string   `json:"host"`
	Index         string   `json:"index"`
	LineCount     string   `json:"lincount"`
	LogType       string   `json:"log_type"`
	Source        string   `json:"source"`
	SourceType    string   `json:"souretype"`
	SplunkServer  string   `json:"splunk_sever"`
}

// The splunk search result
type splunkSearchResult struct {
	Preview    bool              `json:"preview"`
	InitOffset float64           `json:"init_offset"`
	Fields     []interface{}     `json:"fields"`
	Messages   []interface{}     `json:"messages"`
	Results    []splunkLogRecord `json:"results"`
}

/* runtime-config.yaml for Loki when overriding spec's in LokiStack CR
---
overrides:
  application:
    ingestion_rate_mb: 10
    ingestion_burst_size_mb: 6
    max_label_name_length: 1024
    max_label_value_length: 2048
    max_label_names_per_series: 30
    max_line_size: 256000
    per_stream_rate_limit: 3MB
    per_stream_rate_limit_burst: 15MB
    max_entries_limit_per_query: 5000
    max_chunks_per_query: 2000000
    max_query_series: 500
    query_timeout: 3m
    cardinality_limit: 100000
    retention_period: 1d
    retention_stream:
    - selector: '{kubernetes_namespace_name=~"test.+"}'
      priority: 1
      period: 1d
    ruler_alertmanager_config:
      alertmanager_url: https://_web._tcp.alertmanager-operated.openshift-user-workload-monitoring.svc
      enable_alertmanager_v2: true
      enable_alertmanager_discovery: true
      alertmanager_refresh_interval: 1m
      alertmanager_client:
        tls_ca_path: /var/run/ca/alertmanager/service-ca.crt
        tls_server_name: alertmanager-user-workload.openshift-user-workload-monitoring.svc.cluster.local
        type: Bearer
        credentials_file: /var/run/secrets/kubernetes.io/serviceaccount/token
  audit:
    ingestion_rate_mb: 20
    ingestion_burst_size_mb: 6
    max_label_name_length: 1024
    max_label_value_length: 2048
    max_label_names_per_series: 30
    max_line_size: 256000
    per_stream_rate_limit: 3MB
    per_stream_rate_limit_burst: 15MB
    max_entries_limit_per_query: 5000
    max_chunks_per_query: 2000000
    max_query_series: 500
    query_timeout: 3m
    cardinality_limit: 100000
    retention_period: 1d
    retention_stream:
    - selector: '{kubernetes_namespace_name=~"openshift-logging.+"}'
      priority: 1
      period: 10d
  infrastructure:
    ingestion_rate_mb: 15
    ingestion_burst_size_mb: 6
    max_label_name_length: 1024
    max_label_value_length: 2048
    max_label_names_per_series: 30
    max_line_size: 256000
    per_stream_rate_limit: 3MB
    per_stream_rate_limit_burst: 15MB
    max_entries_limit_per_query: 5000
    max_chunks_per_query: 2000000
    max_query_series: 500
    query_timeout: 3m
    cardinality_limit: 100000
    retention_period: 5d
    retention_stream:
    - selector: '{kubernetes_namespace_name=~"openshift-cluster.+"}'
      priority: 1
      period: 1d
*/

type RuntimeConfig struct {
	Overrides *Overrides `yaml:"overrides,omitempty"`
}

type Overrides struct {
	Application    *OverridesConfig `yaml:"application,omitempty"`
	Audit          *OverridesConfig `yaml:"audit,omitempty"`
	Infrastructure *OverridesConfig `yaml:"infrastructure,omitempty"`
}
type RetentionStream struct {
	Selector string `yaml:"selector"`
	Priority *int   `yaml:"priority,omitempty"`
	Period   string `yaml:"period"`
}
type RulerAlertmanagerConfig struct {
	AlertmanagerURL             string             `yaml:"alertmanager_url"`
	EnableAlertmanagerV2        bool               `yaml:"enable_alertmanager_v2"`
	EnableAlertmanagerDiscovery bool               `yaml:"enable_alertmanager_discovery"`
	AlertmanagerRefreshInterval string             `yaml:"alertmanager_refresh_interval"`
	AlertmanagerClient          AlertmanagerClient `yaml:"alertmanager_client"`
}
type AlertmanagerClient struct {
	TLSCaPath       string `yaml:"tls_ca_path"`
	TLSServerName   string `yaml:"tls_server_name"`
	Type            string `yaml:"type"`
	CredentialsFile string `yaml:"credentials_file"`
}
type OverridesConfig struct {
	IngestionRateMb         *int                     `yaml:"ingestion_rate_mb,omitempty"`
	IngestionBurstSizeMb    *int                     `yaml:"ingestion_burst_size_mb,omitempty"`
	MaxLabelNameLength      *int                     `yaml:"max_label_name_length,omitempty"`
	MaxLabelValueLength     *int                     `yaml:"max_label_value_length,omitempty"`
	MaxLabelNamesPerSeries  *int                     `yaml:"max_label_names_per_series,omitempty"`
	MaxLineSize             *int                     `yaml:"max_line_size,omitempty"`
	MaxGlobalStreamsPerUser *int                     `yaml:"max_global_streams_per_user,omitempty"`
	PerStreamRateLimit      *string                  `yaml:"per_stream_rate_limit,omitempty"`
	PerStreamRateLimitBurst *string                  `yaml:"per_stream_rate_limit_burst,omitempty"`
	MaxEntriesLimitPerQuery *int                     `yaml:"max_entries_limit_per_query,omitempty"`
	MaxChunksPerQuery       *int                     `yaml:"max_chunks_per_query,omitempty"`
	MaxQuerySeries          *int                     `yaml:"max_query_series,omitempty"`
	QueryTimeout            *string                  `yaml:"query_timeout,omitempty"`
	CardinalityLimit        *int                     `yaml:"cardinality_limit,omitempty"`
	RetentionPeriod         *string                  `yaml:"retention_period,omitempty"`
	RetentionStream         *[]RetentionStream       `yaml:"retention_stream,omitempty"`
	RulerAlertmanagerConfig *RulerAlertmanagerConfig `yaml:"ruler_alertmanager_config,omitempty"`
	ShardStreams            struct {
		Enabled     bool   `yaml:"enabled"`
		DesiredRate string `yaml:"desired_rate"`
	} `yaml:"shard_streams"`
	OtlpConfig OtlpConfig `yaml:"otlp_config,omitempty"`
}

/*
Loki Schema Config

schema_config:
  configs:
    - from: "2023-10-15"
      index:
        period: 24h
        prefix: index_
      object_store: s3
      schema: v13
      store: tsdb

*/

type StorageSchemaConfig struct {
	SchemaConfig SchemaConfig `yaml:"schema_config"`
}

type SchemaConfig struct {
	Configs []ConfigEntry `yaml:"configs"`
}

type ConfigEntry struct {
	From        string `yaml:"from"`
	Index       Index  `yaml:"index"`
	ObjectStore string `yaml:"object_store"`
	Schema      string `yaml:"schema"`
	Store       string `yaml:"store"`
}

type Index struct {
	Period string `yaml:"period"`
	Prefix string `yaml:"prefix"`
}

/*
Loki limits config
*/

type LokiLimitsConfig struct {
	LimitsConfig LimitsConfig `yaml:"limits_config"`
}

type LimitsConfig struct {
	IngestionRateStrategy      string            `yaml:"ingestion_rate_strategy"`
	IngestionRateMB            int               `yaml:"ingestion_rate_mb"`
	IngestionBurstSizeMB       int               `yaml:"ingestion_burst_size_mb"`
	MaxLabelNameLength         int               `yaml:"max_label_name_length"`
	MaxLabelValueLength        int               `yaml:"max_label_value_length"`
	MaxLabelNamesPerSeries     int               `yaml:"max_label_names_per_series"`
	RejectOldSamples           bool              `yaml:"reject_old_samples"`
	RejectOldSamplesMaxAge     string            `yaml:"reject_old_samples_max_age"`
	CreationGracePeriod        string            `yaml:"creation_grace_period"`
	MaxStreamsPerUser          int               `yaml:"max_streams_per_user"`
	MaxLineSize                int               `yaml:"max_line_size"`
	MaxEntriesLimitPerQuery    int               `yaml:"max_entries_limit_per_query"`
	DiscoverServiceName        []string          `yaml:"discover_service_name"`
	DiscoverLogLevels          bool              `yaml:"discover_log_levels"`
	MaxGlobalStreamsPerUser    int               `yaml:"max_global_streams_per_user"`
	MaxChunksPerQuery          int               `yaml:"max_chunks_per_query"`
	MaxQueryLength             string            `yaml:"max_query_length"`
	MaxQueryParallelism        int               `yaml:"max_query_parallelism"`
	TsdbMaxQueryParallelism    int               `yaml:"tsdb_max_query_parallelism"`
	MaxQuerySeries             int               `yaml:"max_query_series"`
	CardinalityLimit           int               `yaml:"cardinality_limit"`
	MaxStreamsMatchersPerQuery int               `yaml:"max_streams_matchers_per_query"`
	QueryTimeout               string            `yaml:"query_timeout"`
	VolumeEnabled              bool              `yaml:"volume_enabled"`
	VolumeMaxSeries            int               `yaml:"volume_max_series"`
	RetentionPeriod            string            `yaml:"retention_period"`
	RetentionStream            []RetentionStream `yaml:"retention_stream"`
	MaxCacheFreshnessPerQuery  string            `yaml:"max_cache_freshness_per_query"`
	PerStreamRateLimit         string            `yaml:"per_stream_rate_limit"`
	PerStreamRateLimitBurst    string            `yaml:"per_stream_rate_limit_burst"`
	SplitQueriesByInterval     string            `yaml:"split_queries_by_interval"`
	ShardStreams               struct {
		Enabled     bool   `yaml:"enabled"`
		DesiredRate string `yaml:"desired_rate"`
	} `yaml:"shard_streams"`
	OtlpConfig              OtlpConfig `yaml:"otlp_config,omitempty"`
	AllowStructuredMetadata bool       `yaml:"allow_structured_metadata"`
}

type OtlpConfig struct {
	ResourceAttributes struct {
		AttributesConfig []struct {
			Action     string   `yaml:"action"`
			Attributes []string `yaml:"attributes"`
			Regex      string   `yaml:"regex,omitempty"`
		} `yaml:"attributes_config"`
	} `yaml:"resource_attributes"`
	LogAttributes []struct {
		Action     string   `yaml:"action"`
		Attributes []string `yaml:"attributes,omitempty"`
		Regex      string   `yaml:"regex,omitempty"`
	} `yaml:"log_attributes,omitempty"`
}

/*
A amqstream instance struct which can share data in kafka intance, CLF and cases
*/
type amqInstance struct {
	name         string // amqstream kakfa clustername
	namespace    string // amqstream kafka namespace
	user         string // amqstream user for topicPrefix
	password     string // amqstream user password
	service      string // amqstream kafka service name for internal service using sasl Plain auth
	route        string // amqstream kakfa broker external route using sasl ssl auth
	routeCA      string // amqstream kafka route ca
	topicPrefix  string // amqstream topicPrefix, only topic with this prefix are allowed.
	instanceType string // the my kafka instance type: kafka-no-auth-cluster,kafka-sasl-cluster
}
