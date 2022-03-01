package cloudwatch

import (
	"testing"

	"github.com/openshift/cluster-logging-operator/internal/generator"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/security"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	corev1 "k8s.io/api/core/v1"
)

var _ = Describe("Generate Vector config", func() {
	inputPipeline := []string{"application-logs"}
	prefix := "all-logs"
	var f = func(clspec logging.ClusterLoggingSpec, secrets map[string]*corev1.Secret, clfspec logging.ClusterLogForwarderSpec, op generator.Options) []generator.Element {
		return Conf(clfspec.Outputs[0], inputPipeline, secrets[clfspec.Outputs[0].Name], op)
	}
	DescribeTable("For Cloudwatch output", generator.TestGenerateConfWith(f),
		Entry("with auth.access_key_id auth.secret_access_key auth.credentials_file", generator.ConfGenerateTest{
			CLFSpec: logging.ClusterLogForwarderSpec{
				Outputs: []logging.OutputSpec{
					{
						Type: logging.OutputTypeCloudwatch,
						Name: "cw",
						URL:  "https://cw.svc.all.cluster:9200",
						OutputTypeSpec: logging.OutputTypeSpec{
							Cloudwatch: &logging.Cloudwatch{
								Region:      "us-east-1",
								GroupBy:     "logType",
								GroupPrefix: &prefix,
							},
						},

						Secret: &logging.OutputSecretSpec{
							Name: "cw",
						},
					},
				},
			},
			Secrets: map[string]*corev1.Secret{
				"cw": {
					Data: map[string][]byte{
						"aws_access_key_id":     []byte("xXyYzZ"),
						"aws_secret_access_key": []byte("sSxXyYzZ"),
					},
				},
			},
			ExpectedConf: `
			# Adding group_name and stream_name field
[transforms.cw_add_grpandstream]
type = "remap"
inputs = ["application-logs"]
source = """
.GroupBy = "logType"
.LogGroupPrefix = "all-logs"
  if (.file != null) {
       filepath = replace!(.file, "/", ".")
       .file = filepath
     } else {
       .file = ".journald.system."
     }
     
     if ( .host == null ) {
        .host = .kubernetes.pod_ips
     }
  
     if (.GroupBy == "logType") {
        .Cw_groupName =  ( .LogGroupPrefix + "." + .log_type ) ?? {}
       if (.log_type == "application" ) {
        .Cw_streamName =  .file
       } else {
        .Cw_streamName = ( .host + .file )  ?? {} 
       }
     } 
  
     if ( .GroupBy == "namespaceName"  || .GroupBy == "namespaceUUID" ) {
       if (.log_type == "application" ) {
        .Cw_groupName =  ( .LogGroupPrefix + "." + .kubernetes.namespace_name ) ?? {}
        .Cw_streamName = .file
       } else {
        .Cw_groupName =  ( .LogGroupPrefix + "." + .log_type ) ?? {}
        .Cw_streamName = (  .host + .file )  ?? {} 
       }
     }
"""

[sinks.cw]
type = "aws_cloudwatch_logs"
inputs = ["cw_add_grpandstream"]
region = "us-east-1"
create_missing_group = true
create_missing_stream = true
compression = "none"
encoding.codec = "json"
request.concurrency = 2
group_name = "{{ Cw_groupName }}"
stream_name = "{{ Cw_streamName }}"
auth.access_key_id = "xXyYzZ"
auth.secret_access_key = "sSxXyYzZ"
`,
		}),
		Entry("without security", generator.ConfGenerateTest{
			CLFSpec: logging.ClusterLogForwarderSpec{
				Outputs: []logging.OutputSpec{
					{
						Type: logging.OutputTypeCloudwatch,
						Name: "cw",
						URL:  "https://cw.svc.all.cluster:9200",
						OutputTypeSpec: logging.OutputTypeSpec{
							Cloudwatch: &logging.Cloudwatch{
								Region:      "us-east-1",
								GroupBy:     "logType",
								GroupPrefix: &prefix,
							},
						},
						Secret: nil,
					},
				},
			},
			Secrets: security.NoSecrets,
			ExpectedConf: `
			# Adding group_name and stream_name field
[transforms.cw_add_grpandstream]
type = "remap"
inputs = ["application-logs"]
source = """
.GroupBy = "logType"
.LogGroupPrefix = "all-logs"
  if (.file != null) {
       filepath = replace!(.file, "/", ".")
       .file = filepath
     } else {
       .file = ".journald.system."
     }
     
     if ( .host == null ) {
        .host = .kubernetes.pod_ips
     }
  
     if (.GroupBy == "logType") {
        .Cw_groupName =  ( .LogGroupPrefix + "." + .log_type ) ?? {}
       if (.log_type == "application" ) {
        .Cw_streamName =  .file
       } else {
        .Cw_streamName = ( .host + .file )  ?? {} 
       }
     } 
  
     if ( .GroupBy == "namespaceName"  || .GroupBy == "namespaceUUID" ) {
       if (.log_type == "application" ) {
        .Cw_groupName =  ( .LogGroupPrefix + "." + .kubernetes.namespace_name ) ?? {}
        .Cw_streamName = .file
       } else {
        .Cw_groupName =  ( .LogGroupPrefix + "." + .log_type ) ?? {}
        .Cw_streamName = (  .host + .file )  ?? {} 
       }
     }
"""

[sinks.cw]
type = "aws_cloudwatch_logs"
inputs = ["cw_add_grpandstream"]
region = "us-east-1"
create_missing_group = true
create_missing_stream = true
compression = "none"
encoding.codec = "json"
request.concurrency = 2
group_name = "{{ Cw_groupName }}"
stream_name = "{{ Cw_streamName }}"
`,
		}),
	)
})

func TestVectorConfGenerator(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Vector Conf Generation")
}
