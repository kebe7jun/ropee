package storage

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/kebe7jun/ropee/metrics"
	"github.com/prometheus/prometheus/prompb"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"sort"
	"strconv"
	"strings"
	"time"
)

type RemoteClient interface {
	Read(*prompb.ReadRequest) (*prompb.ReadResponse, error)
	Write(*prompb.WriteRequest) error
	MetricLabels(string) []string
	LabelValues(string) []string
}

type Client struct {
	url              string
	user             string
	password         string
	client           *http.Client
	timeout          time.Duration
	index            string
	hecUrl, hecToken string
	sourcetype       string
	log              log.Logger
}

func NewClient(
	url, user, password,
	index, sourcetype string,
	hecUrl, hecToken string,
	timeout time.Duration, log log.Logger) (RemoteClient, error) {
	transCfg := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // ignore expired SSL certificates
	}
	return &Client{
		url:        url,
		user:       user,
		password:   password,
		client:     &http.Client{Transport: transCfg},
		timeout:    timeout,
		index:      index,
		hecUrl:     hecUrl,
		hecToken:   hecToken,
		sourcetype: sourcetype,
		log:        log,
	}, nil
}


func (c *Client) Write(req *prompb.WriteRequest) error {
	events := make([]SplunkMetricEvent, 0)
	for _, series := range req.Timeseries {
		es := TimeSeriesToPromMetrics(series)
		events = append(events, es...)
		// todo slice events
	}
	err := c.splunkHECEvents(events)
	if err != nil {
		metrics.SplunkEventsWroteFailed.Add(float64(len(events)))
		return err
	}
	metrics.SplunkEventsWrote.Add(float64(len(events)))
	return nil
}

func (c *Client) Read(req *prompb.ReadRequest) (*prompb.ReadResponse, error) {
	queryResults := make([]*prompb.QueryResult, 0)
	for _, q := range req.Queries {
		search, err := MakeSPL(q, c, c.index)
		if err != nil {
			level.Error(c.log).Log("msg", err)
			return nil, err
		}
		level.Debug(c.log).Log("rendered_search", search, "earliest", q.StartTimestampMs, "latest", q.EndTimestampMs)
		timeStarted := time.Now()
		res, err := c.runSearchWithResult(search, q.StartTimestampMs, q.EndTimestampMs)
		if err != nil {
			level.Error(c.log).Log("msg", err)
			return nil, err
		}
		metrics.SplunkJobLatency.Observe(float64(time.Now().Sub(timeStarted) / time.Second))
		var resPreview map[string][]map[string]string
		json.Unmarshal(res, &resPreview)
		if _, ok := resPreview["fields"]; !ok {
			level.Error(c.log).Log("msg", "search result error from splunk")
			return nil, err
		}
		results := resPreview["results"]
		keysMap := make(map[string]*prompb.TimeSeries)

		for _, result := range results {
			var labelValueList []string
			key := ""
			l := make([]prompb.Label, 0)
			var t time.Time
			var value float64
			for k, v := range result {
				if k == CommonMetricName {
					k = "__name__"
				}
				if k == "_time" {
					t, _ = time.Parse(time.RFC3339, v)
					continue
				}
				if k == CommonMetricValue {
					value, _ = strconv.ParseFloat(v, 64)
					continue
				}
				l = append(l, prompb.Label{
					Name:  k,
					Value: v,
				})
				labelValueList = append(labelValueList, k+"="+v)
			}
			sort.Strings(labelValueList)
			key = strings.Join(labelValueList, ",")
			if _, ok := keysMap[key]; !ok {
				tv := make([]prompb.Sample, 0)
				tv = append(tv, prompb.Sample{Timestamp: t.Unix() * 1000, Value: value})
				keysMap[key] = &prompb.TimeSeries{
					Labels:  l,
					Samples: tv,
				}
			} else {
				s := keysMap[key]
				s.Samples = append(keysMap[key].Samples, prompb.Sample{Timestamp: t.Unix() * 1000, Value: value})
				keysMap[key] = s
			}
		}
		timeSeries := make([]*prompb.TimeSeries, 0)
		for _, value := range keysMap {
			timeSeries = append(timeSeries, value)
		}
		queryResults = append(queryResults, &prompb.QueryResult{
			Timeseries: timeSeries,
		})
	}
	return &prompb.ReadResponse{
		Results: queryResults,
	}, nil
}

func urlJoin(baseUrl, reqPath string) (string, error) {
	u, err := url.Parse(baseUrl)
	if err != nil {
		return "", err
	}
	u.Path = path.Join(u.Path, reqPath)
	return u.String(), nil
}

func (c *Client) splunkHECEvents(events []SplunkMetricEvent) error {
	var buffer bytes.Buffer
	var reqUrl string
	if _url, err := urlJoin(c.hecUrl, "/services/collector"); err == nil {
		reqUrl = _url
	} else {
		return err
	}
	for _, event := range events {
		e, _ := json.Marshal(map[string]string{
			"index":      c.index,
			"sourcetype": c.sourcetype,
			"time":       strconv.FormatFloat(float64(event.Time)/1000.0, 'f', -1, 64),
			"event":      event.MetricStr,
			"source":     "ropee-client/1.0",
		})
		buffer.Write(e)
	}
	httpReq, err := http.NewRequest("POST", reqUrl, strings.NewReader(buffer.String()))
	if err != nil {
		level.Error(c.log).Log("type", "hec-events", "err", err)
	}
	httpReq.Header.Set("User-Agent", "ropee client/1.0")
	httpReq.SetBasicAuth("x", c.hecToken)

	ctx := context.Background()

	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	httpResp, err := c.client.Do(httpReq.WithContext(ctx))
	if err != nil {
		return err
	}
	if httpResp.StatusCode >= 400 {
		level.Warn(c.log).Log("type", "hec-events-resp", "status", httpResp.StatusCode)
	}
	return nil
}

func (c *Client) splunkRESTRequest(method, reqPath string, params, body map[string]string) ([]byte, error) {
	var b io.Reader = nil
	if body != nil {
		p := url.Values{}
		for k, v := range body {
			p.Add(k, v)
		}
		b = strings.NewReader(p.Encode())
	}
	var reqUrl string
	if _url, err := urlJoin(c.url, reqPath); err == nil {
		reqUrl = _url
	} else {
		return nil, err
	}
	httpReq, err := http.NewRequest(method, reqUrl, b)
	httpReq.SetBasicAuth(c.user, c.password)
	q := httpReq.URL.Query()
	q.Add("output_mode", "json")
	q.Add("count", "50000")
	for k, v := range params {
		q.Add(k, v)
	}
	httpReq.URL.RawQuery = q.Encode()
	httpReq.Header.Set("User-Agent", "ropee client/1.0")

	ctx := context.Background()

	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	httpResp, err := c.client.Do(httpReq.WithContext(ctx))
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()
	return ioutil.ReadAll(httpResp.Body)
}

type Metric struct {
	Name string `json:"name"`
}

type MetricLabel struct {
	Name string `json:"name"`
}

type LabelValue struct {
	Name string `json:"name"`
}

func (c *Client) GetMetrics() []string {
	var params = map[string]string{
		"filter": "index=" + c.index,
	}

	res, _ := c.splunkRESTRequest("GET", "/services/catalog/metricstore/metrics", params, nil)
	var result map[string][]Metric
	json.Unmarshal(res, &result)
	ls := make([]string, 0)
	for l := 0; l < len(result["entry"]); l ++ {
		ls = append(ls, result["entry"][l].Name)
	}
	return ls
}

func (c *Client) MetricLabels(metricName string) []string {
	var params = map[string]string{
		"filter":      "index=" + c.index,
		"metric_name": metricName,
	}

	res, _ := c.splunkRESTRequest("GET", "/services/catalog/metricstore/dimensions", params, nil)
	var result map[string][]MetricLabel
	json.Unmarshal(res, &result)
	ls := make([]string, 0)
	for l := 0; l < len(result["entry"]); l ++ {
		if result["entry"][l].Name == "source" || result["entry"][l].Name == "sourcetype" {
			continue
		}
		ls = append(ls, result["entry"][l].Name)
	}
	return ls
}

func (c *Client) LabelValues(labelName string) []string {
	if labelName == "__name__" {
		return c.GetMetrics()
	}
	var params = map[string]string{
		"filter":      "index=" + c.index,
		"metric_name": "*",
	}

	res, _ := c.splunkRESTRequest("GET",
		"/services/catalog/metricstore/dimensions/"+labelName+"/values", params, nil)
	var result map[string][]LabelValue
	json.Unmarshal(res, &result)
	ls := make([]string, 0)
	for l := 0; l < len(result["entry"]); l ++ {
		ls = append(ls, result["entry"][l].Name)
	}
	return ls
}

func (c *Client) runSearchWithResult(search string, start, end int64) ([]byte, error) {
	body := map[string]string{
		"search":        search,
		"latest_time":   strconv.FormatInt(int64(end)/1000, 10),
		"earliest_time": strconv.FormatInt(int64(start)/1000, 10),
	}
	var result map[string]string
	res, err := c.splunkRESTRequest("POST", "/services/search/jobs", nil, body)
	if err != nil {
		return nil, err
	}
	json.Unmarshal(res, &result)
	sid := result["sid"]
	for {
		time.Sleep(100 * time.Millisecond)
		var jobResult map[string][]map[string]map[string]bool
		res, _ := c.splunkRESTRequest("GET", "/services/search/jobs/"+sid, nil, body)

		json.Unmarshal(res, &jobResult)
		jobs := jobResult["entry"]
		if len(jobs) < 1 {
			return nil, fmt.Errorf("get job error")
		}
		if jobs[0]["content"]["isDone"] {
			break
		}
	}
	return c.splunkRESTRequest("GET", "/servicesNS/nobody/-/search/jobs/"+sid+"/results_preview", nil, nil)
}
