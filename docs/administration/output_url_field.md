# Using the clusterlogforwarder.output.url field

The `clusterlogforwarder.output.url` field requires a valid absolute URL.  An
'absolute' URL is one with non-empty `scheme` and `host:port` parts, in other
words it must start with "something://hostname". User and password parts
(e.g. http://user:pass@hostname) are *not allowed*, credentials should be
provided in the `output.secret` field.

In some cases an output type may provide an alternative way to configure
connections, e.g. the `output.kafka.brokers` field. In such cases the output.url
can be omitted. Using the output.url field is preferred whenever possible, but
some output features may not be available via URL (e.g. kafka allows multiple
failover addresses via `output.kafka.brokers`).

Some output types define a URL scheme, for example elasticsearch uses `http` and
`https`. Other output types (e.g. syslog) don't define a "natural" URL
scheme. For those output types we use these special schemes:

* tcp: insecure TCP connection.
* tls: secure TLS over TCP connection.
* udp: insecure UDP packets.
* udps: secure TLS over UDP packets.

If the url scheme is a TLS secure scheme (https, tls, udps) then the
`output.secret` MUST NOT be empty, it provides the TLS certificates. If the URL
scheme is insecure, then `output.secret` is normally empty, but MAY be used for
username/password credentials (e.g. for `fluentForward` secure connections)



