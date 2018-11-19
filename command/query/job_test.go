package query

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/hashicorp/go-uuid"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-cleanhttp"
	"gitlab.morningconsult.com/mci/go-elasticsearch-alerts/command/alert"
	"gitlab.morningconsult.com/mci/go-elasticsearch-alerts/command/alert/file"
)

var SkipElasticSearchTests bool = false

const ElasticSearchURL string = "http://127.0.0.1:9200"

func TestStateIndexExists(t *testing.T) {
	if SkipElasticSearchTests {
		t.Skipf("unable to connect to ElasticSearch at %s. Skipping test.", ElasticSearchURL)
	}

	id, err := uuid.GenerateUUID()
	if err != nil {
		t.Fatal(err)
	}

	stateIndex := ElasticSearchURL+"/"+id
	client := cleanhttp.DefaultClient()

	delIndexFunc := func() {
		req, err := http.NewRequest("DELETE", stateIndex, nil)
		if err != nil {
			t.Fatal(err)
		}
		_, err = client.Do(req)
		if err != nil {
			t.Fatal(err)
		}
	}
	createIndexFunc := func() {
		req, err := http.NewRequest("PUT", stateIndex, nil)
		if err != nil {
			t.Fatal(err)
		}
		_, err = client.Do(req)
		if err != nil {
			t.Fatal(err)
		}
	}

	qh := &QueryHandler{
		stateURL: stateIndex,
		client:   client,
		logger:   hclog.NewNullLogger(),
	}

	cases := []struct{
		name   string
		exists bool
		err    bool
	}{
		{
			"does-not-exist",
			false,
			false,
		},
		{
			"bad-url",
			false,
			true,
		},
		{
			"exists",
			true,
			false,
		},
	}
	
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			delIndexFunc()
			if tc.name == "exists" {
				createIndexFunc()
				defer delIndexFunc()
			}
			if tc.name == "bad-url" {
				id, err := uuid.GenerateUUID()
				if err != nil {
					t.Fatal(err)
				}
				currentURL := qh.stateURL
				qh.stateURL = "http://"+id+".com"
				defer func() {
					qh.stateURL = currentURL
				}()
			}

			ctx := context.Background()
			exists, err := qh.stateIndexExists(ctx)
			if tc.err {
				if err == nil {
					t.Fatal("expected an error but didn't receive one")
				}
				return
			}

			if err != nil {
				t.Fatal(err)
			}

			if exists != tc.exists {
				t.Fatalf("returned unexpected value (got %t, expected %t)", exists, tc.exists)
			}
		})
	}
}

func TestCreateStateIndex(t *testing.T) {
	if SkipElasticSearchTests {
		t.Skipf("unable to connect to ElasticSearch at %q. Skipping test.", ElasticSearchURL)
	}

	id, err := uuid.GenerateUUID()
	if err != nil {
		t.Fatal(err)
	}

	stateIndex := ElasticSearchURL+"/"+id
	client := cleanhttp.DefaultClient()

	delIndexFunc := func() {
		req, err := http.NewRequest("DELETE", stateIndex, nil)
		if err != nil {
			t.Fatal(err)
		}
		_, err = client.Do(req)
		if err != nil {
			t.Fatal(err)
		}
	}
	createIndexFunc := func() {
		req, err := http.NewRequest("PUT", stateIndex, nil)
		if err != nil {
			t.Fatal(err)
		}
		_, err = client.Do(req)
		if err != nil {
			t.Fatal(err)
		}
	}

	qh := &QueryHandler{
		stateURL: stateIndex,
		client:   client,
		logger:   hclog.NewNullLogger(),
	}

	cases := []struct{
		name   string
		exists bool
		err    bool
	}{
		{
			"creates",
			false,
			false,
		},
		{
			"bad-url",
			false,
			true,
		},
		{
			"already-exists",
			true,
			true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			delIndexFunc()
			if tc.name == "already-exists" {
				createIndexFunc()
				defer delIndexFunc()
			}
			if tc.name == "bad-url" {
				id, err := uuid.GenerateUUID()
				if err != nil {
					t.Fatal(err)
				}
				currentURL := qh.stateURL
				qh.stateURL = "http://"+id+".com"
				defer func() {
					qh.stateURL = currentURL
				}()
			}

			ctx := context.Background()
			err = qh.createStateIndex(ctx)
			if tc.err {
				if err == nil {
					t.Fatal("expected an error but didn't receive one")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}

			resp, err := client.Get(qh.stateURL)
			if err != nil {
				t.Fatal(err)
			}
			if resp.StatusCode != 200 {
				t.Fatalf("ElasticSearch index %q should have been created, but wasn't", qh.stateURL)
			}
		})
	}
}

func TestQuery(t *testing.T) {
	if SkipElasticSearchTests {
		t.Skipf("unable to connect to ElasticSearch at %q. Skipping test.", ElasticSearchURL)
	}

	id, err := uuid.GenerateUUID()
	if err != nil {
		t.Fatal(err)
	}

	queryURL := ElasticSearchURL+"/"+id
	client := cleanhttp.DefaultClient()

	req, err := http.NewRequest("PUT", queryURL, nil)
	if err != nil {
		t.Fatal(err)
	}
	_, err = client.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	// write some data
	payload := `{
    "hello": "darkness",
    "my": {
        "old": "friend"
    }
}`
	_, err = client.Post(queryURL+"/_doc", "application/json", bytes.NewBufferString(payload))
    	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		req, err := http.NewRequest("DELETE", queryURL, nil)
		if err != nil {
			t.Fatal(err)
		}
		_, err = client.Do(req)
		if err != nil {
			t.Fatal(err)
		}
	}()

	qh := &QueryHandler{
		queryURL:  queryURL,
		client:    client,
		logger:    hclog.NewNullLogger(),
		queryData: map[string]interface{}{
			"query": map[string]interface{}{
				"term": map[string]interface{}{
					"old": "darkness",
				},
			},
		},
	}

	ctx := context.Background()
	_, err = qh.query(ctx)
	if err != nil {
		t.Fatal(err)
	}
	// TODO
}

func TestRun(t *testing.T) {
	if SkipElasticSearchTests {
		t.Skipf("unable to connect to ElasticSearch at %q. Skipping test.", ElasticSearchURL)
	}

	client := cleanhttp.DefaultClient()

	randomUUID := func() string {
		id, err := uuid.GenerateUUID()
		if err != nil {
			t.Fatal(err)
		}
		return id
	}

	changeIndexFunc := func(method, index string, body io.Reader) {
		req, err := http.NewRequest(method, ElasticSearchURL+"/"+index, body)
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Add("Content-Type", "application/json")

		_, err = client.Do(req)
		if err != nil {
			t.Fatal(err)
		}
	}

	queryIndex := randomUUID()
	stateIndex := randomUUID()

	changeIndexFunc("PUT", queryIndex, nil)
	payload := bytes.NewBufferString(`{"hello": "world", "this": {"is": "a-test"}}`)
	changeIndexFunc("POST", queryIndex+"/_doc", payload)
	defer changeIndexFunc("DELETE", queryIndex, nil)
	defer changeIndexFunc("DELETE", stateIndex, nil)

	fileAM, err := file.NewFileAlertMethod(&file.FileAlertMethodConfig{
		OutputFilepath: "/tmp/testfile.log",
	})
	if err != nil {
		t.Fatal(err)
	}

	qh, err := NewQueryHandler(&QueryHandlerConfig{
		Name:   "test_errors",
		Logger: hclog.NewNullLogger(),
		Client: client,
		ESUrl:  ElasticSearchURL,
		QueryData: map[string]interface{}{
			"query": map[string]interface{}{
				"term": map[string]interface{}{
					"hello": "world",
				},
			},
		},
		QueryIndex: queryIndex,
		Schedule: "@every 10s",
		StateIndex: stateIndex,
		AlertMethods: []alert.AlertMethod{fileAM},
	})
	if err != nil {
		t.Fatal(err)
	}

	ctx, _ := context.WithTimeout(context.Background(), 2 * time.Minute)
	outputCh := make(chan *alert.Alert, 1)
	var wg sync.WaitGroup
	go qh.Run(ctx, outputCh, &wg)

	select {
	case <-ctx.Done():
		t.Fatal("context timeout")
	case a := <-outputCh:
		if a.RuleName != qh.name {
			t.Fatalf("bad alert rule name (expected: %q, got: %q)", qh.name, a.RuleName)
		}
		if len(a.Methods) != 1 {
			t.Fatal("alert should have just one alert method (file)")
		}
		if _, ok := a.Methods[0].(*file.FileAlertMethod); !ok {
			t.Fatalf("alert method should be of the FileAlertMethod type (got type: %T)", a.Methods[0])
		}
		if len(a.Records) != 1 {
			t.Fatalf("expected only one alert.Record (got %d records)", len(a.Records))
		}

		expected := `{
    "hello": "world",
    "this": {
        "is": "a-test"
    }
}`
		if a.Records[0].Text != expected {
			t.Fatalf("unexpected alert.Record[0].Text value (got %q, expected %q)", a.Records[0].Text, expected)
		}
	}
}

func TestMain(m *testing.M) {
	client := cleanhttp.DefaultClient()
	resp, err := client.Get(ElasticSearchURL)
	if err != nil {
		SkipElasticSearchTests = true
	} else {
		if resp.StatusCode != 200 {
			SkipElasticSearchTests = true
		}
	}
	os.Exit(m.Run())
}