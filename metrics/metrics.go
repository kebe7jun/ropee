package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	WriteRequestCounter = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "ropee_write_request_count",
		},
	)
	ReadRequestCounter = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "ropee_read_request_count",
		},
	)
	SplunkJobLatency = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "ropee_splunk_job_latency",
		Buckets: prometheus.LinearBuckets(0.1, .5, 5),
	})
	SplunkEventsWrote = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "ropee_splunk_events_wrote_count",
		},
	)
	SplunkEventsWroteFailed = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "ropee_splunk_events_wrote_failed_count",
		},
	)
	uptime = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "ropee_uptime",
	})
)

func init() {
	prometheus.MustRegister(WriteRequestCounter)
	prometheus.MustRegister(ReadRequestCounter)
	prometheus.MustRegister(SplunkJobLatency)
	prometheus.MustRegister(SplunkEventsWrote)
	prometheus.MustRegister(SplunkEventsWroteFailed)
	prometheus.MustRegister(uptime)
	uptime.SetToCurrentTime()
}
