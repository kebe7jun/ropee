# Ropee -- A prometheus remote storage adapter for splunk

[![Build Status](https://travis-ci.org/kebe7jun/ropee.svg)](https://travis-ci.org/kebe7jun/ropee)
[![GolangCI](https://golangci.com/badges/github.com/kebe7jun/ropee.svg)](https://golangci.com/r/github.com/kebe7jun/ropee)


With this remote storage adapter, Prometheus can use Splunk as a long-term store for time-series metrics.


## Docker instructions

A docker image for the splunk storage adapter is available on Docker Hub at kebe/ropee.

### Start with docker

```console
# You must edit the following command for your env.
$ docker run -d --name ropee -p 9970:9970 \
    -e LISTEN_ADDR=0.0.0.0:9970 \
    -e SPLUNK_METRICS_INDEX=metrics \
    -e SPLUNK_METRICS_SOURCETYPE=DaoCloud_promu_metrics \
    -e SPLUNK_HEC_TOKEN=asddsa1-12312312-3123-2 \
    -e SPLUNK_HEC_URL=https://192.168.1.1:8088 \
    -e SPLUNK_URL=https://192.168.1.1:8089 \
    -e TIMEOUT=60 \
    -e DEBUG=0 \
    kebe/ropee:latest
```

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
  -splunk-url string
    	Splunk Manage Url. (default "https://127.0.0.1:8089")
  -timeout int
    	API timeout seconds. (default 60)
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
REGEX = ([a-zA-Z_][a-zA-Z0-9_]*)="([^"]*)"[, ]*
FORMAT = $1::"$2"
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
# for remote read, you should set the basic auth which belongs splunk's user.

remote_write:
  - url: "http://127.0.0.1:9970/write"

```

### Building

```
go mod download
go run main.go
```
