package storage

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/prometheus/prometheus/prompb"
)

func MakeSPL(query *prompb.Query, c RemoteClient, index string) (string, error) {
	metricName := ""
	for _, m := range query.Matchers {
		if m.Name == "__name__" {
			if m.Type != prompb.LabelMatcher_EQ {
				return "", fmt.Errorf("metric_name label macher type error, only euqals supported")
			}
			metricName = m.Value
			break
		}
	}
	if metricName == "" {
		return "", fmt.Errorf("__name__ is required")
	}
	step := query.Hints.StepMs / 1000
	if step < 10 {
		step = 10
	}
	ls := strings.Join(c.MetricLabels(metricName), " ")
	search := fmt.Sprintf("| mstats latest(_value) as %s where index=%s AND metric_name=%s span=%ds by metric_name %s",
		CommonMetricValue, index, metricName, step, ls)
	for _, m := range query.Matchers {
		if m.Name == "__name__" {
			m.Name = "metric_name"
		}
		switch m.Type {
		case prompb.LabelMatcher_RE:
			//search += fmt.Sprintf("| regex " + m.Name + "=" + strconv.Quote(m.Value))
			search += fmt.Sprintf("| regex %s=%q", m.Name, m.Value)
		case prompb.LabelMatcher_NRE:
			search += fmt.Sprintf("| regex %s!=%q", m.Name, m.Value)
		case prompb.LabelMatcher_EQ:
			search += fmt.Sprintf("| where %s=%q", m.Name, m.Value)
		case prompb.LabelMatcher_NEQ:
			search += fmt.Sprintf("| where %s!=%q", m.Name, m.Value)
		}
	}
	search += "| rename metric_name as " + CommonMetricName
	return search, nil
}

type SplunkMetricEvent struct {
	Time      int64
	MetricStr string
}

func TimeSeriesToPromMetrics(series prompb.TimeSeries) []SplunkMetricEvent {
	res := make([]SplunkMetricEvent, 0, len(series.Samples))
	labels := []string{}
	metricName := ""
	for _, label := range series.Labels {
		if label.Name == "__name__" {
			metricName = label.Value
			continue
		}
		labels = append(labels, label.Name+"="+strconv.Quote(label.Value))
	}
	if metricName == "" {
		return nil
	}
	mergedKey := fmt.Sprintf("%s{%s}", metricName, strings.Join(labels, ","))
	for _, sample := range series.Samples {
		valueTime := strconv.FormatFloat(sample.Value, 'f', -1, 64)
		res = append(res, SplunkMetricEvent{
			Time:      sample.Timestamp,
			MetricStr: fmt.Sprintf("%s %s", mergedKey, valueTime),
		})
	}
	return res
}
