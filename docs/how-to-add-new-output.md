# How to add a new output type

## Understanding the API

The ClusterLogForwarder API is defined by a collection of Go struct types.

[type ClusterLogForwarderSpec struct][cluster_log_forwarder] specifies
* inputs to select logs using InputSpec
* outputs to forward logs to remote targets using OutputSpec, OutputTypeSpec and OutputSecretSpec
* pipelines to connect inputs to outputs using PipelineSpec

It supports an open-ended set of "output types", such as Loki, Kafka, Elasticsearch and so on.

[type OutputSpec struct][cluster_log_forwarder] specifies common Name, Type, URL, Secret fields used by all outputs. The Secret fields references a K8s Secret object used to store any sensitive authentication or authorization data needed by the output.

[type OutputTypeSpec struct][output_types] MAY add extra fields that are unique to the new output type by defining a struct for the new type, and adding it to the OutputTypeSpec.
It is not required if URL and Secret are sufficient.

Do not use the OutputTypeSpec to create extra fields for:
* Information that can be provided in the URL
* Private authentication or authorization data, that should be kept in the Secret.

[apis]: ../apis/logging/v1
[cluster_log_forwarder]: ../apis/logging/v1/cluster_log_forwarder_types.go
[output_types]: ../apis/logging/v1/output_types.go

## Coding a new Output type

Adding a output involves two steps:
1. Add types to the clusterlogforwarder spec.
2. Add code to generate configuration for the supported collectors.

Note: As well as reading this guide, pick an existing output type (e.g. Kafka or CloudWatch) that most resembles your new output type and follow it through the code as an example.

### Add types to the clusterlogforwarder spec

The output type code takes data from ClusterLogForwarderOutputSpec and OutputType struct types and
provides [Go templates][template] to generate a fragment of a collector configuration file that
forwards logs to a single destination. Supported collectors are vector and fluentd.

The generator combines this fragment with others to create the complete collector configuration file.

Relevant source files:

[../apis/logging/v1/cluster_log_forwarder_types.go](../apis/logging/v1/cluster_log_forwarder_types.go)
* Add the name of your output type to the `+kubebuilder:validation` comment on the `OutputSpec.Type` field.

[../apis/logging/v1/output_types.go](../apis/logging/v1/output_types.go)
* Add a constant for your output type name to the top of the file.
* If necessary, add a struct for output-specific fields to `OutputTypeSpec`.

[../internal/constants/constants.go](../internal/constants/constants.go)_
* Keys used in the Secret for authentication/authorization data.
* If your output uses TLS, user name/password, SASL or other common security mechanisms it MUST use the keywords already defined in constants.go.
* If your output has unique security needs that can't be expressed in common terms, you MAY add output-specific keywords.

### Add code to generate configuration for the supported collectors

You need to add/edit the following files:

[../internal/generator/fluentd/output/***your_output***](../internal/generator/fluentd/output/) \
[../internal/generator/vector/output/***your_output***](../internal/generator/vector/output/)
* The *collector* is `vector` or `fluentd`, do both if your output supports both.
* Define types implementing [type Element interface][generator] with a Go template and data to instantiate it.
* Add unit tests to verify your templates produce the expected configuration.
* For example see [../internal/generator/vector/output/kafka](../internal/generator/vector/output/kafka) or other.

[../internal/generator/fluentd/outputs.go](../internal/generator/fluentd/outputs.go) \
[../internal/generator/vector/outputs.go](../internal/generator/vector/outputs.go)
* Add the entry-point function for your vector and/or fluentd output to the switch.

[../test/functional/outputs/***your_output***_test.go](../test/functional/outputs)
* Add a functional test to verify your output can connect and forward logs.

[generator]: /home/aconway/src/cluster-logging-operator/internal/generator/generator.go
[template]: https://pkg.go.dev/text/template
