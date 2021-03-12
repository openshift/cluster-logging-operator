# Unmanaged Configurations for Cluster Logging
This document describes the various settings that can be tuned in cluster logging but are not
managed by the `cluster-logging-operator`.  Administrators assume responsibility of managing and
maintaining the deployment by modifying the cluster logging CRD to be unmanaged.

```
  spec:
    managementState: "Unmanaged"
```
**NOTE:** These changes are no longer supported in a managed capacity and are
subject to change in the future.

## Fluentd
### MERGE_JSON_LOG
This setting configures Fluentd to inspect each log message to determine if it's format is 'JSON' and to merge
it into the JSON payload document posted to Elasticsearch.  This setting is `false` by default and is a change
from 3.11 and earlier deployments where it was `true`.  

![WARNING](./images/warn.jpg) **WARNING**

**NOTE:** Enabling this feature comes with [risks](https://github.com/openshift/origin-aggregated-logging/issues/1492) summarized here:
* Possible log loss due to Elasticsearch rejecting documents when there are inconsistent type mappings
  * (e.g. document 1 field foo=1, document 2 field foo={"bar":1}) 
* Potential buffer storage leak caused by rejected message cycling
* Overwrite of envelope data for log fields with same names
* Exceeding the maximum configured [field limits](https://www.elastic.co/guide/en/elasticsearch/reference/6.8/mapping.html#mapping-limit-settings).


### Making the Change
Enabling for fluentd config with no clusterlogforwarder:

`oc edit configmap/fluentd`, find the following filter and update the `merge_json_log` attribute to "true":
```
...
   <filter kubernetes.var.log.containers.**>
     @type parse_json_field
     merge_json_log 'true'
     preserve_json_log 'true'
     json_fields 'log,MESSAGE'
   </filter>
...

```
or
use the following patch:
```
259c259
<     merge_json_log 'false'
---
>     merge_json_log 'true'

```

```
$ oc extract configmap/fluentd
$ patch fluent.conf fluent.conf.merge_json.patch
$ oc set data configmap/fluentd --from-file=fluent.conf=fluent.conf
$ oc delete pod -l component=fluentd

```

### Troubleshooting
#### Experiencing Issues with Field Limits
Exceeding the field limits will manifest itself by the collector's inability to push log records.  The collector logs will exhibit errors like:

```
{"error":{"root_cause":[{"type":"illegal_argument_exception","reason":"Limit of total fields [3] in index [foo-001] has been exceeded"}],"type":"illegal_argument_exception","reason":"Limit of total fields [3] in index [foo-001] has been exceeded"},"status":400}
```

The Elasticsearch deployment will additionally log failed ingestion:
```
[2021-03-11T18:40:26,162][DEBUG][o.e.a.a.i.m.p.TransportPutMappingAction] [elasticsearch-cdm-b65klge3-1] failed to put mappings on indices [[[foo-001/58MVEaP8THu7aQEPZMppLQ]]], type [_doc]
java.lang.IllegalArgumentException: Limit of total fields [3] in index [app-00001] has been exceeded

```
It is possible to expand the limit for new indices but doing so will subject the Elasticsearch cluster to potential memory issues as identified by the warning.

* Modify the field limit for new indices(e.g. max limit = 2000)
```
echo '{"index_patterns":["app-*"],
  "order": 1000,
  "settings": { 
    "index.mapping.total_fields.limit": 2000
   }
 }' | oc exec -i -c elasticsearch  $ES_POD -- es_util --query=_template/fieldlimits -d @-
```
* Manually rollover the indices to apply the new template
```
echo '{
  "conditions": {
    "max_docs":  0,
  }
}' | oc exec -i -c elasticsearch  $ES_POD -- es_util --query=app-write/_rollover -d @-
```
#### Experiencing Issues with Type Mappings
Type mapping issues will manifest when the collector and Elasticsearch begin logging errors for certain documents.  This is likely to occur if applications utilize the same key to represent different value types and is exposed in the collector logs with errors similar to:

```
{"error":{"root_cause":[{"type":"mapper_parsing_exception","reason":"failed to parse field [foo1] of type [text] in document with id 'mum2IngBojMwEPumrn_-'"}],"type":"mapper_parsing_exception","reason":"failed to parse field [foo1] of type [text] in document with id 'mum2IngBojMwEPumrn_-'","caused_by":{"type":"illegal_state_exception","reason":"Can't get text on a START_OBJECT at 1:9"}},"status":400}
```

The Elasticsearch deployment will log a similar error:
```
[2021-03-11T19:15:00,991][DEBUG][o.e.a.b.TransportShardBulkAction] [elasticsearch-cdm-b65klge3-1] [foo-001][3] failed to execute bulk item (index) index {[app-00001][_doc][mum2IngBojMwEPumrn_-], source[{"foo1":{"bar":"xyz"}}]}
org.elasticsearch.index.mapper.MapperParsingException: failed to parse field [foo1] of type [text] in document with id 'mum2IngBojMwEPumrn_-'
	at org.elasticsearch.index.mapper.FieldMapper.parse(FieldMapper.java:303) ~[elasticsearch-6.8.1.jar:6.8.1]
...
Caused by: java.lang.IllegalStateException: Can't get text on a START_OBJECT at 1:9

```
Possible ways to resolve:
* Find the offending applications and modify them to consistently generate logs with the same value type for a given key.
* Introduce a filter in the collector configuration to [rename](https://docs.fluentd.org/filter/record_transformer#example-configurations) or [drop](https://docs.fluentd.org/filter/record_transformer#remove_keys) the field for a given application by editing the `fluent.conf` entry in `configmap/fluentd`.  Note the following example adds a filter specifically for pod named `foo` in namespace `bar`:
```
...

    # Ship logs to specific outputs
    <label @DEFAULT>
      
      #start new filter
      <filter kubernetes.var.log.containers.foo*_bar_*.log>
        @type record_transformer
        remove_keys $.my.offensive.key
      </filter>
      #end new filter

      <match retry_default>

...
```
**Note**: Any updates to `fluent.conf` require the fluentd pods to be restarted:
```
oc -n openshift-logging delete pod -l component=fluentd
```