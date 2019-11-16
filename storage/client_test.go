package storage

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/kebe7jun/ropee/test"
	"github.com/prometheus/prometheus/prompb"
)

type fakeClient struct {
	expectBody string
	status     int
	body       string
}

func (f *fakeClient) Do(req *http.Request) (*http.Response, error) {
	if req.Body != nil {
		bs, err := ioutil.ReadAll(req.Body)
		if err != nil {
			return nil, err
		}
		if f.expectBody != "" && string(bs) != f.expectBody {
			return nil, fmt.Errorf("req expect error, body: %s, want: %s", bs, f.expectBody)
		}
	}
	return &http.Response{
		StatusCode: f.status,
		Body:       test.NewBody(f.body),
	}, nil
}

func TestClient_Write(t *testing.T) {
	cases := []struct {
		name      string
		events    prompb.WriteRequest
		wannaBody string
	}{
		{
			"normal events",
			prompb.WriteRequest{
				Timeseries: []prompb.TimeSeries{
					{
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
				},
			},
			`{"event":"test{test=\"1\"} 1","index":"","source":"ropee-client/1.0","sourcetype":"","time":"0.001"}`,
		},
		{
			"multi events",
			prompb.WriteRequest{
				Timeseries: []prompb.TimeSeries{
					{
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
				},
			},
			`{"event":"test{test=\"1\"} 1","index":"","source":"ropee-client/1.0","sourcetype":"","time":"0.001"}{"event":"test{test=\"1\"} 2","index":"","source":"ropee-client/1.0","sourcetype":"","time":"0.002"}`,
		},
	}
	for i, c := range cases {
		t.Run(fmt.Sprintf("test-%d-%s", i, c.name), func(t *testing.T) {
			client := Client{
				url: "http://test.com",
				client: &fakeClient{
					expectBody: c.wannaBody,
					status:     200,
				},
			}
			err := client.Write(&c.events)
			if err != nil {
				t.Fatal(err)
			}
		})
	}
}

type fakeReadClient struct {
	status   int
	bodyChan chan string
}

func (f *fakeReadClient) Do(req *http.Request) (*http.Response, error) {
	return &http.Response{
		StatusCode: f.status,
		Body:       test.NewBody(<-f.bodyChan),
	}, nil
}

func TestClient_Read(t *testing.T) {
	cases := []struct {
		name      string
		req       prompb.ReadRequest
		splunkRes string
		bodys     []string
		wannaRes  string
	}{
		{
			"normal read",
			prompb.ReadRequest{
				Queries: []*prompb.Query{
					{
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
				},
			},
			`{}`,
			[]string{
				"[]",
				`{"sid":"1"}`,
				`{"sid":"1","entry":[{"content":{"isDone":true}}]}`,
				`{"fields":["ropee_metric_name","ropee_metric_value","_time"],"rows":[["test","test","1970-01-01T00:00:01Z"]]}`,
			},
			`{"results":[{"timeseries":[{"labels":[{"name":"__name__","value":"test"}],"samples":[{"timestamp":1000}]}]}]}`,
		},
	}
	for i, c := range cases {
		t.Run(fmt.Sprintf("test-%d-%s", i, c.name), func(t *testing.T) {
			bodyChan := make(chan string, len(c.bodys))
			for _, s := range c.bodys {
				bodyChan <- s
			}
			client := Client{
				url: "http://test.com",
				client: &fakeReadClient{
					status:   200,
					bodyChan: bodyChan,
				},
				log: test.Logger(),
			}
			res, err := client.Read(&c.req)
			if err != nil {
				t.Fatal(err)
			}
			resb, err := json.Marshal(res)
			if string(resb) != c.wannaRes {
				t.Fatalf("unexpected res: %s, want: %s", resb, c.wannaRes)
			}
		})
	}
}
