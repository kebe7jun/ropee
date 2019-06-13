# Ropee -- A prometheus remote storage adapter for splunk

With this remote storage adapter, Prometheus can use Splunk as a long-term store for time-series metrics.


## Docker instructions

A docker image for the splunk storage adapter is available on Docker Hub at kebe7jun/ropee.

### Command args
```
Usage of ./ropee:
  -debug
    	Debug mode.
  -listen-addr string
    	Sopee listen addr. (default "127.0.0.1:9970")
  -log-file-path string
    	Log files path. (default "/var/log")
  -splunk-hec-token string
    	Splunk Http event collector token.
  -splunk-hec-url string
    	Splunk Http event collector url. (default "https://127.0.0.1:8088")
  -splunk-metrics-index string
    	Index name. (default "*")
  -splunk-metrics-sourcetype string
    	The prometheus sourcetype name. (default "DaoCloud_promu_metrics")
  -splunk-password string
    	Splunk Manage Password.
  -splunk-url string
    	Splunk Manage Url. (default "https://127.0.0.1:8089")
  -splunk-user string
    	Splunk Manage Username.
  -timeout int
    	API timeout. (default 60)
```

## Configuring Splunk

### HEC(HTTP Event Collector)
Please follow splunk docs.

### Add SourceType for prom metrics

props.conf

```
[DaoCloud_promu_metrics]
DATETIME_CONFIG = CURRENT
TRANSFORMS-prometheus_to_metric = prometheus_metric_name_value, prometheus_metric_dims
NO_BINARY_CHECK = true
description = Prometheus Metrics.
SHOULD_LINEMERGE = false
pulldown_type = 1
category = Metrics
```

transforms.conf
```
[prometheus_metric_name_value]
REGEX = ^([^\s{]+)({[^}]+})? ([-+]?[0-9]*\.?[0-9]+([eE][-+]?[0-9]+)?)
FORMAT = metric_name::$1 ::$2 _value::$3
WRITE_META = true

[prometheus_metric_dims]
REGEX = ([^{=]+)="([^"]*)",?
FORMAT = $1::$2
REPEAT_MATCH = true
WRITE_META = true
```

### Add a metric index
Please follow splunk docs.


## Configuring Prometheus

```
...
remote_read:
  - url: "http://127.0.0.1:9970/read"
    read_recent: true

remote_write:
  - url: "http://127.0.0.1:9970/write"

```

### Building

```
go mod download
go run cmd/ropee/main.go
```
