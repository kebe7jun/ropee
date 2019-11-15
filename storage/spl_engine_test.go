package storage

import (
	"fmt"
	"testing"

	"github.com/prometheus/prometheus/prompb"
)

type rClient struct {
	RemoteClient
	labels []string
}

func (c *rClient) MetricLabels(string) []string {
	return c.labels
}

func TestMakeSPL(t *testing.T) {
	cases := []struct {
		name     string
		q        prompb.Query
		cli      rClient
		index    string
		wannaRes string
		wannaErr error
	}{
		{
			"metric equals",
			prompb.Query{
				StartTimestampMs: 0,
				EndTimestampMs:   10,
				Matchers: []*prompb.LabelMatcher{
					{
						Type:  prompb.LabelMatcher_EQ,
						Name:  "__name__",
						Value: "test",
					},
				},
				Hints: &prompb.ReadHints{
					StepMs: 0,
				},
			},
			rClient{
				labels: []string{"test"},
			},
			"test",
			`| mstats latest(_value) as ropee_metric_value where index=test AND metric_name=test span=10s by metric_name test| where metric_name="test"| rename metric_name as ropee_metric_name`,
			nil,
		},
		{
			"step ms 100000",
			prompb.Query{
				StartTimestampMs: 0,
				EndTimestampMs:   10,
				Matchers: []*prompb.LabelMatcher{
					{
						Type:  prompb.LabelMatcher_EQ,
						Name:  "__name__",
						Value: "test",
					},
				},
				Hints: &prompb.ReadHints{
					StepMs: 100000,
				},
			},
			rClient{
				labels: []string{"test"},
			},
			"test",
			`| mstats latest(_value) as ropee_metric_value where index=test AND metric_name=test span=100s by metric_name test| where metric_name="test"| rename metric_name as ropee_metric_name`,
			nil,
		},
		{
			"multi labels",
			prompb.Query{
				StartTimestampMs: 0,
				EndTimestampMs:   10,
				Matchers: []*prompb.LabelMatcher{
					{
						Type:  prompb.LabelMatcher_EQ,
						Name:  "__name__",
						Value: "test",
					},
				},
				Hints: &prompb.ReadHints{
					StepMs: 0,
				},
			},
			rClient{
				labels: []string{"test", "q"},
			},
			"test",
			`| mstats latest(_value) as ropee_metric_value where index=test AND metric_name=test span=10s by metric_name test q| where metric_name="test"| rename metric_name as ropee_metric_name`,
			nil,
		},
		{
			"multi types",
			prompb.Query{
				StartTimestampMs: 0,
				EndTimestampMs:   10,
				Matchers: []*prompb.LabelMatcher{
					{
						Type:  prompb.LabelMatcher_EQ,
						Name:  "__name__",
						Value: "test",
					},
					{
						Type:  prompb.LabelMatcher_NEQ,
						Name:  "test1",
						Value: "test",
					},
					{
						Type:  prompb.LabelMatcher_RE,
						Name:  "test2",
						Value: ".*test$",
					},
					{
						Type:  prompb.LabelMatcher_NRE,
						Name:  "test3",
						Value: ".*test$",
					},
				},
				Hints: &prompb.ReadHints{
					StepMs: 0,
				},
			},
			rClient{
				labels: []string{"test1", "test2", "test3"},
			},
			"test",
			`| mstats latest(_value) as ropee_metric_value where index=test AND metric_name=test span=10s by metric_name test1 test2 test3| where metric_name="test"| where test1!="test"| regex test2=".*test$"| regex test3!=".*test$"| rename metric_name as ropee_metric_name`,
			nil,
		},
		{
			"missing __name__",
			prompb.Query{
				StartTimestampMs: 0,
				EndTimestampMs:   10,
				Matchers: []*prompb.LabelMatcher{
					{
						Type:  prompb.LabelMatcher_EQ,
						Name:  "t",
						Value: "test",
					},
				},
				Hints: &prompb.ReadHints{
					StepMs: 0,
				},
			},
			rClient{
				labels: []string{"test"},
			},
			"test",
			"",
			fmt.Errorf("__name__ is required"),
		},
		{
			"__name__ not eq",
			prompb.Query{
				StartTimestampMs: 0,
				EndTimestampMs:   10,
				Matchers: []*prompb.LabelMatcher{
					{
						Type:  prompb.LabelMatcher_NRE,
						Name:  "__name__",
						Value: "test",
					},
				},
				Hints: &prompb.ReadHints{
					StepMs: 0,
				},
			},
			rClient{
				labels: []string{"test"},
			},
			"test",
			"",
			fmt.Errorf("metric_name label macher type error, only euqals supported"),
		},
	}
	for i, c := range cases {
		t.Run(fmt.Sprintf("test-%d-%s", i, c.name), func(t *testing.T) {
			res, err := MakeSPL(&c.q, &c.cli, c.index)
			if res != c.wannaRes || (err != nil && err.Error() != c.wannaErr.Error()) {
				t.Fatalf("res: %s, %v, want: %s, %v", res, err, c.wannaRes, c.wannaErr)
			}
		})
	}
}

func TestTimeSeriesToPromMetrics(t *testing.T) {
	cases := []struct {
		name string
		ts   prompb.TimeSeries
		want []SplunkMetricEvent
	}{
		{
			"one row test",
			prompb.TimeSeries{
				Labels: []prompb.Label{
					{
						Name:  "__name__",
						Value: "test",
					},
					{
						Name:  "test",
						Value: "1",
					},
				},
				Samples: []prompb.Sample{
					{
						Value:     1,
						Timestamp: 1,
					},
				},
			},
			[]SplunkMetricEvent{
				{
					Time:      1,
					MetricStr: "test{test=\"1\"} 1",
				},
			},
		},
		{
			"multi values",
			prompb.TimeSeries{
				Labels: []prompb.Label{
					{
						Name:  "__name__",
						Value: "test",
					},
					{
						Name:  "test",
						Value: "1",
					},
				},
				Samples: []prompb.Sample{
					{
						Value:     1,
						Timestamp: 1,
					},
					{
						Value:     2,
						Timestamp: 2,
					},
				},
			},
			[]SplunkMetricEvent{
				{
					Time:      1,
					MetricStr: "test{test=\"1\"} 1",
				},
				{
					Time:      2,
					MetricStr: "test{test=\"1\"} 2",
				},
			},
		},
		{
			"missing __name__",
			prompb.TimeSeries{
				Labels: []prompb.Label{
					{
						Name:  "test",
						Value: "1",
					},
				},
				Samples: []prompb.Sample{
					{
						Value:     1,
						Timestamp: 1,
					},
				},
			},
			[]SplunkMetricEvent{},
		},
	}
	for i, c := range cases {
		t.Run(fmt.Sprintf("test-%d-%s", i, c.name), func(t *testing.T) {
			res := TimeSeriesToPromMetrics(c.ts)
			for j, r := range res {
				if r.Time != c.want[j].Time || r.MetricStr != c.want[j].MetricStr {
					t.Fatalf("unexpect res: %v, want: %v", res, c.want)
				}
			}
		})
	}
}
