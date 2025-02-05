# Vector Config Generator Design
This document describes the general design of translating the intention of a 
ClusterLogForwarder to vector config.

# Overview
Historically, the internal model by which log events are processed by the collector are based
upon the [ViaQ data model](../../docs/reference/datamodels/viaq/v1.adoc).  This was changed
in order to facilitate alternate data models, specifically Otel while keeping the consistency
needed to continue to support the existing feature set (e.g. prune filters).

The minimal set of attributes required to support the existing feature set are:

* log_type
* log_source
* timestamp
* message
* level
* kubernetes_metadata (container sources)

Logs generally flow through the collector and processed like:

```
  
  collect from source 
      -> move all fields to '._internal' 
          -> transform
             -> apply output datamodel (copy required fields from ._internal.* to root) 
                  -> apply sink required changes
                      -> send  
```

