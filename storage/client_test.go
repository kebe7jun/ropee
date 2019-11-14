package storage

import (
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
	bs, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}
	if f.expectBody != "" && string(bs) != f.expectBody {
		return nil, fmt.Errorf("req expect error, body: %s, want: %s", bs, f.expectBody)
	}
	return &http.Response{
		StatusCode: f.status,
		Body:       test.NewBody(f.body),
	}, nil
}

func TestClient_Write(t *testing.T) {
	testEvents := prompb.WriteRequest{
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
	}
	client := Client{
		url:      "http://test.com",
		user:     "",
		password: "",
		client: &fakeClient{
			expectBody: `{"event":"test{test=\"1\"} 1","index":"","source":"ropee-client/1.0","sourcetype":"","time":"0.001"}`,
			status:     200,
		},
		timeout:    0,
		index:      "",
		hecUrl:     "",
		hecToken:   "",
		sourcetype: "",
		log:        nil,
	}
	err := client.Write(&testEvents)
	if err != nil {
		t.Fatal(err)
	}
}
