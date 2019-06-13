# ropee

提供给 Prometheus 的远程读写 Splunk 数据的组件。

## 开发

```bash

go mod vendor
go run cmd/ropee/main.go

```

## Splunk 配置

### 开启 HEC(HTTP Event Collector)
1. 登录 Splunk，打开管理、数据输入、HTTP 事件收集。

### 添加 SourceType

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

## Prometheus 配置

